package api

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/mgkbadola/disqus-comment-exporter/api/repository"
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
	comments          []repository.Comment
	threads           []repository.Thread
	baseUrl           string
}

func NewDisqusAPIWrapper(config Config) *DisqusAPIWrapper {
	return &DisqusAPIWrapper{
		config:            config,
		comments:          make([]repository.Comment, 0),
		threads:           make([]repository.Thread, 0),
		requestsAvailable: 1000,
		resetUnix:         time.Now().Unix(),
		baseUrl:           "https://disqus.com/api/3.0",
	}
}

func (w *DisqusAPIWrapper) BeginCommentExport() {
	exportedParent, err := w.FetchComments("", true)
	if err != nil {
		log.Printf("%+v", err)
	}
	nextCursor := exportedParent.Cursor.Next
	w.comments = append(w.comments, exportedParent.Response...)
	for more := exportedParent.Cursor.More; more; more = exportedParent.Cursor.More {
		exportedParent, err = w.FetchComments(nextCursor, true)
		if err != nil {
			log.Printf("%+v", err)
		}
		w.comments = append(w.comments, exportedParent.Response...)
		nextCursor = exportedParent.Cursor.Next
	}

	visited := make(map[string]bool)
	var posts []repository.Post

	for _, comment := range w.comments {
		if !visited[comment.Thread] {
			visited[comment.Thread] = true
		}
		var pid string
		if comment.Parent != nil {
			pid = strconv.Itoa(*comment.Parent)
		} else {
			pid = ""
		}
		post := repository.Post{
			UID:            comment.ID,
			Message:        repository.CDATAString{Val: comment.Message},
			CreatedAt:      comment.CreatedAt.Time,
			AuthorName:     comment.Author.Name,
			AuthorUserName: comment.Author.Username,
			Tid:            repository.Uid{Val: comment.Thread},
			Pid:            repository.Uid{Val: pid},
			IsSpam:         comment.IsSpam,
			Deleted:        comment.IsDeleted,
		}
		posts = append(posts, post)
	}

	d := repository.DisqusExport{
		XMLNS:          "http://disqus.com",
		Dsq:            "http://disqus.com/disqus-internals",
		Xsi:            "http://www.w3.org/2001/XMLSchema-instance",
		SchemaLocation: "http://disqus.com/api/schemas/1.0/disqus.xsd http://disqus.com/api/schemas/1.0/disqus-internals.xsd",
		Posts:          posts,
	}
	out, err := xml.MarshalIndent(d, "", "	")
	if err != nil {
		log.Println("%+v", err)
	}
	fmt.Println(xml.Header + string(out))
}

func (w *DisqusAPIWrapper) FetchComments(cursor string, isDemo bool) (repository.CommentApiResponse, error) {
	//convert from JSON to XML
	if w.requestsAvailable == 0 && time.Now().Unix() < w.resetUnix {
		//TODO
		//we have to wait for cooldown
		//add condition to pause based on time left
	}
	var r io.Reader
	if !isDemo {
		client := &http.Client{}
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/posts/list.json", w.baseUrl), nil)
		if err != nil {
			return repository.CommentApiResponse{}, err
		}
		q := req.URL.Query()
		q.Add("api_key", w.config.APIKey)
		q.Add("forum", w.config.Forum)
		q.Add("limit", strconv.Itoa(int(w.config.FetchLimit)))
		//TODO remove when final
		q.Add("start", "1729056409")
		if cursor != "" {
			q.Add("cursor", cursor)
		}
		req.URL.RawQuery = q.Encode()
		req.Header.Set("Accept", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			return repository.CommentApiResponse{}, err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return repository.CommentApiResponse{}, err
		}
		//TODO ratelimiting
		//requestsAvailable, err := strconv.ParseInt(resp.Header.Get("X-Ratelimit-Remaining"), 10, 16)
		//if err != nil {
		//	fmt.Printf("REQ")
		//	return repository.CommentApiResponse{}, err
		//}
		//resetUnix, err := strconv.ParseInt(resp.Header.Get("X-Ratelimit-Reset"), 10, 64)
		//if err != nil {
		//	fmt.Printf("LIM")
		//	return repository.CommentApiResponse{}, err
		//}
		//w.requestsAvailable = int16(requestsAvailable)
		//w.resetUnix = resetUnix
		r = strings.NewReader(string(body))
	} else {
		str := "YOUR JSON STRING HERE"
		r = strings.NewReader(str)
	}
	decoder := json.NewDecoder(r)
	var exportedParent repository.CommentApiResponse
	err := decoder.Decode(&exportedParent)
	if err != nil {
		return repository.CommentApiResponse{}, err
	}
	return exportedParent, nil
}
