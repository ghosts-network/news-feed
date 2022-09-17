package news

type Publication struct {
	Id        string             `json:"id" bson:"_id"`
	Content   string             `json:"content" bson:"content"`
	Author    *PublicationAuthor `json:"author" bson:"author"`
	CreatedOn int64              `json:"createOn" bson:"createOn"`
	UpdatedOn int64              `json:"updateOn" bson:"updateOn"`
	Media     []*Media           `json:"media" bson:"media"`
}

type PublicationAuthor struct {
	Id        string `json:"id" bson:"_id"`
	FullName  string `json:"fullName" bson:"fullName"`
	AvatarUrl string `json:"avatarUrl" bson:"avatarUrl"`
}

type Media struct {
	Link string `json:"link" bson:"link"`
}
