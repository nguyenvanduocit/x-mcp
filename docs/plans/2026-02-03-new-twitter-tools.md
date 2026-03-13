# New Twitter Tools Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add 13 new MCP tools to the X MCP server: like/unlike, retweet/unretweet, delete tweet, get mentions, quote tweets lookup, mute/unmute, block/unblock, get list tweets, get user lists.

**Architecture:** Each tool follows the existing pattern — typed input struct, tool registration function, handler function, YAML output. Tools requiring the authenticated user's ID (like, retweet, mute, block) use a shared `AuthUserID()` singleton in the services package. New tools are grouped into new files by domain: `like.go`, `retweet.go`, `mention.go`, `quote.go`, `moderation.go` (mute+block), `list.go`. Each file has its own `Register*Tools` function called from `main.go`.

**Tech Stack:** Go 1.25, go-twitter/v2, mcp-go, OAuth 1.0a


## Patterns from Existing Code

Every tool follows this exact pattern:

```go
// 1. Input struct with json + validate tags
type FooInput struct {
    Field string `json:"field" validate:"required"`
}

// 2. Register function creates tool with mcp.NewTool and adds handler
func RegisterFooTools(s *server.MCPServer) {
    tool := mcp.NewTool("x_tool_name",
        mcp.WithDescription("..."),
        mcp.WithString("field", mcp.Required(), mcp.Description("...")),
    )
    s.AddTool(tool, mcp.NewTypedToolHandler(fooHandler))
}

// 3. Handler gets client, calls API, builds map, marshals YAML
func fooHandler(ctx context.Context, request mcp.CallToolRequest, input FooInput) (*mcp.CallToolResult, error) {
    client := services.TwitterClient()
    // ... call API ...
    out, err := yaml.Marshal(result)
    if err != nil {
        return nil, fmt.Errorf("marshal: %w", err)
    }
    return mcp.NewToolResultText(string(out)), nil
}
```

Imports are always:
```go
import (
    "context"
    "fmt"

    twitter "github.com/g8rswimmer/go-twitter/v2"
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
    "github.com/nguyenvanduocit/x-mcp/services"
    "gopkg.in/yaml.v3"
)
```

## Prerequisite: AuthUserID

Many tools (like, retweet, mute, block) need the authenticated user's numeric ID. The go-twitter library provides `client.AuthUserLookup()` for this. We add a `sync.OnceValue` singleton in `services/twitter.go` — same pattern as `TwitterClient`.

---

### Task 1: Add AuthUserID to services

**Files:**
- Modify: `services/twitter.go`

**Step 1: Add AuthUserID singleton**

Add this after the existing `TwitterClient` var at line 49:

```go
var AuthUserID = sync.OnceValue(func() string {
	client := TwitterClient()

	opts := twitter.UserLookupOpts{
		UserFields: []twitter.UserField{
			twitter.UserFieldUserName,
		},
	}

	resp, err := client.AuthUserLookup(context.Background(), opts)
	if err != nil {
		panic(fmt.Sprintf("failed to get authenticated user: %v", err))
	}

	dictionaries := resp.Raw.UserDictionaries()
	for _, dict := range dictionaries {
		return dict.User.ID
	}

	panic("no authenticated user found")
})
```

Also add `"context"` to the imports.

**Step 2: Build and verify**

Run: `cd /Volumes/Data/Projects/claudeserver/x-mcp && go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add services/twitter.go
git commit -m "feat: add AuthUserID singleton for authenticated user lookup"
```

---

### Task 2: Add like/unlike tools

**Files:**
- Create: `tools/like.go`
- Modify: `main.go` (add `tools.RegisterLikeTools(mcpServer)`)

**Step 1: Create `tools/like.go`**

