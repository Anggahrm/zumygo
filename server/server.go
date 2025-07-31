package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
	"zumygo/config"
	"zumygo/database"
	"zumygo/systems"
	"zumygo/plugins"
)

// Server represents the web server
type Server struct {
	config        *config.BotConfig
	database      *database.Database
	miningSystem  *systems.MiningSystem
	pluginManager *plugins.PluginManager
	startTime     time.Time
	port          int
}

// Response represents a standard API response
type Response struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Author  string      `json:"author"`
}

// BotStatus represents bot status information
type BotStatus struct {
	BotName       string                 `json:"botName"`
	Version       string                 `json:"version"`
	Uptime        string                 `json:"uptime"`
	UptimeSeconds int64                  `json:"uptimeSeconds"`
	Users         int                    `json:"users"`
	Chats         int                    `json:"chats"`
	Messages      int64                  `json:"messages"`
	Commands      int                    `json:"commands"`
	Plugins       int                    `json:"plugins"`
	Memory        map[string]interface{} `json:"memory"`
	System        map[string]interface{} `json:"system"`
}

// NewServer creates a new web server instance
func NewServer(cfg *config.BotConfig, db *database.Database, ms *systems.MiningSystem, pm *plugins.PluginManager) *Server {
	port := 8080
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		if p, err := strconv.Atoi(portEnv); err == nil {
			port = p
		}
	}

	return &Server{
		config:        cfg,
		database:      db,
		miningSystem:  ms,
		pluginManager: pm,
		startTime:     time.Now(),
		port:          port,
	}
}

// Start starts the web server
func (s *Server) Start() {
	mux := http.NewServeMux()

	// Setup routes
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/stats", s.handleStats)
	mux.HandleFunc("/plugins", s.handlePlugins)
	mux.HandleFunc("/commands", s.handleCommands)
	mux.HandleFunc("/users", s.handleUsers)
	mux.HandleFunc("/health", s.handleHealth)

	// Add CORS middleware
	handler := s.corsMiddleware(mux)

	fmt.Printf("🌐 Web server starting on port %d\n", s.port)
	
	// Try to start server, if port is busy, try next port
	for attempts := 0; attempts < 10; attempts++ {
		addr := fmt.Sprintf(":%d", s.port+attempts)
		fmt.Printf("🔗 Trying to bind to %s\n", addr)
		
		if err := http.ListenAndServe(addr, handler); err != nil {
			if attempts < 9 {
				fmt.Printf("⚠️ Port %d is busy, trying port %d\n", s.port+attempts, s.port+attempts+1)
				continue
			}
			log.Fatalf("❌ Failed to start server after 10 attempts: %v", err)
		}
		break
	}
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// handleRoot handles the root endpoint
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Status:  true,
		Message: "Bot Successfully Activated!",
		Author:  s.config.NameOwner,
		Data: map[string]interface{}{
			"botName":    s.config.NameBot,
			"version":    "2.0",
			"uptime":     s.getUptimeString(),
			"timestamp":  time.Now().Unix(),
		},
	}
	
	s.sendJSON(w, response)
}

// handleStatus handles the status endpoint
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)
	
	uptime := time.Since(s.startTime)
	
	status := BotStatus{
		BotName:       s.config.NameBot,
		Version:       "2.0",
		Uptime:        s.getUptimeString(),
		UptimeSeconds: int64(uptime.Seconds()),
		Users:         len(s.database.Users),
		Chats:         len(s.database.Chats),
		Messages:      s.database.Stats.TotalMessages,
		Commands:      len(s.pluginManager.GetCommands()),
		Plugins:       len(s.pluginManager.GetCommands()),
		Memory: map[string]interface{}{
			"used":      bToMb(m.Alloc),
			"total":     bToMb(m.TotalAlloc),
			"sys":       bToMb(m.Sys),
			"numGC":     m.NumGC,
		},
		System: map[string]interface{}{
			"platform":   runtime.GOOS,
			"arch":       runtime.GOARCH,
			"version":    runtime.Version(),
			"goroutines": runtime.NumGoroutine(),
			"cpus":       runtime.NumCPU(),
		},
	}
	
	response := Response{
		Status:  true,
		Message: "Bot status retrieved successfully",
		Author:  s.config.NameOwner,
		Data:    status,
	}
	
	s.sendJSON(w, response)
}

