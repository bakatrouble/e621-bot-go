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

func AddCommandHandler(ctx *th.Context, message telego.Message) error {
	store := ctx.Value("store").(*storage.Storage)
	logger := ctx.Value("logger").(utils.Logger)
	//config := ctx.Value("config").(*utils.Config)
	bot := ctx.Bot()

	if gtfo(ctx, message) {
		return nil
	}

	logger.Info("AddCommandHandler called")

	args, err := argsParser.ParseString("", message.Text)
	if err != nil {
		logger.With("err", err).Error("failed to parse args")
		return err
	}

	if len(args.Args) < 2 {
		if _, err = bot.SendMessage(ctx, tu.Message(
			message.Chat.ChatID(),
			`Usage: /add <subscription1> <subscription2> ...\nExample: /add tag1 "tag2 tag3" -tag4`,
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

	addedSubs := make([]string, 0)
	skippedSubs := make([]string, 0)
	invalidSubs := make([]string, 0)

	for _, arg := range args.Args[1:] {
		arg = strings.ToLower(arg)
		if _, exists := existingSubs[arg]; exists {
			skippedSubs = append(skippedSubs, arg)
			continue
		}
		if _, err = utils.QueryParser.ParseString("", arg); err != nil {
			invalidSubs = append(invalidSubs, arg)
			continue
		}
		if err = store.AddSub(ctx, arg); err != nil {
			logger.With("err", err).Error("failed to add sub")
			return err
		}
		addedSubs = append(addedSubs, arg)
	}

	reply := ""
	if len(addedSubs) > 0 {
		reply += fmt.Sprintf("Added subscriptions:\n%s\n\n", strings.Join(addedSubs, "\n"))
	}
	if len(skippedSubs) > 0 {
		reply += fmt.Sprintf("Skipped existing subscriptions:\n%s\n\n", strings.Join(skippedSubs, "\n"))
	}
	if len(invalidSubs) > 0 {
		reply += fmt.Sprintf("Invalid subscriptions:\n%s", strings.Join(invalidSubs, "\n"))
	}

	if reply == "" {
		reply = "No subscriptions were added."
	}

	if _, err = bot.SendMessage(ctx, tu.Message(
		message.Chat.ChatID(),
		reply,
	)); err != nil {
		logger.With("err", err).Error("failed to send message")
	}

	return nil
}
