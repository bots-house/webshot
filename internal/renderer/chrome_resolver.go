package renderer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/xerrors"
)

// ChromeResolver define interface to retrieve browser websocket url.
type ChromeResolver interface {
	BrowserWebSocketURL(ctx context.Context) (string, error)
}

func NewChromeResolver(addr string) (ChromeResolver, error) {
	u, err := url.ParseRequestURI(addr)
	if err != nil {
		return nil, xerrors.Errorf("invalid addr")
	}

	if u.Scheme == "ws" || u.Scheme == "wss" {
		return &ChromeResolverStatic{WebSocketURL: u.String()}, nil
	} else if u.Scheme == "http" || u.Scheme == "https" {
		u.Path = strings.TrimSuffix(u.Path, "/")
		return &ChromeResolverAPI{Addr: u.String()}, nil
	}

	return nil, xerrors.Errorf("unsupported scheme '%s'", u.Scheme)
}

// ChromeResolverStatic accepts WS connection string and resolve to it everytime.
type ChromeResolverStatic struct {
	WebSocketURL string
}

func (r *ChromeResolverStatic) BrowserWebSocketURL(ctx context.Context) (string, error) {
	return r.WebSocketURL, nil
}

// ChromeResolverAPI uses Chrome debug API to retrieve WebSocket URL.
type ChromeResolverAPI struct {
	Addr   string
	Client *http.Client
}

func (r *ChromeResolverAPI) getClient() *http.Client {
	if r.Client == nil {
		return http.DefaultClient
	}

	return r.Client
}

func (r *ChromeResolverAPI) BrowserWebSocketURL(ctx context.Context) (string, error) {
	// otherwise, resolve websocket url via API
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/json/version", r.Addr),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}

	res, err := r.getClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("do devtools url lookup request: %w", err)
	}
	defer res.Body.Close()

	var chromeInfo struct {
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}

	if err := json.NewDecoder(res.Body).Decode(&chromeInfo); err != nil {
		return "", fmt.Errorf("decode devtools url lookup response: %w", err)
	}

	return chromeInfo.WebSocketDebuggerURL, nil
}
