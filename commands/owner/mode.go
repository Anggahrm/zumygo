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
			currentMode := os.Getenv("PUBLIC")
			var newMode string
			var message string

			if currentMode == "false" {
				newMode = "true"
				message = "The bot is now in public mode."
			} else if currentMode == "true" {
				newMode = "false"
				message = "The bot is now in private mode."
			} else {
				newMode = "false"
				message = "The bot is now in private mode."
			}

			// First update the .env file
			err := helpers.UpdateEnvFile("PUBLIC", newMode)
			if err != nil {
				m.Reply("Failed to update .env file: " + err.Error())
				return false
			}

			// Only update memory after successful file update
			os.Setenv("PUBLIC", newMode)
			m.Reply(message)

			return true
		},
	})
}
