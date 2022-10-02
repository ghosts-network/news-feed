package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ProfilesClient struct {
	client  *http.Client
	baseUrl string
}

func NewProfilesClient(baseUrl string, client *http.Client) *ProfilesClient {
	return &ProfilesClient{
		client:  client,
		baseUrl: baseUrl,
	}
}

func (c ProfilesClient) GetProfiles(ctx context.Context, skip int, take int) ([]Profile, error) {
	url := fmt.Sprintf("%s/profiles?skip=%d&take=%d", c.baseUrl, skip, take)
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

	ps := make([]Profile, take)
	err = json.Unmarshal(rb, &ps)

	return ps, err
}

type Profile struct {
	Id        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}
