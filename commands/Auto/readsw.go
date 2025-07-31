package commands

import (
	"math/rand"
	"time"
	"zumygo/libs"
	"zumygo/config"

	"go.mau.fi/whatsmeow/types"
)

func init() {
	libs.NewCommands(&libs.ICommand{
		Before: func(conn *libs.IClient, m *libs.IMessage) {
			// Check if this is a status message
			if m.Info.Chat.String() == "status@broadcast" {
				cfg := config.Config
				
				// Check if auto-read status is enabled
				if cfg.ReadStatus {
					// Mark status as read
					err := conn.WA.MarkRead([]types.MessageID{m.Info.ID}, m.Info.Timestamp, m.Info.Chat, m.Info.Sender)
					if err != nil {
						// Log error but don't panic
						return
					}

					// Check if auto-react status is enabled
					if cfg.ReactStatus {
						// List of emojis for random reactions
						emojis := []string{
							"😀", "😃", "😄", "😁", "😆", "🥹", "😅", "😂", "🤣", "🥲", "☺️", "😊", "😇", "🙂", "🙃", "😉", "😌", "😍", "🥰", "😘", "😗", "😙", "😚", "😋", "😛", "😝", "🤪", "🤨", "🧐", "🤓", "😎", "🥸", "🤩", "🥳", "😏", "😒", "😞", "😔", "😟", "😕", "🙁", "☹️", "😣", "😖", "😫", "😩", "🥺", "😢", "😭", "😤", "😠", "😡", "🤬", "🤯", "😳", "🥵", "🥶", "😶‍🌫️", "😱", "😨", "😰", "😥", "😓", "🤗", "🤔", "🫣", "🤭", "🫢", "🫡", "🤫", "🫠", "🤥", "😶", "🫥", "😐", "🫤", "😑", "😬", "🙄", "😯", "😦", "😧", "😮", "😲", "🥱", "😴", "🤤", "😪", "😮‍💨", "😵", "😵‍💫", "🤐", "🥴", "🤢", "🤮", "🤧", "😷", "🤒", "🤕", "🤑", "🤡", "💩", "👻", "💀", "☠️", "🙌", "👏", "👍", "👎", "👊", "✊", "🤛", "🤞", "✌️", "🫰", "🤟", "🤘", "👌", "🤏", "☝️", "✋", "🤚", "🖖", "👋", "🤙", "🫲", "🫱", "💪", "🖕", "✍️", "🙏", "🫵", "🦶", "👣", "👀", "🧠",
						}

						// Use modern random generation (Go 1.20+)
						r := rand.New(rand.NewSource(time.Now().UnixNano()))
						randomEmoji := emojis[r.Intn(len(emojis))]

						// React to status with random emoji
						_, err = m.React(randomEmoji)
						if err != nil {
							// Log error but don't panic
							return
						}
					}
				}
			}
		},
	})
}
