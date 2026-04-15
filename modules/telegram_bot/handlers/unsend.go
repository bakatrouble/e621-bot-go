package handlers

import (
	"e621-bot-go/utils"
	"fmt"
	"strings"

	"github.com/imroc/req/v3"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type unsendApiResponse struct {
	Status string `json:"status"`
}

func UnsendCallbackHandler(ctx *th.Context, callback telego.CallbackQuery) error {
	logger := ctx.Value("logger").(utils.Logger)
	config := ctx.Value("config").(*utils.Config)
	bot := ctx.Bot()

	var err error

	cmd := strings.Split(callback.Data, ":")[1]
	args := strings.Split(cmd, " ")
	destination := args[0]
	apiBase := ""
	switch destination {
	case "nsfw":
		apiBase = config.Destinations.Nsfw
	case "sfw":
		apiBase = config.Destinations.Sfw
	default:
		logger.Error("invalid destination")
		return fmt.Errorf("invalid destination")
	}

	uploadID := args[1]
	cachedName := args[2]
	message := callback.Message.Message()

	var resp *req.Response
	if resp, err = req.R().
		SetContext(ctx).
		SetBodyJsonMarshal(map[string]string{"upload_id": uploadID}).
		Delete(fmt.Sprintf("%s/internalDelete", apiBase)); err != nil {
		logger.With("err", err).Error("failed to call unsend api")
		return err
	}

	var apiResp unsendApiResponse
	if err = resp.Into(&apiResp); err != nil {
		logger.With("err", err).Error("error parsing response from api")
		return err
	}
	responseText := ""
	switch apiResp.Status {
	case "ok":
		responseText = "Unsent"
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

	if apiResp.Status == "ok" {
		kbd := message.ReplyMarkup
		switch args[0] {
		case "nsfw":
			kbd.InlineKeyboard[0][0].Text = "NSFW"
			kbd.InlineKeyboard[0][0].CallbackData = fmt.Sprintf("send:nsfw %s", cachedName)
		case "sfw":
			kbd.InlineKeyboard[0][1].Text = "SFW"
			kbd.InlineKeyboard[0][1].CallbackData = fmt.Sprintf("send:sfw %s", cachedName)
		}
		if _, err = bot.EditMessageReplyMarkup(ctx, &telego.EditMessageReplyMarkupParams{
			ChatID:      message.Chat.ChatID(),
			MessageID:   message.MessageID,
			ReplyMarkup: kbd,
		}); err != nil {
			logger.With("err", err).Error("failed to update reply markup")
			return err
		}
	}

	return nil
}