```go
package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/x-mcp/services"
	"gopkg.in/yaml.v3"
)

type LikeTweetInput struct {
	TweetID string `json:"tweet_id" validate:"required"`
}

type UnlikeTweetInput struct {
	TweetID string `json:"tweet_id" validate:"required"`
}

func RegisterLikeTools(s *server.MCPServer) {
	likeTool := mcp.NewTool("x_like_tweet",
		mcp.WithDescription("Like a tweet on X/Twitter"),
		mcp.WithString("tweet_id", mcp.Required(), mcp.Description("The tweet ID to like")),
	)
	s.AddTool(likeTool, mcp.NewTypedToolHandler(likeTweetHandler))

	unlikeTool := mcp.NewTool("x_unlike_tweet",
		mcp.WithDescription("Unlike a previously liked tweet on X/Twitter"),
		mcp.WithString("tweet_id", mcp.Required(), mcp.Description("The tweet ID to unlike")),
	)
	s.AddTool(unlikeTool, mcp.NewTypedToolHandler(unlikeTweetHandler))
}

func likeTweetHandler(ctx context.Context, request mcp.CallToolRequest, input LikeTweetInput) (*mcp.CallToolResult, error) {
	client := services.TwitterClient()
	userID := services.AuthUserID()

	_, err := client.UserLikes(ctx, userID, input.TweetID)
	if err != nil {
		return nil, fmt.Errorf("like tweet: %w", err)
	}

	out, err := yaml.Marshal(map[string]string{
		"status":   "liked",
		"tweet_id": input.TweetID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return mcp.NewToolResultText(string(out)), nil
}

func unlikeTweetHandler(ctx context.Context, request mcp.CallToolRequest, input UnlikeTweetInput) (*mcp.CallToolResult, error) {
	client := services.TwitterClient()
	userID := services.AuthUserID()

	_, err := client.DeleteUserLikes(ctx, userID, input.TweetID)
	if err != nil {
		return nil, fmt.Errorf("unlike tweet: %w", err)
	}

	out, err := yaml.Marshal(map[string]string{
		"status":   "unliked",
		"tweet_id": input.TweetID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return mcp.NewToolResultText(string(out)), nil
}
```

**Step 2: Register in main.go**

Add `tools.RegisterLikeTools(mcpServer)` after the existing `RegisterThreadTools` call in `main.go`.

**Step 3: Build and verify**

Run: `cd /Volumes/Data/Projects/claudeserver/x-mcp && go build ./...`
Expected: No errors

**Step 4: Commit**

```bash
git add tools/like.go main.go
git commit -m "feat: add x_like_tweet and x_unlike_tweet tools"
```

---

### Task 3: Add retweet/unretweet tools

**Files:**
- Create: `tools/retweet.go`
- Modify: `main.go` (add `tools.RegisterRetweetTools(mcpServer)`)

**Step 1: Create `tools/retweet.go`**

```go
package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/x-mcp/services"
	"gopkg.in/yaml.v3"
)

type RetweetInput struct {
	TweetID string `json:"tweet_id" validate:"required"`
}

type UnretweetInput struct {
	TweetID string `json:"tweet_id" validate:"required"`
}

func RegisterRetweetTools(s *server.MCPServer) {
	retweetTool := mcp.NewTool("x_retweet",
		mcp.WithDescription("Retweet a tweet on X/Twitter"),
		mcp.WithString("tweet_id", mcp.Required(), mcp.Description("The tweet ID to retweet")),
	)
	s.AddTool(retweetTool, mcp.NewTypedToolHandler(retweetHandler))

	unretweetTool := mcp.NewTool("x_unretweet",
		mcp.WithDescription("Undo a retweet on X/Twitter"),
		mcp.WithString("tweet_id", mcp.Required(), mcp.Description("The tweet ID to unretweet")),
	)
	s.AddTool(unretweetTool, mcp.NewTypedToolHandler(unretweetHandler))
}

func retweetHandler(ctx context.Context, request mcp.CallToolRequest, input RetweetInput) (*mcp.CallToolResult, error) {
	client := services.TwitterClient()
	userID := services.AuthUserID()

	_, err := client.UserRetweet(ctx, userID, input.TweetID)
	if err != nil {
		return nil, fmt.Errorf("retweet: %w", err)
	}

	out, err := yaml.Marshal(map[string]string{
		"status":   "retweeted",
		"tweet_id": input.TweetID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return mcp.NewToolResultText(string(out)), nil
}

func unretweetHandler(ctx context.Context, request mcp.CallToolRequest, input UnretweetInput) (*mcp.CallToolResult, error) {
	client := services.TwitterClient()
	userID := services.AuthUserID()

	_, err := client.DeleteUserRetweet(ctx, userID, input.TweetID)
	if err != nil {
		return nil, fmt.Errorf("unretweet: %w", err)
	}

	out, err := yaml.Marshal(map[string]string{
		"status":   "unretweeted",
		"tweet_id": input.TweetID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return mcp.NewToolResultText(string(out)), nil
}
```

