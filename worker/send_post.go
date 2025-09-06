package worker

import (
	"context"
	"e621-bot-go/e621"
	"e621-bot-go/utils"
	"errors"
	"fmt"
	_ "image/gif"
	_ "image/png"
	"regexp"
	"strings"
	"time"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	_ "golang.org/x/image/webp"
)

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

func sendAsVideo(ctx context.Context, postId int, bytes []byte, caption string) error {
	bot := ctx.Value("bot").(*telego.Bot)
	config := ctx.Value("config").(*utils.Config)
	logger := ctx.Value("logger").(utils.Logger)

	if len(bytes) < 50*1024*1024 {
		_, err := bot.SendVideo(ctx,
			tu.Video(
				tu.ID(config.ChatId),
				tu.FileFromBytes(bytes, "file.mp4"),
			).
				WithSupportsStreaming().
				WithCaption(caption).
				WithParseMode("html"),
		)
		return err
	} else {
		logger.Info("file is too large, uploading to S3")
		url, err := utils.UploadToS3(ctx,
			config,
			fmt.Sprintf("e621-go-%d-%d.mp4", postId, time.Now().Unix()),
			bytes,
			"video/mp4",
		)
		logger.With("url", url).Info("uploaded to S3")
		if err != nil {
			return err
		}
		caption = fmt.Sprintf("%s\n\n%s", caption, url)
		_, err = bot.SendMessage(ctx,
			tu.Message(
				tu.ID(config.ChatId),
				caption,
			).WithParseMode("html"),
		)
		return err
	}
}

func sendAsPhoto(ctx context.Context, bytes []byte, caption string, kb *telego.InlineKeyboardMarkup) error {
	bot := ctx.Value("bot").(*telego.Bot)
	config := ctx.Value("config").(*utils.Config)

	_, err := bot.SendPhoto(ctx,
		tu.Photo(
			tu.ID(config.ChatId),
			tu.FileFromBytes(bytes, "image.jpg"),
		).
			WithCaption(caption).
			WithParseMode("html").
			WithReplyMarkup(kb),
	)
	return err
}

func SendPost(ctx context.Context, client *e621.E621, postId int, query *utils.QueryInfo) error {
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

	kb := tu.InlineKeyboard([]telego.InlineKeyboardButton{
		{
			Text:         "NSFW",
			CallbackData: "send:nsfw",
		},
		{
			Text:         "SFW",
			CallbackData: "send:sfw",
		},
	})

	switch post.File.Ext {
	case "jpg", "png", "webp":
		mediaBytes, err = utils.ResizeImage(mediaBytes)
		logger.With("size", len(mediaBytes)).Debug("resized image")
		if err != nil {
			return err
		}
		if err = sendAsPhoto(ctx, mediaBytes, caption, kb); err != nil {
			return err
		}
	case "gif", "mp4", "webm":
		mediaBytes, err = utils.ConvertToMp4(ctx, mediaBytes)
		if err != nil {
			return err
		}
		if err = sendAsVideo(ctx, postId, mediaBytes, caption); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported file extension: %s", post.File.Ext)
	}

	return nil
}
