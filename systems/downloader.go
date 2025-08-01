package systems

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"zumygo/config"
	"zumygo/database"
	"zumygo/helpers"
)

// DownloaderSystem handles media downloads from various platforms
type DownloaderSystem struct {
	cfg    *config.BotConfig
	db     *database.Database
	logger *helpers.Logger
}

// DownloadResult represents the result of a download operation
type DownloadResult struct {
	Success  bool   `json:"success"`
	URL      string `json:"url"`
	Title    string `json:"title"`
	Size     string `json:"size"`
	Type     string `json:"type"`
	ID       string `json:"id"`
	Duration string `json:"duration"`
	Views    string `json:"views"`
	Error    string `json:"error,omitempty"`
}

// VideoInfo represents video information
type VideoInfo struct {
	Title       string `json:"title"`
	Duration    string `json:"duration"`
	Quality     string `json:"quality"`
	Size        string `json:"size"`
	Thumbnail   string `json:"thumbnail"`
	DownloadURL string `json:"download_url"`
}

// InitializeDownloaderSystem creates a new downloader system
func InitializeDownloaderSystem(cfg *config.BotConfig, db *database.Database, logger *helpers.Logger) *DownloaderSystem {
	return &DownloaderSystem{
		cfg:    cfg,
		db:     db,
		logger: logger,
	}
}

// DownloadMedia handles downloading media from various platforms
func (ds *DownloaderSystem) DownloadMedia(platform, url string) (*DownloadResult, error) {
	ds.logger.Info(fmt.Sprintf("Starting download from %s: %s", platform, url))
	
	switch strings.ToLower(platform) {
	case "youtube", "yt":
		return ds.downloadYouTube(url)
	case "instagram", "ig":
		return ds.downloadInstagram(url)
	case "tiktok", "tt":
		return ds.downloadTikTok(url)
	case "facebook", "fb":
		return ds.downloadFacebook(url)
	case "twitter", "x":
		return ds.downloadTwitter(url)
	case "telegram":
		return ds.downloadTelegram(url)
	default:
		return ds.downloadGeneric(url)
	}
}

// downloadYouTube downloads YouTube videos/audio
func (ds *DownloaderSystem) downloadYouTube(url string) (*DownloadResult, error) {
	// Use betabotz API for audio download
	apiURL := fmt.Sprintf("https://api.betabotz.eu.org/api/download/ytmp3?url=%s&apikey=%s", 
		url, ds.cfg.APIKeys["lann"])
	
	ds.logger.Info(fmt.Sprintf("Calling API: %s", apiURL))
	
	resp, err := http.Get(apiURL)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to fetch video info: %v", err)
		ds.logger.Error(errorMsg)
		return &DownloadResult{Success: false, Error: errorMsg}, err
	}
	defer resp.Body.Close()
	
	// Read the full response body for debugging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to read response body: %v", err)
		ds.logger.Error(errorMsg)
		return &DownloadResult{Success: false, Error: errorMsg}, err
	}
	
	ds.logger.Info(fmt.Sprintf("API Response: %s", string(bodyBytes)))
	

	
	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		errorMsg := fmt.Sprintf("Failed to parse response: %v. Response: %s", err, string(bodyBytes))
		ds.logger.Error(errorMsg)
		return &DownloadResult{Success: false, Error: errorMsg}, err
	}
	
	// Check if success field exists and is true
	success, successExists := result["success"]
	if !successExists {
		// If success field doesn't exist, check for status field
		if status, statusExists := result["status"]; statusExists {
			if status == "success" {
				success = true
			} else {
				errorMsg := fmt.Sprintf("API returned status: %v", status)
				ds.logger.Error(errorMsg)
				return &DownloadResult{Success: false, Error: errorMsg}, nil
			}
		} else {
			errorMsg := "API response missing success/status field"
			ds.logger.Error(errorMsg)
			return &DownloadResult{Success: false, Error: errorMsg}, nil
		}
	}
	
	if success != true {
		errorMsg := "API returned error"
		if errStr, ok := result["error"].(string); ok {
			errorMsg = fmt.Sprintf("API Error: %s", errStr)
		} else if errStr, ok := result["message"].(string); ok {
			errorMsg = fmt.Sprintf("API Message: %s", errStr)
		} else {
			errorMsg = fmt.Sprintf("API returned error. Full response: %s", string(bodyBytes))
		}
		ds.logger.Error(errorMsg)
		return &DownloadResult{Success: false, Error: errorMsg}, nil
	}
	
	data := result["result"].(map[string]interface{})
	
	// Extract data from betabotz API response
	mp3URL := ""
	if mp3, ok := data["mp3"].(string); ok {
		mp3URL = mp3
	}
	
	title := ""
	if titleVal, ok := data["title"].(string); ok && titleVal != "" {
		title = titleVal
	} else {
		title = "Unknown Title"
	}
	
	id := ""
	if idVal, ok := data["id"].(string); ok {
		id = idVal
	}
	
	duration := ""
	if durationVal, ok := data["duration"].(string); ok && durationVal != "" {
		duration = durationVal
	} else {
		duration = "Unknown"
	}
	
	views := "Unknown" // betabotz API doesn't provide views
	
	return &DownloadResult{
		Success:  true,
		URL:      mp3URL,
		Title:    title,
		Type:     "audio",
		ID:       id,
		Duration: duration,
		Views:    views,
		Size:     fmt.Sprintf("ID: %s, Duration: %s, Views: %s", id, duration, views),
	}, nil
}

