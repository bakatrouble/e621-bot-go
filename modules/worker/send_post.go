package worker

import (
	"context"
	"e621-bot-go/e621"
	"e621-bot-go/utils"
	"fmt"
	_ "image/gif"
	_ "image/png"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/go-errors/errors"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	_ "golang.org/x/image/webp"
)

func tagToHashtag(tag string) string {
	r := regexp.MustCompile("[^a-zA-Z0-9_]")
	return "#" + r.ReplaceAllString(tag, "_")
}

func buildCaption(post *e621.Post, matches []*utils.QueryInfo, queryTags map[string]struct{}) (string, bool) {
	if len(matches) == 0 {
		return "", true
	}
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
		return "", true
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

	slices.Sort(monitoredTags)
	slices.Sort(artistTags)
	slices.Sort(characterTags)
	slices.Sort(copyrightTags)

	var result []string
	result = append(result, fmt.Sprintf("Monitored tags: <b>%s</b>", strings.Join(monitoredTags, " ")))
	result = append(result, "Matched queries:")
	for _, match := range matches {
		result = append(result, fmt.Sprintf(" - <code>%s</code>", match.Raw))
	}
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
	return strings.Join(result, "\n"), false
}

func buildKeyboard(fileName string) *telego.InlineKeyboardMarkup {
	return tu.InlineKeyboard([]telego.InlineKeyboardButton{
		{
			Text:         "NSFW",
			CallbackData: fmt.Sprintf("send:nsfw %s", fileName),
		},
		{
			Text:         "SFW",
			CallbackData: fmt.Sprintf("send:sfw %s", fileName),
		},
	})
}

func sendAsVideo(ctx context.Context, postId int, bytes []byte, caption string) error {
	bot := ctx.Value("bot").(*telego.Bot)
	config := ctx.Value("config").(*utils.Config)
	logger := ctx.Value("logger").(utils.Logger)

	if len(bytes) < 50*1024*1024 {
		cachedName := fmt.Sprintf("%d-%d.mp4", postId, time.Now().Unix())
		_, _ = utils.CacheFile(ctx, bytes, cachedName)
		kb := buildKeyboard(cachedName)

		var hasAudio bool
		var err error
		if hasAudio, err = utils.Mp4HasAudio(ctx, bytes); err != nil {
			logger.With("err", err).Error("failed to check if mp4 has audio")
			hasAudio = true
		} else if hasAudio {
			_, err = bot.SendVideo(ctx,
				&telego.SendVideoParams{
					ChatID:            tu.ID(config.ChatId),
					Video:             tu.FileFromBytes(bytes, "file.mp4"),
					SupportsStreaming: true,
					Caption:           caption,
					ParseMode:         "html",
					ReplyMarkup:       kb,
				},
			)
		} else {
			_, err = bot.SendAnimation(ctx,
				&telego.SendAnimationParams{
					ChatID:      tu.ID(config.ChatId),
					Animation:   tu.FileFromBytes(bytes, "file.mp4"),
					Caption:     caption,
					ParseMode:   "html",
					ReplyMarkup: kb,
				},
			)
		}
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

func sendAsPhoto(ctx context.Context, postId int, bytes []byte, caption string) error {
	bot := ctx.Value("bot").(*telego.Bot)
	config := ctx.Value("config").(*utils.Config)

	cachedName := fmt.Sprintf("%d-%d.jpg", postId, time.Now().Unix())
	_, _ = utils.CacheFile(ctx, bytes, cachedName)
	kb := buildKeyboard(cachedName)

	if _, err := bot.SendPhoto(ctx,
		tu.Photo(
			tu.ID(config.ChatId),
			tu.FileFromBytes(bytes, "image.jpg"),
		).
			WithCaption(caption).
			WithParseMode("html").
			WithReplyMarkup(kb),
	); err != nil {
		return errors.New(err)
	}
	return nil
}

func sendAsFile(ctx context.Context, postId int, bytes []byte, ext string, caption string) error {
	bot := ctx.Value("bot").(*telego.Bot)
	config := ctx.Value("config").(*utils.Config)

	cachedName := fmt.Sprintf("%d-%d.jpg", postId, time.Now().Unix())
	_, _ = utils.CacheFile(ctx, bytes, cachedName)
	kb := buildKeyboard(cachedName)

	if _, err := bot.SendDocument(ctx,
		tu.Document(
			tu.ID(config.ChatId),
			tu.FileFromBytes(bytes, fmt.Sprintf("file.%s", ext)),
		).
			WithCaption(caption).
			WithParseMode("html").
			WithReplyMarkup(kb),
	); err != nil {
		return errors.New(err)
	}
	return nil
}

func SendPost(ctx context.Context, client *e621.E621, postId int, matches []*utils.QueryInfo, queries []*utils.QueryInfo) error {
	logger := ctx.Value("logger").(utils.Logger)

	post, err := client.GetPost(ctx, postId)
	if err != nil {
		return errors.New(err)
	}
	if post.File.Url == nil || *post.File.Url == "" {
		logger.With("file", post.File).Error("file url is empty")
		return errors.New("post has no file url")
	}

	monitoredTags := make(map[string]struct{}, 0)
	for _, query := range queries {
		if query.Check(post.FlatTagsMap()) {
			for tag := range query.Query.MentionedTags(false) {
				monitoredTags[tag] = struct{}{}
			}
		}
	}

	caption, abort := buildCaption(post, matches, monitoredTags)
	if abort {
		logger.Info("no monitored tags found in post, skipping")
		return nil
	}

	var mediaBytes []byte
	var skip bool
	if mediaBytes, err, skip = client.DownloadFile(ctx, *post.File.Url); err != nil {
		if skip {
			logger.With("err", err, "url", *post.File.Url).Warn("skipping send")
			return nil
		}
		return errors.New(err)
	}

	switch post.File.Ext {
	case "jpg", "png", "webp":
		if mediaBytes, err = utils.ResizeImage(mediaBytes); err != nil {
			if strings.Contains(err.Error(), "invalid checksum") || strings.Contains(err.Error(), "invalid format") {
				return nil
			}
			return errors.New(err)
		}
		logger.With("size", len(mediaBytes)).Debug("resized image")

		if err = sendAsPhoto(ctx, postId, mediaBytes, caption); err != nil {
			return errors.New(err)
		}
	case "gif", "mp4", "webm":
		if mediaBytes, err = utils.ConvertToMp4(ctx, mediaBytes); err != nil {
			return errors.New(err)
		}
		logger.With("size", len(mediaBytes)).Debug("converted to mp4")

		if err = sendAsVideo(ctx, postId, mediaBytes, caption); err != nil {
			return errors.New(err)
		}
	case "swf":
		if err = sendAsFile(ctx, postId, mediaBytes, "swf", caption); err != nil {
			return errors.New(err)
		}
	default:
		return errors.Errorf("unsupported file extension: %s", post.File.Ext)
	}

	return nil
}
