package handlers

import (
	"context"
	"fmt"
	"zumygo/libs"
	"zumygo/config"
	"regexp"
	"strings"
	"time"
	"sync"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type IHandler struct {
	Container *store.Device
}

// Performance optimizations
var (
	commandCache     = make(map[string]*regexp.Regexp)
	commandCacheMutex sync.RWMutex
	messageQueue     = make(chan *libs.IMessage, 1000) // Buffer for concurrent processing
	workerCount      = 5 // Number of worker goroutines
)

func NewHandler(container *sqlstore.Container) *IHandler {
	ctx := context.Background()
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		fmt.Printf("Failed to get first device: %v\n", err)
		return nil
	}
	
	// Start worker goroutines for concurrent message processing
	startMessageWorkers()
	
	return &IHandler{
		Container: deviceStore,
	}
}

// startMessageWorkers starts worker goroutines for concurrent message processing
func startMessageWorkers() {
	for i := 0; i < workerCount; i++ {
		go func(workerID int) {
			for msg := range messageQueue {
				processMessage(msg, workerID)
			}
		}(i)
	}
}

// processMessage processes a single message
func processMessage(m *libs.IMessage, workerID int) {
	// Add recovery mechanism for message processing
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Worker %d recovered from message processing panic: %v\n", workerID, r)
		}
	}()
	
	// Process the message
	if m.Command != "" && libs.HasCommand(m.Command) {
		ExecuteCommand(nil, m) // Pass nil for client as it's not needed in this context
	}
}

func (h *IHandler) Client() *whatsmeow.Client {
	clientLog := waLog.Stdout("Client", "ERROR", true)
	conn := whatsmeow.NewClient(h.Container, clientLog)
	conn.AddEventHandler(h.RegisterHandler(conn))
	return conn
}

func (h *IHandler) RegisterHandler(conn *whatsmeow.Client) func(evt interface{}) {
	return func(evt interface{}) {
		sock := libs.SerializeClient(conn)
		switch v := evt.(type) {
		case *events.Message:
			m := libs.SerializeMessage(v, sock)

			// skip deleted message
			if m.Message.GetProtocolMessage() != nil && m.Message.GetProtocolMessage().GetType() == 0 {
				return
			}

			// log (use async logging for better performance)
			if m.Body != "" {
				go func() {
					fmt.Println("\x1b[94mFrom :", v.Info.PushName, m.Info.Sender.User, "\x1b[39m")
					if libs.HasCommand(m.Command) {
						fmt.Println("\x1b[93mCommand :", m.Command, "\x1b[39m")
					}
					if len(m.Body) < 350 {
						fmt.Print("\x1b[92mMessage : ", m.Body, "\x1b[39m", "\n")
					} else {
						fmt.Print("\x1b[92mMessage : ", m.Info.Type, "\x1b[39m", "\n")
					}
				}()
			}

			// Get command and queue for processing
			if m.Command != "" && libs.HasCommand(m.Command) {
				// Send to message queue for concurrent processing
				select {
				case messageQueue <- m:
					// Message queued successfully
				default:
					// Queue is full, process immediately
					go ExecuteCommand(sock, m)
				}
			}
			return
		case *events.Connected, *events.PushNameSetting:
			if len(conn.Store.PushName) == 0 {
				return
			}
			conn.SendPresence(types.PresenceAvailable)
		}
	}
}

// getCachedRegex returns a cached compiled regex or compiles and caches it
func getCachedRegex(pattern string) *regexp.Regexp {
	commandCacheMutex.RLock()
	if re, exists := commandCache[pattern]; exists {
		commandCacheMutex.RUnlock()
		return re
	}
	commandCacheMutex.RUnlock()
	
	// Compile and cache the regex
	commandCacheMutex.Lock()
	defer commandCacheMutex.Unlock()
	
	// Double-check after acquiring write lock
	if re, exists := commandCache[pattern]; exists {
		return re
	}
	
	var compiled *regexp.Regexp
	if strings.ContainsAny(pattern, "|*+?()[]{}") {
		// Use as regex pattern
		compiled = regexp.MustCompile(`^` + pattern + `$`)
	} else {
		// Use exact match with quote meta for safety
		compiled = regexp.MustCompile(`^` + regexp.QuoteMeta(pattern) + `$`)
	}
	
	commandCache[pattern] = compiled
	return compiled
}

// cleanupCommandCache removes old cache entries to prevent memory leaks
func cleanupCommandCache() {
	commandCacheMutex.Lock()
	defer commandCacheMutex.Unlock()
	
	// Keep only the most recent 1000 patterns
	if len(commandCache) > 1000 {
		// Simple cleanup: clear all and let them be recompiled as needed
		commandCache = make(map[string]*regexp.Regexp)
	}
}

func ExecuteCommand(c *libs.IClient, m *libs.IMessage) {
	// Add recovery mechanism for command execution
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from command execution panic: %v\n", r)
			if m != nil {
				m.Reply("An error occurred while executing the command")
			}
		}
	}()

	// Get the command name (already processed in SerializeMessage)
	commandName := m.Command
	if commandName == "" {
		return
	}
	
	// Extract prefix from the original message body
	prefix, hasPrefix := libs.ExtractPrefix(m.Body)
	if !hasPrefix {
		return
	}
	
	lists := libs.GetList()
	
	// Use a more efficient loop with early exit
	for _, cmd := range lists {
		// Execute Before hook if exists
		if cmd.Before != nil {
			cmd.Before(c, m)
		}
		
		// Get cached regex for command matching
		re := getCachedRegex(cmd.Name)
		
		if valid := len(re.FindAllString(commandName, -1)) > 0; valid {
			if cmd.Execute != nil {
				// Check public mode
				if !config.Config.PublicMode && !m.IsOwner {
					return
				}

				// Check prefix requirements
				var cmdWithPref bool
				var cmdWithoutPref bool
				if cmd.IsPrefix && prefix != "" {
					cmdWithPref = true
				} else {
					cmdWithPref = false
				}

				if !cmd.IsPrefix {
					cmdWithoutPref = true
				} else {
					cmdWithoutPref = false
				}

				if !cmdWithPref && !cmdWithoutPref {
					continue
				}

				// Check owner requirement
				if cmd.IsOwner && !m.IsOwner {
					continue
				}

				// Check query requirement
				if cmd.IsQuery && m.Text == "" {
					m.Reply("Query Required")
					continue
				}

				// Check group requirement
				if cmd.IsGroup && !m.Info.IsGroup {
					m.Reply("Commands only work in Group Chat")
					continue
				}

				// Check private requirement
				if cmd.IsPrivate && m.Info.IsGroup {
					m.Reply("Commands only work in Private Chat")
					continue
				}

				// Check media requirement
				if cmd.IsMedia && m.IsMedia == "" {
					m.Reply("Reply to Media Message, or send Media with Command")
					continue
				}

				// Show wait indicator
				if cmd.IsWait {
					m.React("⏳")
				}

				// Execute command
				ok := cmd.Execute(c, m)

				// Handle wait indicator
				if cmd.IsWait && !ok {
					m.React("❌")
				}

				if cmd.IsWait && ok {
					if c != nil && c.WA != nil {
						c.WA.MarkRead([]string{m.Info.ID}, time.Now(), m.Info.Chat, m.Info.Sender)
					}
					m.React("")
				}
				
				// Return after executing command to avoid multiple executions
				return
			}
		}
	}
	
	// Cleanup cache periodically
	if time.Now().Unix()%1000 == 0 { // Every 1000th command
		go cleanupCommandCache()
	}
}