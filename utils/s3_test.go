package utils

import (
	"os"
	"testing"
)

func TestS3(t *testing.T) {
	ctx := t.Context()

	config, err := ParseConfig("../config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	f, _ := os.ReadFile("/tmp/2154936000.mp4")
	url, err := UploadToS3(ctx, config, "2154936000.mp4", f, "video/mp4")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Upload success: %s", url)
}
