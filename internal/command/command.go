package command
import (
	"strconv"
	"context"
	"github.com/google/uuid"
	"fmt"
	"time"
	"database/sql"
	"github.com/IArtMediums/blog_aggregator/internal/config"
	"github.com/IArtMediums/blog_aggregator/internal/database"
	"github.com/IArtMediums/blog_aggregator/internal/rssfetch"
)

type State struct {
	CfgPtr *config.Config
	DbPtr *database.Queries
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	Data map[string]func(*State, Command) error
}

func HandlerBrowse(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Expected 1 argument, received %v\n", len(cmd.Args))
	}
	ctx := context.Background()
	limit64, err := strconv.ParseInt(cmd.Args[0], 10, 32)
	if err != nil || limit64 <= 0 {
		return fmt.Errorf("Limit must be positive integer, got %v\n", cmd.Args[0])
	}
	limit := int32(limit64)
	param := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit: limit,
	}
	posts, err := s.DbPtr.GetPostsForUser(ctx, param)
	if err != nil {
		return fmt.Errorf("%v\n", err)
	}
	for _, post := range posts {
		fmt.Printf("---Feed %v---\n", post.FeedName)
		fmt.Printf("	* Title: %v\n", post.Title.String)
		fmt.Printf("	* Url: %v\n", post.Url)
		fmt.Printf("	* Description: %v\n", post.Description.String)
		fmt.Printf("	* Published At: %v\n", post.PublishedAt.Time)
	}
	return nil
}

func HandlerUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Expected 1 argument, received %v\n", len(cmd.Args))
	}
	ctx := context.Background()
	feed, err := s.DbPtr.GetFeedByUrl(ctx, cmd.Args[0])
	if err != nil {
		return fmt.Errorf("%v\n", err)
	}
	params := database.UnfollowFeedParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	if err := s.DbPtr.UnfollowFeed(ctx, params); err != nil {
		return fmt.Errorf("%v\n", err)
	}
	return nil
}

func HandlerFollowing(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("Expected 0 arguments, received %v\n", len(cmd.Args))
	}
	ctx := context.Background()
	feeds, err := s.DbPtr.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("%v\n", err)
	}
	fmt.Printf("---Feeds Followed by %v---\n", user.Name)
	for _, feed := range feeds {
		fmt.Printf("	* %s\n", feed)
	}
	return nil
}

func HandlerFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Expected 1 argument, received %v\n", len(cmd.Args))
	}
	ctx := context.Background()
	feed, err := s.DbPtr.GetFeedByUrl(ctx, cmd.Args[0])
	if err != nil {
		return fmt.Errorf("%v\n", err)
	}
	params := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feed.ID,
	}
	res, err := s.DbPtr.CreateFeedFollow(ctx, params)
	if err != nil {
		return fmt.Errorf("%v\n", err)
	}
	fmt.Printf("User: %v, Follows: %v\n", res.UserName, res.FeedName)
	return nil
}

func HandlerFeeds(s *State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("Expected 0 arguments, received %v\n", len(cmd.Args))
	}
	ctx := context.Background()
	feeds, err := s.DbPtr.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("%v\n", err)
	}
	for i, feed := range feeds {
		fmt.Printf("---Feed %v---\n", i)
		fmt.Printf("	- Name: %v\n", feed.Name)
		fmt.Printf("	- Url: %v\n", feed.Url)
		user, err := s.DbPtr.GetUserByID(ctx, feed.UserID)
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Printf("	- User: %v\n", user.Name)
		}
	}
	return nil
}

func HandlerAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("Expected 2 arguments, received %v\n", len(cmd.Args))
	}
	ctx := context.Background()
	req := database.AddFeedParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: cmd.Args[0],
		Url: cmd.Args[1],
		UserID: user.ID,
	}
	feed, err := s.DbPtr.AddFeed(ctx, req)
	if err != nil {
		return fmt.Errorf("%v\n", err)
	}
	fmt.Printf("Feed ID: %v\n", feed.ID)
	fmt.Printf("Feed Created at: %v\n", feed.CreatedAt)
	fmt.Printf("Feed Updated at: %v\n", feed.UpdatedAt)
	fmt.Printf("Feed Name: %v\n", feed.Name)
	fmt.Printf("Feed Url: %v\n", feed.Url)
	fmt.Printf("Feed User ID: %v\n", feed.UserID)
	cmd.Args = cmd.Args[1:]
	if err := HandlerFollow(s, cmd, user); err != nil {
		return fmt.Errorf("%v\n", err)
	}
	return nil
}

