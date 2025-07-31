package database

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// User represents a user in the database
type User struct {
	// Basic Info
	Name         string    `json:"name"`
	Age          int       `json:"age"`
	RegTime      int64     `json:"regTime"`
	Registered   bool      `json:"registered"`
	
	// Economy
	Money        int64     `json:"money"`
	ZC           int64     `json:"zc"`           // ZumyCoin
	ATM          int64     `json:"atm"`
	Bank         int64     `json:"bank"`
	
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
	LastClaim    int64     `json:"lastclaim"`
	LastRob      int64     `json:"lastrob"`
	LastWork     int64     `json:"lastwork"`
	LastPM       int64     `json:"lastpm"`
	AFK          int64     `json:"afk"`
	AFKReason    string    `json:"afkReason"`
	
	// Premium
	Premium      bool      `json:"premium"`
	PremiumTime  int64     `json:"premiumTime"`
	PremiumDate  int64     `json:"premiumDate"`
	
	// Mining System
	Mining       *MiningData `json:"mining,omitempty"`
	
	// Health System
	Health       *HealthData `json:"health,omitempty"`
	
	// Inventory
	Inventory    map[string]int64 `json:"inventory,omitempty"`
}

// MiningData represents user mining information
type MiningData struct {
	// Pickaxes
	WoodenPickaxe int64 `json:"woodenPickaxe"`
	StonePickaxe  int64 `json:"stonePickaxe"`
	IronPickaxe   int64 `json:"ironPickaxe"`
	GoldPickaxe   int64 `json:"goldPickaxe"`
	DiamondPickaxe int64 `json:"diamondPickaxe"`
	
	// Ores
	Coal        int64 `json:"coal"`
	Iron        int64 `json:"iron"`
	Gold        int64 `json:"gold"`
	Diamond     int64 `json:"diamond"`
	Emerald     int64 `json:"emerald"`
	
	// Mining Stats
	LastMine    int64 `json:"lastMine"`
	TotalMined  int64 `json:"totalMined"`
	MiningLevel int   `json:"miningLevel"`
	MiningExp   int64 `json:"miningExp"`
}

// HealthData represents user health information
type HealthData struct {
	Health         int64 `json:"health"`
	MaxHealth      int64 `json:"maxHealth"`
	LastRegenTime  int64 `json:"lastRegenTime"`
	HealthPotions  int64 `json:"healthPotions"`
	LastDamage     int64 `json:"lastDamage"`
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
	
	// Economy
	Economy     bool  `json:"economy"`
	
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

// OreStock represents ore market stock
type OreStock struct {
	Coal        int64   `json:"coal"`
	Iron        int64   `json:"iron"`
	Gold        int64   `json:"gold"`
	Diamond     int64   `json:"diamond"`
	Emerald     int64   `json:"emerald"`
	LastUpdate  int64   `json:"lastUpdate"`
	Prices      map[string]int64 `json:"prices"`
}

// PickaxeShop represents pickaxe shop stock
type PickaxeShop struct {
	WoodenPickaxe  int64 `json:"woodenPickaxe"`
	StonePickaxe   int64 `json:"stonePickaxe"`
	IronPickaxe    int64 `json:"ironPickaxe"`
	GoldPickaxe    int64 `json:"goldPickaxe"`
	DiamondPickaxe int64 `json:"diamondPickaxe"`
	LastUpdate     int64 `json:"lastUpdate"`
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
	OreStock           []OreStock       `json:"oreStock"`
	LastOreStockUpdate int64            `json:"lastOreStockUpdate"`
	PickaxeShopStock   []PickaxeShop    `json:"pickaxeShopStock"`
	LastPickaxeUpdate  int64            `json:"lastPickaxeShopUpdate"`
	
	// Internal
	mutex    sync.RWMutex `json:"-"`
	filename string       `json:"-"`
}

var DB *Database

// InitDatabase initializes the database
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
		OreStock:           []OreStock{},
		PickaxeShopStock:   []PickaxeShop{},
		filename:           filename,
	}
	
	// Load existing data if file exists
	if _, err := os.Stat(filename); err == nil {
		if err := DB.Load(); err != nil {
			return nil, fmt.Errorf("failed to load database: %v", err)
		}
	}
	
	// Initialize default data if needed
	DB.initializeDefaults()
	
	return DB, nil
}

