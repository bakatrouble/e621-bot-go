package e621

import (
	"context"

	"github.com/go-errors/errors"
)

func (e *E621) GetTagAliases(ctx context.Context, tag string) (result []string, err error) {
	rq := e.httpClient.Get("/tag_aliases.json").
		SetQueryParam("search[name]", tag).
		Do(ctx)

	var resp struct {
		TagAliases []string `json:"tag_aliases"`
	}
	if err = rq.Into(&resp); err == nil {
		return // if tag_aliases key exists, results are empty
	}

	resp2 := make([]*TagAlias, 0)
	err = rq.Into(&resp2)
	if err = rq.Into(&resp2); err == nil { // if response is a slice of TagAlias, there are aliases
		for _, alias := range resp2 {
			if alias.Status == "active" {
				result = append(result, alias.ConsequentName)
			}
		}
		return
	}

	err = errors.New(err)
	return
}
