package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	twitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/nguyenvanduocit/x-mcp/services"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "get-tweet":
		runGetTweet(os.Args[2:])
	case "post-tweet":
		runPostTweet(os.Args[2:])
	case "delete-tweet":
		runDeleteTweet(os.Args[2:])
	case "search":
		runSearch(os.Args[2:])
	case "get-user":
		runGetUser(os.Args[2:])
	case "get-user-timeline":
		runGetUserTimeline(os.Args[2:])
	case "post-thread":
		runPostThread(os.Args[2:])
	case "like-tweet":
		runLikeTweet(os.Args[2:])
	case "unlike-tweet":
		runUnlikeTweet(os.Args[2:])
	case "retweet":
		runRetweet(os.Args[2:])
	case "unretweet":
		runUnretweet(os.Args[2:])
	case "get-mentions":
		runGetMentions(os.Args[2:])
	case "get-quote-tweets":
		runGetQuoteTweets(os.Args[2:])
	case "mute-user":
		runMuteUser(os.Args[2:])
	case "unmute-user":
		runUnmuteUser(os.Args[2:])
	case "block-user":
		runBlockUser(os.Args[2:])
	case "unblock-user":
		runUnblockUser(os.Args[2:])
	case "get-user-lists":
		runGetUserLists(os.Args[2:])
	case "get-list-tweets":
		runGetListTweets(os.Args[2:])
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`x-cli - X/Twitter command-line interface

Usage:
  x-cli <command> [flags]

Commands:
  get-tweet          Get a tweet by ID
  post-tweet         Post a new tweet
  delete-tweet       Delete a tweet
  search             Search recent tweets
  get-user           Get user profile by username
  get-user-timeline  Get a user's recent tweets
  post-thread        Post a thread of tweets
  like-tweet         Like a tweet
  unlike-tweet       Unlike a tweet
  retweet            Retweet a tweet
  unretweet          Undo a retweet
  get-mentions       Get mentions for a user
  get-quote-tweets   Get quote tweets for a tweet
  mute-user          Mute a user
  unmute-user        Unmute a user
  block-user         Block a user
  unblock-user       Unblock a user
  get-user-lists     Get lists owned by a user
  get-list-tweets    Get tweets from a list

Global Flags:
  --env string     Path to .env file
  --output string  Output format: text or json (default "text")

Required Environment Variables:
  X_API_KEY
  X_API_SECRET
  X_ACCESS_TOKEN
  X_ACCESS_TOKEN_SECRET`)
}

func loadEnv(envFile string) {
	if envFile != "" {
		godotenv.Load(envFile)
	}
}

func outputResult(data interface{}, outputFmt string) {
	if outputFmt == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(data); err != nil {
			fmt.Fprintf(os.Stderr, "error encoding json: %v\n", err)
			os.Exit(1)
		}
		return
	}
	// text output: pretty-print as key: value
	printMap(data, 0)
}

func printMap(v interface{}, indent int) {
	prefix := strings.Repeat("  ", indent)
	switch val := v.(type) {
	case map[string]interface{}:
		for k, vv := range val {
			switch inner := vv.(type) {
			case map[string]interface{}:
				fmt.Printf("%s%s:\n", prefix, k)
				printMap(inner, indent+1)
			case []interface{}:
				fmt.Printf("%s%s:\n", prefix, k)
				for i, item := range inner {
					fmt.Printf("%s  [%d]:\n", prefix, i)
					printMap(item, indent+2)
				}
			default:
				fmt.Printf("%s%s: %v\n", prefix, k, vv)
			}
		}
	default:
		fmt.Printf("%s%v\n", prefix, v)
	}
}