**Step 2: Register in main.go**

Add `tools.RegisterRetweetTools(mcpServer)` in `main.go`.

**Step 3: Build and verify**

Run: `cd /Volumes/Data/Projects/claudeserver/x-mcp && go build ./...`
Expected: No errors

**Step 4: Commit**

```bash
git add tools/retweet.go main.go
git commit -m "feat: add x_retweet and x_unretweet tools"
```

---

### Task 4: Add delete tweet tool

**Files:**
- Modify: `tools/tweet.go` (add delete tool to existing RegisterTweetTools)

**Step 1: Add DeleteTweetInput and handler to `tools/tweet.go`**

Add input struct:

```go
type DeleteTweetInput struct {
	TweetID string `json:"tweet_id" validate:"required"`
}
```

Add tool registration inside `RegisterTweetTools`:

```go
deleteTool := mcp.NewTool("x_delete_tweet",
    mcp.WithDescription("Delete a tweet you posted on X/Twitter"),
    mcp.WithString("tweet_id", mcp.Required(), mcp.Description("The tweet ID to delete")),
)
s.AddTool(deleteTool, mcp.NewTypedToolHandler(deleteTweetHandler))
```

Add handler:

```go
func deleteTweetHandler(ctx context.Context, request mcp.CallToolRequest, input DeleteTweetInput) (*mcp.CallToolResult, error) {
	client := services.TwitterClient()

	_, err := client.DeleteTweet(ctx, input.TweetID)
	if err != nil {
		return nil, fmt.Errorf("delete tweet: %w", err)
	}

	out, err := yaml.Marshal(map[string]string{
		"status":   "deleted",
		"tweet_id": input.TweetID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return mcp.NewToolResultText(string(out)), nil
}
```

**Step 2: Build and verify**

Run: `cd /Volumes/Data/Projects/claudeserver/x-mcp && go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add tools/tweet.go
git commit -m "feat: add x_delete_tweet tool"
```

---

### Task 5: Add get mentions tool

**Files:**
- Create: `tools/mention.go`
- Modify: `main.go` (add `tools.RegisterMentionTools(mcpServer)`)

**Step 1: Create `tools/mention.go`**

```go
package tools

import (
	"context"
	"fmt"

	twitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/x-mcp/services"
	"gopkg.in/yaml.v3"
)

type GetMentionsInput struct {
	UserID     string `json:"user_id" validate:"required"`
	MaxResults int    `json:"max_results,omitempty"`
}

func RegisterMentionTools(s *server.MCPServer) {
	tool := mcp.NewTool("x_get_mentions",
		mcp.WithDescription("Get recent tweets that mention a user by their user ID. Use x_get_user first to get the user ID from a username."),
		mcp.WithString("user_id", mcp.Required(), mcp.Description("The user's ID (numeric). Use x_get_user to get the ID from a username.")),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of mentions to return (5-100, default: 10)")),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(getMentionsHandler))
}

func getMentionsHandler(ctx context.Context, request mcp.CallToolRequest, input GetMentionsInput) (*mcp.CallToolResult, error) {
	client := services.TwitterClient()

	maxResults := input.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}
	if maxResults < 5 {
		maxResults = 5
	}
	if maxResults > 100 {
		maxResults = 100
	}

	opts := twitter.UserMentionTimelineOpts{
		Expansions: []twitter.Expansion{
			twitter.ExpansionAuthorID,
			twitter.ExpansionEntitiesMentionsUserName,
		},
		TweetFields: []twitter.TweetField{
			twitter.TweetFieldCreatedAt,
			twitter.TweetFieldPublicMetrics,
			twitter.TweetFieldConversationID,
		},
		UserFields: []twitter.UserField{
			twitter.UserFieldUserName,
			twitter.UserFieldName,
		},
		MaxResults: maxResults,
	}

	resp, err := client.UserMentionTimeline(ctx, input.UserID, opts)
	if err != nil {
		return nil, fmt.Errorf("get mentions: %w", err)
	}

	dictionaries := resp.Raw.TweetDictionaries()

	results := make([]map[string]interface{}, 0, len(dictionaries))
	for _, dict := range dictionaries {
		tweet := map[string]interface{}{
			"id":   dict.Tweet.ID,
			"text": dict.Tweet.Text,
		}
		if dict.Tweet.CreatedAt != "" {
			tweet["created_at"] = dict.Tweet.CreatedAt
		}
		if dict.Tweet.PublicMetrics != nil {
			tweet["metrics"] = map[string]int{
				"likes":    dict.Tweet.PublicMetrics.Likes,
				"retweets": dict.Tweet.PublicMetrics.Retweets,
				"replies":  dict.Tweet.PublicMetrics.Replies,
				"quotes":   dict.Tweet.PublicMetrics.Quotes,
			}
		}
		if dict.Author != nil {
			tweet["author"] = map[string]string{
				"id":       dict.Author.ID,
				"username": dict.Author.UserName,
				"name":     dict.Author.Name,
			}
		}
		results = append(results, tweet)
	}

	out, err := yaml.Marshal(map[string]interface{}{
		"user_id":  input.UserID,
		"count":    len(results),
		"mentions": results,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal mentions: %w", err)
	}

	return mcp.NewToolResultText(string(out)), nil
}
```

