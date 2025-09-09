package e621

import (
	"context"
	"e621-bot-go/utils"
	"fmt"
	"strconv"

	"github.com/imroc/req/v3"
)

type PostVersionsRequest struct {
	httpClient *req.Client
	afterID    int
	beforeID   int
	postID     int
	limit      int
}

func (e *E621) GetPostVersions() *PostVersionsRequest {
	return &PostVersionsRequest{
		httpClient: e.httpClient,
		afterID:    0,
		beforeID:   0,
		postID:     0,
		limit:      320,
	}
}

func (r *PostVersionsRequest) WithAfterID(afterID int) *PostVersionsRequest {
	r.afterID = afterID
	return r
}

func (r *PostVersionsRequest) WithBeforeID(beforeID int) *PostVersionsRequest {
	r.beforeID = beforeID
	return r
}

func (r *PostVersionsRequest) WithLimit(limit int) *PostVersionsRequest {
	r.limit = limit
	return r
}

func (r *PostVersionsRequest) WithPostID(postID int) *PostVersionsRequest {
	r.postID = postID
	return r
}

func (r *PostVersionsRequest) Send(ctx context.Context) (postVersions []*PostVersion, err error) {
	logger := ctx.Value("logger").(utils.Logger)

	rq := r.httpClient.Get("/post_versions.json").
		SetQueryParam("limit", strconv.Itoa(r.limit))

	if r.beforeID > 0 {
		rq = rq.SetQueryParam("page", fmt.Sprintf("b%d", r.beforeID))
	} else if r.afterID > 0 {
		rq = rq.SetQueryParam("page", fmt.Sprintf("a%d", r.afterID))
	}

	if r.postID > 0 {
		rq = rq.SetQueryParam("search[post_id]", strconv.Itoa(r.postID))
	}

	rs := rq.Do(ctx)

	var resp struct {
		Success bool `json:"success"`
	}
	if err = rs.Into(&resp); err == nil {
		return // if response is a success flag, return empty slice
	}

	if err = rs.Into(&postVersions); err == nil {
		return // then return the slice
	}
	logger.With("response", rs.String()).Error("invalid api response")
	return
}