func die(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// --- get-tweet ---

func runGetTweet(args []string) {
	fs := flag.NewFlagSet("get-tweet", flag.ExitOnError)
	tweetID := fs.String("tweet-id", "", "Tweet ID to look up (required)")
	envFile := fs.String("env", "", "Path to .env file")
	outputFmt := fs.String("output", "text", "Output format: text or json")
	fs.Parse(args)

	loadEnv(*envFile)
	if *tweetID == "" {
		die("--tweet-id is required")
	}

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

	resp, err := client.TweetLookup(context.Background(), []string{*tweetID}, opts)
	if err != nil {
		die("tweet lookup: %v", err)
	}

	dictionaries := resp.Raw.TweetDictionaries()
	if len(dictionaries) == 0 {
		die("tweet %s not found", *tweetID)
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
			tweet["metrics"] = map[string]interface{}{
				"likes":    dict.Tweet.PublicMetrics.Likes,
				"retweets": dict.Tweet.PublicMetrics.Retweets,
				"replies":  dict.Tweet.PublicMetrics.Replies,
				"quotes":   dict.Tweet.PublicMetrics.Quotes,
			}
		}
		if dict.Author != nil {
			tweet["author"] = map[string]interface{}{
				"id":       dict.Author.ID,
				"username": dict.Author.UserName,
				"name":     dict.Author.Name,
			}
		}
		outputResult(tweet, *outputFmt)
		return
	}
}

// --- post-tweet ---

func runPostTweet(args []string) {
	fs := flag.NewFlagSet("post-tweet", flag.ExitOnError)
	text := fs.String("text", "", "Tweet text (required)")
	replyToID := fs.String("reply-to-id", "", "Tweet ID to reply to")
	envFile := fs.String("env", "", "Path to .env file")
	outputFmt := fs.String("output", "text", "Output format: text or json")
	fs.Parse(args)

	loadEnv(*envFile)
	if *text == "" {
		die("--text is required")
	}

	client := services.TwitterClient()
	req := twitter.CreateTweetRequest{Text: *text}
	if *replyToID != "" {
		req.Reply = &twitter.CreateTweetReply{InReplyToTweetID: *replyToID}
	}

	resp, err := client.CreateTweet(context.Background(), req)
	if err != nil {
		die("post tweet: %v", err)
	}

	outputResult(map[string]interface{}{
		"id":   resp.Tweet.ID,
		"text": *text,
	}, *outputFmt)
}

// --- delete-tweet ---

func runDeleteTweet(args []string) {
	fs := flag.NewFlagSet("delete-tweet", flag.ExitOnError)
	tweetID := fs.String("tweet-id", "", "Tweet ID to delete (required)")
	envFile := fs.String("env", "", "Path to .env file")
	outputFmt := fs.String("output", "text", "Output format: text or json")
	fs.Parse(args)

	loadEnv(*envFile)
	if *tweetID == "" {
		die("--tweet-id is required")
	}

	client := services.TwitterClient()
	_, err := client.DeleteTweet(context.Background(), *tweetID)
	if err != nil {
		die("delete tweet: %v", err)
	}

	outputResult(map[string]interface{}{
		"status":   "deleted",
		"tweet_id": *tweetID,
	}, *outputFmt)
}

// --- search ---

func runSearch(args []string) {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	query := fs.String("query", "", "Search query (required)")
	maxResults := fs.Int("max-results", 10, "Maximum results (10-100)")
	envFile := fs.String("env", "", "Path to .env file")
	outputFmt := fs.String("output", "text", "Output format: text or json")
	fs.Parse(args)

	loadEnv(*envFile)
	if *query == "" {
		die("--query is required")
	}

	if *maxResults < 10 {
		*maxResults = 10
	}
	if *maxResults > 100 {
		*maxResults = 100
	}

	client := services.TwitterClient()
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
		MaxResults: *maxResults,
	}

	resp, err := client.TweetRecentSearch(context.Background(), *query, opts)
	if err != nil {
		die("search tweets: %v", err)
	}

	dictionaries := resp.Raw.TweetDictionaries()
	results := make([]interface{}, 0, len(dictionaries))
	for _, dict := range dictionaries {
		tweet := map[string]interface{}{
			"id":   dict.Tweet.ID,
			"text": dict.Tweet.Text,
		}
		if dict.Tweet.CreatedAt != "" {
			tweet["created_at"] = dict.Tweet.CreatedAt
		}
		if dict.Tweet.PublicMetrics != nil {
			tweet["metrics"] = map[string]interface{}{
				"likes":    dict.Tweet.PublicMetrics.Likes,
				"retweets": dict.Tweet.PublicMetrics.Retweets,
				"replies":  dict.Tweet.PublicMetrics.Replies,
				"quotes":   dict.Tweet.PublicMetrics.Quotes,
			}
		}
		if dict.Author != nil {
			tweet["author"] = map[string]interface{}{
				"id":       dict.Author.ID,
				"username": dict.Author.UserName,
				"name":     dict.Author.Name,
			}
		}
		results = append(results, tweet)
	}

	outputResult(map[string]interface{}{
		"query":   *query,
		"count":   len(results),
		"results": results,
	}, *outputFmt)
}

