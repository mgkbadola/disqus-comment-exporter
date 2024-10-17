package repository

import "time"

type Post struct {
	UID            string    `xml:"id,attr"`
	ID             string    `xml:"id"`
	Message        string    `xml:"message"`
	CreatedAt      time.Time `xml:"createdAt"`
	AuthorEmail    string    `xml:"author>email"`
	AuthorName     string    `xml:"author>name"`
	AuthorUserName string    `xml:"author>username"`
	IP             string    `xml:"ipAddress"`
	Tid            Uid       `xml:"thread"`
	Pid            Uid       `xml:"parent"`
	IsSpam         bool      `xml:"isSpam"`
	Deleted        bool      `xml:"isDeleted"`
}

type Uid struct {
	Val string `xml:"id,attr"`
}
