package commands

import (
	"fmt"
	"zumygo/libs"
	"time"
)

func init() {
	libs.NewCommands(&libs.ICommand{
		Name:     "(ping|p)",
		As:       []string{"ping"},
		Tags:     "main",
		IsPrefix: true,
		Execute: func(conn *libs.IClient, m *libs.IMessage) bool {
			start := time.Now()
			messageTime := time.Unix(m.Info.Timestamp.Unix(), 0)
			ping := start.Sub(messageTime).Seconds()
			response := fmt.Sprintf("*Ping :* %.2f Detik\n", ping)
			m.Reply(response)
			return true
		},
	})
}
