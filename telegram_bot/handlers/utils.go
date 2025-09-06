package handlers

import (
	"e621-bot-go/utils"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

type argsStruct struct {
	Args []string `(@QuotedArg | @Arg)*`
}

var (
	argsLexer = lexer.MustSimple([]lexer.SimpleRule{
		{`QuotedArg`, `".+?"`},
		{`Arg`, `[^\s]+`},
		{"whitespace", `\s+`},
	})

	argsParser = participle.MustBuild[argsStruct](
		participle.Lexer(argsLexer),
		participle.Unquote("QuotedArg"),
	)
)

func reactToMessage(ctx *th.Context, message *telego.Message) {
	_ = ctx.Bot().SetMessageReaction(ctx, &telego.SetMessageReactionParams{
		ChatID:    message.Chat.ChatID(),
		MessageID: message.MessageID,
		Reaction: []telego.ReactionType{&telego.ReactionTypeEmoji{
			Type:  telego.ReactionEmoji,
			Emoji: "👍",
		}},
	})
}

func gtfo(ctx *th.Context, message telego.Message) bool {
	bot := ctx.Bot()
	config := ctx.Value("config").(*utils.Config)

	if config.ChatId == message.Chat.ID {
		return false
	}

	_, _ = bot.SendMessage(ctx, tu.Message(
		message.Chat.ChatID(),
		"GTFO",
	))
	return true
}
