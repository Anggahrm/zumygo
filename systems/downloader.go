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
	"sync"
	"context"
	"zumygo/config"
	"zumygo/database"
	"zumygo/helpers"
)

// DownloaderSystem handles media downloads from various platforms
type DownloaderSystem struct {
	cfg    *config.BotConfig
	db     *database.Database
	logger *helpers.Logger
	
	// Performance optimizations
	httpClient *http.Client
	cache      map[string]*DownloadResult
	cacheMutex sync.RWMutex
	cacheTTL   time.Duration
}

// DownloadResult represents the result of a download operation
type DownloadResult struct {
	Success   bool     `json:"success"`
	URL       string   `json:"url"`
	URLs      []string `json:"urls,omitempty"`      // For multiple videos (slides)
	AudioURLs []string `json:"audio_urls,omitempty"` // For audio files in slides
	Title     string   `json:"title"`
	Size      string   `json:"size"`
	Type      string   `json:"type"`
	ID        string   `json:"id"`
	Duration  string   `json:"duration"`
	Views     string   `json:"views"`
	Error     string   `json:"error,omitempty"`
	CachedAt  time.Time `json:"cached_at,omitempty"`
	IsSlide   bool      `json:"is_slide,omitempty"` // Indicates if this is a slide/video collection
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

// SearchResult represents a YouTube search result
type SearchResult struct {
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

// SearchResponse represents the API response structure
type SearchResponse struct {
	Status  bool `json:"status"`
	Creator string `json:"creator"`
	Result  []struct {
		VideoID     string `json:"videoId"`
		URL         string `json:"url"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Thumbnail   string `json:"thumbnail"`
		Duration    string `json:"duration"`
		Published   string `json:"published_at"`
		Views       int64  `json:"views"`
		IsLive      bool   `json:"isLive"`
		Author      struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"author"`
	} `json:"result"`
}

// InitializeDownloaderSystem creates a new downloader system with performance optimizations
func InitializeDownloaderSystem(cfg *config.BotConfig, db *database.Database, logger *helpers.Logger) *DownloaderSystem {
	// Create optimized HTTP client with connection pooling and generous timeouts
	transport := &http.Transport{
		MaxIdleConns:        200,                    // Increased from 100
		MaxIdleConnsPerHost: 20,                     // Increased from 10
		IdleConnTimeout:     60 * time.Second,       // Reduced from 90s
		DisableCompression:  false,
		DisableKeepAlives:   false,
		ForceAttemptHTTP2:   true,                   // Enable HTTP/2 for better performance
		MaxConnsPerHost:     50,                     // Limit connections per host
		ResponseHeaderTimeout: 45 * time.Second,     // Increased from 25s to 45s for very slow APIs
		ExpectContinueTimeout: 15 * time.Second,     // Increased from 10s to 15s
	}
	
	client := &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second, // Increased from 30s to 60s for better reliability
	}
	
	return &DownloaderSystem{
		cfg:        cfg,
		db:         db,
		logger:     logger,
		httpClient: client,
		cache:      make(map[string]*DownloadResult),
		cacheTTL:   15 * time.Minute, // Increased from 10min to 15min for better cache hit rate
	}
}

// getCachedResult retrieves cached download result
func (ds *DownloaderSystem) getCachedResult(key string) (*DownloadResult, bool) {
	ds.cacheMutex.RLock()
	defer ds.cacheMutex.RUnlock()
	
	if result, exists := ds.cache[key]; exists {
		if time.Since(result.CachedAt) < ds.cacheTTL {
			return result, true
		}
		// Remove expired cache entry
		delete(ds.cache, key)
	}
	return nil, false
}

// setCachedResult stores download result in cache
func (ds *DownloaderSystem) setCachedResult(key string, result *DownloadResult) {
	ds.cacheMutex.Lock()
	defer ds.cacheMutex.Unlock()
	
	result.CachedAt = time.Now()
	ds.cache[key] = result
	
	// Optimized cleanup: only cleanup if cache gets very large
	if len(ds.cache) > 2000 { // Increased from 1000
		go ds.cleanupCache() // Run cleanup in background goroutine
	}
}

// cleanupCache removes old cache entries
func (ds *DownloaderSystem) cleanupCache() {
	ds.cacheMutex.Lock()
	defer ds.cacheMutex.Unlock()
	
	cutoff := time.Now().Add(-ds.cacheTTL)
	removed := 0
	
	for key, result := range ds.cache {
		if result.CachedAt.Before(cutoff) {
			delete(ds.cache, key)
			removed++
		}
	}
	
	// Log cleanup stats for monitoring
	if removed > 0 && ds.logger != nil {
		ds.logger.Info(fmt.Sprintf("Cache cleanup: removed %d expired entries, remaining: %d", removed, len(ds.cache)))
	}
}

// DownloadMedia handles downloading media from various platforms with caching
func (ds *DownloaderSystem) DownloadMedia(platform, url string) (*DownloadResult, error) {
	// Add nil checks for safety
	if ds == nil {
		return &DownloadResult{Success: false, Error: "Downloader system is nil"}, fmt.Errorf("downloader system is nil")
	}
	
	if ds.logger == nil {
		return &DownloadResult{Success: false, Error: "Logger is not initialized"}, fmt.Errorf("logger is not initialized")
	}
	
	if ds.httpClient == nil {
		return &DownloadResult{Success: false, Error: "HTTP client is not initialized"}, fmt.Errorf("http client is not initialized")
	}
	
	// Check cache first (fast path)
	cacheKey := fmt.Sprintf("%s:%s", platform, url)
	if cached, exists := ds.getCachedResult(cacheKey); exists {
		ds.logger.Info(fmt.Sprintf("Cache hit for %s: %s", platform, url))
		return cached, nil
	}
	
	ds.logger.Info(fmt.Sprintf("Starting download from %s: %s", platform, url))
	
	var result *DownloadResult
	var err error
	
	// Use context with timeout for faster failure detection
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second) // Slightly less than client timeout
	defer cancel()
	
	// Create channel for concurrent processing
	resultChan := make(chan *DownloadResult, 1)
	errChan := make(chan error, 1)
	
	// Start download in goroutine
	go func() {
		var downloadResult *DownloadResult
		var downloadErr error
		
		switch strings.ToLower(platform) {
		case "youtube", "yt":
			downloadResult, downloadErr = ds.downloadYouTube(url)
		case "instagram", "ig":
			downloadResult, downloadErr = ds.downloadInstagram(url)
		case "tiktok", "tt":
			downloadResult, downloadErr = ds.downloadTikTok(url)
		case "facebook", "fb":
			downloadResult, downloadErr = ds.downloadFacebook(url)
		case "twitter", "x":
			downloadResult, downloadErr = ds.downloadTwitter(url)
		case "telegram":
			downloadResult, downloadErr = ds.downloadTelegram(url)
		default:
			downloadResult, downloadErr = ds.downloadGeneric(url)
		}
		
		select {
		case resultChan <- downloadResult:
		case <-ctx.Done():
			return
		}
		
		select {
		case errChan <- downloadErr:
		case <-ctx.Done():
			return
		}
	}()
	
	// Wait for result or timeout
	select {
	case result = <-resultChan:
		err = <-errChan
	case err = <-errChan:
		result = <-resultChan
	case <-ctx.Done():
		return &DownloadResult{Success: false, Error: "Download timeout"}, fmt.Errorf("download timeout")
	}
	
	// Cache the result
	if result != nil {
		ds.setCachedResult(cacheKey, result)
	}
	
	return result, err
}

