package main

type Publication struct {
	Id      string
	Content string
	Author  *PublicationAuthor
}

type PublicationAuthor struct {
	Id        string
	FullName  string
	AvatarUrl string
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
