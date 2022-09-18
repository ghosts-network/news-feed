package infrastructure

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ProfilesClient struct {
	baseUrl string
}

func NewProfilesClient(baseUrl string) *ProfilesClient {
	return &ProfilesClient{baseUrl: baseUrl}
}

func (c ProfilesClient) GetProfiles(skip int, take int) ([]Profile, error) {
	url := fmt.Sprintf("%s/profiles?skip=%d&take=%d", c.baseUrl, skip, take)
	resp, err := http.Get(url)

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
