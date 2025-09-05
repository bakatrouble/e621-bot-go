package worker

import (
	"os"
	"testing"
)

func TestFfmpeg(t *testing.T) {
	ctx := t.Context()
	f, _ := os.ReadFile("/tmp/2154936000.mp4")
	_, err := convertToMp4(f, false, ctx)
	if err != nil {
		t.Error(err)
	}
}
