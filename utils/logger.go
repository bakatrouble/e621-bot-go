package utils

import (
	"context"
	"e621-bot-go/utils/tint"
	"fmt"
	"log/slog"
	"os"
	"path"

	"github.com/cappuccinotm/slogx"
	"github.com/gookit/slog/rotatefile"
)

// ApplyHandler wraps slog.Handler as Middleware.
func applyHandler(handler slog.Handler) slogx.Middleware {
	return func(next slogx.HandleFunc) slogx.HandleFunc {
		return func(ctx context.Context, rec slog.Record) error {
			err := handler.Handle(ctx, rec)
			if err != nil {
				return err
			}

			return next(ctx, rec)
		}
	}
}

type Logger = *slog.Logger

func NewLogger(module string) Logger {
	level := slog.LevelDebug

	prefix := ""
	if module != "" {
		prefix = fmt.Sprintf("[%s] ", module)
	}

	textHandler := tint.NewHandler(os.Stdout, &tint.Options{
		Level:  level,
		Prefix: prefix,
	})

	var handlers []slogx.Middleware

	if module != "" {
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
		handlers = append(handlers,
			applyHandler(slog.NewTextHandler(writer, &slog.HandlerOptions{Level: level})),
		)
	}

	return slog.New(slogx.Accumulator(slogx.NewChain(textHandler, handlers...)))
}
