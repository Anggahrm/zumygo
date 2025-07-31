package commands

import (
	"zumygo/libs"
	"zumygo/config"
)

func init() {
	libs.NewCommands(&libs.ICommand{
		Name:     `mode`,
		As:       []string{"mode"},
		Tags:     "owner",
		IsPrefix: true,
		IsOwner:  true,
		Execute: func(conn *libs.IClient, m *libs.IMessage) bool {
			cfg := config.Config
			var message string

			if cfg.PublicMode {
				cfg.PublicMode = false
				message = "The bot is now in private mode."
			} else {
				cfg.PublicMode = true
				message = "The bot is now in public mode."
			}

			m.Reply(message)
			return true
		},
	})
}
