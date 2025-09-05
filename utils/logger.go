package utils

import (
	"e621-bot-go/utils/tint"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/gookit/slog/rotatefile"
	"github.com/samber/slog-multi"
)

type Logger = *slog.Logger

func createLogsDir(name string) {
	err := os.MkdirAll("logs", os.ModePerm)
	if err != nil {
		panic("failed to create logs directory: " + err.Error())
	}
}

func NewLogger(module string) Logger {

	level := slog.LevelDebug
	writer, err := rotatefile.NewConfig(
		path.Join("logs", fmt.Sprintf("%s.log", module)),
		func(c *rotatefile.Config) {
			c.MaxSize = 10 * 1024 * 1024 // 10 MB
			c.BackupNum = 5
			c.RotateTime = rotatefile.EveryMonth
			c.Compress = true
		},
	).Create()
	if err != nil {
		panic("failed to create log file: " + err.Error())
	}

	consoleHandler := tint.NewHandler(os.Stdout, &tint.Options{
		Level:  level,
		Prefix: fmt.Sprintf("[%s]", module),
	})

	return slog.New(
		slogmulti.Fanout(
			consoleHandler,
			slog.NewTextHandler(writer, &slog.HandlerOptions{Level: level}),
		),
	)
}
