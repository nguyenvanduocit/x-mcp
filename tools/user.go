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

type GetUserInput struct {
	Username string `json:"username" validate:"required"`
}

func RegisterUserTools(s *server.MCPServer) {
	tool := mcp.NewTool("x_get_user",
		mcp.WithDescription("Get a user's profile information on X/Twitter by username"),
		mcp.WithString("username", mcp.Required(), mcp.Description("The username to look up (without @ symbol)")),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(getUserHandler))
}

func getUserHandler(ctx context.Context, request mcp.CallToolRequest, input GetUserInput) (*mcp.CallToolResult, error) {
	client := services.TwitterClient()

	opts := twitter.UserLookupOpts{
		Expansions: []twitter.Expansion{
			twitter.ExpansionPinnedTweetID,
		},
		UserFields: []twitter.UserField{
			twitter.UserFieldCreatedAt,
			twitter.UserFieldDescription,
			twitter.UserFieldPublicMetrics,
			twitter.UserFieldProfileImageURL,
			twitter.UserFieldVerified,
			twitter.UserFieldLocation,
			twitter.UserFieldURL,
		},
	}

	resp, err := client.UserNameLookup(ctx, []string{input.Username}, opts)
	if err != nil {
		return nil, fmt.Errorf("user lookup: %w", err)
	}

	dictionaries := resp.Raw.UserDictionaries()
	if len(dictionaries) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("User @%s not found", input.Username)), nil
	}

	for _, dict := range dictionaries {
		user := map[string]interface{}{
			"id":       dict.User.ID,
			"username": dict.User.UserName,
			"name":     dict.User.Name,
		}
		if dict.User.Description != "" {
			user["bio"] = dict.User.Description
		}
		if dict.User.Location != "" {
			user["location"] = dict.User.Location
		}
		if dict.User.URL != "" {
			user["url"] = dict.User.URL
		}
		if dict.User.CreatedAt != "" {
			user["created_at"] = dict.User.CreatedAt
		}
		if dict.User.Verified {
			user["verified"] = true
		}
		if dict.User.PublicMetrics != nil {
			user["metrics"] = map[string]int{
				"followers":  dict.User.PublicMetrics.Followers,
				"following":  dict.User.PublicMetrics.Following,
				"tweets":     dict.User.PublicMetrics.Tweets,
				"listed":     dict.User.PublicMetrics.Listed,
			}
		}
		if dict.PinnedTweet != nil {
			user["pinned_tweet"] = map[string]string{
				"id":   dict.PinnedTweet.ID,
				"text": dict.PinnedTweet.Text,
			}
		}

		out, err := yaml.Marshal(user)
		if err != nil {
			return nil, fmt.Errorf("marshal user: %w", err)
		}
		return mcp.NewToolResultText(string(out)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("User @%s not found", input.Username)), nil
}