**Step 2: Register in main.go**

Add `tools.RegisterMentionTools(mcpServer)` in `main.go`.

**Step 3: Build and verify**

Run: `cd /Volumes/Data/Projects/claudeserver/x-mcp && go build ./...`
Expected: No errors

**Step 4: Commit**

```bash
git add tools/mention.go main.go
git commit -m "feat: add x_get_mentions tool"
```

---

### Task 6: Add quote tweets lookup tool

**Files:**
- Create: `tools/quote.go`
- Modify: `main.go` (add `tools.RegisterQuoteTools(mcpServer)`)

**Step 1: Create `tools/quote.go`**

```go
package tools

import (
	"context"
	"fmt"

	twitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/x-mcp/services"
	"gopkg.in/yaml.v3"
)

type GetQuoteTweetsInput struct {
	TweetID    string `json:"tweet_id" validate:"required"`
	MaxResults int    `json:"max_results,omitempty"`
}

func RegisterQuoteTools(s *server.MCPServer) {
	tool := mcp.NewTool("x_get_quote_tweets",
		mcp.WithDescription("Get tweets that quote a specific tweet on X/Twitter"),
		mcp.WithString("tweet_id", mcp.Required(), mcp.Description("The tweet ID to find quotes for")),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of results to return (10-100, default: 10)")),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(getQuoteTweetsHandler))
}

func getQuoteTweetsHandler(ctx context.Context, request mcp.CallToolRequest, input GetQuoteTweetsInput) (*mcp.CallToolResult, error) {
	client := services.TwitterClient()

	maxResults := input.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}
	if maxResults > 100 {
		maxResults = 100
	}

	opts := twitter.QuoteTweetsLookupOpts{
		Expansions: []twitter.Expansion{
			twitter.ExpansionAuthorID,
		},
		TweetFields: []twitter.TweetField{
			twitter.TweetFieldCreatedAt,
			twitter.TweetFieldPublicMetrics,
		},
		UserFields: []twitter.UserField{
			twitter.UserFieldUserName,
			twitter.UserFieldName,
		},
		MaxResults: maxResults,
	}

	resp, err := client.QuoteTweetsLookup(ctx, input.TweetID, opts)
	if err != nil {
		return nil, fmt.Errorf("quote tweets lookup: %w", err)
	}

	dictionaries := resp.Raw.TweetDictionaries()

	results := make([]map[string]interface{}, 0, len(dictionaries))
	for _, dict := range dictionaries {
		tweet := map[string]interface{}{
			"id":   dict.Tweet.ID,
			"text": dict.Tweet.Text,
		}
		if dict.Tweet.CreatedAt != "" {
			tweet["created_at"] = dict.Tweet.CreatedAt
		}
		if dict.Tweet.PublicMetrics != nil {
			tweet["metrics"] = map[string]int{
				"likes":    dict.Tweet.PublicMetrics.Likes,
				"retweets": dict.Tweet.PublicMetrics.Retweets,
				"replies":  dict.Tweet.PublicMetrics.Replies,
				"quotes":   dict.Tweet.PublicMetrics.Quotes,
			}
		}
		if dict.Author != nil {
			tweet["author"] = map[string]string{
				"id":       dict.Author.ID,
				"username": dict.Author.UserName,
				"name":     dict.Author.Name,
			}
		}
		results = append(results, tweet)
	}

	out, err := yaml.Marshal(map[string]interface{}{
		"tweet_id": input.TweetID,
		"count":    len(results),
		"quotes":   results,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal quotes: %w", err)
	}

	return mcp.NewToolResultText(string(out)), nil
}
```

