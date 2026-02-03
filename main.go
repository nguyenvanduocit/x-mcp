package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nguyenvanduocit/x-mcp/tools"
)

func main() {
	envFile := flag.String("env", "", "Path to environment file")
	httpPort := flag.String("http_port", "", "Port for HTTP server")
	flag.Parse()

	if *envFile != "" {
		if err := godotenv.Load(*envFile); err != nil {
			fmt.Printf("Warning: Error loading env file %s: %v\n", *envFile, err)
		}
	}

	requiredEnvs := []string{"X_API_KEY", "X_API_SECRET", "X_ACCESS_TOKEN", "X_ACCESS_TOKEN_SECRET"}
	var missing []string
	for _, env := range requiredEnvs {
		if os.Getenv(env) == "" {
			missing = append(missing, env)
		}
	}

	if len(missing) > 0 {
		fmt.Println("Missing required environment variables:")
		for _, env := range missing {
			fmt.Printf("  - %s\n", env)
		}
		fmt.Println()
		fmt.Println("Get your API keys from: https://developer.x.com/en/portal/dashboard")
		os.Exit(1)
	}

	mcpServer := server.NewMCPServer(
		"X MCP",
		"1.0.0",
		server.WithLogging(),
		server.WithRecovery(),
	)

	tools.RegisterSearchTools(mcpServer)
	tools.RegisterUserTools(mcpServer)
	tools.RegisterTimelineTools(mcpServer)
	tools.RegisterTweetTools(mcpServer)
	tools.RegisterThreadTools(mcpServer)

	if *httpPort != "" {
		log.Printf("X MCP server available at http://localhost:%s/mcp", *httpPort)
		httpServer := server.NewStreamableHTTPServer(mcpServer, server.WithEndpointPath("/mcp"))
		if err := httpServer.Start(fmt.Sprintf(":%s", *httpPort)); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}
