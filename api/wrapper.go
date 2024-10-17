package api

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	repo "github.com/mgkbadola/disqus-comment-exporter/api/repository"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	FetchLimit int8   `json:"fetch_limit"`
	APIKey     string `json:"api_key"`
	Forum      string `json:"forum"`
}

type DisqusAPIWrapper struct {
	config            Config
	requestsAvailable int16
	resetUnix         int64
	baseUrl           string
}

func NewDisqusAPIWrapper(config Config) *DisqusAPIWrapper {
	return &DisqusAPIWrapper{
		config:            config,
		requestsAvailable: 1000,
		resetUnix:         time.Now().Unix(),
		baseUrl:           "https://disqus.com/api/3.0",
	}
}

func (w *DisqusAPIWrapper) BeginCommentExport() {
	//XML
	posts := make([]repo.Post, 0)
	threads := make([]repo.Thread, 0)

	//reference maps
	visited := make(map[string]bool)
	authors := make(map[string]repo.Author)

	//transient
	comments := make([]repo.Comment, 0)
	page := repo.Page{}
	author := repo.Author{}
	next := ""

	commentsParent, err := w.FetchComments("", true)
	if err != nil {
		log.Printf("%+v", err)
	}
	next = commentsParent.Cursor.Next
	comments = append(comments, commentsParent.Response...)

	for more := commentsParent.Cursor.More; more; more = commentsParent.Cursor.More {
		commentsParent, err = w.FetchComments(next, true)
		if err != nil {
			log.Printf("%+v\n", err)
		}
		comments = append(comments, commentsParent.Response...)
		next = commentsParent.Cursor.Next
	}

	for _, comment := range comments {
		if !visited[comment.Thread] {
			visited[comment.Thread] = true
			pageParent, err2 := w.GetPageDetails(comment.Thread)
			if err2 != nil {
				log.Printf("%+v\n", err2)
			}
			page = pageParent.Response
			author = authors[page.AuthorId]

			if author.Username == "" && author.Name == "" {
				authorParent, err3 := w.GetAuthorDetails(pageParent.Response.AuthorId)
				if err3 != nil {
					log.Printf("%+v\n", err2)
				}
				author = authorParent.Response
				authors[page.AuthorId] = author
			}
			threads = append(threads, repo.Thread{
				UID:            comment.Thread,
				Forum:          page.Forum,
				Category:       repo.Uid{Val: page.Category},
				Link:           page.Link,
				Title:          page.Title,
				Message:        page.Message,
				CreatedAt:      page.CreatedAt.Time,
				AuthorName:     author.Name,
				AuthorUsername: author.Username,
				AuthorAnon:     false,
				IsClosed:       page.IsClosed,
				IsDeleted:      page.IsDeleted,
			})
		}
		var pid string
		if comment.Parent != nil {
			pid = strconv.Itoa(*comment.Parent)
		} else {
			pid = ""
		}
		post := repo.Post{
			UID:            comment.ID,
			Message:        repo.CDATAString{Val: comment.Message},
			CreatedAt:      comment.CreatedAt.Time,
			AuthorName:     comment.Author.Name,
			AuthorUserName: comment.Author.Username,
			Tid:            repo.Uid{Val: comment.Thread},
			Pid:            repo.Uid{Val: pid},
			IsSpam:         comment.IsSpam,
			Deleted:        comment.IsDeleted,
		}
		posts = append(posts, post)
	}
	category := repo.Category{
		UID:       "8310048",
		Forum:     w.config.Forum,
		Title:     "General",
		IsDefault: true,
	}
	d := repo.DisqusExport{
		XMLNS:          "http://disqus.com",
		Dsq:            "http://disqus.com/disqus-internals",
		Xsi:            "http://www.w3.org/2001/XMLSchema-instance",
		SchemaLocation: "http://disqus.com/api/schemas/1.0/disqus.xsd http://disqus.com/api/schemas/1.0/disqus-internals.xsd",
		Categories:     append(make([]repo.Category, 0), category),
		Threads:        threads,
		Posts:          posts,
	}
	out, err := xml.MarshalIndent(d, "", "	")
	if err != nil {
		log.Printf("%+v\n", err)
	}
	fmt.Println(xml.Header + string(out))
}

func (w *DisqusAPIWrapper) FetchComments(cursor string, isDemo bool) (repo.CommentApiResponse, error) {
	queryParams := make(map[string]string)
	queryParams["forum"] = w.config.Forum
	queryParams["limit"] = strconv.Itoa(int(w.config.FetchLimit))
	//TODO remove when final
	queryParams["start"] = "1729056409"
	if cursor != "" {
		queryParams["cursor"] = cursor
	}
	if isDemo {
		queryParams["isDemo"] = "1"
	}
	res, err := w.executeGetRequest("posts", "list", queryParams)
	if err != nil {
		return repo.CommentApiResponse{}, err
	}
	var obj repo.CommentApiResponse
	err = res.Decode(&obj)
	if err != nil {
		return repo.CommentApiResponse{}, errors.New("decoding failed")
	}
	return obj, nil
}

func (w *DisqusAPIWrapper) GetPageDetails(threadId string) (repo.DetailsResponse[repo.Page], error) {
	res, err := w.executeGetRequest("threads", "details", map[string]string{"thread": threadId})
	if err != nil {
		return repo.DetailsResponse[repo.Page]{}, err
	}
	var obj repo.DetailsResponse[repo.Page]
	err = res.Decode(&obj)
	if err != nil {
		return repo.DetailsResponse[repo.Page]{}, errors.New("decoding failed")
	}
	return obj, nil
}

func (w *DisqusAPIWrapper) GetAuthorDetails(authorId string) (repo.DetailsResponse[repo.Author], error) {
	res, err := w.executeGetRequest("users", "details", map[string]string{"user": authorId})
	if err != nil {
		return repo.DetailsResponse[repo.Author]{}, err
	}
	var obj repo.DetailsResponse[repo.Author]
	err = res.Decode(&obj)
	if err != nil {
		return repo.DetailsResponse[repo.Author]{}, errors.New("decoding failed")
	}
	return obj, nil
}

func (w *DisqusAPIWrapper) executeGetRequest(resource string, operation string, queryParams map[string]string) (*json.Decoder, error) {
	var r *strings.Reader
	if resource == "posts" && queryParams["isDemo"] == "1" {
		str := "YOUR JSON STRING HERE"
		r = strings.NewReader(str)
	} else {
		if w.requestsAvailable == 0 && time.Now().Unix() < w.resetUnix {
			//TODO
			//we have to wait for cooldown
			//add condition to pause based on time left
		}
		client := &http.Client{}
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s/%s.json", w.baseUrl, resource, operation), nil)
		if err != nil {
			return nil, err
		}
		q := req.URL.Query()
		if queryParams != nil {
			for key, value := range queryParams {
				q.Add(key, value)
			}
		}
		q.Add("api_key", w.config.APIKey)
		req.URL.RawQuery = q.Encode()
		req.Header.Set("Accept", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		requestsAvailable, err := strconv.Atoi(resp.Header.Get("X-Ratelimit-Remaining"))
		if err != nil {
			return nil, err
		}
		resetUnix, err := strconv.Atoi(resp.Header.Get("X-Ratelimit-Reset"))
		if err != nil {
			return nil, err
		}
		w.requestsAvailable = int16(requestsAvailable)
		w.resetUnix = int64(resetUnix)
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		r = strings.NewReader(string(body))
	}
	decoder := json.NewDecoder(r)
	return decoder, nil
}
