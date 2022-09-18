package infrastructure

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type RelationsClient struct {
	baseUrl string
}

func NewRelationsClient(baseUrl string) *RelationsClient {
	return &RelationsClient{baseUrl: baseUrl}
}

func (c RelationsClient) GetFriends(user string, skip int, take int) ([]string, error) {
	url := fmt.Sprintf("%s/relations/%s/friends?skip=%d&take=%d", c.baseUrl, user, skip, take)
	resp, err := http.Get(url)

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

func (c RelationsClient) GetOutgoingRequests(user string, skip int, take int) ([]string, error) {
	url := fmt.Sprintf("%s/relations/%s/friends/outgoing-requests?skip=%d&take=%d", c.baseUrl, user, skip, take)
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ids := make([]string, take)
	err = json.Unmarshal(rb, &ids)

	return ids, err
}
