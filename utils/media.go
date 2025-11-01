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

	"github.com/nfnt/resize"
	"github.com/vimeo/go-magic/magic"
	"gopkg.in/vansante/go-ffprobe.v2"
)

func ConvertToMp4(ctx context.Context, mediaBytes []byte) ([]byte, error) {
	logger := ctx.Value("logger").(Logger)

	mime := magic.MimeFromBytes(mediaBytes)

	pattern := ""
	switch mime {
	case "video/mp4":
		pattern = "*.mp4"
	case "image/gif":
		pattern = "*.gif"
	case "video/webm", "application/octet-stream":
		pattern = "*.webm"
	default:
		return nil, fmt.Errorf("unsupported media type: %s", mime)
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
	width := uint(imConfig.Width)
	height := uint(imConfig.Height)
	if imConfig.Width+imConfig.Height > 10000 {
		scale := float64(10000) / (float64(width) + float64(height))
		width = uint(math.Floor(float64(imConfig.Width) * scale))
		height = uint(math.Floor(float64(imConfig.Height) * scale))
		im = resize.Thumbnail(width, height, im, resize.Lanczos3)
		fmt.Printf("Resized image to %dx%d\n", width, height)
	}
	buf := new(bytes.Buffer)
	for {
		if err = jpeg.Encode(buf, im, &jpeg.Options{Quality: 95}); err != nil {
			return nil, err
		}
		bufLen := len(buf.Bytes())
		if bufLen > 10*1024*1024 {
			buf.Reset()
			width = width * 95 / 100
			height = height * 95 / 100
			im = resize.Thumbnail(width, height, im, resize.Lanczos3)
			fmt.Printf("Buf is %d, resized image to %dx%d\n", bufLen, width, height)
		} else {
			break
		}
	}
	return buf.Bytes(), nil
}

func Mp4HasAudio(ctx context.Context, mediaBytes []byte) (bool, error) {
	var data *ffprobe.ProbeData
	var err error
	if data, err = ffprobe.ProbeReader(ctx, bytes.NewReader(mediaBytes)); err != nil {
		return false, err
	}
	if data == nil || data.Streams == nil {
		return false, fmt.Errorf("ffprobe returned nil data")
	}
	for _, stream := range data.Streams {
		if stream.CodecType == "audio" {
			return true, nil
		}
	}
	return false, nil
}
