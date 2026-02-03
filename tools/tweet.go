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

type GetTweetInput struct {
	TweetID string `json:"tweet_id" validate:"required"`
}

type PostTweetInput struct {
	Text      string `json:"text" validate:"required"`
	ReplyToID string `json:"reply_to_id,omitempty"`
}

func RegisterTweetTools(s *server.MCPServer) {
	getTool := mcp.NewTool("x_get_tweet",
		mcp.WithDescription("Get a specific tweet by its ID with full details including metrics and author info"),
		mcp.WithString("tweet_id", mcp.Required(), mcp.Description("The tweet ID to look up")),
	)
	s.AddTool(getTool, mcp.NewTypedToolHandler(getTweetHandler))

	postTool := mcp.NewTool("x_post_tweet",
		mcp.WithDescription("Post a new tweet on X/Twitter. Can also reply to an existing tweet."),
		mcp.WithString("text", mcp.Required(), mcp.Description("The text content of the tweet (max 280 characters)")),
		mcp.WithString("reply_to_id", mcp.Description("Optional tweet ID to reply to")),
	)
	s.AddTool(postTool, mcp.NewTypedToolHandler(postTweetHandler))
}

func getTweetHandler(ctx context.Context, request mcp.CallToolRequest, input GetTweetInput) (*mcp.CallToolResult, error) {
	client := services.TwitterClient()

	opts := twitter.TweetLookupOpts{
		Expansions: []twitter.Expansion{
			twitter.ExpansionAuthorID,
			twitter.ExpansionEntitiesMentionsUserName,
		},
		TweetFields: []twitter.TweetField{
			twitter.TweetFieldCreatedAt,
			twitter.TweetFieldPublicMetrics,
			twitter.TweetFieldConversationID,
			twitter.TweetFieldEntities,
			twitter.TweetFieldAttachments,
			twitter.TweetFieldContextAnnotations,
		},
		UserFields: []twitter.UserField{
			twitter.UserFieldUserName,
			twitter.UserFieldName,
			twitter.UserFieldVerified,
		},
	}

	resp, err := client.TweetLookup(ctx, []string{input.TweetID}, opts)
	if err != nil {
		return nil, fmt.Errorf("tweet lookup: %w", err)
	}

	dictionaries := resp.Raw.TweetDictionaries()
	if len(dictionaries) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("Tweet %s not found", input.TweetID)), nil
	}

	for _, dict := range dictionaries {
		tweet := map[string]interface{}{
			"id":   dict.Tweet.ID,
			"text": dict.Tweet.Text,
		}
		if dict.Tweet.CreatedAt != "" {
			tweet["created_at"] = dict.Tweet.CreatedAt
		}
		if dict.Tweet.ConversationID != "" {
			tweet["conversation_id"] = dict.Tweet.ConversationID
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

		out, err := yaml.Marshal(tweet)
		if err != nil {
			return nil, fmt.Errorf("marshal tweet: %w", err)
		}
		return mcp.NewToolResultText(string(out)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Tweet %s not found", input.TweetID)), nil
}

func postTweetHandler(ctx context.Context, request mcp.CallToolRequest, input PostTweetInput) (*mcp.CallToolResult, error) {
	client := services.TwitterClient()

	req := twitter.CreateTweetRequest{
		Text: input.Text,
	}

	if input.ReplyToID != "" {
		req.Reply = &twitter.CreateTweetReply{
			InReplyToTweetID: input.ReplyToID,
		}
	}

	resp, err := client.CreateTweet(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("post tweet: %w", err)
	}

	out, err := yaml.Marshal(map[string]string{
		"id":   resp.Tweet.ID,
		"text": input.Text,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal response: %w", err)
	}

	return mcp.NewToolResultText(string(out)), nil
}
