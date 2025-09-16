package e621

import (
	"bufio"
	"bytes"
	"context"
	"fmt"

	"github.com/imroc/req/v3"
)

func (e *E621) DownloadFile(ctx context.Context, url string) ([]byte, error, bool) {
	var b bytes.Buffer
	var r *req.Response
	buf := bufio.NewWriter(&b)
	if r = e.httpClient.Get(url).SetOutput(buf).Do(ctx); r.Err != nil {
		return nil, r.Err, false
	}
	if r.IsErrorState() && r.StatusCode == 404 {
		return nil, fmt.Errorf("server responded 404"), true
	}
	return b.Bytes(), nil, false
}
