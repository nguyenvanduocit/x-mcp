package services

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/dghubble/oauth1"
	twitter "github.com/g8rswimmer/go-twitter/v2"
)

// oauth1Transport is a no-op Authorizer because the oauth1 http.Client
// already signs requests at the transport level.
type oauth1Transport struct{}

func (a *oauth1Transport) Add(req *http.Request) {}

var TwitterClient = sync.OnceValue(func() *twitter.Client {
	apiKey := os.Getenv("X_API_KEY")
	apiSecret := os.Getenv("X_API_SECRET")
	accessToken := os.Getenv("X_ACCESS_TOKEN")
	accessTokenSecret := os.Getenv("X_ACCESS_TOKEN_SECRET")

	config := oauth1.NewConfig(apiKey, apiSecret)
	token := oauth1.NewToken(accessToken, accessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	proxyURL := os.Getenv("PROXY_URL")
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			panic(fmt.Sprintf("failed to parse PROXY_URL: %v", err))
		}
		transport := httpClient.Transport.(*oauth1.Transport)
		transport.Base = &http.Transport{
			Proxy:           http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return &twitter.Client{
		Authorizer: &oauth1Transport{},
		Client:     httpClient,
		Host:       "https://api.twitter.com",
	}
})
