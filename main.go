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
	cfg            *config.BotConfig
	db             *database.Database
	miningSystem   *systems.MiningSystem
	healthSystem   *systems.HealthSystem
	economySystem  *systems.EconomySystem
	levelingSystem *systems.LevelingSystem
	pluginManager  *plugins.PluginManager
	logger         *helpers.Logger
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

	// Initialize all systems
	miningSystem = systems.InitializeMiningSystem(db)
	logger.Info("Mining system initialized successfully")
	
	healthSystem = systems.InitializeHealthSystem(db)
	logger.Info("Health system initialized successfully")
	
	economySystem = systems.InitializeEconomySystem(db)
	logger.Info("Economy system initialized successfully")
	
	levelingSystem = systems.InitializeLevelingSystem(db)
	logger.Info("Leveling system initialized successfully")

	// Initialize plugin system
	pluginContext := &plugins.PluginContext{
		Config:         cfg,
		Database:       db,
		MiningSystem:   miningSystem,
		HealthSystem:   healthSystem,
		EconomySystem:  economySystem,
		LevelingSystem: levelingSystem,
		Logger:         logger,
		Prefix:         cfg.Prefix,
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
	server.StartWebServer(cfg, db, miningSystem, healthSystem, economySystem, levelingSystem, pluginManager)

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
	fmt.Printf("║ Mining System: %-21s ║\n", "✅ Active")
	fmt.Printf("║ Health System: %-21s ║\n", "✅ Active")
	fmt.Printf("║ Economy System: %-20s ║\n", "✅ Active")
	fmt.Printf("║ Leveling System: %-19s ║\n", "✅ Active")
	fmt.Printf("║ Plugins: %-27s ║\n", "✅ Loaded")
	fmt.Printf("║ Web Server: %-24s ║\n", "✅ Running")
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Println()
	
	// Show system features
	fmt.Println("🎮 Available Features:")
	fmt.Println("  ⛏️  Mining System - Mine ores and buy pickaxes")
	fmt.Println("  ❤️  Health System - Manage HP and potions")
	fmt.Println("  💰 Economy System - Work, shop, and trade")
	fmt.Println("  ⭐ Leveling System - Gain XP and unlock roles")
	fmt.Println("  🔌 Plugin System - Hot-reloadable commands")
	fmt.Println("  🌐 Web Dashboard - Real-time monitoring")
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
	
	// Show built-in commands count
	builtinCommands := []string{
		"mine", "mining", "pickaxeshop", "buypickaxe", "sellore",
		"health", "usepotion", "potionshop", "buypotion", "upgradehealth",
		"work", "daily", "shop", "buy", "inventory", "transfer", "rob", "deposit", "withdraw",
		"level", "leaderboard", "roles", "autolevelup",
		"balance", "stats", "toplevel", "topmoney", "tophealth",
	}
	fmt.Printf("⚡ Built-in commands: %d\n", len(builtinCommands))
	fmt.Printf("🌐 Web dashboard: http://localhost:%s\n", os.Getenv("PORT"))
	if os.Getenv("PORT") == "" {
		fmt.Printf("🌐 Web dashboard: http://localhost:8080\n")
	}
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

// GetGlobalMiningSystem returns the global mining system
func GetGlobalMiningSystem() *systems.MiningSystem {
	return miningSystem
}

// GetGlobalHealthSystem returns the global health system
func GetGlobalHealthSystem() *systems.HealthSystem {
	return healthSystem
}

// GetGlobalEconomySystem returns the global economy system
func GetGlobalEconomySystem() *systems.EconomySystem {
	return economySystem
}

// GetGlobalLevelingSystem returns the global leveling system
func GetGlobalLevelingSystem() *systems.LevelingSystem {
	return levelingSystem
}

// GetGlobalPluginManager returns the global plugin manager
func GetGlobalPluginManager() *plugins.PluginManager {
	return pluginManager
}

// GetGlobalLogger returns the global logger
func GetGlobalLogger() *helpers.Logger {
	return logger
}
