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

type GetUserTimelineInput struct {
	UserID     string `json:"user_id" validate:"required"`
	MaxResults int    `json:"max_results,omitempty"`
}

func RegisterTimelineTools(s *server.MCPServer) {
	tool := mcp.NewTool("x_get_user_timeline",
		mcp.WithDescription("Get recent tweets from a user's timeline by their user ID. Use x_get_user first to get the user ID from a username."),
		mcp.WithString("user_id", mcp.Required(), mcp.Description("The user's ID (numeric). Use x_get_user to get the ID from a username.")),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of tweets to return (5-100, default: 10)")),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(getUserTimelineHandler))
}

func getUserTimelineHandler(ctx context.Context, request mcp.CallToolRequest, input GetUserTimelineInput) (*mcp.CallToolResult, error) {
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

	opts := twitter.UserTweetTimelineOpts{
		Expansions: []twitter.Expansion{
			twitter.ExpansionAuthorID,
		},
		TweetFields: []twitter.TweetField{
			twitter.TweetFieldCreatedAt,
			twitter.TweetFieldPublicMetrics,
			twitter.TweetFieldConversationID,
			twitter.TweetFieldEntities,
		},
		UserFields: []twitter.UserField{
			twitter.UserFieldUserName,
			twitter.UserFieldName,
		},
		MaxResults: maxResults,
	}

	resp, err := client.UserTweetTimeline(ctx, input.UserID, opts)
	if err != nil {
		return nil, fmt.Errorf("user timeline: %w", err)
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
		results = append(results, tweet)
	}

	out, err := yaml.Marshal(map[string]interface{}{
		"user_id": input.UserID,
		"count":   len(results),
		"tweets":  results,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal timeline: %w", err)
	}

	return mcp.NewToolResultText(string(out)), nil
}
