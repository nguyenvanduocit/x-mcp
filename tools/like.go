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
