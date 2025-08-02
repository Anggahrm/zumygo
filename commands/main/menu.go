package commands

import (
	"fmt"
	"sort"
	"strings"
	"zumygo/helpers"
	"zumygo/libs"
)

type item struct {
	Name        []string
	IsPrefix    bool
	Description string
}

type tagSlice []string

func (t tagSlice) Len() int {
	return len(t)
}

func (t tagSlice) Less(i int, j int) bool {
	return t[i] < t[j]
}

func (t tagSlice) Swap(i int, j int) {
	t[i], t[j] = t[j], t[i]
}

// createCategoryMenu creates a professional menu for a specific category
func createCategoryMenu(conn *libs.IClient, m *libs.IMessage, category string) bool {
	var str strings.Builder
	
	// Professional header
	str.WriteString("🎯 *ZUMYGO BOT MENU*\n")
	str.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	str.WriteString(fmt.Sprintf("👤 User: %s\n", m.Info.PushName))
	str.WriteString(fmt.Sprintf("📱 Category: %s\n", strings.ToUpper(category)))
	str.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	
	// Get commands for the specific category
	var commands []item
	for _, list := range libs.GetList() {
		if strings.ToLower(list.Tags) == strings.ToLower(category) {
			commands = append(commands, item{
				Name:        list.As,
				IsPrefix:    list.IsPrefix,
				Description: list.Description,
			})
		}
	}
	
	if len(commands) == 0 {
		str.WriteString("❌ No commands found for this category.\n")
		str.WriteString("💡 Try: .menu to see all categories\n")
		m.Reply(str.String())
		return true
	}
	
	// Display commands with descriptions
	counter := 1
	for _, cmd := range commands {
		var prefix string
		if cmd.IsPrefix {
			prefix, _ = libs.ExtractPrefix(m.Body)
		}
		
		for _, name := range cmd.Name {
			description := cmd.Description
			if description == "" {
				description = "No description available"
			}
			
			str.WriteString(fmt.Sprintf("%d. *%s%s*\n", counter, prefix, name))
			str.WriteString(fmt.Sprintf("   └ %s\n\n", description))
			counter++
		}
	}
	
	// Footer with navigation
	str.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	str.WriteString("💡 *Navigation:*\n")
	str.WriteString("• .menu - Show main menu\n")
	str.WriteString("• .menu [category] - Show specific category\n")
	str.WriteString("• .help - Show detailed help\n")
	str.WriteString("• Example: .menu downloader\n\n")
	str.WriteString("🔧 Powered by ZUMYGO Bot")
	
	m.Reply(str.String())
	return true
}

// getAvailableCategories returns all available command categories
func getAvailableCategories() []string {
	categories := make(map[string]bool)
	
	for _, list := range libs.GetList() {
		if list.Tags != "" {
			categories[list.Tags] = true
		}
	}
	
	var result []string
	for category := range categories {
		result = append(result, category)
	}
	
	sort.Strings(result)
	return result
}

// getCategoryDisplayName returns a user-friendly name for categories
func getCategoryDisplayName(category string) string {
	displayNames := map[string]string{
		"main":       "🏠 Main",
		"downloader": "📥 Download",
		"owner":      "⚙️ Owner",
		"auto":       "🤖 Auto",
		"tools":      "🛠️ Tools",
		"fun":        "🎮 Fun",
		"info":       "ℹ️ Info",
	}
	
	if displayName, exists := displayNames[strings.ToLower(category)]; exists {
		return displayName
	}
	
	return helpers.CapitalizeWords(category)
}

// getCommandCount returns the number of commands in a category
func getCommandCount(category string) int {
	count := 0
	for _, list := range libs.GetList() {
		if strings.ToLower(list.Tags) == strings.ToLower(category) {
			count += len(list.As)
		}
	}
	return count
}

// mainMenu shows the main menu with category overview
func mainMenu(conn *libs.IClient, m *libs.IMessage) bool {
	var str strings.Builder
	
	// Professional header with emojis and styling
	str.WriteString("🎯 *ZUMYGO BOT - MAIN MENU*\n")
	str.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	str.WriteString(fmt.Sprintf("👤 Welcome, %s!\n", m.Info.PushName))
	str.WriteString("🤖 Your AI-powered WhatsApp assistant\n")
	str.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	
	// Get all categories
	categories := getAvailableCategories()
	
	// Display category overview with command counts
	str.WriteString("*📋 Available Categories:*\n\n")
	
	counter := 1
	for _, category := range categories {
		displayName := getCategoryDisplayName(category)
		commandCount := getCommandCount(category)
		
		str.WriteString(fmt.Sprintf("%d. %s\n", counter, displayName))
		str.WriteString(fmt.Sprintf("   └ %d commands available\n\n", commandCount))
		counter++
	}
	
	// Footer with instructions and status
	str.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	str.WriteString("💡 *How to use:*\n")
	str.WriteString("• .menu [category] - Explore specific category\n")
	str.WriteString("• .help - Show detailed help\n")
	str.WriteString("• Example: .menu downloader\n")
	str.WriteString("• Example: .menu owner\n\n")
	str.WriteString("🔧 *Bot Status:* Online ✅\n")
	str.WriteString("📊 *Version:* 2.0 Professional\n")
	str.WriteString("🌟 *Features:* Interactive & Fast")
	
	m.Reply(str.String())
	return true
}

