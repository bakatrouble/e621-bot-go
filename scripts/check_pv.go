package scripts

import (
	"context"
	"e621-bot-go/e621"
	"e621-bot-go/storage"
	"e621-bot-go/utils"
	"e621-bot-go/worker"
	"fmt"
	"os"
	"os/signal"
	"strconv"

	go_console "github.com/DrSmithFr/go-console"
)

func CheckPvScript(cmd *go_console.Script) go_console.ExitCode {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	config, err := utils.ParseConfig("config.yaml")
	if err != nil {
		panic("Failed to parse config file: " + err.Error())
	}

	postIdStr := cmd.Input.Argument("post_id")
	postId, err := strconv.Atoi(postIdStr)
	if err != nil {
		fmt.Printf("Invalid post_id: %s\n", err.Error())
		return go_console.ExitError
	}

	logger := utils.NewLogger("check_pv")
	client := e621.NewE621(logger)
	store := storage.NewStorage(config.Redis)

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
		query := worker.CheckPostVersion(pv, queries)
		if query != nil {
			logger.With("post_version_id", pv.ID, "query", query.Raw).Info("Post version matched")
		} else {
			logger.With("post_version_id", pv.ID).Info("Post version did not match any query")
		}
	}

	return go_console.ExitSuccess
}
