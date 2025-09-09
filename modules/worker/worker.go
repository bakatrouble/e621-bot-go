package worker

import (
	"context"
	"e621-bot-go/e621"
	"e621-bot-go/modules/telegram_bot"
	"e621-bot-go/utils"
	"sync"
	"time"

	tu "github.com/mymmrac/telego/telegoutil"
)

func StartWorker(ctx context.Context) {
	config := ctx.Value("config").(*utils.Config)
	wg := ctx.Value("wg").(*sync.WaitGroup)

	logger := utils.NewLogger("worker")
	logger.Info("starting worker")

	defer wg.Done()

	client := e621.NewE621(logger)
	ctx = context.WithValue(ctx, "e621", client)

	bot, err := telegram_bot.CreateBot(ctx, logger)
	if err != nil {
		logger.With("err", err).Error("failed to create bot")
		panic(err)
		return
	}
	ctx = context.WithValue(ctx, "bot", bot)
	ctx = context.WithValue(ctx, "logger", logger)

	go utils.CacheCleaner(ctx)

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			logger.Info("checking updates")
			if err = checkPosts(ctx); err != nil {
				_, _ = bot.SendMessage(ctx, tu.Message(
					tu.ID(config.ChatId),
					"Error occurred while checking posts: "+err.Error(),
				))
				ticker.Reset(config.Interval)
			}
		case <-ctx.Done():
			logger.Info("stopping worker")
			return
		}
	}
}
