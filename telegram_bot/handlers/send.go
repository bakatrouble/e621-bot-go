package handlers

import (
	"e621-bot-go/utils"
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func SendCallbackHandler(ctx *th.Context, callback telego.CallbackQuery) error {
	logger := ctx.Value("logger").(utils.Logger)
	config := ctx.Value("config").(*utils.Config)
	bot := ctx.Bot()

	cmd := strings.Split(callback.Data, ":")[1]
	message := callback.Message.Message()
	media := message.Photo[len(message.Photo)-1].FileID

	_, err := bot.SendPhoto(ctx, tu.Photo(
		tu.ID(config.SharedChatId),
		tu.FileFromID(media),
	).WithCaption(fmt.Sprintf("/%s", cmd)))
	if err != nil {
		logger.With("err", err).Error("failed to send photo to shared chat")
		_ = bot.AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
			CallbackQueryID: callback.ID,
			Text:            "Error",
		})
		return err
	}

	if err = bot.AnswerCallbackQuery(ctx, &telego.AnswerCallbackQueryParams{
		CallbackQueryID: callback.ID,
		Text:            "Sent",
	}); err != nil {
		logger.With("err", err).Error("failed to answer callback query")
		return err
	}

	return nil
}