// --- get-user ---

func runGetUser(args []string) {
	fs := flag.NewFlagSet("get-user", flag.ExitOnError)
	username := fs.String("username", "", "Username to look up (required, without @)")
	envFile := fs.String("env", "", "Path to .env file")
	outputFmt := fs.String("output", "text", "Output format: text or json")
	fs.Parse(args)

	loadEnv(*envFile)
	if *username == "" {
		die("--username is required")
	}

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

	resp, err := client.UserNameLookup(context.Background(), []string{*username}, opts)
	if err != nil {
		die("user lookup: %v", err)
	}

	dictionaries := resp.Raw.UserDictionaries()
	if len(dictionaries) == 0 {
		die("user @%s not found", *username)
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
			user["metrics"] = map[string]interface{}{
				"followers": dict.User.PublicMetrics.Followers,
				"following": dict.User.PublicMetrics.Following,
				"tweets":    dict.User.PublicMetrics.Tweets,
				"listed":    dict.User.PublicMetrics.Listed,
			}
		}
		if dict.PinnedTweet != nil {
			user["pinned_tweet"] = map[string]interface{}{
				"id":   dict.PinnedTweet.ID,
				"text": dict.PinnedTweet.Text,
			}
		}
		outputResult(user, *outputFmt)
		return
	}
}

// --- get-user-timeline ---

func runGetUserTimeline(args []string) {
	fs := flag.NewFlagSet("get-user-timeline", flag.ExitOnError)
	userID := fs.String("user-id", "", "User ID (required)")
	maxResults := fs.Int("max-results", 10, "Maximum results (5-100)")
	envFile := fs.String("env", "", "Path to .env file")
	outputFmt := fs.String("output", "text", "Output format: text or json")
	fs.Parse(args)

	loadEnv(*envFile)
	if *userID == "" {
		die("--user-id is required")
	}

	if *maxResults < 5 {
		*maxResults = 5
	}
	if *maxResults > 100 {
		*maxResults = 100
	}

	client := services.TwitterClient()
	opts := twitter.UserTweetTimelineOpts{
		Expansions: []twitter.Expansion{
			twitter.ExpansionAuthorID,
		},
		TweetFields: []twitter.TweetField{
			twitter.TweetFieldCreatedAt,
			twitter.TweetFieldPublicMetrics,
			twitter.TweetFieldConversationID,
			twitter.TweetFieldEntities,
		},
		UserFields: []twitter.UserField{
			twitter.UserFieldUserName,
			twitter.UserFieldName,
		},
		MaxResults: *maxResults,
	}

	resp, err := client.UserTweetTimeline(context.Background(), *userID, opts)
	if err != nil {
		die("user timeline: %v", err)
	}

	dictionaries := resp.Raw.TweetDictionaries()
	results := make([]interface{}, 0, len(dictionaries))
	for _, dict := range dictionaries {
		tweet := map[string]interface{}{
			"id":   dict.Tweet.ID,
			"text": dict.Tweet.Text,
		}
		if dict.Tweet.CreatedAt != "" {
			tweet["created_at"] = dict.Tweet.CreatedAt
		}
		if dict.Tweet.PublicMetrics != nil {
			tweet["metrics"] = map[string]interface{}{
				"likes":    dict.Tweet.PublicMetrics.Likes,
				"retweets": dict.Tweet.PublicMetrics.Retweets,
				"replies":  dict.Tweet.PublicMetrics.Replies,
				"quotes":   dict.Tweet.PublicMetrics.Quotes,
			}
		}
		results = append(results, tweet)
	}

	outputResult(map[string]interface{}{
		"user_id": *userID,
		"count":   len(results),
		"tweets":  results,
	}, *outputFmt)
}

