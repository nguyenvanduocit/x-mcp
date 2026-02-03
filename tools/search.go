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

type SearchTweetsInput struct {
	Query      string `json:"query" validate:"required"`
	MaxResults int    `json:"max_results,omitempty"`
}

func RegisterSearchTools(s *server.MCPServer) {
	tool := mcp.NewTool("x_search_tweets",
		mcp.WithDescription("Search recent tweets on X/Twitter using Twitter search syntax. Returns tweets from the last 7 days."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query using Twitter search syntax (e.g., 'from:elonmusk', '#golang', 'AI lang:en')")),
		mcp.WithNumber("max_results", mcp.Description("Maximum number of results to return (10-100, default: 10)")),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(searchTweetsHandler))
}

func searchTweetsHandler(ctx context.Context, request mcp.CallToolRequest, input SearchTweetsInput) (*mcp.CallToolResult, error) {
	client := services.TwitterClient()

	maxResults := input.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}
	if maxResults > 100 {
		maxResults = 100
	}

	opts := twitter.TweetRecentSearchOpts{
		Expansions: []twitter.Expansion{
			twitter.ExpansionAuthorID,
			twitter.ExpansionEntitiesMentionsUserName,
		},
		TweetFields: []twitter.TweetField{
			twitter.TweetFieldCreatedAt,
			twitter.TweetFieldPublicMetrics,
			twitter.TweetFieldAuthorID,
			twitter.TweetFieldConversationID,
			twitter.TweetFieldEntities,
		},
		UserFields: []twitter.UserField{
			twitter.UserFieldUserName,
			twitter.UserFieldName,
		},
		MaxResults: maxResults,
	}

	resp, err := client.TweetRecentSearch(ctx, input.Query, opts)
	if err != nil {
		return nil, fmt.Errorf("search tweets: %w", err)
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
		"query":   input.Query,
		"count":   len(results),
		"results": results,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal results: %w", err)
	}

	return mcp.NewToolResultText(string(out)), nil
}
