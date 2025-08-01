package main

import (
	"context"
	"fmt"
	"zumygo/handlers"
	"zumygo/helpers"
	"zumygo/systems"
	"zumygo/database"
	"zumygo/config"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	_ "zumygo/commands/main"      // Import main commands
	_ "zumygo/commands/owner"     // Import owner commands
	_ "zumygo/commands/Auto"      // Import auto commands
	_ "zumygo/commands/downloader" // Import downloader commands

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
	waproto "go.mau.fi/whatsmeow/binary/proto"
)

var (
	clientLogger helpers.Logger
)

// CommandMessage represents a command message
type CommandMessage struct {
	ID        string
	From      string
	Chat      string
	Text      string
	Command   string
	Args      []string
	IsGroup   bool
	IsOwner   bool
	IsAdmin   bool
	IsPremium bool
	User      *database.User
	ChatData  *database.Chat
	Reply     func(string) error
	React     func(string) error
	Delete    func() error
}

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
		clientLogger.Warn(fmt.Sprintf("Connection attempt %d failed, retrying in %d seconds...", i+1, i+1))
		time.Sleep(time.Second * time.Duration(i+1))
	}
	return fmt.Errorf("failed to connect after %d attempts", maxRetries)
}

func StartClient() {
	// Add recovery mechanism
	defer func() {
		if r := recover(); r != nil {
			clientLogger.Error(fmt.Sprintf("Recovered from panic: %v", r))
		}
	}()

	// Get global systems
	cfg := GetGlobalConfig()
	db := GetGlobalDatabase()
	downloaderSystem := GetGlobalDownloaderSystem()
	
	// Validate configuration
	if len(cfg.Owner) == 0 {
		clientLogger.Error("OWNER configuration is required")
		os.Exit(1)
	}
	if cfg.Prefix == "" {
		clientLogger.Error("PREFIX configuration is required")
		os.Exit(1)
	}

	clientLogger.Info("Configuration validation passed")
	clientLogger.Info("Starting WhatsApp bot with enhanced features...")
	
	ctx := context.Background()
	dbLog := waLog.Stdout("Database", "ERROR", true)
	container, err := sqlstore.New(ctx, "sqlite3", "file:session.db?_foreign_keys=on", dbLog)
	if err != nil {
		clientLogger.Error("Failed to create database container: " + err.Error())
		os.Exit(1)
	}

	handler := handlers.NewHandler(container)
	if handler == nil {
		clientLogger.Error("Failed to create message handler")
		os.Exit(1)
	}

	clientLogger.Info("Connecting Socket")
	conn := handler.Client()
	conn.PrePairCallback = func(jid types.JID, platform, businessName string) bool {
		clientLogger.Info("Connected Socket")
		return true
	}

	if conn.Store.ID == nil {
		// No ID stored, new login
		pairingNumber := cfg.PairingNumber

		if pairingNumber != "" {
			pairingNumber = regexp.MustCompile(`\D+`).ReplaceAllString(pairingNumber, "")

			if err := connectWithRetry(conn, 3); err != nil {
				clientLogger.Error("Failed to connect for pairing: " + err.Error())
				os.Exit(1)
			}

			ctx := context.Background()
			code, err := conn.PairPhone(ctx, pairingNumber, true, whatsmeow.PairClientChrome, "Edge (Linux)")
			if err != nil {
				clientLogger.Error("Failed to pair phone: " + err.Error())
				os.Exit(1)
			}

			fmt.Printf("🔗 Pairing Code: %s\n", code)
		} else {
			qrChan, _ := conn.GetQRChannel(context.Background())
			if err := connectWithRetry(conn, 3); err != nil {
				clientLogger.Error("Failed to connect for QR: " + err.Error())
				os.Exit(1)
			}

			for evt := range qrChan {
				switch string(evt.Event) {
				case "code":
					qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
					clientLogger.Info("QR Code generated - scan with WhatsApp")
				}
			}
		}
	} else {
		// Already logged in, just connect
		if err := connectWithRetry(conn, 3); err != nil {
			clientLogger.Error("Failed to connect: " + err.Error())
			os.Exit(1)
		}
		clientLogger.Info("Connected to WhatsApp successfully")
	}

	// Bio system is auto-managed via Before hook
	clientLogger.Info("Bio system auto-managed via Before hook")

	// Set up enhanced message handler
	setupEnhancedMessageHandler(conn, cfg, db, downloaderSystem)

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	clientLogger.Info("Shutting down gracefully...")
	
	// Bio system is auto-managed, no need to stop
	clientLogger.Info("Bio system auto-managed, no cleanup needed")
	
	// Save database before shutdown
	if db := GetGlobalDatabase(); db != nil {
		if err := db.Save(); err != nil {
			clientLogger.Error("Failed to save database: " + err.Error())
		}
	}
	
	conn.Disconnect()
}



// setupEnhancedMessageHandler sets up message handling
func setupEnhancedMessageHandler(conn *whatsmeow.Client, cfg *config.BotConfig, db *database.Database, downloaderSystem *systems.DownloaderSystem) {
	conn.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			handleEnhancedMessage(v, conn, cfg, db, downloaderSystem)
		}
	})
}

// handleEnhancedMessage handles incoming messages
func handleEnhancedMessage(evt *events.Message, conn *whatsmeow.Client, cfg *config.BotConfig, db *database.Database, downloaderSystem *systems.DownloaderSystem) {
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
	
	// Update performance metrics
	if monitor := GetGlobalPerformanceMonitor(); monitor != nil {
		monitor.IncrementMessageCount()
	}
	
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
	cmdMsg := &CommandMessage{
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
			if err != nil {
				return err
			}
			return nil
		},
		React: func(emoji string) error {
			_, err := conn.SendMessage(context.Background(), evt.Info.Chat, &waproto.Message{
				ReactionMessage: &waproto.ReactionMessage{
					Key: &waproto.MessageKey{
						RemoteJID: proto.String(evt.Info.Chat.String()),
						FromMe:    proto.Bool(false),
						ID:        proto.String(evt.Info.ID),
					},
					Text: proto.String(emoji),
				},
			})
			if err != nil {
				return err
			}
			return nil
		},
		Delete: func() error {
			_, err := conn.SendMessage(context.Background(), evt.Info.Chat, &waproto.Message{
				ProtocolMessage: &waproto.ProtocolMessage{
					Type: waproto.ProtocolMessage_REVOKE.Enum(),
					Key: &waproto.MessageKey{
						RemoteJID: proto.String(evt.Info.Chat.String()),
						FromMe:    proto.Bool(false),
						ID:        proto.String(evt.Info.ID),
					},
				},
			})
			if err != nil {
				return err
			}
			return nil
		},
	}
	
	// Handle built-in commands
			handleBuiltinCommands(cmdMsg, cfg, db, downloaderSystem)
	
	// Update command statistics
	db.IncrementCommand(command)
	
	// Update performance metrics
	if monitor := GetGlobalPerformanceMonitor(); monitor != nil {
		monitor.IncrementCommandCount()
	}
}

// handleBuiltinCommands handles built-in commands
func handleBuiltinCommands(msg *CommandMessage, cfg *config.BotConfig, db *database.Database, downloaderSystem *systems.DownloaderSystem) {
	switch msg.Command {






	}
}
