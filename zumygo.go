package main

import (
	"context"
	"fmt"
	"zumygo/handlers"
	"zumygo/helpers"
	"zumygo/plugins"
	"zumygo/systems"
	"zumygo/database"
	"zumygo/config"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"strconv"
	"syscall"
	"time"

	_ "zumygo/commands/main"   // Import main commands
	_ "zumygo/commands/owner"  // Import owner commands
	_ "zumygo/commands/Auto"   // Import auto commands

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
	waproto "go.mau.fi/whatsmeow/binary/proto"
)

var (
	log helpers.Logger
	statusUpdateTicker *time.Ticker
)

func init() {
	store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_EDGE.Enum()
	store.DeviceProps.Os = proto.String("Linux")
}

// connectWithRetry attempts to connect with retry mechanism
func connectWithRetry(conn *whatsmeow.Client, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		if err := conn.Connect(); err == nil {
			return nil
		}
		log.Warn(fmt.Sprintf("Connection attempt %d failed, retrying in %d seconds...", i+1, i+1))
		time.Sleep(time.Second * time.Duration(i+1))
	}
	return fmt.Errorf("failed to connect after %d attempts", maxRetries)
}

func StartClient() {
	// Add recovery mechanism
	defer func() {
		if r := recover(); r != nil {
			log.Error(fmt.Sprintf("Recovered from panic: %v", r))
		}
	}()

	// Get global systems
	cfg := GetGlobalConfig()
	db := GetGlobalDatabase()
	miningSystem := GetGlobalMiningSystem()
	pluginManager := GetGlobalPluginManager()
	
	// Validate configuration
	if len(cfg.Owner) == 0 {
		log.Error("OWNER configuration is required")
		os.Exit(1)
	}
	if cfg.Prefix == "" {
		log.Error("PREFIX configuration is required")
		os.Exit(1)
	}

	log.Info("Configuration validation passed")
	log.Info("Starting WhatsApp bot with enhanced features...")
	
	ctx := context.Background()
	dbLog := waLog.Stdout("Database", "ERROR", true)
	container, err := sqlstore.New(ctx, "sqlite3", "file:session.db?_foreign_keys=on", dbLog)
	if err != nil {
		log.Error("Failed to create database container: " + err.Error())
		os.Exit(1)
	}

	handler := handlers.NewHandler(container)
	if handler == nil {
		log.Error("Failed to create message handler")
		os.Exit(1)
	}

	log.Info("Connecting Socket")
	conn := handler.Client()
	conn.PrePairCallback = func(jid types.JID, platform, businessName string) bool {
		log.Info("Connected Socket")
		return true
	}

	if conn.Store.ID == nil {
		// No ID stored, new login
		pairingNumber := cfg.PairingNumber

		if pairingNumber != "" {
			pairingNumber = regexp.MustCompile(`\D+`).ReplaceAllString(pairingNumber, "")

			if err := connectWithRetry(conn, 3); err != nil {
				log.Error("Failed to connect for pairing: " + err.Error())
				os.Exit(1)
			}

			ctx := context.Background()
			code, err := conn.PairPhone(ctx, pairingNumber, true, whatsmeow.PairClientChrome, "Edge (Linux)")
			if err != nil {
				log.Error("Failed to pair phone: " + err.Error())
				os.Exit(1)
			}

			fmt.Printf("🔗 Pairing Code: %s\n", code)
		} else {
			qrChan, _ := conn.GetQRChannel(context.Background())
			if err := connectWithRetry(conn, 3); err != nil {
				log.Error("Failed to connect for QR: " + err.Error())
				os.Exit(1)
			}

			for evt := range qrChan {
				switch string(evt.Event) {
				case "code":
					qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
					log.Info("QR Code generated - scan with WhatsApp")
				}
			}
		}
	} else {
		// Already logged in, just connect
		if err := connectWithRetry(conn, 3); err != nil {
			log.Error("Failed to connect: " + err.Error())
			os.Exit(1)
		}
		log.Info("Connected to WhatsApp successfully")
	}

	// Start periodic status updates
	startStatusUpdates(conn, cfg, db)

	// Set up enhanced message handler with plugin support
	setupEnhancedMessageHandler(conn, cfg, db, miningSystem, pluginManager)

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Info("Shutting down gracefully...")
	
	// Stop status updates
	if statusUpdateTicker != nil {
		statusUpdateTicker.Stop()
	}
	
	// Save database before shutdown
	if db := GetGlobalDatabase(); db != nil {
		if err := db.Save(); err != nil {
			log.Error("Failed to save database: " + err.Error())
		}
	}
	
	conn.Disconnect()
}

