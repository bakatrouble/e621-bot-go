package utils

import (
	"context"
	"os"
	"testing"
)

func TestFfmpeg(t *testing.T) {
	ctx := t.Context()
	logger := NewLogger("")
	ctx = context.WithValue(ctx, "logger", logger)

	f, _ := os.ReadFile("/tmp/2154936000.mp4")
	_, err := ConvertToMp4(ctx, f)
	if err != nil {
		t.Error(err)
	}
}
