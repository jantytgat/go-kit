package application

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/jantytgat/go-kit/pkg/slogd"
)

func NewHttpClient(useragent string, timeout int, followRedirects bool) *http.Client {
	return &http.Client{
		Transport:     NewHttpTransport(useragent),
		CheckRedirect: ConfigureRedirectPolicy(followRedirects),
		Jar:           nil,
		Timeout:       time.Duration(timeout) * time.Second,
	}
}

func ConfigureRedirectPolicy(state bool) func(req *http.Request, via []*http.Request) error {
	switch state {
	case true:
		return nil
	case false:
		return DoNotFollowHttpRedirects
	default:
		return nil
	}
}

// DoNotFollowHttpRedirects information at https://go.dev/src/net/http/client.go - line 72
func DoNotFollowHttpRedirects(req *http.Request, via []*http.Request) error {
	slogd.FromContext(req.Context()).LogAttrs(req.Context(), slogd.LevelTrace, "do not follow redirects", slog.String("url", req.URL.String()))
	return http.ErrUseLastResponse
}
