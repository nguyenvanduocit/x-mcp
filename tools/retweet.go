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
