package repository

import (
	"strings"
	"time"
)

type CustomTime struct {
	time.Time
}

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)
	parsedTime, err := time.Parse("2006-01-02T15:04:05", str)
	if err != nil {
		return err
	}
	ct.Time = parsedTime
	return nil
}

type Cursor struct {
	Next string `json:"next"`
	More bool   `json:"more"`
}

type Author struct {
	Username string `json:"username"`
	Name     string `json:"name"`
}

type Comment struct {
	Dislikes  int        `json:"dislikes"`
	Likes     int        `json:"likes"`
	Message   string     `json:"message"`
	IsSpam    bool       `json:"isSpam"`
	Author    Author     `json:"author"`
	ID        string     `json:"id"`
	IsDeleted bool       `json:"isDeleted"`
	Parent    *int       `json:"parent"`
	Thread    string     `json:"thread"`
	CreatedAt CustomTime `json:"createdAt"`
}

type CommentApiResponse struct {
	Cursor   Cursor    `json:"cursor"`
	Code     int       `json:"code"`
	Response []Comment `json:"response"`
}

type Thread struct {
	Link string `json:"link"`
}

type ThreadApiResponse struct {
	Code     int    `json:"code"`
	Response Thread `json:"response"`
}
