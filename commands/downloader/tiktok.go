package downloader

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
	"zumygo/libs"
	"zumygo/systems"
)

func init() {
	libs.NewCommands(&libs.ICommand{
		Name:        "(tiktok|ttdl|tiktokdl|tiktoknowm|tt)",
		As:          []string{"tiktok"},
		Tags:        "downloader",
		IsPrefix:    true,
		IsQuery:     true,
		Description: "Download TikTok videos without watermark",
		Execute: func(conn *libs.IClient, m *libs.IMessage) bool {
			// Check if URL is provided
			if len(m.Args) == 0 {
				m.Reply("✳️ Masukkan tautan Tiktok\n\n 📌 Contoh : .tiktok https://vm.tiktok.com/ZMYG92bUh/")
				return false
			}

			url := m.Args[0]

			// Validate TikTok URL
			if !strings.Contains(strings.ToLower(url), "tiktok") {
				m.Reply("❎ Verifikasi bahwa tautan tersebut berasal dari TikTok")
				return false
			}

			// Send processing reaction
			m.React("⏱️")

			// Get downloader system from global systems with optimized retry mechanism
			downloaderSystem := systems.EnsureGlobalDownloaderSystem(500 * time.Millisecond) // Reduced from 2s to 500ms
			
			if downloaderSystem == nil {
				m.Reply("❎ Downloader system not available. Please try again.")
				return false
			}

			// Check cache first
			tiktokID := extractTikTokID(url)
			tiktokInfo := getCachedTikTokInfo(tiktokID)
			
			// Download TikTok video
			result, err := downloaderSystem.DownloadMedia("tiktok", url)
			if err != nil {
				m.Reply("❎ Kesalahan mengunduh video")
				return false
			}

			if !result.Success {
				m.Reply("❎ " + result.Error)
				return false
			}

			// Cache the result if we have a TikTok ID
			if tiktokID != "" && tiktokInfo == nil {
				tiktokInfo = &TikTokInfo{
					VideoID:     tiktokID,
					URL:         url,
					Title:       result.Title,
					Description: result.Title,
				}
				cacheTikTokInfo(tiktokID, tiktokInfo)
			}

			// Download video data
			videoData, err := conn.GetBytes(result.URL)
			if err != nil {
				m.Reply("❎ Gagal mengunduh data video")
				return false
			}

			// Create caption with detailed information
			title := result.Title
			if title == "" {
				title = "TikTok Video"
			}
			
			caption := fmt.Sprintf(`┌─⊷ TIKTOK
▢ *Deskripsi:* %s
▢ *URL:* %s
└───────────`, title, url)

			// Check if client is available
			if conn == nil {
				m.Reply("❎ Client not available for sending media")
				return false
			}

			// Send video file
			_, err = conn.SendVideo(m.Info.Chat, videoData, caption, nil)
			if err != nil {
				m.Reply("❎ Gagal mengirim video")
				return false
			}

			// Send success reaction
			m.React("✅")
			return true
		},
	})
}

// TikTokInfo holds detailed information about a TikTok video
type TikTokInfo struct {
	VideoID     string `json:"videoId"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	Duration    string `json:"duration"`
	Published   string `json:"published_at"`
	Views       int64  `json:"views"`
	Author      string `json:"author"`
	AuthorURL   string `json:"authorUrl"`
}

// Simple cache for TikTok video information
var (
	tiktokCache = make(map[string]*TikTokInfo)
	tiktokCacheMutex sync.RWMutex
)

// cacheTikTokInfo stores TikTok video information in cache
func cacheTikTokInfo(videoID string, info *TikTokInfo) {
	tiktokCacheMutex.Lock()
	defer tiktokCacheMutex.Unlock()
	tiktokCache[videoID] = info
}

// getCachedTikTokInfo retrieves TikTok video information from cache
func getCachedTikTokInfo(videoID string) *TikTokInfo {
	tiktokCacheMutex.RLock()
	defer tiktokCacheMutex.RUnlock()
	return tiktokCache[videoID]
}

// extractTikTokID extracts the video ID from a TikTok URL
func extractTikTokID(url string) string {
	// TikTok URL patterns
	patterns := []string{
		`(?:tiktok\.com\/@[^\/]+\/video\/|vm\.tiktok\.com\/)([^&\n?#\/]+)`,
		`tiktok\.com\/@[^\/]+\/video\/([^&\n?#\/]+)`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
} 