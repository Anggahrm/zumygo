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

			// Get downloader system from global systems with optimized retry mechanism
			downloaderSystem := systems.EnsureGlobalDownloaderSystem(500 * time.Millisecond) // Reduced from 2s to 500ms
			
			if downloaderSystem == nil {
				m.Reply("❎ Downloader system not available. Please try again.")
				return false
			}

			// Check if it's a YouTube URL or text query
			isYouTubeURL := strings.Contains(strings.ToLower(query), "youtube") || strings.Contains(strings.ToLower(query), "youtu.be")
			
			var downloadResult *systems.DownloadResult
			var downloadErr error
			var videoInfo *VideoInfo

			if isYouTubeURL {
				// For direct YouTube URLs, first check cache, then search if needed
				videoID := extractYouTubeID(query)
				videoInfo = getCachedVideoInfo(videoID)
				
				if videoInfo == nil {
					// If not in cache, try to find the video in search results
					searchResult, _ := downloaderSystem.SearchYouTubeByURL(query)
					if searchResult != nil {
						videoInfo = &VideoInfo{
							Title:     searchResult.Title,
							Duration:  searchResult.Duration,
							Views:     searchResult.Views, // Keep as int64
							Author:    searchResult.Author,
							Published: searchResult.Published,
							URL:       searchResult.URL,
						}
						// Cache the result
						cacheVideoInfo(videoID, videoInfo)
					}
				}
				
				// Download using the original URL
				downloadResult, downloadErr = downloaderSystem.DownloadMedia("youtube", query)
			} else {
				// Text query - search for the song first, then download
				// Check cache first using the query as key
				videoInfo = getCachedVideoInfo(query)
				
				if videoInfo == nil {
					// If not in cache, search for the song
					searchResult, searchErr := downloaderSystem.SearchYouTube(query)
					if searchErr != nil {
						m.Reply(fmt.Sprintf("❎ Gagal mencari video: %v", searchErr))
						return false
					}
					
					// Convert SearchResult to VideoInfo
					videoInfo = &VideoInfo{
						Title:     searchResult.Title,
						Duration:  searchResult.Duration,
						Views:     searchResult.Views, // Keep as int64
						Author:    searchResult.Author,
						Published: searchResult.Published,
						URL:       searchResult.URL,
					}
					
					// Cache the result using query as key
					cacheVideoInfo(query, videoInfo)
					
					// Also cache using video ID if available
					videoID := extractYouTubeID(searchResult.URL)
					if videoID != "" {
						cacheVideoInfo(videoID, videoInfo)
					}
				}
				
				// Download the found video
				downloadResult, downloadErr = downloaderSystem.DownloadMedia("youtube", videoInfo.URL)
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

			// Create caption with detailed information from search results
			var title, duration, views, author, published, videoId string
			
			if videoInfo != nil {
				// Use detailed info from search results
				title = videoInfo.Title
				duration = videoInfo.Duration
				views = downloaderSystem.FormatViews(videoInfo.Views)
				author = videoInfo.Author
				published = videoInfo.Published
				videoId = videoInfo.VideoID
				
				// If some fields are empty, try to get from download result
				if title == "" && downloadResult.Title != "" {
					title = downloadResult.Title
				}
				if duration == "" && downloadResult.Duration != "" {
					duration = downloadResult.Duration
				}
				if videoId == "" && downloadResult.ID != "" {
					videoId = downloadResult.ID
				}
			} else {
				// Use info from download result (for direct URLs)
				title = downloadResult.Title
				duration = downloadResult.Duration
				views = downloadResult.Views
				author = "Unknown"
				published = "Unknown"
				videoId = downloadResult.ID
			}
			
			// Fallback untuk field yang masih kosong
			if title == "" {
				title = "Unknown Title"
			}
			if duration == "" {
				duration = "Unknown"
			}
			if views == "" {
				views = "Unknown"
			}
			if author == "" {
				author = "Unknown"
			}
			if published == "" {
				published = "Unknown"
			}
			if videoId == "" {
				videoId = "Unknown"
			}
			
			caption := fmt.Sprintf(`*🎵 YT PLAY*

◦ VideoID : %s
◦ Title : %s
◦ Duration : %s
◦ Views : %s
◦ Author : %s
◦ Published : %s
◦ URL : %s`, videoId, title, duration, views, author, published, 
				fmt.Sprintf("https://youtu.be/%s", videoId))

			// Check if client is available
			if conn == nil {
				m.Reply("❎ Client not available for sending media")
				return false
			}

			// Send audio file as document
			_, err = conn.SendDocument(m.Info.Chat, audioData, fmt.Sprintf("%s.mp3", downloaderSystem.CleanFileName(title)), caption, nil)
			if err != nil {
				m.Reply("❎ Gagal mengirim audio document")
				return false
			}

			// Also send as audio message
			_, err = conn.SendAudio(m.Info.Chat, audioData, fmt.Sprintf("%s.mp3", downloaderSystem.CleanFileName(title)), nil)
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

// VideoInfo holds detailed information about a video from search results
type VideoInfo struct {
	VideoID     string `json:"videoId"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Thumbnail   string `json:"thumbnail"`
	Duration    string `json:"duration"`
	Published   string `json:"published_at"`
	Views       int64  `json:"views"`
	IsLive      bool   `json:"isLive"`
	Author      string `json:"author"`
	AuthorURL   string `json:"authorUrl"`
}



// Simple cache for video information
var (
	videoCache = make(map[string]*VideoInfo)
	cacheMutex sync.RWMutex
)

// cacheVideoInfo stores video information in cache
func cacheVideoInfo(videoID string, info *VideoInfo) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	videoCache[videoID] = info
}

// getCachedVideoInfo retrieves video information from cache
func getCachedVideoInfo(videoID string) *VideoInfo {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	return videoCache[videoID]
}





// extractYouTubeID extracts the video ID from a YouTube URL
func extractYouTubeID(url string) string {
	// YouTube URL patterns
	patterns := []string{
		`(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)([^&\n?#]+)`,
		`youtube\.com\/watch\?.*v=([^&\n?#]+)`,
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

