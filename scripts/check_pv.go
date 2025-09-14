package scripts

import (
	"context"
	"e621-bot-go/e621"
	"e621-bot-go/modules/telegram_bot"
	"e621-bot-go/modules/worker"
	"e621-bot-go/storage"
	"e621-bot-go/utils"
	"fmt"
	"os"
	"os/signal"
	"strconv"

	"github.com/DrSmithFr/go-console"
)

func CheckPvScript(cmd *go_console.Script) go_console.ExitCode {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	config, err := utils.ParseConfig("config.yaml")
	if err != nil {
		panic("Failed to parse config file: " + err.Error())
	}
	ctx = context.WithValue(ctx, "config", config)

	postIdStr := cmd.Input.Argument("post_id")
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		fmt.Printf("Invalid post_id: %s\n", err.Error())
		return go_console.ExitError
	}

	logger := utils.NewLogger("")
	ctx = context.WithValue(ctx, "logger", logger)
	bot, _ := telegram_bot.CreateBot(ctx, logger)
	ctx = context.WithValue(ctx, "bot", bot)
	client := e621.NewE621(logger)
	ctx = context.WithValue(ctx, "e621", client)
	store := storage.NewStorage(config.Redis)
	ctx = context.WithValue(ctx, "store", store)

	pvs, err := client.GetPostVersions().WithPostID(postId).Send(ctx)
	if err != nil {
		fmt.Printf("Failed to get post versions: %s\n", err.Error())
		return go_console.ExitError
	}

	if len(pvs) == 0 {
		fmt.Printf("No post versions found for post ID %d\n", postId)
		return go_console.ExitSuccess
	}

	queries, err := utils.GetQueries(store, ctx)
	if err != nil {
		logger.With("err", err).Error("failed to get queries")
		return go_console.ExitError
	}

	for _, pv := range pvs {
		if match := worker.CheckPostVersion(pv, queries); match != nil {
			logger.With("post_version_id", pv.ID, "query", match.Raw).Info("Post version matched")
			if err = worker.SendPost(ctx, client, pv.PostID, queries); err != nil {
				logger.With("err", err).Error("failed to send post")
				return go_console.ExitError
			}
		} else {
			logger.With("post_version_id", pv.ID).Info("Post version did not match any query")
		}
	}

	return go_console.ExitSuccess
}
