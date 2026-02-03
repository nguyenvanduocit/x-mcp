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
