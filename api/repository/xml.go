package repository

import (
	"encoding/xml"
	"time"
)

type CDATAString struct {
	Val string `xml:",cdata"`
}

type Uid struct {
	Val string `xml:"dsq:id,attr,omitempty"`
}

type Post struct {
	UID            string      `xml:"dsq:id,attr"`
	Message        CDATAString `xml:"message"`
	CreatedAt      time.Time   `xml:"createdAt"`
	AuthorName     string      `xml:"author>name"`
	AuthorUserName string      `xml:"author>username"`
	IP             string      `xml:"ipAddress"`
	Tid            Uid         `xml:"thread,omitempty"`
	Pid            Uid         `xml:"parent,omitempty"`
	IsSpam         bool        `xml:"isSpam"`
	Deleted        bool        `xml:"isDeleted"`
}

type Category struct {
	UID       string `xml:"dsq:id,attr"`
	Forum     string `xml:"forum"`
	Title     string `xml:"title"`
	IsDefault bool   `xml:"isDefault"`
}

type Thread struct {
	UID            string    `xml:"dsq:id,attr"`
	Forum          string    `xml:"forum"`
	Category       Uid       `xml:"category"`
	Link           string    `xml:"link"`
	Title          string    `xml:"title"`
	Message        string    `xml:"message"`
	CreatedAt      time.Time `xml:"createdAt"`
	AuthorName     string    `xml:"author>name"`
	AuthorUsername string    `xml:"author>username"`
	AuthorAnon     bool      `xml:"author>isAnonymous"`
	IsClosed       bool      `xml:"isClosed"`
	IsDeleted      bool      `xml:"isDeleted"`
}

type DisqusExport struct {
	XMLName        xml.Name   `xml:"disqus"`
	XMLNS          string     `xml:"xmlns,attr"`
	Dsq            string     `xml:"xmlns:dsq,attr"`
	Xsi            string     `xml:"xmlns:xsi,attr"`
	SchemaLocation string     `xml:"xmlns:schemaLocation,attr"`
	Categories     []Category `xml:"category"`
	Threads        []Thread   `xml:"thread"`
	Posts          []Post     `xml:"post"`
}
