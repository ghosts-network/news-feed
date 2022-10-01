package infrastructure

import (
	"fmt"
	"github.com/ghosts-network/news-feed/utils"
	"net/http"
	"time"
)

func NewScopedClient(logger *utils.Logger) *http.Client {
	return &http.Client{
		Transport: &LogRoundTripper{logger: logger},
	}
}

type LogRoundTripper struct {
	logger *utils.Logger
}

func (t LogRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	st := time.Now()
	t.logger.Info(fmt.Sprintf("Outgoing http request started %s %s", req.Method, req.URL.String()))
	resp, err := http.DefaultTransport.RoundTrip(req)
	t.logger.
		WithValue("elapsedMilliseconds", time.Now().Sub(st).Milliseconds()).
		WithValue("statusCode", resp.StatusCode).
		Info(fmt.Sprintf("Outgoing http request finished %s %s", req.Method, req.URL.String()))

	return resp, err
}
