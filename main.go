package main

import (
	"fmt"
	"os"
	"zumygo/config"
	"zumygo/database"
	"zumygo/systems"
	"zumygo/helpers"
)

var (
	cfg            *config.BotConfig
	db             *database.Database
	downloaderSystem *systems.DownloaderSystem
	logger         *helpers.Logger
)

func main() {
	// Initialize logger
	logger = &helpers.Logger{}
	
	// Load configuration
	cfg = config.LoadConfig()
	logger.Info("Configuration loaded successfully")

	// Initialize database
	var err error
	dbFile := "database.json"
	if cfg.DatabaseURL != "" {
		dbFile = cfg.DatabaseURL
	}
	
	db, err = database.InitDatabase(dbFile)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to initialize database: %v", err))
		os.Exit(1)
	}
	logger.Info("Database initialized successfully")

	// Start auto-save for database
	db.AutoSave()

	// Initialize all systems
	downloaderSystem = systems.InitializeDownloaderSystem(cfg, db, logger)
	systems.SetGlobalDownloaderSystem(downloaderSystem)
	logger.Info("Downloader system initialized successfully")

	// Bio system is auto-initialized via Before hook
	logger.Info("Bio system auto-initialized via Before hook")

	// Print startup information
	printStartupInfo()

	// Start WhatsApp client
	logger.Info("Starting WhatsApp bot...")
	StartClient()
}

func printStartupInfo() {
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Printf("║            %s v2.0                ║\n", cfg.NameBot)
	fmt.Println("║        Enhanced Go Edition           ║")
	fmt.Println("╠══════════════════════════════════════╣")
	fmt.Printf("║ Owner: %-29s ║\n", cfg.NameOwner)
	fmt.Printf("║ Prefix: %-28s ║\n", cfg.Prefix)
	fmt.Printf("║ Database: %-26s ║\n", "✅ Active")

	fmt.Printf("║ Downloader System: %-17s ║\n", "✅ Active")
	fmt.Printf("║ Bio System: %-17s ║\n", "✅ Active")
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Println()
	
	// Show system features
	fmt.Println("🎮 Available Features:")
	fmt.Println("  📥 Downloader System - Download media from various platforms")
	fmt.Println("  📝 Bio System - Auto update profile bio")
	fmt.Println()
	
	// Show built-in commands count
	builtinCommands := []string{
		// Commands are now auto-detected from the command system
	}
	fmt.Printf("⚡ Built-in commands: %d\n", len(builtinCommands))
	fmt.Println()
}

// GetGlobalConfig returns the global configuration
func GetGlobalConfig() *config.BotConfig {
	return cfg
}

// GetGlobalDatabase returns the global database
func GetGlobalDatabase() *database.Database {
	return db
}





// GetGlobalDownloaderSystem returns the global downloader system
func GetGlobalDownloaderSystem() *systems.DownloaderSystem {
	return downloaderSystem
}



// GetGlobalLogger returns the global logger
func GetGlobalLogger() *helpers.Logger {
	return logger
}
