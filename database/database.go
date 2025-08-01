package database

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
	"bytes"
	"compress/gzip"
)

// User represents a user in the database
type User struct {
	// Basic Info
	Name         string    `json:"name"`
	Age          int       `json:"age"`
	RegTime      int64     `json:"regTime"`
	Registered   bool      `json:"registered"`
	
	// Experience & Level
	Exp          int64     `json:"exp"`
	Level        int       `json:"level"`
	Role         string    `json:"role"`
	AutoLevelUp  bool      `json:"autolevelup"`
	
	// Social
	Pasangan     string    `json:"pasangan"`
	
	// Moderation
	Warn         int       `json:"warn"`
	Banned       bool      `json:"banned"`
	BannedUser   bool      `json:"Banneduser"`
	BannedReason string    `json:"BannedReason"`
	
	// Activity
	LastPM       int64     `json:"lastpm"`
	AFK          int64     `json:"afk"`
	AFKReason    string    `json:"afkReason"`
	
	// Premium
	Premium      bool      `json:"premium"`
	PremiumTime  int64     `json:"premiumTime"`
	PremiumDate  int64     `json:"premiumDate"`
}

// Chat represents a chat/group in the database
type Chat struct {
	// Basic Info
	ID          string `json:"id"`
	Name        string `json:"name"`
	
	// Settings
	IsBanned    bool   `json:"isBanned"`
	Welcome     bool   `json:"welcome"`
	Detect      bool   `json:"detect"`
	SWelcome    string `json:"sWelcome"`
	SBye        string `json:"sBye"`
	SPromote    string `json:"sPromote"`
	SDemote     string `json:"sDemote"`
	Delete      bool   `json:"delete"`
	AntiLink    bool   `json:"antiLink"`
	AntiLink2   bool   `json:"antiLink2"`
	AntiToxic   bool   `json:"antiToxic"`
	AntiVirtex  bool   `json:"antiVirtex"`
	Viewonce    bool   `json:"viewonce"`
	
	// Activity
	LastActivity int64 `json:"lastActivity"`
	MessageCount int64 `json:"messageCount"`
	
	// Games
	Game        bool  `json:"game"`
}

// Stats represents bot statistics
type Stats struct {
	TotalUsers    int64            `json:"totalUsers"`
	TotalChats    int64            `json:"totalChats"`
	TotalMessages int64            `json:"totalMessages"`
	StartTime     int64            `json:"startTime"`
	Commands      map[string]int64 `json:"commands"`
}

// Database represents the main database structure
type Database struct {
	Users              map[string]*User `json:"users"`
	Chats              map[string]*Chat `json:"chats"`
	Stats              *Stats           `json:"stats"`
	Messages           map[string]interface{} `json:"msgs"`
	Stickers           map[string]interface{} `json:"sticker"`
	Settings           map[string]interface{} `json:"settings"`
	Responses          map[string]interface{} `json:"respon"`

	
	// Internal
	mutex           sync.RWMutex `json:"-"`
	filename        string       `json:"-"`
	dirty           bool         `json:"-"` // Track if data has been modified
	lastSave        time.Time    `json:"-"` // Track last save time
	saveInterval    time.Duration `json:"-"` // Auto-save interval
	maxUsers        int          `json:"-"` // Maximum number of users to keep in memory
	maxChats        int          `json:"-"` // Maximum number of chats to keep in memory
	cleanupInterval time.Duration `json:"-"` // Cleanup interval
}

var DB *Database

// InitDatabase initializes the database with performance optimizations
func InitDatabase(filename string) (*Database, error) {
	DB = &Database{
		Users:              make(map[string]*User),
		Chats:              make(map[string]*Chat),
		Stats:              &Stats{
			StartTime: time.Now().Unix(),
			Commands:  make(map[string]int64),
		},
		Messages:           make(map[string]interface{}),
		Stickers:           make(map[string]interface{}),
		Settings:           make(map[string]interface{}),
		Responses:          make(map[string]interface{}),

		filename:        filename,
		dirty:           false,
		lastSave:        time.Now(),
		saveInterval:    5 * time.Minute, // Save every 5 minutes instead of 30 seconds
		maxUsers:        10000,           // Keep max 10k users in memory
		maxChats:        1000,            // Keep max 1k chats in memory
		cleanupInterval: 1 * time.Hour,   // Cleanup every hour
	}
	
	// Load existing data if file exists
	if _, err := os.Stat(filename); err == nil {
		if err := DB.Load(); err != nil {
			return nil, fmt.Errorf("failed to load database: %v", err)
		}
	}
	
	return DB, nil
}

// Load loads the database from file with compression support
func (db *Database) Load() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	
	data, err := os.ReadFile(db.filename)
	if err != nil {
		return err
	}
	
	// Try to decompress if it's gzipped
	if len(data) > 2 && data[0] == 0x1f && data[1] == 0x8b {
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return err
		}
		defer reader.Close()
		
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(reader); err != nil {
			return err
		}
		data = buf.Bytes()
	}
	
	return json.Unmarshal(data, db)
}

