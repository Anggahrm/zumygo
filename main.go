package main

import (
	"fmt"
	"os"
	"zumygo/config"
	"zumygo/database"
	"zumygo/systems"
	"zumygo/helpers"
	"github.com/subosito/gotenv"
)

var (
	cfg            *config.BotConfig
	db             *database.Database
	miningSystem   *systems.MiningSystem
	healthSystem   *systems.HealthSystem
	economySystem  *systems.EconomySystem
	levelingSystem *systems.LevelingSystem
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
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Println()
	
	// Show system features
	fmt.Println("🎮 Available Features:")
	fmt.Println("  ⛏️  Mining System - Mine ores and buy pickaxes")
	fmt.Println("  ❤️  Health System - Manage HP and potions")
	fmt.Println("  💰 Economy System - Work, shop, and trade")
	fmt.Println("  ⭐ Leveling System - Gain XP and unlock roles")
	fmt.Println()
	
	// Show built-in commands count
	builtinCommands := []string{
		"mine", "mining", "pickaxeshop", "buypickaxe", "sellore",
		"health", "usepotion", "potionshop", "buypotion", "upgradehealth",
		"work", "daily", "shop", "buy", "inventory", "transfer", "rob", "deposit", "withdraw",
		"level", "leaderboard", "roles", "autolevelup",
		"balance", "stats", "toplevel", "topmoney", "tophealth",
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

// GetGlobalLogger returns the global logger
func GetGlobalLogger() *helpers.Logger {
	return logger
}
