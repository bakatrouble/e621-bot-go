package telegram_bot

import (
	"context"
	"e621-bot-go/e621"
	"e621-bot-go/storage"
	"e621-bot-go/telegram_bot/handlers"
	"e621-bot-go/utils"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

type botLogger struct {
	utils.Logger
	replacer *strings.Replacer
}

func (b *botLogger) Debugf(format string, args ...interface{}) {
	b.Debug(b.replacer.Replace(fmt.Sprintf(format, args...)))
}

func (b *botLogger) Errorf(format string, args ...interface{}) {
	b.Error(b.replacer.Replace(fmt.Sprintf(format, args...)))
}

func CreateBot(ctx context.Context, logger utils.Logger) (*telego.Bot, error) {
	config := ctx.Value("config").(*utils.Config)

	l := botLogger{logger, strings.NewReplacer(config.BotToken, "****")}

	return telego.NewBot(
		config.BotToken,
		telego.WithLogger(&l),
	)
}

func StartBot(ctx context.Context) {
	config := ctx.Value("config").(*utils.Config)
	store := ctx.Value("store").(*storage.Storage)
	wg := ctx.Value("wg").(*sync.WaitGroup)

	logger := utils.NewLogger("telegram-bot")
	logger.Info("starting telegram bot")

	defer wg.Done()

	bot, err := CreateBot(ctx, logger)
	if err != nil {
		logger.With("err", err).Error("failed to create bot")
		panic(err)
	}

	client := e621.NewE621(logger)
	ctx = context.WithValue(ctx, "e621", client)

	updates, _ := bot.UpdatesViaLongPolling(ctx, nil)
	bh, _ := th.NewBotHandler(bot, updates)
	defer func() {
		_ = bh.Stop()
	}()

	bh.Use(func(ctx *th.Context, update telego.Update) error {
		// Add context values for handlers
		ctx = ctx.WithValue("config", config)
		ctx = ctx.WithValue("logger", logger)
		ctx = ctx.WithValue("store", store)

		return ctx.Next(update)
	})
	bh.HandleMessage(handlers.AddCommandHandler, messageCommands([]string{"add"}))
	bh.HandleMessage(handlers.DelCommandHandler, messageCommands([]string{"del", "delete", "rm", "rem", "remove"}))
	bh.HandleCallbackQuery(handlers.SendCallbackHandler, th.CallbackDataPrefix("send:"))

	// Initialize done chan
	done := make(chan struct{}, 1)

	go func() {
		<-ctx.Done()
		logger.Warn("stopping telegram bot")
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer stopCancel()

	out:
		for len(updates) > 0 {
			select {
			case <-stopCtx.Done():
				break out
			case <-time.After(100 * time.Millisecond):
				// continue
			}
		}
		logger.Info("long polling done")
		_ = bh.StopWithContext(stopCtx)
		done <- struct{}{}
	}()

	go func() { _ = bh.Start() }()
	logger.Info("handling updates")

	<-done
	logger.Info("telegram bot stopped")
}

func messageCommands(commands []string) th.Predicate {
	return func(_ context.Context, update telego.Update) bool {
		if update.Message == nil {
			return false
		}

		matches := th.CommandRegexp.FindStringSubmatch(update.Message.Text)
		if len(matches) != th.CommandMatchGroupsLen {
			return false
		}

		for _, command := range commands {
			if strings.EqualFold(matches[th.CommandMatchCmdGroup], command) {
				return true
			}
		}
		return false
	}
}
