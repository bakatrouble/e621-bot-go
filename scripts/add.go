package scripts

import (
	"context"
	"e621-bot-go/storage"
	"e621-bot-go/utils"
	"fmt"
	"os"
	"os/signal"
	"strings"

	go_console "github.com/DrSmithFr/go-console"
)

func AddScript(cmd *go_console.Script) go_console.ExitCode {
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

	addedSubs := make([]string, 0)
	failedSubs := make([]string, 0)
	for _, sub := range cmd.Input.ArgumentList("subs") {
		sub = strings.ToLower(sub)

		if _, err = utils.QueryParser.ParseString(sub, sub); err != nil {
			fmt.Printf("Failed to parse query: %s\n", err.Error())
			failedSubs = append(failedSubs, sub)
			continue
		}
		if _, exists := existingSubs[sub]; exists {
			fmt.Printf("Sub already exists: %s\n", sub)
			continue
		}

		err = store.AddSub(ctx, sub)
		addedSubs = append(addedSubs, sub)
	}

	fmt.Printf("Added %d subs\n", len(addedSubs))
	if len(failedSubs) > 0 {
		fmt.Printf("Failed to add %d subs\n", len(failedSubs))
		for _, sub := range failedSubs {
			fmt.Printf(" - %s\n", sub)
		}
	}

	return go_console.ExitSuccess
}