// --- post-thread ---

func runPostThread(args []string) {
	fs := flag.NewFlagSet("post-thread", flag.ExitOnError)
	texts := fs.String("texts", "", "Thread tweets separated by ||| (required)")
	envFile := fs.String("env", "", "Path to .env file")
	outputFmt := fs.String("output", "text", "Output format: text or json")
	fs.Parse(args)

	loadEnv(*envFile)
	if *texts == "" {
		die("--texts is required")
	}

	parts := strings.Split(*texts, "|||")
	if len(parts) < 2 {
		die("thread must contain at least 2 tweets separated by '|||'")
	}
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	client := services.TwitterClient()
	type postedTweet struct {
		Index int    `json:"index"`
		ID    string `json:"id"`
		Text  string `json:"text"`
	}

	posted := make([]postedTweet, 0, len(parts))
	var lastTweetID string

	for i, text := range parts {
		if text == "" {
			continue
		}
		req := twitter.CreateTweetRequest{Text: text}
		if lastTweetID != "" {
			req.Reply = &twitter.CreateTweetReply{InReplyToTweetID: lastTweetID}
		}
		resp, err := client.CreateTweet(context.Background(), req)
		if err != nil {
			die("failed to post tweet %d: %v", i+1, err)
		}
		lastTweetID = resp.Tweet.ID
		posted = append(posted, postedTweet{Index: i + 1, ID: resp.Tweet.ID, Text: text})
	}

	// convert to []interface{} for outputResult
	tweetsOut := make([]interface{}, len(posted))
	for i, p := range posted {
		tweetsOut[i] = map[string]interface{}{
			"index": p.Index,
			"id":    p.ID,
			"text":  p.Text,
		}
	}

	outputResult(map[string]interface{}{
		"thread_length": len(posted),
		"tweets":        tweetsOut,
	}, *outputFmt)
}

// --- like-tweet ---

func runLikeTweet(args []string) {
	fs := flag.NewFlagSet("like-tweet", flag.ExitOnError)
	tweetID := fs.String("tweet-id", "", "Tweet ID to like (required)")
	envFile := fs.String("env", "", "Path to .env file")
	outputFmt := fs.String("output", "text", "Output format: text or json")
	fs.Parse(args)

	loadEnv(*envFile)
	if *tweetID == "" {
		die("--tweet-id is required")
	}

	client := services.TwitterClient()
	userID := services.AuthUserID()

	_, err := client.UserLikes(context.Background(), userID, *tweetID)
	if err != nil {
		die("like tweet: %v", err)
	}

	outputResult(map[string]interface{}{
		"status":   "liked",
		"tweet_id": *tweetID,
	}, *outputFmt)
}

// --- unlike-tweet ---

func runUnlikeTweet(args []string) {
	fs := flag.NewFlagSet("unlike-tweet", flag.ExitOnError)
	tweetID := fs.String("tweet-id", "", "Tweet ID to unlike (required)")
	envFile := fs.String("env", "", "Path to .env file")
	outputFmt := fs.String("output", "text", "Output format: text or json")
	fs.Parse(args)

	loadEnv(*envFile)
	if *tweetID == "" {
		die("--tweet-id is required")
	}

	client := services.TwitterClient()
	userID := services.AuthUserID()

	_, err := client.DeleteUserLikes(context.Background(), userID, *tweetID)
	if err != nil {
		die("unlike tweet: %v", err)
	}

	outputResult(map[string]interface{}{
		"status":   "unliked",
		"tweet_id": *tweetID,
	}, *outputFmt)
}

// --- retweet ---

func runRetweet(args []string) {
	fs := flag.NewFlagSet("retweet", flag.ExitOnError)
	tweetID := fs.String("tweet-id", "", "Tweet ID to retweet (required)")
	envFile := fs.String("env", "", "Path to .env file")
	outputFmt := fs.String("output", "text", "Output format: text or json")
	fs.Parse(args)

	loadEnv(*envFile)
	if *tweetID == "" {
		die("--tweet-id is required")
	}

	client := services.TwitterClient()
	userID := services.AuthUserID()

	_, err := client.UserRetweet(context.Background(), userID, *tweetID)
	if err != nil {
		die("retweet: %v", err)
	}

	outputResult(map[string]interface{}{
		"status":   "retweeted",
		"tweet_id": *tweetID,
	}, *outputFmt)
}

