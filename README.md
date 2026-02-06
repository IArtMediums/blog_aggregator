# blog_aggregator

Simple RSS blog aggregator with a CLI client (`gator`).

## Prerequisites

- Postgres (database)
- Go (toolchain)

## Install the CLI

```bash
go install ./...
```

This installs the `blog_aggregator` binary into your `GOBIN` (or `~/go/bin` by default).

## Configure

Create a config file at `~/.gatorconfig.json`:

```json
{
  "db_url": "postgres://USER:PASSWORD@localhost:5432/gator?sslmode=disable",
  "current_user_name": "your_name"
}
```

Update the connection string for your Postgres instance.

## Run

```bash
blog_aggregator
```

## Commands

Examples:

- `blog_aggregator register <name>`: create a new user
- `blog_aggregator login <name>`: switch the current user
- `blog_aggregator addfeed <url>`: add an RSS feed
- `blog_aggregator follow <url>`: follow an existing feed
- `blog_aggregator unfollow <url>`: unfollow a feed
- `blog_aggregator agg <interval>`: fetch and aggregate new posts
- `blog_aggregator browse [limit]`: list recent posts
