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