// --- unretweet ---

func runUnretweet(args []string) {
	fs := flag.NewFlagSet("unretweet", flag.ExitOnError)
	tweetID := fs.String("tweet-id", "", "Tweet ID to unretweet (required)")
	envFile := fs.String("env", "", "Path to .env file")
	outputFmt := fs.String("output", "text", "Output format: text or json")
	fs.Parse(args)

	loadEnv(*envFile)
	if *tweetID == "" {
		die("--tweet-id is required")
	}

	client := services.TwitterClient()
	userID := services.AuthUserID()

	_, err := client.DeleteUserRetweet(context.Background(), userID, *tweetID)
	if err != nil {
		die("unretweet: %v", err)
	}

	outputResult(map[string]interface{}{
		"status":   "unretweeted",
		"tweet_id": *tweetID,
	}, *outputFmt)
}

// --- get-mentions ---

func runGetMentions(args []string) {
	fs := flag.NewFlagSet("get-mentions", flag.ExitOnError)
	userID := fs.String("user-id", "", "User ID (required)")
	maxResults := fs.Int("max-results", 10, "Maximum results (5-100)")
	envFile := fs.String("env", "", "Path to .env file")
	outputFmt := fs.String("output", "text", "Output format: text or json")
	fs.Parse(args)

	loadEnv(*envFile)
	if *userID == "" {
		die("--user-id is required")
	}

	if *maxResults < 5 {
		*maxResults = 5
	}
	if *maxResults > 100 {
		*maxResults = 100
	}

	client := services.TwitterClient()
	opts := twitter.UserMentionTimelineOpts{
		Expansions: []twitter.Expansion{
			twitter.ExpansionAuthorID,
			twitter.ExpansionEntitiesMentionsUserName,
		},
		TweetFields: []twitter.TweetField{
			twitter.TweetFieldCreatedAt,
			twitter.TweetFieldPublicMetrics,
			twitter.TweetFieldConversationID,
		},
		UserFields: []twitter.UserField{
			twitter.UserFieldUserName,
			twitter.UserFieldName,
		},
		MaxResults: *maxResults,
	}

	resp, err := client.UserMentionTimeline(context.Background(), *userID, opts)
	if err != nil {
		die("get mentions: %v", err)
	}

	dictionaries := resp.Raw.TweetDictionaries()
	results := make([]interface{}, 0, len(dictionaries))
	for _, dict := range dictionaries {
		tweet := map[string]interface{}{
			"id":   dict.Tweet.ID,
			"text": dict.Tweet.Text,
		}
		if dict.Tweet.CreatedAt != "" {
			tweet["created_at"] = dict.Tweet.CreatedAt
		}
		if dict.Tweet.PublicMetrics != nil {
			tweet["metrics"] = map[string]interface{}{
				"likes":    dict.Tweet.PublicMetrics.Likes,
				"retweets": dict.Tweet.PublicMetrics.Retweets,
				"replies":  dict.Tweet.PublicMetrics.Replies,
				"quotes":   dict.Tweet.PublicMetrics.Quotes,
			}
		}
		if dict.Author != nil {
			tweet["author"] = map[string]interface{}{
				"id":       dict.Author.ID,
				"username": dict.Author.UserName,
				"name":     dict.Author.Name,
			}
		}
		results = append(results, tweet)
	}

	outputResult(map[string]interface{}{
		"user_id":  *userID,
		"count":    len(results),
		"mentions": results,
	}, *outputFmt)
}

// --- get-quote-tweets ---

