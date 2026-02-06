-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
	$1,
	$2,
	$3,
	$4
	)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE name = $1;

-- name: ClearDb :exec
DELETE FROM users;

-- name: GetUsers :many
SELECT * FROM users;

-- name: AddFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
	$1,
	$2,
	$3,
	$4, 
	$5,
	$6
	)
RETURNING *;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
	INSERT INTO feed_follows(id, created_at, updated_at, user_id, feed_id)
	VALUES (
		$1,
		$2,
		$3,
		$4,
		$5
		)
	RETURNING *
)
SELECT inserted_feed_follow.*,
	feeds.name AS feed_name,
	users.name AS user_name
FROM inserted_feed_follow
INNER JOIN feeds ON inserted_feed_follow.feed_id = feeds.id
INNER JOIN users ON inserted_feed_follow.user_id = users.id;

-- name: GetFeedByUrl :one
SELECT * FROM feeds WHERE url = $1;

-- name: GetFeedFollowsForUser :many
SELECT feeds.name AS feed_name
FROM feed_follows
INNER JOIN feeds ON feed_follows.feed_id = feeds.id
WHERE feed_follows.user_id = $1;

-- name: UnfollowFeed :exec
DELETE FROM feed_follows WHERE user_id = $1 AND feed_id = $2;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched = NOW(),
	updated_at = NOW()
WHERE id = $1;

-- name: GetNextFeedToFetch :one
SELECT id FROM feeds
WHERE last_fetched IS NULL
	OR last_fetched < $1
ORDER BY last_fetched NULLS FIRST
LIMIT 1;

-- name: GetUrlByID :one
SELECT url FROM feeds
WHERE id = $1;

-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES (
	$1,
	NOW(),
	NOW(), 
	$2,
	$3,
	$4,
	$5,
	$6
	)
RETURNING *;

-- name: GetPostsForUser :many
SELECT posts.title,
	posts.url, 
	posts.description,
	posts.published_at,
	feeds.name AS feed_name
FROM posts
JOIN feed_follows ON posts.feed_id = feed_follows.feed_id
JOIN feeds ON posts.feed_id = feeds.id
WHERE feed_follows.user_id = $1
ORDER BY posts.published_at DESC NULLS LAST
LIMIT $2;