**Step 2: Register in main.go**

Add `tools.RegisterQuoteTools(mcpServer)` in `main.go`.

**Step 3: Build and verify**

Run: `cd /Volumes/Data/Projects/claudeserver/x-mcp && go build ./...`
Expected: No errors

**Step 4: Commit**

```bash
git add tools/quote.go main.go
git commit -m "feat: add x_get_quote_tweets tool"
```

---

### Task 7: Add mute/unmute and block/unblock tools

**Files:**
- Create: `tools/moderation.go`
- Modify: `main.go` (add `tools.RegisterModerationTools(mcpServer)`)

**Step 1: Create `tools/moderation.go`**

```go
package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/x-mcp/services"
	"gopkg.in/yaml.v3"
)

type MuteUserInput struct {
	TargetUserID string `json:"target_user_id" validate:"required"`
}

type UnmuteUserInput struct {
	TargetUserID string `json:"target_user_id" validate:"required"`
}

type BlockUserInput struct {
	TargetUserID string `json:"target_user_id" validate:"required"`
}

type UnblockUserInput struct {
	TargetUserID string `json:"target_user_id" validate:"required"`
}

func RegisterModerationTools(s *server.MCPServer) {
	muteTool := mcp.NewTool("x_mute_user",
		mcp.WithDescription("Mute a user on X/Twitter. Use x_get_user first to get the user ID from a username."),
		mcp.WithString("target_user_id", mcp.Required(), mcp.Description("The user ID to mute (numeric). Use x_get_user to get the ID.")),
	)
	s.AddTool(muteTool, mcp.NewTypedToolHandler(muteUserHandler))

	unmuteTool := mcp.NewTool("x_unmute_user",
		mcp.WithDescription("Unmute a previously muted user on X/Twitter. Use x_get_user first to get the user ID from a username."),
		mcp.WithString("target_user_id", mcp.Required(), mcp.Description("The user ID to unmute (numeric). Use x_get_user to get the ID.")),
	)
	s.AddTool(unmuteTool, mcp.NewTypedToolHandler(unmuteUserHandler))

	blockTool := mcp.NewTool("x_block_user",
		mcp.WithDescription("Block a user on X/Twitter. Use x_get_user first to get the user ID from a username."),
		mcp.WithString("target_user_id", mcp.Required(), mcp.Description("The user ID to block (numeric). Use x_get_user to get the ID.")),
	)
	s.AddTool(blockTool, mcp.NewTypedToolHandler(blockUserHandler))

	unblockTool := mcp.NewTool("x_unblock_user",
		mcp.WithDescription("Unblock a previously blocked user on X/Twitter. Use x_get_user first to get the user ID from a username."),
		mcp.WithString("target_user_id", mcp.Required(), mcp.Description("The user ID to unblock (numeric). Use x_get_user to get the ID.")),
	)
	s.AddTool(unblockTool, mcp.NewTypedToolHandler(unblockUserHandler))
}

func muteUserHandler(ctx context.Context, request mcp.CallToolRequest, input MuteUserInput) (*mcp.CallToolResult, error) {
	client := services.TwitterClient()
	userID := services.AuthUserID()

	_, err := client.UserMutes(ctx, userID, input.TargetUserID)
	if err != nil {
		return nil, fmt.Errorf("mute user: %w", err)
	}

	out, err := yaml.Marshal(map[string]string{
		"status":         "muted",
		"target_user_id": input.TargetUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return mcp.NewToolResultText(string(out)), nil
}

func unmuteUserHandler(ctx context.Context, request mcp.CallToolRequest, input UnmuteUserInput) (*mcp.CallToolResult, error) {
	client := services.TwitterClient()
	userID := services.AuthUserID()

	_, err := client.DeleteUserMutes(ctx, userID, input.TargetUserID)
	if err != nil {
		return nil, fmt.Errorf("unmute user: %w", err)
	}

	out, err := yaml.Marshal(map[string]string{
		"status":         "unmuted",
		"target_user_id": input.TargetUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return mcp.NewToolResultText(string(out)), nil
}

func blockUserHandler(ctx context.Context, request mcp.CallToolRequest, input BlockUserInput) (*mcp.CallToolResult, error) {
	client := services.TwitterClient()
	userID := services.AuthUserID()

	_, err := client.UserBlocks(ctx, userID, input.TargetUserID)
	if err != nil {
		return nil, fmt.Errorf("block user: %w", err)
	}

	out, err := yaml.Marshal(map[string]string{
		"status":         "blocked",
		"target_user_id": input.TargetUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return mcp.NewToolResultText(string(out)), nil
}

func unblockUserHandler(ctx context.Context, request mcp.CallToolRequest, input UnblockUserInput) (*mcp.CallToolResult, error) {
	client := services.TwitterClient()
	userID := services.AuthUserID()

	_, err := client.DeleteUserBlocks(ctx, userID, input.TargetUserID)
	if err != nil {
		return nil, fmt.Errorf("unblock user: %w", err)
	}

	out, err := yaml.Marshal(map[string]string{
		"status":         "unblocked",
		"target_user_id": input.TargetUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	return mcp.NewToolResultText(string(out)), nil
}
```

