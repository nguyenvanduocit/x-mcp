## X MCP

An MCP server that gives AI assistants direct access to X/Twitter — search tweets, post content, manage threads, and retrieve user profiles.

Built for real workflows: monitor conversations, publish threads, and gather intelligence from X without leaving your AI assistant.

## Available tools

### Tweets
- **x_search_tweets** - Search recent tweets using Twitter search syntax
- **x_get_tweet** - Get a specific tweet by ID with full details and metrics
- **x_post_tweet** - Post a new tweet or reply to an existing tweet
- **x_post_thread** - Post a multi-tweet thread (tweets separated by `|||`)
- **x_delete_tweet** - Delete a tweet you posted
- **x_get_quote_tweets** - Get tweets that quote a specific tweet

### Engagement
- **x_like_tweet** - Like a tweet
- **x_unlike_tweet** - Unlike a tweet
- **x_retweet** - Retweet a tweet
- **x_unretweet** - Undo a retweet

### Users
- **x_get_user** - Get user profile information by username
- **x_get_user_timeline** - Get recent tweets from a user's timeline
- **x_get_mentions** - Get recent tweets mentioning a user

### Lists
- **x_get_user_lists** - Get lists owned by a user
- **x_get_list_tweets** - Get recent tweets from a list

### Moderation
- **x_mute_user** - Mute a user
- **x_unmute_user** - Unmute a user
- **x_block_user** - Block a user
- **x_unblock_user** - Unblock a user

## Installation

Copy this prompt to your AI assistant:

```
Install the X MCP server (https://github.com/nguyenvanduocit/x-mcp) for my Claude Desktop or Cursor IDE. Read the MCP documentation carefully and guide me through the installation step by step.
```

If your AI assistant cannot help with this installation, it indicates either a misconfiguration or an ineffective AI tool. A capable AI assistant should be able to guide you through MCP installation.

## Quick start

### 1) Get API credentials
Create your keys at [X Developer Portal](https://developer.x.com/en/portal/dashboard).

### 2) Add to Cursor

#### Binary
```json
{
  "mcpServers": {
    "x": {
      "command": "x-mcp",
      "args": ["-env", "/path/to/your/.env"],
      "env": {
        "X_API_KEY": "your-api-key",
        "X_API_SECRET": "your-api-secret",
        "X_ACCESS_TOKEN": "your-access-token",
        "X_ACCESS_TOKEN_SECRET": "your-access-token-secret"
      }
    }
  }
}
```

Install the binary:
```bash
go install github.com/nguyenvanduocit/x-mcp@latest
```

### 3) Add to Claude Code
```bash
claude mcp add x-mcp -- x-mcp -env /path/to/your/.env
```

### 4) Try it
- "Search for tweets about golang from the last week"
- "Post a tweet saying Hello from X MCP!"
- "Get user profile for elonmusk"
- "Post a thread about the top 3 Go features"
- "Like this tweet: 123456789"
- "Get my recent mentions"
- "Mute user @spammer"
- "Show me tweets from my 'Tech News' list"

## Configuration

Required environment variables:
- **X_API_KEY** - Your X API key
- **X_API_SECRET** - Your X API secret
- **X_ACCESS_TOKEN** - Your X access token
- **X_ACCESS_TOKEN_SECRET** - Your X access token secret

Optional:
- **PROXY_URL** - HTTP proxy URL (e.g., `http://localhost:8080`)

`.env` file example:
```bash
X_API_KEY=your-api-key
X_API_SECRET=your-api-secret
X_ACCESS_TOKEN=your-access-token
X_ACCESS_TOKEN_SECRET=your-access-token-secret
```

HTTP mode (optional, for debugging):
```bash
x-mcp -env .env -http_port 8080
```

## Verify credentials

```bash
git clone https://github.com/nguyenvanduocit/x-mcp.git
cd x-mcp
go run scripts/verify-token/main.go \
  -api-key=$X_API_KEY \
  -api-secret=$X_API_SECRET \
  -access-token=$X_ACCESS_TOKEN \
  -access-token-secret=$X_ACCESS_TOKEN_SECRET
```

## CLI Usage

In addition to the MCP server, `x-mcp` ships a standalone CLI binary (`x-cli`) for direct terminal use — no MCP client needed.

### Installation

```bash
just install-cli
# or
go install github.com/nguyenvanduocit/x-mcp/cmd/x-cli@latest
```

### Quick Start

```bash
export X_API_KEY=your-api-key
export X_API_SECRET=your-api-secret
export X_ACCESS_TOKEN=your-access-token
export X_ACCESS_TOKEN_SECRET=your-access-token-secret
# or
x-cli --env .env <command> [flags]
```

### Commands

| Command | Description |
|---------|-------------|
| `post-tweet` | Post a new tweet |
| `delete-tweet` | Delete a tweet |
| `search` | Search tweets |
| `get-user` | Get user profile |
| `get-user-timeline` | Get user's timeline |
| `post-thread` | Post a thread |
| `like-tweet` | Like a tweet |
| `unlike-tweet` | Unlike a tweet |
| `retweet` | Retweet a tweet |
| `unretweet` | Undo a retweet |
| `get-mentions` | Get mentions |
| `mute-user` | Mute a user |
| `block-user` | Block a user |

### Examples

```bash
# Post a tweet
x-cli post-tweet --text "Hello from x-cli!"

# Search tweets
x-cli search --query "golang MCP" --max-results 10

# Get user info
x-cli get-user --username myusername

# Like a tweet
x-cli like-tweet --tweet-id 1234567890

# JSON output
x-cli search --query "golang" --output json | jq '.[].text'
```

### Flags

Every command accepts:
- `--env string` — Path to `.env` file
- `--output string` — Output format: `text` (default) or `json`

## License
MIT
