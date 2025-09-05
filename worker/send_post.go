package worker

import (
	"bytes"
	"context"
	"e621-bot-go/e621"
	"e621-bot-go/utils"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/nfnt/resize"
	_ "golang.org/x/image/webp"
)

func convertToMp4(mediaBytes []byte, gif bool, ctx context.Context) ([]byte, error) {
	logger := ctx.Value("logger").(utils.Logger)

	pattern := "*.mp4"
	if gif {
		pattern = "*.gif"
	}

	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return nil, err
	}
	//defer func(name string) {
	//	_ = os.Remove(name)
	//}(file.Name())

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
	)

	if !gif {
		cmd.Args = append(cmd.Args,
			"-c:a", "aac",
			"-b:a", "128k",
		)
	}

	cmd.Args = append(cmd.Args, file.Name()+".mp4", "-y")

	logger.With("cmd", strings.Join(cmd.Args, " ")).
		Debug("running ffmpeg command")
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stdout

	if err = cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg error: %w", err)
	}

	if err != nil {
		return nil, err
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

func resizeImage(imageBytes []byte) ([]byte, error) {
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
	if err = jpeg.Encode(buf, im, &jpeg.Options{Quality: 85}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func tagToHashtag(tag string) string {
	r := regexp.MustCompile("[^a-zA-Z0-9_]")
	return "#" + r.ReplaceAllString(tag, "_")
}

func buildCaption(post *e621.Post, query *utils.Query) string {
	queryTags := query.MentionedTags()
	monitoredTags := make([]string, 0)
	artistTags := make([]string, 0)
	characterTags := make([]string, 0)
	copyrightTags := make([]string, 0)

	for _, tag := range post.FlatTags() {
		if _, ok := queryTags[tag]; ok {
			monitoredTags = append(monitoredTags, tagToHashtag(tag))
		}
	}
	if len(monitoredTags) == 0 {
		monitoredTags = append(monitoredTags, "(none)")
	}
	for _, tag := range post.Tags.Artist {
		artistTags = append(artistTags, tagToHashtag(tag))
	}
	for _, tag := range post.Tags.Character {
		characterTags = append(characterTags, tagToHashtag(tag))
	}
	for _, tag := range post.Tags.Copyright {
		copyrightTags = append(copyrightTags, tagToHashtag(tag))
	}

	var result []string
	result = append(result, fmt.Sprintf("Monitored tags: <b>%s</b>", strings.Join(monitoredTags, " ")))
	if len(artistTags) > 0 {
		result = append(result, fmt.Sprintf("Artist: <b>%s</b>", strings.Join(artistTags, " ")))
	}
	if len(characterTags) > 0 {
		result = append(result, fmt.Sprintf("Character: <b>%s</b>", strings.Join(characterTags, " ")))
	}
	if len(copyrightTags) > 0 {
		result = append(result, fmt.Sprintf("Copyright: <b>%s</b>", strings.Join(copyrightTags, " ")))
	}
	result = append(result, "")
	result = append(result, fmt.Sprintf("https://e621.net/posts/%d", post.ID))
	return strings.Join(result, "\n")
}

func SendPost(ctx context.Context, client *e621.E621, postId int, query *utils.QueryInfo) error {
	bot := ctx.Value("bot").(*telego.Bot)
	config := ctx.Value("config").(*utils.Config)
	logger := ctx.Value("logger").(utils.Logger)

	post, err := client.GetPost(ctx, postId)
	if err != nil {
		return err
	}
	if post.File.Url == nil || *post.File.Url == "" {
		logger.With("file", post.File).Error("file url is empty")
		return errors.New("post has no file url")
	}

	caption := buildCaption(post, query.Query)

	mediaBytes, err := client.DownloadFile(ctx, *post.File.Url)
	if err != nil {
		return err
	}

	chatId := tu.ID(config.ChatId)
	switch post.File.Ext {
	case "jpg", "png", "webp":
		mediaBytes, err = resizeImage(mediaBytes)
		if err != nil {
			return err
		}
		if _, err = bot.SendPhoto(ctx,
			tu.Photo(
				chatId,
				tu.FileFromBytes(mediaBytes, "image.jpg"),
			).WithCaption(caption).WithParseMode("html"),
		); err != nil {
			return err
		}
	case "gif":
		mediaBytes, err = convertToMp4(mediaBytes, true, ctx)
		if err != nil {
			return err
		}
		if _, err = bot.SendDocument(ctx,
			tu.Document(
				chatId,
				tu.FileFromBytes(mediaBytes, "image.mp4"),
			).WithCaption(caption).WithParseMode("html"),
		); err != nil {
			return err
		}
	case "mp4", "webm":
		mediaBytes, err = convertToMp4(mediaBytes, false, ctx)
		if err != nil {
			return err
		}
		if _, err = bot.SendDocument(ctx,
			tu.Document(
				chatId,
				tu.FileFromBytes(mediaBytes, "image.mp4"),
			).WithCaption(caption).WithParseMode("html"),
		); err != nil {
			return err
		}
	}

	return nil
}
