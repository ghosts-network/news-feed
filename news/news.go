package news

type Publication struct {
	Id      string             `json:"id"`
	Content string             `json:"content"`
	Author  *PublicationAuthor `json:"author"`
}

type PublicationAuthor struct {
	Id        string `json:"id"`
	FullName  string `json:"fullName"`
	AvatarUrl string `json:"avatarUrl"`
}
