package downloader

import (
	"fmt"
	"strings"
	"zumygo/libs"
	"zumygo/systems"
)



func init() {
	libs.NewCommands(&libs.ICommand{
		Name:        "(ytsearch|yts|search)",
		As:          []string{"ytsearch"},
		Tags:        "downloader",
		IsPrefix:    true,
		IsQuery:     true,
		Description: "Search YouTube videos",
		Execute: func(conn *libs.IClient, m *libs.IMessage) bool {
			// Check if query is provided
			if len(m.Args) == 0 {
				m.Reply("Masukan kata kunci pencarian!\n\ncontoh:\n.ytsearch despacito\n.ytsearch alex mica dalinda")
				return false
			}

			query := strings.Join(m.Args, " ")

			// Send processing reaction
			m.React("⏱️")

			// Get downloader system from global systems
			downloaderSystem := systems.GetGlobalDownloaderSystem()
			if downloaderSystem == nil {
				m.Reply("❎ Downloader system not available")
				return false
			}

			// Search for videos using the downloader system
			searchResults, err := downloaderSystem.SearchYouTubeMultiple(query)
			if err != nil {
				m.Reply(fmt.Sprintf("❎ Gagal melakukan pencarian: %v", err))
				return false
			}

			// Create search results message
			var results []string
			results = append(results, fmt.Sprintf("*🔍 YouTube Search Results*\n*Query:* %s\n", query))

			// Limit to first 10 results
			maxResults := 10
			if len(searchResults) < maxResults {
				maxResults = len(searchResults)
			}

			for i := 0; i < maxResults; i++ {
				result := searchResults[i]
				
				// Format duration and views
				duration := result.Duration
				if duration == "" {
					duration = "Unknown"
				}
				
				views := downloaderSystem.FormatViews(result.Views)
				if result.Views == 0 {
					views = "Unknown"
				}

				// Create result entry
				entry := fmt.Sprintf(`*%d.* %s
⏱️ *Duration:* %s
👁️ *Views:* %s
📅 *Published:* %s
👤 *Author:* %s
🔗 *URL:* %s

`, i+1, result.Title, duration, views, result.Published, result.Author, result.URL)
				
				results = append(results, entry)
			}

			// Join all results
			fullMessage := strings.Join(results, "")

			// Add footer
			fullMessage += fmt.Sprintf("\n*Total Results:* %d\n*Use .play <URL> to download audio*", len(searchResults))

			// Send the search results
			m.Reply(fullMessage)

			// Send success reaction
			m.React("✅")
			return true
		},
	})
} 