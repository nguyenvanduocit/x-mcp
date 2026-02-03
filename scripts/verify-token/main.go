package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dghubble/oauth1"
	twitter "github.com/g8rswimmer/go-twitter/v2"
)

type oauth1Auth struct{}

func (a *oauth1Auth) Add(req *http.Request) {}

func main() {
	apiKey := flag.String("api-key", "", "X API Key (Consumer Key)")
	apiSecret := flag.String("api-secret", "", "X API Secret (Consumer Secret)")
	accessToken := flag.String("access-token", "", "X Access Token")
	accessTokenSecret := flag.String("access-token-secret", "", "X Access Token Secret")
	envOutput := flag.String("output", ".env", "Path to save .env file")
	flag.Parse()

	if *apiKey == "" || *apiSecret == "" || *accessToken == "" || *accessTokenSecret == "" {
		fmt.Println("X/Twitter API Token Verification & Setup")
		fmt.Println()
		fmt.Println("Get your credentials from: https://developer.x.com/en/portal/dashboard")
		fmt.Println("  1. Create a project and app")
		fmt.Println("  2. Go to 'Keys and tokens'")
		fmt.Println("  3. Generate 'Consumer Keys' (API Key & Secret)")
		fmt.Println("  4. Generate 'Authentication Tokens' (Access Token & Secret)")
		fmt.Println("  5. Make sure your app has Read and Write permissions")
		fmt.Println()
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Println("Verifying credentials...")

	config := oauth1.NewConfig(*apiKey, *apiSecret)
	token := oauth1.NewToken(*accessToken, *accessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	client := &twitter.Client{
		Authorizer: &oauth1Auth{},
		Client:     httpClient,
		Host:       "https://api.twitter.com",
	}

	// Verify by looking up the authenticated user
	opts := twitter.UserLookupOpts{
		UserFields: []twitter.UserField{
			twitter.UserFieldUserName,
			twitter.UserFieldName,
			twitter.UserFieldDescription,
			twitter.UserFieldPublicMetrics,
		},
	}

	resp, err := client.AuthUserLookup(context.Background(), opts)
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	dictionaries := resp.Raw.UserDictionaries()
	if len(dictionaries) == 0 {
		log.Fatal("Authentication failed: no user returned")
	}

	for _, dict := range dictionaries {
		enc, _ := json.MarshalIndent(map[string]interface{}{
			"id":       dict.User.ID,
			"username": dict.User.UserName,
			"name":     dict.User.Name,
		}, "", "  ")
		fmt.Println("Authenticated as:")
		fmt.Println(string(enc))
		break
	}

	// Save .env file
	envContent := fmt.Sprintf("X_API_KEY=%s\nX_API_SECRET=%s\nX_ACCESS_TOKEN=%s\nX_ACCESS_TOKEN_SECRET=%s\n",
		*apiKey, *apiSecret, *accessToken, *accessTokenSecret)

	if err := os.WriteFile(*envOutput, []byte(envContent), 0600); err != nil {
		log.Fatalf("Failed to write .env file: %v", err)
	}

	fmt.Printf("\nCredentials saved to %s\n", *envOutput)
}