// startStatusUpdates starts periodic status updates similar to JavaScript base
func startStatusUpdates(conn *whatsmeow.Client, cfg *config.BotConfig, db *database.Database) {
	statusUpdateTicker = time.NewTicker(1 * time.Minute) // Update every minute
	
	go func() {
		defer statusUpdateTicker.Stop()
		
		for range statusUpdateTicker.C {
			if conn.IsConnected() {
				uptime := db.GetUptime()
				hours := uptime / 3600
				minutes := (uptime % 3600) / 60
				
				var uptimeStr string
				if hours > 0 {
					uptimeStr = fmt.Sprintf("%dh %dm", hours, minutes)
				} else {
					uptimeStr = fmt.Sprintf("%dm", minutes)
				}
				
				userCount := len(db.Users)
				mode := "Public-Mode"
				if cfg.Prefix == "." {
					mode = "Self-Mode"
				}
				
				status := fmt.Sprintf("I am %s | Online for %s ⏳ | Mode: %s | Users: %d | Created by %s", 
					cfg.NameBot, uptimeStr, mode, userCount, cfg.NameOwner)
				
				conn.SetStatusMessage(status)
			}
		}
	}()
}

// setupEnhancedMessageHandler sets up message handling with plugin support
func setupEnhancedMessageHandler(conn *whatsmeow.Client, cfg *config.BotConfig, db *database.Database, miningSystem *systems.MiningSystem, pluginManager *plugins.PluginManager) {
	conn.AddEventHandler(func(evt *whatsmeow.Event) {
		switch v := evt.(type) {
		case *whatsmeow.MessageEvent:
			handleEnhancedMessage(v, conn, cfg, db, miningSystem, pluginManager)
		}
	})
}

// handleEnhancedMessage handles incoming messages with plugin support
func handleEnhancedMessage(evt *whatsmeow.MessageEvent, conn *whatsmeow.Client, cfg *config.BotConfig, db *database.Database, miningSystem *systems.MiningSystem, pluginManager *plugins.PluginManager) {
	// Skip if no message content
	if evt.Message == nil {
		return
	}
	
	// Get message text
	var messageText string
	if evt.Message.Conversation != nil {
		messageText = *evt.Message.Conversation
	} else if evt.Message.ExtendedTextMessage != nil && evt.Message.ExtendedTextMessage.Text != nil {
		messageText = *evt.Message.ExtendedTextMessage.Text
	}
	
	if messageText == "" {
		return
	}
	
	// Update database stats
	db.IncrementMessages()
	
	// Get or create user
	user := db.GetUser(evt.Info.Sender.String())
	chat := db.GetChat(evt.Info.Chat.String())
	
	// Update chat activity
	chat.LastActivity = time.Now().Unix()
	chat.MessageCount++
	
	// Check if message starts with prefix
	if !strings.HasPrefix(messageText, cfg.Prefix) {
		return
	}
	
	// Parse command
	parts := strings.Fields(messageText[len(cfg.Prefix):])
	if len(parts) == 0 {
		return
	}
	
	command := strings.ToLower(parts[0])
	args := parts[1:]
	
	// Check permissions
	isOwner := cfg.IsOwner(evt.Info.Sender.User)
	isAdmin := cfg.IsMod(evt.Info.Sender.User) || isOwner
	isPremium := cfg.IsPrem(evt.Info.Sender.User) || user.Premium || isOwner
	isGroup := evt.Info.Chat.Server == "g.us"
	
	// Create command message
	cmdMsg := &plugins.CommandMessage{
		ID:        evt.Info.ID,
		From:      evt.Info.Sender.String(),
		Chat:      evt.Info.Chat.String(),
		Text:      messageText,
		Command:   command,
		Args:      args,
		IsGroup:   isGroup,
		IsOwner:   isOwner,
		IsAdmin:   isAdmin,
		IsPremium: isPremium,
		User:      user,
		ChatData:  chat,
		Reply: func(text string) error {
			_, err := conn.SendMessage(context.Background(), evt.Info.Chat, &waproto.Message{
				Conversation: &text,
			})
			return err
		},
		React: func(emoji string) error {
			return conn.SendMessage(context.Background(), evt.Info.Chat, &waproto.Message{
				ReactionMessage: &waproto.ReactionMessage{
					Key: &waproto.MessageKey{
						RemoteJid: &evt.Info.Chat.String(),
						FromMe:    &evt.Info.FromMe,
						Id:        &evt.Info.ID,
					},
					Text: &emoji,
				},
			})
		},
		Delete: func() error {
			return conn.SendMessage(context.Background(), evt.Info.Chat, &waproto.Message{
				ProtocolMessage: &waproto.ProtocolMessage{
					Type: waproto.ProtocolMessage_REVOKE.Enum(),
					Key: &waproto.MessageKey{
						RemoteJid: &evt.Info.Chat.String(),
						FromMe:    &evt.Info.FromMe,
						Id:        &evt.Info.ID,
					},
				},
			})
		},
	}
	
	// Try to execute command through plugin system
	if err := pluginManager.ExecuteCommand(cmdMsg); err != nil {
		// If plugin command fails, try built-in commands
		handleBuiltinCommands(cmdMsg, cfg, db, miningSystem)
	}
	
	// Update command statistics
	db.IncrementCommand(command)
}

