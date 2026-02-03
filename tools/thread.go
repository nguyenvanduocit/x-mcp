package tools

import (
	"context"
	"fmt"
	"strings"

	twitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/x-mcp/services"
	"gopkg.in/yaml.v3"
)

type PostThreadInput struct {
	Texts string `json:"texts" validate:"required"`
}

func RegisterThreadTools(s *server.MCPServer) {
	tool := mcp.NewTool("x_post_thread",
		mcp.WithDescription("Post a thread (multiple tweets in sequence) on X/Twitter. Each tweet is posted as a reply to the previous one."),
		mcp.WithString("texts", mcp.Required(), mcp.Description("The tweets in the thread, separated by '|||'. Each segment becomes one tweet (max 280 chars each). Example: 'First tweet|||Second tweet|||Third tweet'")),
	)
	s.AddTool(tool, mcp.NewTypedToolHandler(postThreadHandler))
}

func postThreadHandler(ctx context.Context, request mcp.CallToolRequest, input PostThreadInput) (*mcp.CallToolResult, error) {
	client := services.TwitterClient()

	parts := strings.Split(input.Texts, "|||")
	if len(parts) < 2 {
		return nil, fmt.Errorf("thread must contain at least 2 tweets separated by '|||'")
	}

	// Trim whitespace from each part
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	type postedTweet struct {
		Index int    `yaml:"index"`
		ID    string `yaml:"id"`
		Text  string `yaml:"text"`
	}

	posted := make([]postedTweet, 0, len(parts))
	var lastTweetID string

	for i, text := range parts {
		if text == "" {
			continue
		}

		req := twitter.CreateTweetRequest{
			Text: text,
		}

		if lastTweetID != "" {
			req.Reply = &twitter.CreateTweetReply{
				InReplyToTweetID: lastTweetID,
			}
		}

		resp, err := client.CreateTweet(ctx, req)
		if err != nil {
			// Return what we've posted so far plus the error
			result := map[string]interface{}{
				"error":          fmt.Sprintf("failed to post tweet %d: %v", i+1, err),
				"posted_so_far":  posted,
				"failed_at_index": i + 1,
			}
			out, _ := yaml.Marshal(result)
			return mcp.NewToolResultText(string(out)), nil
		}

		lastTweetID = resp.Tweet.ID
		posted = append(posted, postedTweet{
			Index: i + 1,
			ID:    resp.Tweet.ID,
			Text:  text,
		})
	}

	out, err := yaml.Marshal(map[string]interface{}{
		"thread_length": len(posted),
		"tweets":        posted,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal thread: %w", err)
	}

	return mcp.NewToolResultText(string(out)), nil
}
