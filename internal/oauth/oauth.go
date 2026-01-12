package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"golang.org/x/oauth2"

	"hytale-downloader/internal/buildinfo"
)

// Token extends oauth2.Token with branch information
type Token struct {
	*oauth2.Token
	Branch string `json:"branch"`
}

// Config holds the OAuth2 configuration
var cfg *oauth2.Config

func init() {
	baseURL := getBaseURL()

	cfg = &oauth2.Config{
		ClientID: "hytale-downloader",
		Endpoint: oauth2.Endpoint{
			AuthURL:       baseURL + "/oauth2/auth",
			TokenURL:      baseURL + "/oauth2/token",
			DeviceAuthURL: baseURL + "/oauth2/device/auth",
		},
		Scopes: []string{"openid", "offline_access"},
	}
}

// getBaseURL returns the appropriate base URL based on build branch
func getBaseURL() string {
	if buildinfo.Branch == "release" {
		return "https://oauth.accounts.hytale.com"
	}
	return "https://oauth.accounts-dev.hytale.com"
}

// DeviceAuth initiates the device authorization flow
func DeviceAuth(ctx context.Context) (*oauth2.DeviceAuthResponse, error) {
	return cfg.DeviceAuth(ctx)
}

// DeviceAccessToken polls for the token after user authorization
func DeviceAccessToken(ctx context.Context, da *oauth2.DeviceAuthResponse) (*Token, error) {
	tok, err := cfg.DeviceAccessToken(ctx, da)
	if err != nil {
		return nil, err
	}
	return &Token{
		Token:  tok,
		Branch: buildinfo.Branch,
	}, nil
}

// LoadCredentials loads credentials from a file
func LoadCredentials(path string) (*Token, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tok Token
	if err := json.Unmarshal(data, &tok); err != nil {
		return nil, err
	}

	if tok.Branch != buildinfo.Branch {
		return nil, fmt.Errorf("credentials were created for %q environment, but current environment is %q", tok.Branch, buildinfo.Branch)
	}

	return &tok, nil
}

// SaveCredentials saves credentials to a file
func SaveCredentials(path string, tok *Token) error {
	data, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// watchTokenSource wraps a token source and saves tokens when they refresh
type watchTokenSource struct {
	base     oauth2.TokenSource
	path     string
	mu       sync.Mutex
	lastTok  *oauth2.Token
	onRefresh func(*Token)
}

// WatchTokenSource creates a token source that saves refreshed tokens
func WatchTokenSource(path string, tok *Token, onRefresh func(*Token)) oauth2.TokenSource {
	return &watchTokenSource{
		base:      cfg.TokenSource(context.Background(), tok.Token),
		path:      path,
		lastTok:   tok.Token,
		onRefresh: onRefresh,
	}
}

func (w *watchTokenSource) Token() (*oauth2.Token, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	tok, err := w.base.Token()
	if err != nil {
		return nil, err
	}

	// Check if token was refreshed
	if tok.AccessToken != w.lastTok.AccessToken {
		w.lastTok = tok
		newTok := &Token{
			Token:  tok,
			Branch: buildinfo.Branch,
		}
		if w.onRefresh != nil {
			w.onRefresh(newTok)
		}
	}

	return tok, nil
}

// NewClient creates an HTTP client with the token source
func NewClient(ctx context.Context, ts oauth2.TokenSource) *oauth2.Transport {
	return &oauth2.Transport{
		Source: ts,
	}
}