// handleBuiltinCommands handles built-in commands that aren't plugins
func handleBuiltinCommands(msg *plugins.CommandMessage, cfg *config.BotConfig, db *database.Database, miningSystem *systems.MiningSystem) {
	switch msg.Command {
	case "mine":
		if result, err := miningSystem.Mine(msg.From); err == nil {
			msg.Reply(result)
		} else {
			msg.Reply("❌ Mining failed: " + err.Error())
		}
		
	case "mining":
		result := miningSystem.GetMiningInfo(msg.From)
		msg.Reply(result)
		
	case "pickaxeshop":
		result := miningSystem.GetPickaxeShop()
		msg.Reply(result)
		
	case "buypickaxe":
		if len(msg.Args) < 1 {
			msg.Reply("❌ Usage: buypickaxe <type>\nTypes: wooden, stone, iron, gold, diamond")
			return
		}
		
		if result, err := miningSystem.BuyPickaxe(msg.From, msg.Args[0]); err == nil {
			msg.Reply(result)
		} else {
			msg.Reply("❌ Purchase failed: " + err.Error())
		}
		
	case "sellore":
		if len(msg.Args) < 2 {
			msg.Reply("❌ Usage: sellore <type> <amount>\nTypes: coal, iron, gold, diamond, emerald")
			return
		}
		
		amount := int64(1)
		if len(msg.Args) > 1 {
			if parsed, err := strconv.ParseInt(msg.Args[1], 10, 64); err == nil {
				amount = parsed
			}
		}
		
		if result, err := miningSystem.SellOre(msg.From, msg.Args[0], amount); err == nil {
			msg.Reply(result)
		} else {
			msg.Reply("❌ Sale failed: " + err.Error())
		}
		
	case "balance", "bal":
		user := msg.User
		balanceMsg := fmt.Sprintf("💰 *Your Balance*\n\n")
		balanceMsg += fmt.Sprintf("💵 Money: %d coins\n", user.Money)
		balanceMsg += fmt.Sprintf("🪙 ZumyCoin: %d ZC\n", user.ZC)
		balanceMsg += fmt.Sprintf("🏦 ATM: %d coins\n", user.ATM)
		balanceMsg += fmt.Sprintf("⭐ Level: %d\n", user.Level)
		balanceMsg += fmt.Sprintf("✨ Experience: %d\n", user.Exp)
		msg.Reply(balanceMsg)
		
	case "stats":
		stats := db.Stats
		uptime := db.GetUptime()
		hours := uptime / 3600
		minutes := (uptime % 3600) / 60
		
		statsMsg := fmt.Sprintf("📊 *Bot Statistics*\n\n")
		statsMsg += fmt.Sprintf("👥 Total Users: %d\n", stats.TotalUsers)
		statsMsg += fmt.Sprintf("💬 Total Chats: %d\n", stats.TotalChats)
		statsMsg += fmt.Sprintf("📨 Total Messages: %d\n", stats.TotalMessages)
		statsMsg += fmt.Sprintf("⏰ Uptime: %dh %dm\n", hours, minutes)
		statsMsg += fmt.Sprintf("🔧 Commands Used: %d\n", len(stats.Commands))
		msg.Reply(statsMsg)
	}
}
