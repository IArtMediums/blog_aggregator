package main
import (
	_ "github.com/lib/pq"
	"fmt"
	"context"
	"os"
	"database/sql"
	"github.com/IArtMediums/blog_aggregator/internal/config"
	"github.com/IArtMediums/blog_aggregator/internal/command"
	"github.com/IArtMediums/blog_aggregator/internal/database"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	db, err := sql.Open("postgres", cfg.DbUrl)
	dbQueries := database.New(db)
	state := command.State{CfgPtr: &cfg, DbPtr: dbQueries}
	commands := command.Commands{Data: make(map[string]func(*command.State, command.Command) error)}
	registerCommands(&commands)
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Expected at least 2 arguments, received: %v\n", len(os.Args))
		os.Exit(1)
	}
	cmd := command.Command{Name: os.Args[1], Args: os.Args[2:]}
	if err := commands.Run(&state, cmd); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}

func registerCommands(cmds *command.Commands) {
	cmds.Register("login", command.HandlerLogins)
	cmds.Register("register", command.HandlerRegister)
	cmds.Register("reset", command.HandlerReset)
	cmds.Register("users", command.HandlerUsers)
	cmds.Register("agg", command.HandlerAgg)
	cmds.Register("addfeed", middlewareLoggedIn(command.HandlerAddFeed))
	cmds.Register("feeds", command.HandlerFeeds)
	cmds.Register("follow", middlewareLoggedIn(command.HandlerFollow))
	cmds.Register("following", middlewareLoggedIn(command.HandlerFollowing))
	cmds.Register("unfollow", middlewareLoggedIn(command.HandlerUnfollow))
	cmds.Register("browse", middlewareLoggedIn(command.HandlerBrowse))
}

func middlewareLoggedIn(handler func(s *command.State, cmd command.Command, user database.User) error) func(*command.State, command.Command) error {
	return func(s *command.State, cmd command.Command) error {
		if s.CfgPtr.CurrentUserName == "" {
			return fmt.Errorf("You must be logged in to run this command\n")
		}
		ctx := context.Background()
		user, err := s.DbPtr.GetUser(ctx, s.CfgPtr.CurrentUserName)
		if err != nil {
			return fmt.Errorf("%v\n", err)
		}
		return handler(s, cmd, user)
	}
}