// downloadYouTube downloads YouTube videos/audio with optimized HTTP client
func (ds *DownloaderSystem) downloadYouTube(videoURL string) (*DownloadResult, error) {
	// Use betabotz API for audio download
	encodedURL := url.QueryEscape(videoURL)
	apiURL := fmt.Sprintf("https://api.betabotz.eu.org/api/download/ytmp3?url=%s&apikey=%s", 
		encodedURL, ds.cfg.APIKeys["https://api.betabotz.eu.org"])
	
	ds.logger.Info(fmt.Sprintf("Calling API: %s", apiURL))
	
	// Create request with browser-like headers
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to create request: %v", err)
		ds.logger.Error(errorMsg)
		return &DownloadResult{Success: false, Error: errorMsg}, err
	}
	
	// Add browser-like headers to avoid Cloudflare detection
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	
	// Make API request with optimized client
	resp, err := ds.httpClient.Do(req)
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
	status, statusExists := result["status"]
	
	// Determine if the request was successful
	successBool := false
	
	// Check success field first
	if successExists {
		if success == true {
			successBool = true
		} else if successStr, ok := success.(string); ok && successStr == "true" {
			successBool = true
		}
	}
	
	// If success field doesn't exist or is false, check status field
	if !successBool && statusExists {
		if status == true {
			successBool = true
		} else if statusStr, ok := status.(string); ok && statusStr == "true" {
			successBool = true
		}
	}
	
	if !successBool {
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
	
	// Extract result data safely
	resultData, ok := result["result"].(map[string]interface{})
	if !ok {
		errorMsg := "API response missing result data"
		ds.logger.Error(errorMsg)
		return &DownloadResult{Success: false, Error: errorMsg}, nil
	}
	
	// Extract data from betabotz API response
	mp3URL := ""
	if mp3, ok := resultData["mp3"].(string); ok && mp3 != "" {
		mp3URL = mp3
	}
	
	title := ""
	if titleVal, ok := resultData["title"].(string); ok && titleVal != "" {
		title = titleVal
	} else {
		title = "Unknown Title"
	}
	
	id := ""
	if idVal, ok := resultData["id"].(string); ok && idVal != "" {
		id = idVal
	}
	
	duration := ""
	if durationVal, ok := resultData["duration"].(string); ok && durationVal != "" {
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

// downloadInstagram downloads Instagram posts with optimized client
func (ds *DownloaderSystem) downloadInstagram(url string) (*DownloadResult, error) {
	apiURL := ds.cfg.API("tio", "/api/instagram", map[string]string{
		"url": url,
	})
	
	resp, err := ds.httpClient.Get(apiURL)
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

// downloadTikTok downloads TikTok videos with optimized client
func (ds *DownloaderSystem) downloadTikTok(tiktokURL string) (*DownloadResult, error) {
	// Use betabotz API for TikTok download
	encodedURL := url.QueryEscape(tiktokURL)
	apiURL := fmt.Sprintf("https://api.betabotz.eu.org/api/download/tiktok?url=%s&apikey=%s", 
		encodedURL, ds.cfg.APIKeys["https://api.betabotz.eu.org"])
	
	ds.logger.Info(fmt.Sprintf("Calling TikTok API: %s", apiURL))
	
	// Create request with browser-like headers
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to create request: %v", err)
		ds.logger.Error(errorMsg)
		return &DownloadResult{Success: false, Error: errorMsg}, err
	}
	
	// Add browser-like headers to avoid Cloudflare detection
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	
	// Make API request with optimized client
	resp, err := ds.httpClient.Do(req)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to fetch TikTok data: %v", err)
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
	
	ds.logger.Info(fmt.Sprintf("TikTok API Response: %s", string(bodyBytes)))
	
	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		errorMsg := fmt.Sprintf("Failed to parse response: %v. Response: %s", err, string(bodyBytes))
		ds.logger.Error(errorMsg)
		return &DownloadResult{Success: false, Error: errorMsg}, err
	}
	
	// Check if success field exists and is true
	success, successExists := result["success"]
	status, statusExists := result["status"]
	
	// Determine if the request was successful
	successBool := false
	
	// Check success field first
	if successExists {
		if success == true {
			successBool = true
		} else if successStr, ok := success.(string); ok && successStr == "true" {
			successBool = true
		}
	}
	
	// If success field doesn't exist or is false, check status field
	if !successBool && statusExists {
		if status == true {
			successBool = true
		} else if statusStr, ok := status.(string); ok && statusStr == "true" {
			successBool = true
		}
	}
	
	if !successBool {
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
	
	// Extract result data safely
	resultData, ok := result["result"].(map[string]interface{})
	if !ok {
		errorMsg := "API response missing result data"
		ds.logger.Error(errorMsg)
		return &DownloadResult{Success: false, Error: errorMsg}, nil
	}
	
	// Log video and audio counts for debugging
	if videoArray, ok := resultData["video"].([]interface{}); ok {
		ds.logger.Info(fmt.Sprintf("Found %d video URLs in response", len(videoArray)))
	}
	if audioArray, ok := resultData["audio"].([]interface{}); ok {
		ds.logger.Info(fmt.Sprintf("Found %d audio URLs in response", len(audioArray)))
	}
	
	// Extract data from betabotz API response
	videoURLs := []string{}
	isSlide := false
	
	// Handle video array (for slides)
	if videoArray, ok := resultData["video"].([]interface{}); ok && len(videoArray) > 0 {
		isSlide = len(videoArray) > 1
		for _, videoItem := range videoArray {
			if videoURL, ok := videoItem.(string); ok && videoURL != "" {
				videoURLs = append(videoURLs, videoURL)
			}
		}
	} else if video, ok := resultData["video"].(string); ok && video != "" {
		// Fallback for single string
		videoURLs = append(videoURLs, video)
	}
	
	// Handle audio array (for slides with audio)
	audioURLs := []string{}
	if audioArray, ok := resultData["audio"].([]interface{}); ok && len(audioArray) > 0 {
		for _, audioItem := range audioArray {
			if audioURL, ok := audioItem.(string); ok && audioURL != "" {
				audioURLs = append(audioURLs, audioURL)
			}
		}
	} else if audio, ok := resultData["audio"].(string); ok && audio != "" {
		// Fallback for single string
		audioURLs = append(audioURLs, audio)
	}
	
	title := ""
	if titleVal, ok := resultData["title"].(string); ok && titleVal != "" {
		title = titleVal
	} else {
		title = "TikTok Video"
	}
	
	// If no videos found, return error
	if len(videoURLs) == 0 {
		return &DownloadResult{
			Success: false,
			Error:   "No video URLs found in response",
		}, nil
	}
	
	// For slides, return all URLs
	if isSlide {
		return &DownloadResult{
			Success:   true,
			URLs:      videoURLs,
			AudioURLs: audioURLs,
			Title:     title,
			Type:      "video",
			IsSlide:   true,
		}, nil
	}
	
	// For single video, return first URL in URL field for backward compatibility
	return &DownloadResult{
		Success:   true,
		URL:       videoURLs[0],
		URLs:      videoURLs,
		AudioURLs: audioURLs,
		Title:     title,
		Type:      "video",
		IsSlide:   false,
	}, nil
}

// downloadFacebook downloads Facebook videos with optimized client
func (ds *DownloaderSystem) downloadFacebook(url string) (*DownloadResult, error) {
	apiURL := ds.cfg.API("tio", "/api/facebook", map[string]string{
		"url": url,
	})
	
	resp, err := ds.httpClient.Get(apiURL)
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

// downloadTwitter downloads Twitter/X videos with optimized client
func (ds *DownloaderSystem) downloadTwitter(url string) (*DownloadResult, error) {
	apiURL := ds.cfg.API("tio", "/api/twitter", map[string]string{
		"url": url,
	})
	
	resp, err := ds.httpClient.Get(apiURL)
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

// downloadTelegram downloads Telegram media with optimized client
func (ds *DownloaderSystem) downloadTelegram(url string) (*DownloadResult, error) {
	apiURL := ds.cfg.API("tio", "/api/telegram", map[string]string{
		"url": url,
	})
	
	resp, err := ds.httpClient.Get(apiURL)
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

// downloadGeneric handles generic URL downloads with optimized client
func (ds *DownloaderSystem) downloadGeneric(url string) (*DownloadResult, error) {
	// Check if URL is valid
	if !ds.isValidURL(url) {
		return &DownloadResult{Success: false, Error: "Invalid URL"}, nil
	}
	
	// Try to get file info with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return &DownloadResult{Success: false, Error: "Failed to create request"}, err
	}
	
	resp, err := ds.httpClient.Do(req)
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

// DownloadFile downloads a file from URL to local storage with progress tracking
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
	
	// Download the file with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return err
	}
	
	resp, err := ds.httpClient.Do(req)
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

// GetVideoInfo gets information about a video with caching
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

// SearchYouTube searches for YouTube videos using betabotz API and returns first result with caching
func (ds *DownloaderSystem) SearchYouTube(query string) (*SearchResult, error) {
	// Add nil checks for safety
	if ds == nil {
		return nil, fmt.Errorf("downloader system is nil")
	}
	
	if ds.cfg == nil {
		return nil, fmt.Errorf("configuration is not initialized")
	}
	
	if ds.logger == nil {
		return nil, fmt.Errorf("logger is not initialized")
	}
	
	if ds.httpClient == nil {
		return nil, fmt.Errorf("http client is not initialized")
	}
	
	// Check cache first (fast path)
	cacheKey := fmt.Sprintf("search:%s", query)
	if cached, exists := ds.getCachedResult(cacheKey); exists {
		// Convert cached result to SearchResult
		return &SearchResult{
			VideoID:     cached.ID,
			URL:         cached.URL,
			Title:       cached.Title,
			Description: "",
			Thumbnail:   "",
			Duration:    cached.Duration,
			Published:   "",
			Views:       0,
			IsLive:      false,
			Author:      "",
			AuthorURL:   "",
		}, nil
	}
	
	// Use context with timeout for faster failure detection
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	
	// Build search API URL
	searchURL := fmt.Sprintf("https://api.betabotz.eu.org/api/search/yts?query=%s&apikey=%s", 
		url.QueryEscape(query), ds.cfg.APIKeys["https://api.betabotz.eu.org"])

	// Create request with browser-like headers
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	// Add browser-like headers to avoid Cloudflare detection
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	
	// Make search API request with optimized client
	resp, err := ds.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %v", err)
	}
	defer resp.Body.Close()

	// Read response body with timeout
	bodyChan := make(chan []byte, 1)
	errChan := make(chan error, 1)
	
	go func() {
		bodyBytes, err := io.ReadAll(resp.Body)
		select {
		case bodyChan <- bodyBytes:
		case <-ctx.Done():
			return
		}
		select {
		case errChan <- err:
		case <-ctx.Done():
			return
		}
	}()
	
	var bodyBytes []byte
	select {
	case bodyBytes = <-bodyChan:
		err = <-errChan
	case err = <-errChan:
		bodyBytes = <-bodyChan
	case <-ctx.Done():
		return nil, fmt.Errorf("search timeout")
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to read search response: %v", err)
	}

	// Parse JSON response
	var searchResponse SearchResponse
	if err := json.Unmarshal(bodyBytes, &searchResponse); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %v", err)
	}

	// Check if search was successful and has results
	if !searchResponse.Status || len(searchResponse.Result) == 0 {
		return nil, fmt.Errorf("no search results found for: %s", query)
	}

	// Return the first result
	firstResult := searchResponse.Result[0]
	
	searchResult := &SearchResult{
		VideoID:     firstResult.VideoID,
		URL:         firstResult.URL,
		Title:       firstResult.Title,
		Description: firstResult.Description,
		Thumbnail:   firstResult.Thumbnail,
		Duration:    firstResult.Duration,
		Published:   firstResult.Published,
		Views:       firstResult.Views,
		IsLive:      firstResult.IsLive,
		Author:      firstResult.Author.Name,
		AuthorURL:   firstResult.Author.URL,
	}
	
	// Cache the result
	cachedResult := &DownloadResult{
		Success:  true,
		URL:      searchResult.URL,
		Title:    searchResult.Title,
		ID:       searchResult.VideoID,
		Duration: searchResult.Duration,
	}
	ds.setCachedResult(cacheKey, cachedResult)
	
	return searchResult, nil
}

// SearchYouTubeMultiple searches for YouTube videos and returns multiple results
func (ds *DownloaderSystem) SearchYouTubeMultiple(query string) ([]*SearchResult, error) {
	// Build search API URL
	searchURL := fmt.Sprintf("https://api.betabotz.eu.org/api/search/yts?query=%s&apikey=%s", 
		url.QueryEscape(query), ds.cfg.APIKeys["https://api.betabotz.eu.org"])

	// Create request with browser-like headers
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	// Add browser-like headers to avoid Cloudflare detection
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	
	// Make search API request with optimized client
	resp, err := ds.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read search response: %v", err)
	}

	// Parse JSON response
	var searchResponse SearchResponse
	if err := json.Unmarshal(bodyBytes, &searchResponse); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %v", err)
	}

	// Check if search was successful and has results
	if !searchResponse.Status || len(searchResponse.Result) == 0 {
		return nil, fmt.Errorf("no search results found for: %s", query)
	}

	// Convert all results to SearchResult
	var results []*SearchResult
	for _, result := range searchResponse.Result {
		searchResult := &SearchResult{
			VideoID:     result.VideoID,
			URL:         result.URL,
			Title:       result.Title,
			Description: result.Description,
			Thumbnail:   result.Thumbnail,
			Duration:    result.Duration,
			Published:   result.Published,
			Views:       result.Views,
			IsLive:      result.IsLive,
			Author:      result.Author.Name,
			AuthorURL:   result.Author.URL,
		}
		results = append(results, searchResult)
	}
	
	return results, nil
}

