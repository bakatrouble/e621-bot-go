package e621

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-errors/errors"
)

type postResponse struct {
	Post *Post `json:"post"`
}

func (e *E621) GetPost(ctx context.Context, id int) (*Post, error) {
	var resp postResponse
	rq := e.httpClient.Get("/posts/{post_id}.json").
		SetPathParam("post_id", strconv.Itoa(id)).
		Do(ctx)
	err := rq.Into(&resp)
	if resp.Post.File.Url == nil {
		fmt.Println("fixing url")
		fixedUrl := fmt.Sprintf("https://static1.e621.net/data/%s/%s/%s.%s", resp.Post.File.Md5[0:2], resp.Post.File.Md5[2:4], resp.Post.File.Md5, resp.Post.File.Ext)
		resp.Post.File.Url = &fixedUrl
	}
	if err != nil {
		return nil, errors.New(err)
	}
	return resp.Post, nil
}