**Step 2: Register in main.go**

Add `tools.RegisterModerationTools(mcpServer)` in `main.go`.

**Step 3: Build and verify**

Run: `cd /Volumes/Data/Projects/claudeserver/x-mcp && go build ./...`
Expected: No errors

**Step 4: Commit**

```bash
git add tools/moderation.go main.go
git commit -m "feat: add mute/unmute and block/unblock tools"
```

---

### Task 8: Add list tools

**Files:**
- Create: `tools/list.go`
- Modify: `main.go` (add `tools.RegisterListTools(mcpServer)`)

**Step 1: Create `tools/list.go`**

```go
package tools

import (
	"context"
	"fmt"

	twitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/x-mcp/services"
	"gopkg.in/yaml.v3"
)

type GetUserListsInput struct {
	UserID string `json:"user_id" validate:"required"`
}

type GetListTweetsInput struct {
	ListID     string `json:"list_id" validate:"required"`
	MaxResults int    `json:"max_results,omitempty"`
}

func RegisterListTools(s *server.MCPServer) {
	userListsTool := mcp.NewTool("x_get_user_lists",
		mcp.WithDescription("Get lists owned by a user on X/Twitter. Use x_get_user first to get the user ID from a username."),
		mcp.WithString("user_id", mcp.Required(), mcp.Description("The user's ID (numeric). Use x_get_user to get the ID from a username.")),
	)
	s.AddTool(userListsTool, mcp.NewTypedToolHandler(getUserListsHandler))

	listTweetsTool := mcp.NewTool("x_get_list_tweets",
		mcp.WithDescription("Get recent tweets from a X/Twitter list. Use x_get_user_lists first to get the list ID."),
		mcp.WithString("list_id", mcp.Required(), mcp.Description("The list ID to get tweets from")),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of tweets to return (1-100, default: 10)")),
	)
	s.AddTool(listTweetsTool, mcp.NewTypedToolHandler(getListTweetsHandler))
}

func getUserListsHandler(ctx context.Context, request mcp.CallToolRequest, input GetUserListsInput) (*mcp.CallToolResult, error) {
	client := services.TwitterClient()

	opts := twitter.UserListLookupOpts{
		ListFields: []twitter.ListField{
			twitter.ListFieldName,
			twitter.ListFieldDescription,
			twitter.ListFieldMemberCount,
			twitter.ListFieldFollowerCount,
			twitter.ListFieldCreatedAt,
			twitter.ListFieldPrivate,
			twitter.ListFieldOwnerID,
		},
	}

	resp, err := client.UserListLookup(ctx, input.UserID, opts)
	if err != nil {
		return nil, fmt.Errorf("user lists lookup: %w", err)
	}

	results := make([]map[string]interface{}, 0)
	if resp.Raw != nil {
		for _, list := range resp.Raw.Lists {
			l := map[string]interface{}{
				"id":   list.ID,
				"name": list.Name,
			}
			if list.Description != "" {
				l["description"] = list.Description
			}
			if list.MemberCount > 0 {
				l["member_count"] = list.MemberCount
			}
			if list.FollowerCount > 0 {
				l["follower_count"] = list.FollowerCount
			}
			if list.CreatedAt != "" {
				l["created_at"] = list.CreatedAt
			}
			l["private"] = list.Private
			results = append(results, l)
		}
	}

	out, err := yaml.Marshal(map[string]interface{}{
		"user_id": input.UserID,
		"count":   len(results),
		"lists":   results,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal lists: %w", err)
	}

	return mcp.NewToolResultText(string(out)), nil
}

func getListTweetsHandler(ctx context.Context, request mcp.CallToolRequest, input GetListTweetsInput) (*mcp.CallToolResult, error) {
	client := services.TwitterClient()

	maxResults := input.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}
	if maxResults > 100 {
		maxResults = 100
	}

	opts := twitter.ListTweetLookupOpts{
		Expansions: []twitter.Expansion{
			twitter.ExpansionAuthorID,
		},
		TweetFields: []twitter.TweetField{
			twitter.TweetFieldCreatedAt,
			twitter.TweetFieldPublicMetrics,
		},
		UserFields: []twitter.UserField{
			twitter.UserFieldUserName,
			twitter.UserFieldName,
		},
		MaxResults: maxResults,
	}

	resp, err := client.ListTweetLookup(ctx, input.ListID, opts)
	if err != nil {
		return nil, fmt.Errorf("list tweets lookup: %w", err)
	}

	dictionaries := resp.Raw.TweetDictionaries()

	results := make([]map[string]interface{}, 0, len(dictionaries))
	for _, dict := range dictionaries {
		tweet := map[string]interface{}{
			"id":   dict.Tweet.ID,
			"text": dict.Tweet.Text,
		}
		if dict.Tweet.CreatedAt != "" {
			tweet["created_at"] = dict.Tweet.CreatedAt
		}
		if dict.Tweet.PublicMetrics != nil {
			tweet["metrics"] = map[string]int{
				"likes":    dict.Tweet.PublicMetrics.Likes,
				"retweets": dict.Tweet.PublicMetrics.Retweets,
				"replies":  dict.Tweet.PublicMetrics.Replies,
				"quotes":   dict.Tweet.PublicMetrics.Quotes,
			}
		}
		if dict.Author != nil {
			tweet["author"] = map[string]string{
				"id":       dict.Author.ID,
				"username": dict.Author.UserName,
				"name":     dict.Author.Name,
			}
		}
		results = append(results, tweet)
	}

	out, err := yaml.Marshal(map[string]interface{}{
		"list_id": input.ListID,
		"count":   len(results),
		"tweets":  results,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal list tweets: %w", err)
	}

	return mcp.NewToolResultText(string(out)), nil
}
```

