package main

import (
	"fmt"
	"log"
	"os"
	"zumygo/config"
	"zumygo/database"
	"zumygo/systems"
	"zumygo/plugins"
	"zumygo/helpers"
	"zumygo/server"
	"github.com/subosito/gotenv"
)

var (
	cfg           *config.BotConfig
	db            *database.Database
	miningSystem  *systems.MiningSystem
	pluginManager *plugins.PluginManager
	logger        *helpers.Logger
)

func main() {
	// Load environment variables
	gotenv.Load()

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

	// Initialize mining system
	miningSystem = systems.InitializeMiningSystem(db)
	logger.Info("Mining system initialized successfully")

	// Initialize plugin system
	pluginContext := &plugins.PluginContext{
		Config:       cfg,
		Database:     db,
		MiningSystem: miningSystem,
		Logger:       logger,
		Prefix:       cfg.Prefix,
	}

	pluginManager = plugins.NewPluginManager(pluginContext, "plugins")
	if err := pluginManager.LoadAllPlugins(); err != nil {
		logger.Warn(fmt.Sprintf("Failed to load some plugins: %v", err))
	} else {
		logger.Info("Plugin system initialized successfully")
	}

	// Start plugin directory watcher
	pluginManager.WatchPluginDirectory()

	// Start web server
	server.StartWebServer(cfg, db, miningSystem, pluginManager)

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
	fmt.Printf("║ Database: %-26s ║\n", "Initialized")
	fmt.Printf("║ Mining System: %-21s ║\n", "Active")
	fmt.Printf("║ Plugins: %-27s ║\n", "Loaded")
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Println()
	
	// Show loaded plugins
	commands := pluginManager.GetCommands()
	if len(commands) > 0 {
		fmt.Printf("🔌 Loaded %d commands from plugins:\n", len(commands))
		for name := range commands {
			fmt.Printf("  • %s\n", name)
		}
		fmt.Println()
	}
}

// GetGlobalConfig returns the global configuration
func GetGlobalConfig() *config.BotConfig {
	return cfg
}

// GetGlobalDatabase returns the global database
func GetGlobalDatabase() *database.Database {
	return db
}

// GetGlobalMiningSystem returns the global mining system
func GetGlobalMiningSystem() *systems.MiningSystem {
	return miningSystem
}

// GetGlobalPluginManager returns the global plugin manager
func GetGlobalPluginManager() *plugins.PluginManager {
	return pluginManager
}

// GetGlobalLogger returns the global logger
func GetGlobalLogger() *helpers.Logger {
	return logger
}