// downloadInstagram downloads Instagram posts
func (ds *DownloaderSystem) downloadInstagram(url string) (*DownloadResult, error) {
	apiURL := ds.cfg.API("tio", "/api/instagram", map[string]string{
		"url": url,
	})
	
	resp, err := http.Get(apiURL)
	if err != nil {
		return &DownloadResult{Success: false, Error: "Failed to fetch Instagram data"}, err
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &DownloadResult{Success: false, Error: "Failed to parse response"}, err
	}
	
	if result["status"] != "success" {
		return &DownloadResult{Success: false, Error: "API returned error"}, nil
	}
	
	data := result["result"].(map[string]interface{})
	downloadURL := data["url"].(string)
	mediaType := data["type"].(string)
	
	return &DownloadResult{
		Success: true,
		URL:     downloadURL,
		Type:    mediaType,
	}, nil
}

// downloadTikTok downloads TikTok videos
func (ds *DownloaderSystem) downloadTikTok(url string) (*DownloadResult, error) {
	apiURL := ds.cfg.API("lann", "/api/download/tiktok", map[string]string{
		"url": url,
	})
	
	resp, err := http.Get(apiURL)
	if err != nil {
		return &DownloadResult{Success: false, Error: "Failed to fetch TikTok data"}, err
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &DownloadResult{Success: false, Error: "Failed to parse response"}, err
	}
	
	if result["status"] != "success" {
		return &DownloadResult{Success: false, Error: "API returned error"}, nil
	}
	
	data := result["result"].(map[string]interface{})
	downloadURL := data["video"].(string)
	title := data["title"].(string)
	
	return &DownloadResult{
		Success: true,
		URL:     downloadURL,
		Title:   title,
		Type:    "video",
	}, nil
}

// downloadFacebook downloads Facebook videos
func (ds *DownloaderSystem) downloadFacebook(url string) (*DownloadResult, error) {
	apiURL := ds.cfg.API("tio", "/api/facebook", map[string]string{
		"url": url,
	})
	
	resp, err := http.Get(apiURL)
	if err != nil {
		return &DownloadResult{Success: false, Error: "Failed to fetch Facebook data"}, err
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &DownloadResult{Success: false, Error: "Failed to parse response"}, err
	}
	
	if result["status"] != "success" {
		return &DownloadResult{Success: false, Error: "API returned error"}, nil
	}
	
	data := result["result"].(map[string]interface{})
	downloadURL := data["url"].(string)
	
	return &DownloadResult{
		Success: true,
		URL:     downloadURL,
		Type:    "video",
	}, nil
}

// downloadTwitter downloads Twitter/X videos
func (ds *DownloaderSystem) downloadTwitter(url string) (*DownloadResult, error) {
	apiURL := ds.cfg.API("tio", "/api/twitter", map[string]string{
		"url": url,
	})
	
	resp, err := http.Get(apiURL)
	if err != nil {
		return &DownloadResult{Success: false, Error: "Failed to fetch Twitter data"}, err
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &DownloadResult{Success: false, Error: "Failed to parse response"}, err
	}
	
	if result["status"] != "success" {
		return &DownloadResult{Success: false, Error: "API returned error"}, nil
	}
	
	data := result["result"].(map[string]interface{})
	downloadURL := data["url"].(string)
	
	return &DownloadResult{
		Success: true,
		URL:     downloadURL,
		Type:    "video",
	}, nil
}

// downloadTelegram downloads Telegram media
func (ds *DownloaderSystem) downloadTelegram(url string) (*DownloadResult, error) {
	apiURL := ds.cfg.API("tio", "/api/telegram", map[string]string{
		"url": url,
	})
	
	resp, err := http.Get(apiURL)
	if err != nil {
		return &DownloadResult{Success: false, Error: "Failed to fetch Telegram data"}, err
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &DownloadResult{Success: false, Error: "Failed to parse response"}, err
	}
	
	if result["status"] != "success" {
		return &DownloadResult{Success: false, Error: "API returned error"}, nil
	}
	
	data := result["result"].(map[string]interface{})
	downloadURL := data["url"].(string)
	
	return &DownloadResult{
		Success: true,
		URL:     downloadURL,
		Type:    "media",
	}, nil
}

