package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type RelationsClient struct {
	baseUrl string
	client  *http.Client
}

func NewRelationsClient(baseUrl string, client *http.Client) *RelationsClient {
	return &RelationsClient{baseUrl: baseUrl, client: client}
}

func (c RelationsClient) GetFriends(ctx context.Context, user string, skip int, take int) ([]string, error) {
	url := fmt.Sprintf("%s/relations/%s/friends?skip=%d&take=%d", c.baseUrl, user, skip, take)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ids []string
	err = json.Unmarshal(rb, &ids)

	return ids, err
}

func (c RelationsClient) GetOutgoingRequests(ctx context.Context, user string, skip int, take int) ([]string, error) {
	url := fmt.Sprintf("%s/relations/%s/friends/outgoing-requests?skip=%d&take=%d", c.baseUrl, user, skip, take)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, take)
	err = json.Unmarshal(rb, &ids)

	return ids, err
}