// handleStats handles the stats endpoint
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := s.database.Stats
	
	response := Response{
		Status:  true,
		Message: "Statistics retrieved successfully",
		Author:  s.config.NameOwner,
		Data: map[string]interface{}{
			"totalUsers":    stats.TotalUsers,
			"totalChats":    stats.TotalChats,
			"totalMessages": stats.TotalMessages,
			"startTime":     stats.StartTime,
			"uptime":        s.database.GetUptime(),
			"commands":      stats.Commands,
		},
	}
	
	s.sendJSON(w, response)
}

// handlePlugins handles the plugins endpoint
func (s *Server) handlePlugins(w http.ResponseWriter, r *http.Request) {
	commands := s.pluginManager.GetCommands()
	categories := s.pluginManager.GetCommandsByCategory()
	
	pluginData := make(map[string]interface{})
	for category, cmds := range categories {
		commandList := make([]map[string]interface{}, 0)
		for _, cmd := range cmds {
			commandList = append(commandList, map[string]interface{}{
				"name":        cmd.Name,
				"aliases":     cmd.Aliases,
				"description": cmd.Description,
				"usage":       cmd.Usage,
				"ownerOnly":   cmd.OwnerOnly,
				"adminOnly":   cmd.AdminOnly,
				"premiumOnly": cmd.PremiumOnly,
				"groupOnly":   cmd.GroupOnly,
				"privateOnly": cmd.PrivateOnly,
			})
		}
		pluginData[category] = commandList
	}
	
	response := Response{
		Status:  true,
		Message: fmt.Sprintf("Found %d commands in %d categories", len(commands), len(categories)),
		Author:  s.config.NameOwner,
		Data:    pluginData,
	}
	
	s.sendJSON(w, response)
}

// handleCommands handles the commands endpoint
func (s *Server) handleCommands(w http.ResponseWriter, r *http.Request) {
	commands := s.pluginManager.GetCommands()
	
	commandList := make([]map[string]interface{}, 0)
	for _, cmd := range commands {
		commandList = append(commandList, map[string]interface{}{
			"name":        cmd.Name,
			"aliases":     cmd.Aliases,
			"description": cmd.Description,
			"usage":       cmd.Usage,
			"category":    cmd.Category,
			"ownerOnly":   cmd.OwnerOnly,
			"adminOnly":   cmd.AdminOnly,
			"premiumOnly": cmd.PremiumOnly,
			"groupOnly":   cmd.GroupOnly,
			"privateOnly": cmd.PrivateOnly,
		})
	}
	
	response := Response{
		Status:  true,
		Message: fmt.Sprintf("Found %d commands", len(commands)),
		Author:  s.config.NameOwner,
		Data:    commandList,
	}
	
	s.sendJSON(w, response)
}

// handleUsers handles the users endpoint
func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	// Only show basic stats for privacy
	userStats := make(map[string]interface{})
	
	totalUsers := len(s.database.Users)
	premiumUsers := 0
	bannedUsers := 0
	
	for _, user := range s.database.Users {
		if user.Premium {
			premiumUsers++
		}
		if user.Banned {
			bannedUsers++
		}
	}
	
	userStats["total"] = totalUsers
	userStats["premium"] = premiumUsers
	userStats["banned"] = bannedUsers
	userStats["regular"] = totalUsers - premiumUsers - bannedUsers
	
	response := Response{
		Status:  true,
		Message: "User statistics retrieved successfully",
		Author:  s.config.NameOwner,
		Data:    userStats,
	}
	
	s.sendJSON(w, response)
}

// handleHealth handles the health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"uptime":    s.getUptimeString(),
		"version":   "2.0",
		"services": map[string]bool{
			"database":      s.database != nil,
			"miningSystem":  s.miningSystem != nil,
			"pluginManager": s.pluginManager != nil,
		},
	}
	
	response := Response{
		Status:  true,
		Message: "Bot is healthy",
		Author:  s.config.NameOwner,
		Data:    health,
	}
	
	s.sendJSON(w, response)
}

// sendJSON sends a JSON response
func (s *Server) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error encoding JSON: %v", err)
	}
}

// getUptimeString returns formatted uptime string
func (s *Server) getUptimeString() string {
	uptime := time.Since(s.startTime)
	
	days := int(uptime.Hours()) / 24
	hours := int(uptime.Hours()) % 24
	minutes := int(uptime.Minutes()) % 60
	seconds := int(uptime.Seconds()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}

// bToMb converts bytes to megabytes
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// StartWebServer starts the web server in a separate goroutine
func StartWebServer(cfg *config.BotConfig, db *database.Database, ms *systems.MiningSystem, pm *plugins.PluginManager) {
	server := NewServer(cfg, db, ms, pm)
	
	go func() {
		server.Start()
	}()
}