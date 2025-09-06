package scripts

import (
	"context"
	"e621-bot-go/modules/api"
	"e621-bot-go/modules/telegram_bot"
	"e621-bot-go/modules/worker"
	"e621-bot-go/storage"
	"e621-bot-go/utils"
	"os"
	"os/signal"
	"sync"

	"github.com/DrSmithFr/go-console"
)

func StartScript(cmd *go_console.Script) go_console.ExitCode {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	config, err := utils.ParseConfig("config.yaml")
	if err != nil {
		panic("Failed to parse config file: " + err.Error())
	}
	ctx = context.WithValue(ctx, "config", config)

	wg := &sync.WaitGroup{}
	ctx = context.WithValue(ctx, "wg", wg)

	store := storage.NewStorage(config.Redis)
	ctx = context.WithValue(ctx, "store", store)

	go worker.StartWorker(ctx)
	wg.Add(1)
	go telegram_bot.StartBot(ctx)
	wg.Add(1)
	go api.StartAPI(ctx)
	wg.Add(1)

	wg.Wait()

	return go_console.ExitSuccess
}
