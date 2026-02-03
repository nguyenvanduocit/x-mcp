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
