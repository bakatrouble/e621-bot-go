package scripts

import (
	"context"
	"e621-bot-go/storage"
	"e621-bot-go/utils"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/DrSmithFr/go-console"
)

func RemoveScript(cmd *go_console.Script) go_console.ExitCode {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	config, err := utils.ParseConfig("config.yaml")
	if err != nil {
		panic("Failed to parse config file: " + err.Error())
	}
	ctx = context.WithValue(ctx, "config", config)

	store := storage.NewStorage(config.Redis)

	existingSubs, err := store.GetSubsMap(ctx)
	if err != nil {
		fmt.Printf("Failed to get existing subs: %s\n", err.Error())
		return go_console.ExitError
	}

	removedSubs := make([]string, 0)
	for _, sub := range cmd.Input.ArgumentList("subs") {
		sub = strings.ToLower(sub)

		if _, err = utils.QueryParser.ParseString(sub, sub); err != nil {
			fmt.Printf("Failed to parse query: %s\n", err.Error())
			continue
		}
		if _, exists := existingSubs[sub]; !exists {
			fmt.Printf("Sub does not exist: %s\n", sub)
			continue
		}

		err = store.RemoveSub(ctx, sub)
		removedSubs = append(removedSubs, sub)
	}

	fmt.Printf("Removed %d subs\n", len(removedSubs))

	return go_console.ExitSuccess
}
