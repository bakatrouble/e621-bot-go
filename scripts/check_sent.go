package scripts

import (
	"context"
	"e621-bot-go/storage"
	"e621-bot-go/utils"
	"fmt"
	"os"
	"os/signal"
	"strconv"

	go_console "github.com/DrSmithFr/go-console"
)

func CheckSentScript(cmd *go_console.Script) go_console.ExitCode {
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

	store := storage.NewStorage(config.Redis)
	sent, err := store.IsPostSent(ctx, []int{postId})
	if err != nil {
		fmt.Printf("Failed to check if post is sent: %s\n", err.Error())
		return go_console.ExitError
	}

	if sent[postId] {
		fmt.Printf("Post ID %d has been sent\n", postId)
	} else {
		fmt.Printf("Post ID %d has not been sent\n", postId)
	}

	return go_console.ExitSuccess
}
