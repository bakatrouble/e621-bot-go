package scripts

import (
	"context"
	"e621-bot-go/storage"
	"e621-bot-go/utils"
	"fmt"
	"os"
	"os/signal"
	"slices"

	"github.com/DrSmithFr/go-console"
)

func DumpScript(cmd *go_console.Script) go_console.ExitCode {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	config, err := utils.ParseConfig("config.yaml")
	if err != nil {
		panic("Failed to parse config file: " + err.Error())
	}
	ctx = context.WithValue(ctx, "config", config)

	store := storage.NewStorage(config.Redis)

	subs, err := store.GetSubs(ctx)
	if err != nil {
		fmt.Printf("Failed to get subs: %s\n", err.Error())
		return go_console.ExitError
	}

	slices.Sort(subs)

	if len(subs) == 0 {
		fmt.Printf("No subs found\n")
		return go_console.ExitSuccess
	} else {
		for _, sub := range subs {
			fmt.Printf(`"%s" `, sub)
		}
		fmt.Printf("\n")
	}

	return go_console.ExitSuccess
}