**Step 2: Register in main.go**

Add `tools.RegisterListTools(mcpServer)` in `main.go`.

**Step 3: Build and verify**

Run: `cd /Volumes/Data/Projects/claudeserver/x-mcp && go build ./...`
Expected: No errors

**Step 4: Commit**

```bash
git add tools/list.go main.go
git commit -m "feat: add x_get_user_lists and x_get_list_tweets tools"
```

---

### Task 9: Update README

**Files:**
- Modify: `README.md`

**Step 1: Update the Available tools section**

Replace the `## Available tools` section with:

```markdown
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
```

**Step 2: Update the "Try it" section**

Add these examples:

```markdown
- "Like this tweet: 123456789"
- "Get my recent mentions"
- "Mute user @spammer"
- "Show me tweets from my 'Tech News' list"
```

**Step 3: Commit**

```bash
git add README.md
git commit -m "docs: update README with new tools"
```

---

### Task 10: Final build verification

**Step 1: Full build**

Run: `cd /Volumes/Data/Projects/claudeserver/x-mcp && go build ./...`
Expected: No errors

**Step 2: Go vet**

Run: `cd /Volumes/Data/Projects/claudeserver/x-mcp && go vet ./...`
Expected: No issues

## Summary

**New files (6):**
- `tools/like.go` - x_like_tweet, x_unlike_tweet
- `tools/retweet.go` - x_retweet, x_unretweet
- `tools/mention.go` - x_get_mentions
- `tools/quote.go` - x_get_quote_tweets
- `tools/moderation.go` - x_mute_user, x_unmute_user, x_block_user, x_unblock_user
- `tools/list.go` - x_get_user_lists, x_get_list_tweets

**Modified files (3):**
- `services/twitter.go` - Add AuthUserID singleton
- `tools/tweet.go` - Add x_delete_tweet
- `main.go` - Register all new tool groups

**Total new tools: 13**