// Save saves the database to file with compression and performance optimizations
func (db *Database) Save() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	
	// Only save if data is dirty or enough time has passed
	if !db.dirty && time.Since(db.lastSave) < db.saveInterval {
		return nil
	}
	
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}
	
	// Compress data to reduce file size
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		return err
	}
	if err := gw.Close(); err != nil {
		return err
	}
	
	// Write to temporary file first, then rename for atomic operation
	tempFile := db.filename + ".tmp"
	if err := os.WriteFile(tempFile, buf.Bytes(), 0644); err != nil {
		return err
	}
	
	if err := os.Rename(tempFile, db.filename); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return err
	}
	
	db.dirty = false
	db.lastSave = time.Now()
	return nil
}

// GetUser gets or creates a user with performance optimizations
func (db *Database) GetUser(jid string) *User {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	
	user, exists := db.Users[jid]
	if !exists {
		// Check if we need to cleanup old users
		if len(db.Users) >= db.maxUsers {
			db.cleanupOldUsers()
		}
		
		user = &User{
			Name:        "",
			Age:         -1,
			RegTime:     -1,
			Registered:  false,
			Exp:         0,
			Level:       0,
			Role:        "Newbie ㋡",
			AutoLevelUp: true,
			Pasangan:    "",
			Warn:        0,
			Banned:      false,
			BannedUser:  false,
			BannedReason: "",
			LastPM:      0,
			AFK:         -1,
			AFKReason:   "",
			Premium:     false,
			PremiumTime: 0,
			PremiumDate: -1,
		}
		db.Users[jid] = user
		db.Stats.TotalUsers++
		db.dirty = true
	}
	
	return user
}

// GetChat gets or creates a chat with performance optimizations
func (db *Database) GetChat(jid string) *Chat {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	
	chat, exists := db.Chats[jid]
	if !exists {
		// Check if we need to cleanup old chats
		if len(db.Chats) >= db.maxChats {
			db.cleanupOldChats()
		}
		
		chat = &Chat{
			ID:           jid,
			Name:         "",
			IsBanned:     false,
			Welcome:      true,
			Detect:       true,
			SWelcome:     "",
			SBye:         "",
			SPromote:     "",
			SDemote:      "",
			Delete:       true,
			AntiLink:     false,
			AntiLink2:    false,
			AntiToxic:    false,
			AntiVirtex:   false,
			Viewonce:     true,
			LastActivity: time.Now().Unix(),
			MessageCount: 0,
			Game:         true,
		}
		db.Chats[jid] = chat
		db.Stats.TotalChats++
		db.dirty = true
	}
	
	return chat
}

// cleanupOldUsers removes inactive users to free memory
func (db *Database) cleanupOldUsers() {
	now := time.Now().Unix()
	cutoff := now - (30 * 24 * 60 * 60) // 30 days
	
	for jid, user := range db.Users {
		if user.LastPM < cutoff && !user.Premium {
			delete(db.Users, jid)
			db.Stats.TotalUsers--
		}
	}
}

// cleanupOldChats removes inactive chats to free memory
func (db *Database) cleanupOldChats() {
	now := time.Now().Unix()
	cutoff := now - (7 * 24 * 60 * 60) // 7 days
	
	for jid, chat := range db.Chats {
		if chat.LastActivity < cutoff {
			delete(db.Chats, jid)
			db.Stats.TotalChats--
		}
	}
}

// IncrementCommand increments command usage statistics
func (db *Database) IncrementCommand(command string) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	
	db.Stats.Commands[command]++
	db.dirty = true
}

// IncrementMessages increments total message count
func (db *Database) IncrementMessages() {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	
	db.Stats.TotalMessages++
	db.dirty = true
}

// GetUptime returns bot uptime in seconds
func (db *Database) GetUptime() int64 {
	return time.Now().Unix() - db.Stats.StartTime
}

// AutoSave starts automatic saving with improved performance
func (db *Database) AutoSave() {
	go func() {
		ticker := time.NewTicker(db.saveInterval)
		defer ticker.Stop()
		
		cleanupTicker := time.NewTicker(db.cleanupInterval)
		defer cleanupTicker.Stop()
		
		for {
			select {
			case <-ticker.C:
				if err := db.Save(); err != nil {
					fmt.Printf("Error auto-saving database: %v\n", err)
				}
			case <-cleanupTicker.C:
				db.mutex.Lock()
				db.cleanupOldUsers()
				db.cleanupOldChats()
				db.mutex.Unlock()
			}
		}
	}()
}

// ForceSave forces an immediate save regardless of dirty flag
func (db *Database) ForceSave() error {
	db.mutex.Lock()
	db.dirty = true
	db.mutex.Unlock()
	return db.Save()
}

// GetStats returns database statistics
func (db *Database) GetStats() *Stats {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	
	return db.Stats
}

// GetUserCount returns the number of users
func (db *Database) GetUserCount() int {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	
	return len(db.Users)
}

// GetChatCount returns the number of chats
func (db *Database) GetChatCount() int {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	
	return len(db.Chats)
}