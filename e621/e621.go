package e621

import (
	"e621-bot-go/utils"
	"fmt"

	"github.com/imroc/req/v3"
)

type E621 struct {
	httpClient *req.Client
}

type logger struct {
	l utils.Logger
}

func (l logger) Errorf(format string, v ...any) {
	l.l.Info(fmt.Sprintf(format, v...))
}

func (l logger) Warnf(format string, v ...any) {
	l.l.Warn(fmt.Sprintf(format, v...))
}

func (l logger) Debugf(format string, v ...any) {
	l.l.Debug(fmt.Sprintf(format, v...))
}

func NewE621(l utils.Logger) *E621 {
	return &E621{
		httpClient: req.C().
			SetBaseURL("https://e621.net").
			SetCommonHeader("user-agent", "bot/go-2.0 (bakatrouble)").
			SetLogger(logger{l}).
			EnableDebugLog(),
	}
}
