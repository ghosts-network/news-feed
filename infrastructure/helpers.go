package infrastructure

import (
	"fmt"
	"github.com/ghosts-network/news-feed/utils/logger"
	"net/http"
	"time"
)

func NewScopedClient() *http.Client {
	return &http.Client{
		Transport: &LogRoundTripper{
			SetRequestIdRoundTripper{
				http.DefaultTransport,
			},
		},
	}
}

type LogRoundTripper struct {
	Proxied http.RoundTripper
}

func (t LogRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	st := time.Now()
	logger.Info(fmt.Sprintf("Outgoing http request started %s %s", req.Method, req.URL.String()), &map[string]any{
		"correlationId": req.Context().Value("correlationId"),
		"type":          "outgoing:http",
	})

	resp, err := t.Proxied.RoundTrip(req)
	logger.Info(fmt.Sprintf("Outgoing http request finished %s %s", req.Method, req.URL.String()), &map[string]any{
		"correlationId":       req.Context().Value("correlationId"),
		"type":                "outgoing:http",
		"elapsedMilliseconds": time.Now().Sub(st).Milliseconds(),
		"statusCode":          resp.StatusCode,
	})

	return resp, err
}

type SetRequestIdRoundTripper struct {
	Proxied http.RoundTripper
}

func (t SetRequestIdRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-Request-ID", req.Context().Value("correlationId").(string))
	resp, err := t.Proxied.RoundTrip(req)

	return resp, err
}
