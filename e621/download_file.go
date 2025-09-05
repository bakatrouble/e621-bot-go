package e621

import (
	"bufio"
	"bytes"
	"context"
)

func (e *E621) DownloadFile(ctx context.Context, url string) ([]byte, error) {
	var b bytes.Buffer
	buf := bufio.NewWriter(&b)
	if r := e.httpClient.Get(url).SetOutput(buf).Do(ctx); r.Err != nil {
		return nil, r.Err
	}
	return b.Bytes(), nil
}
