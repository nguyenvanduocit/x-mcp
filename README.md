# X MCP Server

A Model Context Protocol (MCP) server that provides AI assistants with access to X/Twitter API. Search tweets, post content, retrieve user profiles, and more.

## Features

- **Search tweets** - Search recent tweets using Twitter search syntax
- **Post tweets** - Create new tweets or reply to existing ones
- **Post threads** - Post multi-tweet threads automatically
- **Get user info** - Retrieve profile information for any user
- **Get user timeline** - Fetch recent tweets from any user
- **Get tweet details** - Retrieve full tweet information with metrics

## Installation

### Setup Credentials

1. Get your API keys from [X Developer Portal](https://developer.x.com/en/portal/dashboard)
2. Create env file at `~/.claude/x-mcp/.env`:

```bash
mkdir -p ~/.claude/x-mcp
cp .env.example ~/.claude/x-mcp/.env
```

3. Fill in your credentials:

```bash
X_API_KEY=your_api_key
X_API_SECRET=your_api_secret
X_ACCESS_TOKEN=your_access_token
X_ACCESS_TOKEN_SECRET=your_access_token_secret

# Optional proxy
# PROXY_URL=http://localhost:8080
```

### Using Claude MCP (stdio)

```bash
go install github.com/nguyenvanduocit/x-mcp@latest

claude mcp add x-mcp stdio -- x-mcp -env ~/.claude/x-mcp/.env
```

Or without installing (slower startup):

```bash
claude mcp add x-mcp stdio -- go run github.com/nguyenvanduocit/x-mcp@latest -env ~/.claude/x-mcp/.env
```

### Using HTTP mode

```bash
x-mcp -env ~/.claude/x-mcp/.env -http_port 8080
```

The server will be available at `http://localhost:8080/mcp`.

### Verify Setup

Clone the repo and run the verification script:

```bash
git clone https://github.com/nguyenvanduocit/x-mcp.git
cd x-mcp
go run scripts/verify-token/main.go \
  -api-key=$X_API_KEY \
  -api-secret=$X_API_SECRET \
  -access-token=$X_ACCESS_TOKEN \
  -access-token-secret=$X_ACCESS_TOKEN_SECRET
```

## Available Tools

| Tool | Description |
|------|-------------|
| `x_search_tweets` | Search recent tweets using Twitter search syntax |
| `x_get_tweet` | Get a specific tweet by ID with full details |
| `x_post_tweet` | Post a new tweet or reply to an existing tweet |
| `x_post_thread` | Post a multi-tweet thread (tweets separated by `|||`) |
| `x_get_user` | Get user profile information by username |
| `x_get_user_timeline` | Get recent tweets from a user's timeline |

## Usage Examples

### Search tweets
```
Search for tweets about golang from the last week
```

### Post a tweet
```
Post a tweet saying "Hello from X MCP!"
```

### Post a thread
```
Post a thread with:
1. First tweet introducing the topic
2. Second tweet with more details
3. Third tweet with conclusion
```

### Get user info
```
Get user profile for elonmusk
```

## License

MIT