// Load loads the database from file
func (db *Database) Load() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	
	data, err := os.ReadFile(db.filename)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(data, db)
}

// Save saves the database to file
func (db *Database) Save() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(db.filename, data, 0644)
}

// GetUser gets or creates a user
func (db *Database) GetUser(jid string) *User {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	
	user, exists := db.Users[jid]
	if !exists {
		user = &User{
			Name:        "",
			Age:         -1,
			RegTime:     -1,
			Registered:  false,
			Money:       100,
			ZC:          15,
			ATM:         0,
			Bank:        0,
			Exp:         0,
			Level:       0,
			Role:        "Newbie ㋡",
			AutoLevelUp: true,
			Pasangan:    "",
			Warn:        0,
			Banned:      false,
			BannedUser:  false,
			BannedReason: "",
			LastClaim:   0,
			LastRob:     0,
			LastWork:    0,
			LastPM:      0,
			AFK:         -1,
			AFKReason:   "",
			Premium:     false,
			PremiumTime: 0,
			PremiumDate: -1,
			Mining:      db.initializeMiningData(),
			Health:      db.initializeHealthData(),
			Inventory:   make(map[string]int64),
		}
		db.Users[jid] = user
		db.Stats.TotalUsers++
	}
	
	return user
}

// GetChat gets or creates a chat
func (db *Database) GetChat(jid string) *Chat {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	
	chat, exists := db.Chats[jid]
	if !exists {
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
			Economy:      true,
			Game:         true,
		}
		db.Chats[jid] = chat
		db.Stats.TotalChats++
	}
	
	return chat
}

// initializeMiningData creates initial mining data for a user
func (db *Database) initializeMiningData() *MiningData {
	return &MiningData{
		WoodenPickaxe:  1, // Start with 1 wooden pickaxe
		StonePickaxe:   0,
		IronPickaxe:    0,
		GoldPickaxe:    0,
		DiamondPickaxe: 0,
		Coal:           0,
		Iron:           0,
		Gold:           0,
		Diamond:        0,
		Emerald:        0,
		LastMine:       0,
		TotalMined:     0,
		MiningLevel:    1,
		MiningExp:      0,
	}
}

// initializeHealthData creates initial health data for a user
func (db *Database) initializeHealthData() *HealthData {
	return &HealthData{
		Health:        100,
		MaxHealth:     100,
		LastRegenTime: time.Now().Unix(),
		HealthPotions: 0,
		LastDamage:    0,
	}
}

// initializeDefaults initializes default data
func (db *Database) initializeDefaults() {
	// Initialize ore stock if empty
	if len(db.OreStock) == 0 {
		db.OreStock = []OreStock{
			{
				Coal:       1000,
				Iron:       500,
				Gold:       200,
				Diamond:    50,
				Emerald:    25,
				LastUpdate: time.Now().Unix(),
				Prices: map[string]int64{
					"coal":    10,
					"iron":    25,
					"gold":    100,
					"diamond": 500,
					"emerald": 1000,
				},
			},
		}
	}
	
	// Initialize pickaxe shop if empty
	if len(db.PickaxeShopStock) == 0 {
		db.PickaxeShopStock = []PickaxeShop{
			{
				WoodenPickaxe:  100,
				StonePickaxe:   50,
				IronPickaxe:    25,
				GoldPickaxe:    10,
				DiamondPickaxe: 5,
				LastUpdate:     time.Now().Unix(),
			},
		}
	}
}

// IncrementCommand increments command usage statistics
func (db *Database) IncrementCommand(command string) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	
	db.Stats.Commands[command]++
}

// IncrementMessages increments total message count
func (db *Database) IncrementMessages() {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	
	db.Stats.TotalMessages++
}

// GetUptime returns bot uptime in seconds
func (db *Database) GetUptime() int64 {
	return time.Now().Unix() - db.Stats.StartTime
}

// AutoSave starts automatic saving every 30 seconds
func (db *Database) AutoSave() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		
		for range ticker.C {
			if err := db.Save(); err != nil {
				fmt.Printf("Error auto-saving database: %v\n", err)
			}
		}
	}()
}