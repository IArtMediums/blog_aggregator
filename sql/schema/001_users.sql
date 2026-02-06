-- +goose Up
CREATE TABLE users (
	id UUID PRIMARY KEY,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	name TEXT UNIQUE NOT NULL
);
CREATE TABLE feeds (
	id UUID PRIMARY KEY,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	name TEXT NOT NULL,
	url TEXT UNIQUE NOT NULL,
	user_id UUID NOT NULL,
	last_fetched TIMESTAMP NULL,
	CONSTRAINT fk_user
		FOREIGN KEY(user_id)
		REFERENCES users(id)
		ON DELETE CASCADE
);
CREATE TABLE feed_follows (
	id UUID PRIMARY KEY,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	user_id UUID NOT NULL,
	feed_id UUID NOT NULL,
	UNIQUE(user_id, feed_id),
	CONSTRAINT fk_user_id
		FOREIGN KEY(user_id)
		REFERENCES users(id)
		ON DELETE CASCADE,
	CONSTRAINT fk_feed_id
		FOREIGN KEY(feed_id)
		REFERENCES feeds(id)
		ON DELETE CASCADE
);
CREATE TABLE posts (
	id UUID PRIMARY KEY,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	title TEXT,
	url TEXT NOT NULL,
	description TEXT,
	published_at TIMESTAMP,
	feed_id UUID NOT NULL,
	CONSTRAINT fk_post_feed_id
		FOREIGN KEY(feed_id)
		REFERENCES feeds(id)
		ON DELETE CASCADE
);
-- +goose Down
DROP TABLE posts;
DROP TABLE feed_follows;
DROP TABLE feeds;
DROP TABLE users;
