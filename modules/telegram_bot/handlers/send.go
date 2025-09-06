package handlers

import (
	"e621-bot-go/utils"
	"fmt"
	"strings"

	"github.com/imroc/req/v3"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

type apiResponse struct {
	Status string `json:"status"`
}

func SendCallbackHandler(ctx *th.Context, callback telego.CallbackQuery) error {
	logger := ctx.Value("logger").(utils.Logger)
	config := ctx.Value("config").(*utils.Config)
	bot := ctx.Bot()

	var err error
	var cachePath string
	var exists bool

	cmd := strings.Split(callback.Data, ":")[1]
	args := strings.Split(cmd, " ")
	destination := args[0]
	switch destination {
	case "nsfw":
		destination = config.Destinations.Nsfw
	case "sfw":
		destination = config.Destinations.Sfw
	default:
		logger.Error("invalid destination")
		return fmt.Errorf("invalid destination")
	}

	cachedName := args[1]
	message := callback.Message.Message()

	if cachePath, exists = utils.IsCached(ctx, cachedName); !exists {
		var fileId string
		var file *telego.File
		var fileData []byte

		if message.Photo != nil && len(message.Photo) > 0 {
			fileId = message.Photo[len(message.Photo)-1].FileID
		} else if message.Document != nil {
			fileId = message.Document.FileID
		} else {
			logger.Error("no photo or document found in message")
			return fmt.Errorf("no photo or document found in message")
		}
		if file, err = bot.GetFile(ctx, &telego.GetFileParams{FileID: fileId}); err != nil {
			logger.With("err", err).Error("error getting image")
			return err
		}
		if fileData, err = tu.DownloadFile(bot.FileDownloadURL(file.FilePath)); err != nil {
			logger.With("err", err).Error("error downloading image")
			return err
		}
		if cachePath, err = utils.CacheFile(ctx, fileData, cachedName); err != nil {
			logger.With("err", err).Error("error caching image")
			return err
		}
	}

	var resp *req.Response
	if resp, err = req.R().
		SetContext(ctx).
		SetBodyJsonMarshal(map[string]string{"path": cachePath}).
		Post(fmt.Sprintf("%s/internalSend", destination)); err != nil {
		logger.With("err", err).Error("error sending image to destination")
		return err
	}

	var apiResp apiResponse
	if err = resp.Into(&apiResp); err != nil {
		logger.With("err", err).Error("error parsing response from destination")
		return err
	}

	responseText := ""
	switch apiResp.Status {
	case "ok":
		responseText = "Sent"
	case "duplicate":
		responseText = "Duplicate"
	default:
		responseText = "Error"
	}

	if err = bot.AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
		CallbackQueryID: callback.ID,
		Text:            responseText,
	}); err != nil {
		logger.With("err", err).Error("failed to answer callback query")
		return err
	}

	return nil
}
