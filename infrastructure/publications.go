package infrastructure

import (
	"encoding/json"
	"fmt"
	"github.com/ghosts-network/news-feed/news"
	"io/ioutil"
	"net/http"
	"time"
)

type PublicationsClient struct {
	baseUrl string
}

func NewPublicationsClient(baseUrl string) *PublicationsClient {
	return &PublicationsClient{baseUrl: baseUrl}
}

func (c PublicationsClient) GetPublications(cursor string, take int) ([]Publication, string, error) {
	url := fmt.Sprintf("%s/publications?cursor=%s&take=%d", c.baseUrl, cursor, take)
	resp, err := http.Get(url)
	defer resp.Body.Close()

	if err != nil {
		return nil, "", err
	}

	rb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	ps := make([]Publication, take)
	err = json.Unmarshal(rb, &ps)

	nextCursor := resp.Header.Get("X-Cursor")

	return ps, nextCursor, err
}

type Publication struct {
	Id        string                  `json:"id"`
	Content   string                  `json:"content"`
	Author    *news.PublicationAuthor `json:"author"`
	CreatedOn time.Time               `json:"createdOn"`
	UpdatedOn time.Time               `json:"updatedOn"`
	Media     []*news.Media           `json:"media"`
}