// helpMenu shows detailed help with examples
func helpMenu(conn *libs.IClient, m *libs.IMessage) bool {
	var str strings.Builder
	
	// Professional header
	str.WriteString("📚 *ZUMYGO BOT - HELP & GUIDE*\n")
	str.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	str.WriteString(fmt.Sprintf("👤 User: %s\n", m.Info.PushName))
	str.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	
	// Quick start guide
	str.WriteString("*🚀 Quick Start Guide:*\n\n")
	str.WriteString("1. *Main Menu:* .menu\n")
	str.WriteString("2. *Category Menu:* .menu [category]\n")
	str.WriteString("3. *Detailed Help:* .help\n\n")
	
	// Popular commands
	str.WriteString("*🔥 Popular Commands:*\n\n")
	str.WriteString("📥 *Download Commands:*\n")
	str.WriteString("• .p [query] - Download YouTube MP3\n")
	str.WriteString("• .tt [url] - Download TikTok video\n")
	str.WriteString("• .yt [query] - Search YouTube\n\n")
	
	str.WriteString("🏠 *Main Commands:*\n")
	str.WriteString("• .menu - Show main menu\n")
	str.WriteString("• .ping - Check bot status\n")
	str.WriteString("• .perf - Show performance stats\n\n")
	
	// Category examples
	str.WriteString("*📂 Category Examples:*\n\n")
	categories := getAvailableCategories()
	for _, category := range categories {
		displayName := getCategoryDisplayName(category)
		str.WriteString(fmt.Sprintf("• .menu %s - %s\n", strings.ToLower(category), displayName))
	}
	
	// Footer
	str.WriteString("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	str.WriteString("💡 *Need more help?*\n")
	str.WriteString("• Contact the bot owner\n")
	str.WriteString("• Check .menu for all commands\n")
	str.WriteString("🔧 Powered by ZUMYGO Bot")
	
	m.Reply(str.String())
	return true
}

// menu handles both main menu and category-specific menus
func menu(conn *libs.IClient, m *libs.IMessage) bool {
	// Check if a specific category is requested
	if len(m.Args) > 0 {
		category := strings.Join(m.Args, " ")
		return createCategoryMenu(conn, m, category)
	}
	
	// Show main menu
	return mainMenu(conn, m)
}

// help command handler
func help(conn *libs.IClient, m *libs.IMessage) bool {
	return helpMenu(conn, m)
}

// createSimpleMenu creates a comprehensive menu with all commands (fallback)
func createSimpleMenu(conn *libs.IClient, m *libs.IMessage) bool {
	var str strings.Builder
	
	// Professional header
	str.WriteString("🎯 *ZUMYGO BOT - COMPLETE COMMAND LIST*\n")
	str.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	str.WriteString(fmt.Sprintf("👤 User: %s\n", m.Info.PushName))
	str.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	
	var tags map[string][]item
	for _, list := range libs.GetList() {
		if tags == nil {
			tags = make(map[string][]item)
		}
		if _, ok := tags[list.Tags]; !ok {
			tags[list.Tags] = []item{}
		}
		tags[list.Tags] = append(tags[list.Tags], item{
			Name:        list.As,
			IsPrefix:    list.IsPrefix,
			Description: list.Description,
		})
	}

	var keys tagSlice
	for key := range tags {
		if key == "" {
			continue
		} else {
			keys = append(keys, key)
		}
	}

	sort.Sort(keys)

	counter := 1
	for _, key := range keys {
		displayName := getCategoryDisplayName(key)
		str.WriteString(fmt.Sprintf("*%s*\n", displayName))
		for _, e := range tags[key] {
			var prefix string
			if e.IsPrefix {
				prefix, _ = libs.ExtractPrefix(m.Body)
			} else {
				prefix = ""
			}
			for _, nm := range e.Name {
				str.WriteString(fmt.Sprintf("%d. ```%s%s```\n", counter, prefix, nm))
				counter++
			}
		}
		str.WriteString("\n")
	}

	// Footer
	str.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	str.WriteString("💡 *Usage:* .menu [category]\n")
	str.WriteString("🔧 Powered by ZUMYGO Bot")

	m.Reply(str.String())
	return true
}

func init() {
	libs.NewCommands(&libs.ICommand{
		Name:        "menu",
		As:          []string{"menu"},
		Tags:        "main",
		IsPrefix:    true,
		Description: "Show bot menu with professional formatting",
		Execute:     menu,
	})
	
	libs.NewCommands(&libs.ICommand{
		Name:        "help",
		As:          []string{"help", "h", "?"},
		Tags:        "main",
		IsPrefix:    true,
		Description: "Show detailed help and usage guide",
		Execute:     help,
	})
}

