package main

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

type RequestSent struct {
	FromUser string
	ToUser   string
}

type RequestCancelled struct {
	FromUser string
	ToUser   string
}

type RequestApproved struct {
	User      string
	Requester string
}

type RequestDeclined struct {
	User      string
	Requester string
}

type Deleted struct {
	User   string
	Friend string
}