func runGetQuoteTweets(args []string) {
	fs := flag.NewFlagSet("get-quote-tweets", flag.ExitOnError)
	tweetID := fs.String("tweet-id", "", "Tweet ID to find quotes for (required)")
	maxResults := fs.Int("max-results", 10, "Maximum results (10-100)")
	envFile := fs.String("env", "", "Path to .env file")
	outputFmt := fs.String("output", "text", "Output format: text or json")
	fs.Parse(args)

	loadEnv(*envFile)
	if *tweetID == "" {
		die("--tweet-id is required")
	}

	if *maxResults < 10 {
		*maxResults = 10
	}
	if *maxResults > 100 {
		*maxResults = 100
	}

	client := services.TwitterClient()
	opts := twitter.QuoteTweetsLookupOpts{
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
		MaxResults: *maxResults,
	}

	resp, err := client.QuoteTweetsLookup(context.Background(), *tweetID, opts)
	if err != nil {
		die("quote tweets lookup: %v", err)
	}

	dictionaries := resp.Raw.TweetDictionaries()
	results := make([]interface{}, 0, len(dictionaries))
	for _, dict := range dictionaries {
		tweet := map[string]interface{}{
			"id":   dict.Tweet.ID,
			"text": dict.Tweet.Text,
		}
		if dict.Tweet.CreatedAt != "" {
			tweet["created_at"] = dict.Tweet.CreatedAt
		}
		if dict.Tweet.PublicMetrics != nil {
			tweet["metrics"] = map[string]interface{}{
				"likes":    dict.Tweet.PublicMetrics.Likes,
				"retweets": dict.Tweet.PublicMetrics.Retweets,
				"replies":  dict.Tweet.PublicMetrics.Replies,
				"quotes":   dict.Tweet.PublicMetrics.Quotes,
			}
		}
		if dict.Author != nil {
			tweet["author"] = map[string]interface{}{
				"id":       dict.Author.ID,
				"username": dict.Author.UserName,
				"name":     dict.Author.Name,
			}
		}
		results = append(results, tweet)
	}

	outputResult(map[string]interface{}{
		"tweet_id": *tweetID,
		"count":    len(results),
		"quotes":   results,
	}, *outputFmt)
}

// --- mute-user ---

func runMuteUser(args []string) {
	fs := flag.NewFlagSet("mute-user", flag.ExitOnError)
	targetUserID := fs.String("target-user-id", "", "User ID to mute (required)")
	envFile := fs.String("env", "", "Path to .env file")
	outputFmt := fs.String("output", "text", "Output format: text or json")
	fs.Parse(args)

	loadEnv(*envFile)
	if *targetUserID == "" {
		die("--target-user-id is required")
	}

	client := services.TwitterClient()
	userID := services.AuthUserID()

	_, err := client.UserMutes(context.Background(), userID, *targetUserID)
	if err != nil {
		die("mute user: %v", err)
	}

	outputResult(map[string]interface{}{
		"status":         "muted",
		"target_user_id": *targetUserID,
	}, *outputFmt)
}

// --- unmute-user ---

func runUnmuteUser(args []string) {
	fs := flag.NewFlagSet("unmute-user", flag.ExitOnError)
	targetUserID := fs.String("target-user-id", "", "User ID to unmute (required)")
	envFile := fs.String("env", "", "Path to .env file")
	outputFmt := fs.String("output", "text", "Output format: text or json")
	fs.Parse(args)

	loadEnv(*envFile)
	if *targetUserID == "" {
		die("--target-user-id is required")
	}

	client := services.TwitterClient()
	userID := services.AuthUserID()

	_, err := client.DeleteUserMutes(context.Background(), userID, *targetUserID)
	if err != nil {
		die("unmute user: %v", err)
	}

	outputResult(map[string]interface{}{
		"status":         "unmuted",
		"target_user_id": *targetUserID,
	}, *outputFmt)
}

// --- block-user ---

func runBlockUser(args []string) {
	fs := flag.NewFlagSet("block-user", flag.ExitOnError)
	targetUserID := fs.String("target-user-id", "", "User ID to block (required)")
	envFile := fs.String("env", "", "Path to .env file")
	outputFmt := fs.String("output", "text", "Output format: text or json")
	fs.Parse(args)

	loadEnv(*envFile)
	if *targetUserID == "" {
		die("--target-user-id is required")
	}

	client := services.TwitterClient()
	userID := services.AuthUserID()

	_, err := client.UserBlocks(context.Background(), userID, *targetUserID)
	if err != nil {
		die("block user: %v", err)
	}

	outputResult(map[string]interface{}{
		"status":         "blocked",
		"target_user_id": *targetUserID,
	}, *outputFmt)
}

