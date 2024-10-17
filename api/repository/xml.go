package repository

import (
	"encoding/xml"
	"time"
)

type CDATAString struct {
	Val string `xml:",cdata"`
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

type Uid struct {
	Val string `xml:"dsq:id,attr,omitempty"`
}

type DisqusExport struct {
	XMLName        xml.Name `xml:"disqus"`
	XMLNS          string   `xml:"xmlns,attr"`
	Dsq            string   `xml:"xmlns:dsq,attr"`
	Xsi            string   `xml:"xmlns:xsi,attr"`
	SchemaLocation string   `xml:"xmlns:schemaLocation,attr"`
	Posts          []Post   `xml:"post"`
}
