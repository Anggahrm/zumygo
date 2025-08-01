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
	"zumygo/systems"
)

func init() {
	libs.NewCommands(&libs.ICommand{
		Name:        "(play|p|song|ds|ytmp3|yta|mp3)",
		As:          []string{"play"},
		Tags:        "downloader",
		IsPrefix:    true,
		IsQuery:     true,
		Description: "Download YouTube videos as MP3 audio",
		Execute: func(conn *libs.IClient, m *libs.IMessage) bool {
			// Check if query is provided
			if len(m.Args) == 0 {
				m.Reply("Masukan URL atau judul lagu!\n\ncontoh:\n.play https://youtu.be/4rDOsvzTicY?si=3Ps-SJyRGzMa83QT\n.play despacito")
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

			// Check if it's a YouTube URL or text query
			isYouTubeURL := strings.Contains(strings.ToLower(query), "youtube") || strings.Contains(strings.ToLower(query), "youtu.be")
			
			var downloadResult *systems.DownloadResult
			var downloadErr error

			if isYouTubeURL {
				// Direct YouTube URL download using betabotz API
				downloadResult, downloadErr = downloaderSystem.DownloadMedia("youtube", query)
			} else {
				// Text query - search for the song first, then download
				videoURL, searchErr := searchAndGetFirstVideo(query)
				if searchErr != nil {
					m.Reply(fmt.Sprintf("❎ Gagal mencari video: %v", searchErr))
					return false
				}
				
				// Download the found video
				downloadResult, downloadErr = downloaderSystem.DownloadMedia("youtube", videoURL)
			}

			if downloadErr != nil {
				m.Reply("❎ Terjadi kesalahan saat mengunduh audio!")
				return false
			}

			if !downloadResult.Success {
				m.Reply("❎ " + downloadResult.Error)
				return false
			}

			// Download audio data
			audioData, err := conn.GetBytes(downloadResult.URL)
			if err != nil {
				m.Reply("❎ Gagal mengunduh data audio")
				return false
			}

			// Create caption
			caption := fmt.Sprintf(`*YT PLAY*

◦ id : %s
◦ title : %s
◦ duration : %s
◦ views : %s`, downloadResult.ID, downloadResult.Title, downloadResult.Duration, downloadResult.Views)

			// Send audio file as document
			_, err = conn.SendDocument(m.Info.Chat, audioData, fmt.Sprintf("%s.mp3", downloadResult.Title), caption, nil)
			if err != nil {
				m.Reply("❎ Gagal mengirim audio document")
				return false
			}

			// Also send as audio message
			_, err = conn.SendAudio(m.Info.Chat, audioData, fmt.Sprintf("%s.mp3", downloadResult.Title), nil)
			if err != nil {
				m.Reply("❎ Gagal mengirim audio message")
				return false
			}

			// Send success reaction
			m.React("✅")
			return true
		},
	})
}

// searchAndGetFirstVideo searches for a video and returns the first result URL
func searchAndGetFirstVideo(query string) (string, error) {
	// Get config for API key
	cfg := config.GetConfig()
	if cfg == nil {
		return "", fmt.Errorf("config not available")
	}

	// Build search API URL
	searchURL := fmt.Sprintf("https://api.betabotz.eu.org/api/search/yts?query=%s&apikey=%s", 
		url.QueryEscape(query), cfg.APIKeys["https://api.betabotz.eu.org"])

	// Make search API request
	resp, err := http.Get(searchURL)
	if err != nil {
		return "", fmt.Errorf("failed to search: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read search response: %v", err)
	}

	// Parse JSON response
	var searchResponse struct {
		Status bool `json:"status"`
		Result []struct {
			URL string `json:"url"`
		} `json:"result"`
	}
	
	if err := json.Unmarshal(bodyBytes, &searchResponse); err != nil {
		return "", fmt.Errorf("failed to parse search response: %v", err)
	}

	// Check if search was successful and has results
	if !searchResponse.Status || len(searchResponse.Result) == 0 {
		return "", fmt.Errorf("no search results found for: %s", query)
	}

	// Return the first video URL
	return searchResponse.Result[0].URL, nil
} 