// SearchYouTubeByURL searches for a specific video by URL
func (ds *DownloaderSystem) SearchYouTubeByURL(targetURL string) (*SearchResult, error) {
	// Extract video ID from the target URL
	targetVideoID := ds.extractYouTubeID(targetURL)
	if targetVideoID == "" {
		return nil, fmt.Errorf("could not extract video ID from URL")
	}

	// Try different search queries to find the video
	searchQueries := []string{
		"video " + targetVideoID,
		"youtube " + targetVideoID,
		targetVideoID,
		"music",
		"song",
	}

	for _, query := range searchQueries {
		// Build search API URL
		searchURL := fmt.Sprintf("https://api.betabotz.eu.org/api/search/yts?query=%s&apikey=%s", 
			url.QueryEscape(query), ds.cfg.APIKeys["https://api.betabotz.eu.org"])

		// Create request with browser-like headers
		req, err := http.NewRequest("GET", searchURL, nil)
		if err != nil {
			continue
		}
		
		// Add browser-like headers
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "application/json, text/plain, */*")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Sec-Fetch-Dest", "empty")
		req.Header.Set("Sec-Fetch-Mode", "cors")
		req.Header.Set("Sec-Fetch-Site", "same-origin")
		
		// Make search API request with optimized client
		resp, err := ds.httpClient.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		// Read response body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		// Parse JSON response
		var searchResponse SearchResponse
		if err := json.Unmarshal(bodyBytes, &searchResponse); err != nil {
			continue
		}

		// Check if search was successful and has results
		if !searchResponse.Status || len(searchResponse.Result) == 0 {
			continue
		}

		// Look through results to find a video that matches our target URL
		for _, result := range searchResponse.Result {
			if result.VideoID == targetVideoID {
				// Found the matching video!
				return &SearchResult{
					VideoID:     result.VideoID,
					URL:         result.URL,
					Title:       result.Title,
					Description: result.Description,
					Thumbnail:   result.Thumbnail,
					Duration:    result.Duration,
					Published:   result.Published,
					Views:       result.Views,
					IsLive:      result.IsLive,
					Author:      result.Author.Name,
					AuthorURL:   result.Author.URL,
				}, nil
			}
		}
	}

	// If we couldn't find the exact video, return error
	return nil, fmt.Errorf("could not find matching video")
}

