package infrastructure

import (
	"fmt"
	"github.com/ghosts-network/news-feed/utils/logger"
	"net/http"
	"time"
)

func NewScopedClient() *http.Client {
	return &http.Client{
		Transport: &LogRoundTripper{},
	}
}

type LogRoundTripper struct{}

func (t LogRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	st := time.Now()
	logger.Info(fmt.Sprintf("Outgoing http request started %s %s", req.Method, req.URL.String()), &map[string]any{
		"correlationId": req.Context().Value("correlationId"),
		"type":          "outgoing:http",
	})

	resp, err := http.DefaultTransport.RoundTrip(req)
	logger.Info(fmt.Sprintf("Outgoing http request finished %s %s", req.Method, req.URL.String()), &map[string]any{
		"correlationId":       req.Context().Value("correlationId"),
		"type":                "outgoing:http",
		"elapsedMilliseconds": time.Now().Sub(st).Milliseconds(),
		"statusCode":          resp.StatusCode,
	})

	return resp, err
}
