package main

import (
	"e621-bot-go/scripts"

	go_console "github.com/DrSmithFr/go-console"
	"github.com/DrSmithFr/go-console/input/argument"
)

func main() {
	script := go_console.Command{
		Description: "e621 subscription bot",
		Scripts: []*go_console.Script{
			{
				Name:        "start",
				Description: "Start the worker",
				Runner:      scripts.StartScript,
			},
			{
				Name:        "add",
				Description: "Add subscriptions",
				Arguments: []go_console.Argument{
					{
						Name:        "subs",
						Description: "Subscriptions to add",
						Value:       argument.List | argument.Required,
					},
				},
				Runner: scripts.AddScript,
			},
			{
				Name:        "remove",
				Description: "Remove subscriptions",
				Arguments: []go_console.Argument{
					{
						Name:        "subs",
						Description: "Subscriptions to remove",
						Value:       argument.List | argument.Required,
					},
				},
				Runner: scripts.RemoveScript,
			},
			{
				Name:        "test",
				Description: "Test sending a post",
				Arguments: []go_console.Argument{
					{
						Name:        "post_id",
						Description: "ID of the post to send",
						Value:       argument.Required,
					},
				},
				Runner: scripts.TestScript,
			},
			{
				Name:        "check-pv",
				Description: "Check post version against subscriptions",
				Arguments: []go_console.Argument{
					{
						Name:        "post_id",
						Description: "ID of the post to check versions for",
						Value:       argument.Required,
					},
				},
				Runner: scripts.CheckPvScript,
			},
			{
				Name:        "dump",
				Description: "Dump all subscriptions",
				Runner:      scripts.DumpScript,
			},
		},
	}
	script.Run()
}