// getYouTubeInfo gets YouTube video information with optimized client
func (ds *DownloaderSystem) getYouTubeInfo(url string) (*VideoInfo, error) {
	apiURL := ds.cfg.API("tio", "/api/youtube/info", map[string]string{
		"url": url,
	})
	
	resp, err := ds.httpClient.Get(apiURL)
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

// getTikTokInfo gets TikTok video information with optimized client
func (ds *DownloaderSystem) getTikTokInfo(url string) (*VideoInfo, error) {
	apiURL := ds.cfg.API("lann", "/api/download/tiktok", map[string]string{
		"url": url,
	})
	
	resp, err := ds.httpClient.Get(apiURL)
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

// FormatViews formats view count for better readability
func (ds *DownloaderSystem) FormatViews(views int64) string {
	if views >= 1000000000 {
		return fmt.Sprintf("%.1fB", float64(views)/1000000000)
	} else if views >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(views)/1000000)
	} else if views >= 1000 {
		return fmt.Sprintf("%.1fK", float64(views)/1000)
	}
	return fmt.Sprintf("%d", views)
}

// CleanFileName removes invalid characters from filename
func (ds *DownloaderSystem) CleanFileName(filename string) string {
	// Remove or replace invalid characters for filenames
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	cleaned := filename
	for _, char := range invalid {
		cleaned = strings.ReplaceAll(cleaned, char, "_")
	}
	return cleaned
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

// ClearCache clears the download cache
func (ds *DownloaderSystem) ClearCache() {
	ds.cacheMutex.Lock()
	defer ds.cacheMutex.Unlock()
	
	ds.cache = make(map[string]*DownloadResult)
}

// GetCacheStats returns cache statistics
func (ds *DownloaderSystem) GetCacheStats() (int, int) {
	ds.cacheMutex.RLock()
	defer ds.cacheMutex.RUnlock()
	
	total := len(ds.cache)
	expired := 0
	cutoff := time.Now().Add(-ds.cacheTTL)
	
	for _, result := range ds.cache {
		if result.CachedAt.Before(cutoff) {
			expired++
		}
	}
	
	return total, expired
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

// EnsureGlobalDownloaderSystem ensures the global downloader system is available
// This function can be called to wait for the system to be initialized
func EnsureGlobalDownloaderSystem(maxWait time.Duration) *DownloaderSystem {
	// First try immediate access (most common case)
	if globalDownloaderSystem != nil {
		return globalDownloaderSystem
	}
	
	// If not available, use exponential backoff for faster response
	start := time.Now()
	backoff := 10 * time.Millisecond // Start with 10ms
	
	for time.Since(start) < maxWait {
		if globalDownloaderSystem != nil {
			return globalDownloaderSystem
		}
		
		// Use shorter sleep intervals for faster response
		time.Sleep(backoff)
		
		// Exponential backoff with cap
		backoff *= 2
		if backoff > 100*time.Millisecond {
			backoff = 100 * time.Millisecond
		}
	}
	return nil
}

// IsGlobalDownloaderSystemReady checks if the global downloader system is ready
func IsGlobalDownloaderSystemReady() bool {
	return globalDownloaderSystem != nil
} 