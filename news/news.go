package news

import "time"

type Publication struct {
	Id        string             `json:"id"`
	Content   string             `json:"content"`
	Author    *PublicationAuthor `json:"author"`
	CreatedOn time.Time          `json:"createdOn"`
	UpdatedOn time.Time          `json:"updatedOn"`
	Media     []*Media           `json:"media"`
}

type PublicationAuthor struct {
	Id        string `json:"id" bson:"_id"`
	FullName  string `json:"fullName" bson:"fullName"`
	AvatarUrl string `json:"avatarUrl" bson:"avatarUrl"`
}

type Media struct {
	Link string `json:"link" bson:"link"`
}