// --- unblock-user ---

func runUnblockUser(args []string) {
	fs := flag.NewFlagSet("unblock-user", flag.ExitOnError)
	targetUserID := fs.String("target-user-id", "", "User ID to unblock (required)")
	envFile := fs.String("env", "", "Path to .env file")
	outputFmt := fs.String("output", "text", "Output format: text or json")
	fs.Parse(args)

	loadEnv(*envFile)
	if *targetUserID == "" {
		die("--target-user-id is required")
	}

	client := services.TwitterClient()
	userID := services.AuthUserID()

	_, err := client.DeleteUserBlocks(context.Background(), userID, *targetUserID)
	if err != nil {
		die("unblock user: %v", err)
	}

	outputResult(map[string]interface{}{
		"status":         "unblocked",
		"target_user_id": *targetUserID,
	}, *outputFmt)
}

// --- get-user-lists ---

func runGetUserLists(args []string) {
	fs := flag.NewFlagSet("get-user-lists", flag.ExitOnError)
	userID := fs.String("user-id", "", "User ID (required)")
	envFile := fs.String("env", "", "Path to .env file")
	outputFmt := fs.String("output", "text", "Output format: text or json")
	fs.Parse(args)

	loadEnv(*envFile)
	if *userID == "" {
		die("--user-id is required")
	}

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

	resp, err := client.UserListLookup(context.Background(), *userID, opts)
	if err != nil {
		die("user lists lookup: %v", err)
	}

	results := make([]interface{}, 0)
	if resp.Raw != nil {
		for _, list := range resp.Raw.Lists {
			l := map[string]interface{}{
				"id":      list.ID,
				"name":    list.Name,
				"private": list.Private,
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
			results = append(results, l)
		}
	}

	outputResult(map[string]interface{}{
		"user_id": *userID,
		"count":   len(results),
		"lists":   results,
	}, *outputFmt)
}

// --- get-list-tweets ---

func runGetListTweets(args []string) {
	fs := flag.NewFlagSet("get-list-tweets", flag.ExitOnError)
	listID := fs.String("list-id", "", "List ID (required)")
	maxResults := fs.Int("max-results", 10, "Maximum results (1-100)")
	envFile := fs.String("env", "", "Path to .env file")
	outputFmt := fs.String("output", "text", "Output format: text or json")
	fs.Parse(args)

	loadEnv(*envFile)
	if *listID == "" {
		die("--list-id is required")
	}

	if *maxResults < 1 {
		*maxResults = 1
	}
	if *maxResults > 100 {
		*maxResults = 100
	}

	client := services.TwitterClient()
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
		MaxResults: *maxResults,
	}

	resp, err := client.ListTweetLookup(context.Background(), *listID, opts)
	if err != nil {
		die("list tweets lookup: %v", err)
	}

	dictionaries := resp.Raw.TweetDictionaries()
	results := make([]interface{}, 0, len(dictionaries))
	for _, dict := range dictionaries {
		tweet := map[string]interface{}{
			"id":   dict.Tweet.ID,
			"text": dict.Tweet.Text,
		}
		if dict.Tweet.CreatedAt != "" {
			tweet["created_at"] = dict.Tweet.CreatedAt
		}
		if dict.Tweet.PublicMetrics != nil {
			tweet["metrics"] = map[string]interface{}{
				"likes":    dict.Tweet.PublicMetrics.Likes,
				"retweets": dict.Tweet.PublicMetrics.Retweets,
				"replies":  dict.Tweet.PublicMetrics.Replies,
				"quotes":   dict.Tweet.PublicMetrics.Quotes,
			}
		}
		if dict.Author != nil {
			tweet["author"] = map[string]interface{}{
				"id":       dict.Author.ID,
				"username": dict.Author.UserName,
				"name":     dict.Author.Name,
			}
		}
		results = append(results, tweet)
	}

	outputResult(map[string]interface{}{
		"list_id": *listID,
		"count":   len(results),
		"tweets":  results,
	}, *outputFmt)
}
