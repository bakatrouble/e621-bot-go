package utils

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"math"
	"os"
	"os/exec"
	"strings"

	"github.com/liamg/magic"
	"github.com/nfnt/resize"
)

func ConvertToMp4(ctx context.Context, mediaBytes []byte) ([]byte, error) {
	logger := ctx.Value("logger").(Logger)

	t, err := magic.Lookup(mediaBytes)
	if err != nil {
		return nil, err
	}

	pattern := ""
	switch t.Extension {
	case "mp4", "gif", "webm":
		pattern = fmt.Sprintf("*.%s", t.Extension)
	default:
		return nil, fmt.Errorf("unsupported media type: %s", t.Extension)
	}

	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return nil, err
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(file.Name())

	_, err = file.Write(mediaBytes)
	if err != nil {
		return nil, err
	}

	err = file.Close()
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx,
		"ffmpeg",
		"-hide_banner",
		"-i", file.Name(),
		"-vf", "pad=width=ceil(iw/2)*2:height=ceil(ih/2)*2:x=0:y=0:color=black",
		"-c:v", "libx264",
		"-crf", "26",
		"-movflags", "+faststart",
		"-c:a", "aac",
		"-b:a", "128k",
		file.Name()+".mp4",
		"-y",
	)

	logger.With("cmd", strings.Join(cmd.Args, " ")).Info("running ffmpeg command")
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stdout

	if err = cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg error: %w", err)
	}

	defer func(name string) {
		_ = os.Remove(name)
	}(file.Name() + ".mp4")

	vfile, err := os.Open(file.Name() + ".mp4")
	if err != nil {
		return nil, err
	}

	mediaBytes, err = io.ReadAll(vfile)
	if err != nil {
		return nil, err
	}

	return mediaBytes, nil
}

func ResizeImage(imageBytes []byte) ([]byte, error) {
	imConfig, _, err := image.DecodeConfig(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, err
	}
	im, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, err
	}
	if imConfig.Width+imConfig.Height > 10000 {
		w := float64(imConfig.Width)
		h := float64(imConfig.Height)
		scale := float64(10000) / (w + h)
		im = resize.Thumbnail(uint(math.Floor(w*scale)), uint(math.Floor(h*scale)), im, resize.Lanczos3)
	}
	buf := new(bytes.Buffer)
	if err = jpeg.Encode(buf, im, &jpeg.Options{Quality: 95}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
