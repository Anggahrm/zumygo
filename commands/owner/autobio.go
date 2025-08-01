package commands

import (
	"fmt"
	"strconv"
	"strings"
	"zumygo/libs"
	Auto "zumygo/commands/Auto"
)

func init() {
	libs.NewCommands(&libs.ICommand{
		Name:        "(autobio|bio)",
		As:          []string{"autobio"},
		Tags:        "owner",
		IsPrefix:    true,
		IsOwner:     true,
		Description: "Control auto update bio system",
		Execute: func(conn *libs.IClient, m *libs.IMessage) bool {
			bioSystem := Auto.GetGlobalBioSystem()
			if bioSystem == nil {
				m.Reply("❎ Bio system not available")
				return false
			}

			// Check if arguments provided
			if len(m.Args) == 0 {
				// Show current status
				status := bioSystem.GetStatus()
				enabled := "❌ Disabled"
				running := "❌ Stopped"
				
				if status["enabled"].(bool) {
					enabled = "✅ Enabled"
				}
				if status["running"].(bool) {
					running = "✅ Running"
				}

				message := fmt.Sprintf(`*📝 Auto Bio System Status*

*Status:* %s
*Running:* %s
*Template:* %s
*Interval:* %d minutes

*Usage:*
• .autobio on/off - Enable/disable auto update
• .autobio template <text> - Set bio template
• .autobio interval <minutes> - Set update interval
• .autobio update - Force update now
• .autobio status - Show this status

*Template Variables:*
• {time} - Current time (HH:MM)
• {status} - Bot status
• {web} - Website URL
• {uptime} - Bot uptime
• {commands} - Command count
• {users} - User count
• {groups} - Group count`, 
					enabled, running, 
					status["template"], 
					status["interval"])

				m.Reply(message)
				return true
			}

			subCommand := strings.ToLower(m.Args[0])

			switch subCommand {
			case "on", "enable":
				enabled := bioSystem.ToggleAutoUpdate()
				if enabled {
					m.Reply("✅ Auto update bio enabled")
				} else {
					m.Reply("❌ Auto update bio disabled")
				}

			case "off", "disable":
				enabled := bioSystem.ToggleAutoUpdate()
				if enabled {
					m.Reply("✅ Auto update bio enabled")
				} else {
					m.Reply("❌ Auto update bio disabled")
				}

			case "template":
				if len(m.Args) < 2 {
					m.Reply("❎ Please provide a template text\n\nExample: .autobio template 🤖 Bot Online | ⏰ {time} | 📊 {status}")
					return false
				}
				
				template := strings.Join(m.Args[1:], " ")
				bioSystem.SetBioTemplate(template)
				m.Reply(fmt.Sprintf("✅ Bio template updated:\n%s", template))

			case "interval":
				if len(m.Args) < 2 {
					m.Reply("❎ Please provide interval in minutes\n\nExample: .autobio interval 30")
					return false
				}
				
				minutes, err := strconv.Atoi(m.Args[1])
				if err != nil || minutes < 1 {
					m.Reply("❎ Invalid interval. Must be a positive number")
					return false
				}
				
				if err := bioSystem.SetBioInterval(minutes); err != nil {
					m.Reply(fmt.Sprintf("❎ Failed to set interval: %v", err))
					return false
				}
				
				m.Reply(fmt.Sprintf("✅ Bio update interval set to %d minutes", minutes))

			case "update", "now":
				if err := bioSystem.UpdateBioNow(); err != nil {
					m.Reply(fmt.Sprintf("❎ Failed to update bio: %v", err))
					return false
				}
				
				m.Reply("✅ Bio updated successfully")

			case "status":
				// Show current status (same as no args)
				status := bioSystem.GetStatus()
				enabled := "❌ Disabled"
				running := "❌ Stopped"
				
				if status["enabled"].(bool) {
					enabled = "✅ Enabled"
				}
				if status["running"].(bool) {
					running = "✅ Running"
				}

				message := fmt.Sprintf(`*📝 Auto Bio System Status*

*Status:* %s
*Running:* %s
*Template:* %s
*Interval:* %d minutes`, 
					enabled, running, 
					status["template"], 
					status["interval"])

				m.Reply(message)

			default:
				m.Reply("❎ Unknown subcommand. Use: .autobio on/off/template/interval/update/status")
			}

			return true
		},
	})
} 