// downloadGeneric handles generic URL downloads
func (ds *DownloaderSystem) downloadGeneric(url string) (*DownloadResult, error) {
	// Check if URL is valid
	if !ds.isValidURL(url) {
		return &DownloadResult{Success: false, Error: "Invalid URL"}, nil
	}
	
	// Try to get file info
	resp, err := http.Head(url)
	if err != nil {
		return &DownloadResult{Success: false, Error: "Failed to access URL"}, err
	}
	defer resp.Body.Close()
	
	contentType := resp.Header.Get("Content-Type")
	size := resp.Header.Get("Content-Length")
	
	return &DownloadResult{
		Success: true,
		URL:     url,
		Size:    size,
		Type:    contentType,
	}, nil
}

// DownloadFile downloads a file from URL to local storage
func (ds *DownloaderSystem) DownloadFile(downloadURL, filename string) error {
	// Create downloads directory if it doesn't exist
	downloadsDir := "downloads"
	if err := os.MkdirAll(downloadsDir, 0755); err != nil {
		return err
	}
	
	// Generate filename if not provided
	if filename == "" {
		filename = fmt.Sprintf("download_%d", time.Now().Unix())
	}
	
	filepath := filepath.Join(downloadsDir, filename)
	
	// Download the file
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	_, err = io.Copy(file, resp.Body)
	return err
}

// GetVideoInfo gets information about a video
func (ds *DownloaderSystem) GetVideoInfo(url string) (*VideoInfo, error) {
	// Extract platform and get info
	platform := ds.detectPlatform(url)
	
	switch platform {
	case "youtube":
		return ds.getYouTubeInfo(url)
	case "tiktok":
		return ds.getTikTokInfo(url)
	default:
		return &VideoInfo{
			Title: "Unknown Video",
		}, nil
	}
}

// getYouTubeInfo gets YouTube video information
func (ds *DownloaderSystem) getYouTubeInfo(url string) (*VideoInfo, error) {
	apiURL := ds.cfg.API("tio", "/api/youtube/info", map[string]string{
		"url": url,
	})
	
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	if result["status"] != "success" {
		return nil, fmt.Errorf("API returned error")
	}
	
	data := result["result"].(map[string]interface{})
	
	return &VideoInfo{
		Title:       data["title"].(string),
		Duration:    data["duration"].(string),
		Quality:     data["quality"].(string),
		Size:        data["size"].(string),
		Thumbnail:   data["thumbnail"].(string),
		DownloadURL: data["url"].(string),
	}, nil
}

// getTikTokInfo gets TikTok video information
func (ds *DownloaderSystem) getTikTokInfo(url string) (*VideoInfo, error) {
	apiURL := ds.cfg.API("lann", "/api/download/tiktok", map[string]string{
		"url": url,
	})
	
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	if result["status"] != "success" {
		return nil, fmt.Errorf("API returned error")
	}
	
	data := result["result"].(map[string]interface{})
	
	return &VideoInfo{
		Title:       data["title"].(string),
		Duration:    data["duration"].(string),
		Quality:     data["quality"].(string),
		Size:        data["size"].(string),
		Thumbnail:   data["cover"].(string),
		DownloadURL: data["video"].(string),
	}, nil
}

// Helper functions
func (ds *DownloaderSystem) extractYouTubeID(url string) string {
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

func (ds *DownloaderSystem) detectPlatform(url string) string {
	url = strings.ToLower(url)
	
	if strings.Contains(url, "youtube.com") || strings.Contains(url, "youtu.be") {
		return "youtube"
	} else if strings.Contains(url, "instagram.com") {
		return "instagram"
	} else if strings.Contains(url, "tiktok.com") {
		return "tiktok"
	} else if strings.Contains(url, "facebook.com") || strings.Contains(url, "fb.com") {
		return "facebook"
	} else if strings.Contains(url, "twitter.com") || strings.Contains(url, "x.com") {
		return "twitter"
	} else if strings.Contains(url, "t.me") {
		return "telegram"
	}
	
	return "generic"
}

func (ds *DownloaderSystem) isValidURL(urlString string) bool {
	_, err := url.Parse(urlString)
	return err == nil
}

// GetSupportedPlatforms returns list of supported platforms
func (ds *DownloaderSystem) GetSupportedPlatforms() []string {
	return []string{
		"YouTube (yt)",
		"Instagram (ig)",
		"TikTok (tt)",
		"Facebook (fb)",
		"Twitter/X (x)",
		"Telegram (tg)",
		"Generic URLs",
	}
}

// Global downloader system instance
var globalDownloaderSystem *DownloaderSystem

// SetGlobalDownloaderSystem sets the global downloader system instance
func SetGlobalDownloaderSystem(ds *DownloaderSystem) {
	globalDownloaderSystem = ds
}

// GetGlobalDownloaderSystem returns the global downloader system instance
func GetGlobalDownloaderSystem() *DownloaderSystem {
	return globalDownloaderSystem
} 