package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ghosts-network/news-feed/news"
	"io/ioutil"
	"net/http"
)

type PublicationsClient struct {
	baseUrl string
	client  *http.Client
}

func NewPublicationsClient(baseUrl string, client *http.Client) *PublicationsClient {
	return &PublicationsClient{
		baseUrl: baseUrl,
		client:  client,
	}
}

func (c PublicationsClient) GetPublications(ctx context.Context, cursor string, take int) ([]news.Publication, string, error) {
	url := fmt.Sprintf("%s/publications?cursor=%s&take=%d", c.baseUrl, cursor, take)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}

	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	rb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	ps := make([]news.Publication, take)
	err = json.Unmarshal(rb, &ps)

	nextCursor := resp.Header.Get("X-Cursor")

	return ps, nextCursor, err
}
