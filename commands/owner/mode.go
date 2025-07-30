package commands

import (
	"zumygo/helpers"
	"zumygo/libs"
	"os"
)

func init() {
	libs.NewCommands(&libs.ICommand{
		Name:     `mode`,
		As:       []string{"mode"},
		Tags:     "owner",
		IsPrefix: true,
		IsOwner:  true,
		Execute: func(conn *libs.IClient, m *libs.IMessage) bool {
			mode := os.Getenv("PUBLIC")

			if mode == "false" {
				mode = "true"
				m.Reply("The bot is now in public mode.")
			} else if mode == "true" {
				mode = "false"
				m.Reply("The bot is now in private mode.")
			} else {
				mode = "false"
				m.Reply("The bot is now in private mode.")
			}

			os.Setenv("PUBLIC", mode)

			// Update the .env file
			err := helpers.UpdateEnvFile("PUBLIC", mode)
			if err != nil {
				m.Reply("Failed to update .env file: " + err.Error())
				return false
			}

			return true
		},
	})
}
