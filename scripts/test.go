package scripts

import (
	"context"
	"e621-bot-go/e621"
	"e621-bot-go/modules/telegram_bot"
	"e621-bot-go/modules/worker"
	"e621-bot-go/utils"
	"fmt"
	"os"
	"os/signal"
	"strconv"

	"github.com/DrSmithFr/go-console"
)

func TestScript(cmd *go_console.Script) go_console.ExitCode {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	postIdStr := cmd.Input.Argument("post_id")
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		fmt.Printf("Invalid post_id: %s\n", err.Error())
		return go_console.ExitError
	}

	config, err := utils.ParseConfig("config.yaml")
	if err != nil {
		panic("Failed to parse config file: " + err.Error())
	}
	ctx = context.WithValue(ctx, "config", config)

	logger := utils.NewLogger("")
	ctx = context.WithValue(ctx, "logger", logger)

	bot, err := telegram_bot.CreateBot(ctx, logger)
	if err != nil {
		logger.With("err", err).Error("failed to create bot")
		return go_console.ExitError
	}
	ctx = context.WithValue(ctx, "bot", bot)

	client := e621.NewE621(logger)

	err = worker.SendPost(ctx, client, postId, []*utils.QueryInfo{})
	if err != nil {
		fmt.Printf("Failed to send post: %s\n", err.Error())
		return go_console.ExitError
	}

	return go_console.ExitSuccess
}