func HandlerAgg(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Expected 1 argument, received %v\n", len(cmd.Args))
	}
	parsed_time, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("%v\n", err)
	}
	fmt.Printf("Collecting feeds every %v\n", cmd.Args[0])
	ticker := time.NewTicker(parsed_time)
	defer ticker.Stop()

	if err := scrapeFeeds(s, cmd.Args[0]); err != nil {
		fmt.Printf("Error during scraping: %v\n", err)
	}
	for ; ; <-ticker.C {
		if err := scrapeFeeds(s, cmd.Args[0]); err != nil {
			fmt.Printf("Error during scraping: %v\n", err)
		}
	}
	return nil
}

func HandlerUsers(s *State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("Expected 0 arguments, received %v\n", len(cmd.Args))
	}
	context := context.Background()
	users, err := s.DbPtr.GetUsers(context)
	if err != nil {
		return fmt.Errorf("%v\n", err)
	}
	for _, user := range users {
		if s.CfgPtr.CurrentUserName == user.Name {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func HandlerReset(s *State, cmd Command) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("Expected 0 arguments, received %v\n", len(cmd.Args))
	}
	context := context.Background()
	err := s.DbPtr.ClearDb(context)
	if err != nil {
		return fmt.Errorf("%v\n", err)
	}
	fmt.Printf("Database Cleared\n")
	return nil
}

func HandlerLogins(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Expected 1 argument, received %v\n", len(cmd.Args))
	}
	context := context.Background()
	user, err := s.DbPtr.GetUser(context, cmd.Args[0])
	if err != nil {
		return fmt.Errorf("%v\n", err)
	}
	if err := s.CfgPtr.SetUser(cmd.Args[0]); err != nil {
		return fmt.Errorf("Error setting User: %s\n", err)
	}
	fmt.Printf("User: %s, Logged in.\n", user.Name)
	return nil
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Expected 1 argument, received %v\n", len(cmd.Args))
	}
	context := context.Background()
	new_entry := database.CreateUserParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: cmd.Args[0],
	}
	user, err := s.DbPtr.CreateUser(context, new_entry)
	if err != nil {
		return fmt.Errorf("%v\n", err)
	}
	if err := HandlerLogins(s, cmd); err != nil {
		return fmt.Errorf("%v\n", err)
	}
	fmt.Printf("New User Created:\n")
	fmt.Printf("	id: %v\n", user.ID)
	fmt.Printf("	created at: %v\n", user.CreatedAt)
	fmt.Printf("	updated at: %v\n", user.UpdatedAt)
	fmt.Printf("	name: %v\n", user.Name)

	return nil
}

func (c *Commands) Run(s *State, cmd Command) error {
	function, ok := c.Data[cmd.Name]
	if !ok {
		return fmt.Errorf("Command: %s missing from commands\n", cmd.Name)
	}
	return function(s, cmd)
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.Data[name] = f
}

func scrapeFeeds(s *State, time_between_reqs string) error {
	ctx := context.Background()
	parsed_time, err := time.ParseDuration(time_between_reqs)
	if err != nil {
		return fmt.Errorf("%v\n", err)
	}
	cutoff := time.Now().Add(-parsed_time)
	time_between := sql.NullTime{
		Time: cutoff,
		Valid: true,
	}
	id, err := s.DbPtr.GetNextFeedToFetch(ctx, time_between)
	if err != nil {
		return fmt.Errorf("%v\n", err)
	}
	url, err := s.DbPtr.GetUrlByID(ctx, id)
	if err != nil {
		return fmt.Errorf("%v\n", err)
	}
	feed, err := rssfetch.FetchFeed(ctx, url)
	if err != nil {
		return fmt.Errorf("%v\n", err)
	}
	fmt.Printf("---Scraping %v---\n", url)
	for _, item := range feed.Channel.Item {
		parse_time, err := time.Parse(time.RFC1123, item.PubDate)
		pub_time := sql.NullTime{Time: parse_time, Valid: true,}
		if err != nil {
			pub_time.Valid = false
		}
		param := database.CreatePostParams{
			ID: uuid.New(),
			Title: sql.NullString{
				String: item.Title,
				Valid: true,
			},
			Url: item.Link,
			Description: sql.NullString{
				String: item.Description,
				Valid: true,
			},
			PublishedAt: pub_time,
			FeedID: id,
		}
		if _, err := s.DbPtr.CreatePost(ctx, param); err != nil {
			return fmt.Errorf("%v\n", err)
		}
		fmt.Printf("Post: %v, added to database\n", item.Title)
	}
	if err := s.DbPtr.MarkFeedFetched(ctx, id); err != nil {
		return fmt.Errorf("%v\n", err)
	}
	return nil
}
