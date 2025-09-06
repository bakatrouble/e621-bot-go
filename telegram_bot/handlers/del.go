package handlers

import (
	"e621-bot-go/storage"
	"e621-bot-go/utils"
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func DelCommandHandler(ctx *th.Context, message telego.Message) error {
	store := ctx.Value("store").(*storage.Storage)
	logger := ctx.Value("logger").(utils.Logger)
	//config := ctx.Value("config").(*utils.Config)
	bot := ctx.Bot()

	if gtfo(ctx, message) {
		return nil
	}

	logger.Info("DelCommandHandler called")

	args, err := argsParser.ParseString("", message.Text)
	if err != nil {
		logger.With("err", err).Error("failed to parse args")
		return err
	}

	if len(args.Args) < 2 {
		if _, err = bot.SendMessage(ctx, tu.Message(
			message.Chat.ChatID(),
			`Usage: /del <subscription1> <subscription2> ...\nExample: /del tag1 "tag2 tag3" -tag4`,
		)); err != nil {
			logger.With("err", err).Error("failed to send message")
			return err
		}
		return nil
	}

	existingSubs, err := store.GetSubsMap(ctx)
	if err != nil {
		logger.With("err", err).Error("failed to get existing subs")
		return err
	}

	removedSubs := make([]string, 0)
	missingSubs := make([]string, 0)

	for _, arg := range args.Args[1:] {
		arg = strings.ToLower(arg)
		if _, exists := existingSubs[arg]; !exists {
			missingSubs = append(missingSubs, arg)
			continue
		}
		if err = store.RemoveSub(ctx, arg); err != nil {
			logger.With("err", err).Error("failed to remove sub")
			return err
		}
		removedSubs = append(removedSubs, arg)
	}

	reply := ""
	if len(removedSubs) > 0 {
		reply += fmt.Sprintf("Removed subscriptions:\n%s\n\n", strings.Join(removedSubs, "\n"))
	}
	if len(missingSubs) > 0 {
		reply += fmt.Sprintf("Subscriptions not found:\n%s\n\n", strings.Join(missingSubs, "\n"))
	}

	if reply == "" {
		reply = "No subscriptions were removed."
	}

	if _, err = bot.SendMessage(ctx, tu.Message(
		message.Chat.ChatID(),
		reply,
	)); err != nil {
		logger.With("err", err).Error("failed to send message")
	}

	return nil
}
