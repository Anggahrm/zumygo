package downloader

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"zumygo/config"
	"zumygo/libs"
)

// YouTubeSearchResult represents a single search result
type YouTubeSearchResult struct {
	VideoID      string `json:"videoId"`
	URL          string `json:"url"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	Thumbnail    string `json:"thumbnail"`
	Duration     string `json:"duration"`
	PublishedAt  string `json:"published_at"`
	Views        int    `json:"views"`
	IsLive       bool   `json:"isLive"`
	Author       struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"author"`
}

// YouTubeSearchResponse represents the API response
type YouTubeSearchResponse struct {
	Status  bool                  `json:"status"`
	Creator string                `json:"creator"`
	Result  []YouTubeSearchResult `json:"result"`
}

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

			// Get config for API key
			cfg := config.GetConfig()
			if cfg == nil {
				m.Reply("❎ Config not available")
				return false
			}

			// Build API URL
			apiURL := fmt.Sprintf("https://api.betabotz.eu.org/api/search/yts?query=%s&apikey=%s", 
				url.QueryEscape(query), cfg.APIKeys["lann"])

			// Make API request
			resp, err := http.Get(apiURL)
			if err != nil {
				m.Reply(fmt.Sprintf("❎ Gagal melakukan pencarian: %v", err))
				return false
			}
			defer resp.Body.Close()

			// Read response body
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				m.Reply("❎ Gagal membaca response API")
				return false
			}



			// Parse JSON response
			var searchResponse YouTubeSearchResponse
			if err := json.Unmarshal(bodyBytes, &searchResponse); err != nil {
				m.Reply(fmt.Sprintf("❎ Gagal parse response: %v\nResponse: %s", err, string(bodyBytes)))
				return false
			}

			// Check if search was successful
			if !searchResponse.Status || len(searchResponse.Result) == 0 {
				m.Reply("❎ Tidak ada hasil ditemukan untuk: " + query)
				return false
			}

			// Create search results message
			var results []string
			results = append(results, fmt.Sprintf("*🔍 YouTube Search Results*\n*Query:* %s\n", query))

			// Limit to first 10 results
			maxResults := 10
			if len(searchResponse.Result) < maxResults {
				maxResults = len(searchResponse.Result)
			}

			for i := 0; i < maxResults; i++ {
				result := searchResponse.Result[i]
				
				// Format duration and views
				duration := result.Duration
				if duration == "" {
					duration = "Unknown"
				}
				
				views := fmt.Sprintf("%d", result.Views)
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

`, i+1, result.Title, duration, views, result.PublishedAt, result.Author.Name, result.URL)
				
				results = append(results, entry)
			}

			// Join all results
			fullMessage := strings.Join(results, "")

			// Add footer
			fullMessage += fmt.Sprintf("\n*Total Results:* %d\n*Use .play <URL> to download audio*", len(searchResponse.Result))

			// Send the search results
			m.Reply(fullMessage)

			// Send success reaction
			m.React("✅")
			return true
		},
	})
} 