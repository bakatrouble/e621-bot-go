package e621

import (
	"context"
	"strconv"

	"github.com/go-errors/errors"
	"github.com/imroc/req/v3"
)

type PostsRequest struct {
	httpClient *req.Client
	tags       string
	page       int
	limit      int
}

type postsResponse struct {
	Posts []*Post `json:"posts"`
}

func (e *E621) GetPosts() *PostsRequest {
	return &PostsRequest{
		httpClient: e.httpClient,
		page:       1,
		limit:      320,
	}
}

func (*PostsRequest) WithTags(tags string) *PostsRequest {
	return &PostsRequest{tags: tags}
}

func (*PostsRequest) WithPage(page int) *PostsRequest {
	return &PostsRequest{page: page}
}

func (*PostsRequest) WithLimit(limit int) *PostsRequest {
	return &PostsRequest{limit: limit}
}

func (r *PostsRequest) Send(ctx context.Context) ([]*Post, error) {
	var resp postsResponse
	err := r.httpClient.Get("/posts.json").
		SetQueryParam("tags", r.tags).
		SetQueryParam("page", strconv.Itoa(r.page)).
		SetQueryParam("limit", strconv.Itoa(r.limit)).
		Do(ctx).
		Into(&resp)
	if err != nil {
		return nil, errors.New(err)
	}
	return resp.Posts, nil
}
