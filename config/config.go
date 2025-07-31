package config

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

// BotConfig holds all bot configuration
type BotConfig struct {
	// Owner and Admin Settings
	Owner       []string `json:"owner"`
	Mods        []string `json:"mods"`
	Prems       []string `json:"prems"`
	MaxWarn     int      `json:"maxwarn"`
	NameOwner   string   `json:"nameowner"`
	NumberOwner string   `json:"numberowner"`
	NumberBot   string   `json:"numberbot"`
	Mail        string   `json:"mail"`
	
	// Payment Info
	Dana   string `json:"dana"`
	Pulsa  string `json:"pulsa"`
	Gopay  string `json:"gopay"`
	
	// Bot Identity
	NameBot     string `json:"namebot"`
	ChName      string `json:"chname"`
	Newsletter  string `json:"newsletter"`
	SGH         string `json:"sgh"`
	GC          string `json:"gc"`
	
	// Panel and Domain Settings
	PanelDomain string `json:"paneldomain"`
	PanelAPI    string `json:"panelapi"`
	PanelCAPI   string `json:"panelcapi"`
	SubZone     string `json:"subzone"`
	SubToken    string `json:"subtoken"`
	SubDomain   string `json:"subdomain"`
	Web         string `json:"web"`
	
	// Media URLs
	Zumy    string `json:"zumy"`
	ZumyGif string `json:"zumygif"`
	Zum     string `json:"zum"`
	Iwel    string `json:"iwel"`
	Ilea    string `json:"ilea"`
	
	// Social Media
	Instagram string `json:"instagram"`
	
	// Bot Messages
	Zums       string `json:"zums"`
	WM         string `json:"wm"`
	Watermark  string `json:"watermark"`
	WM2        string `json:"wm2"`
	Wait       string `json:"wait"`
	EWait      string `json:"ewait"`
	Error      string `json:"eror"`
	Benar      string `json:"benar"`
	EDone      string `json:"edone"`
	Salah      string `json:"salah"`
	EError     string `json:"eerror"`
	StikerWait string `json:"stiker_wait"`
	
	// Sticker Settings
	PackName string `json:"packname"`
	Author   string `json:"author"`
	
	// API Keys
	AlpisKey string            `json:"alpiskey"`
	FG       string            `json:"fg"`
	BTC      string            `json:"btc"`
	Lann     string            `json:"lann"`
	LolKey   string            `json:"lolkey"`
	APIs     map[string]string `json:"apis"`
	APIKeys  map[string]string `json:"apikeys"`
	
	// System Settings
	Multiplier int    `json:"multiplier"`
	Prefix     string `json:"prefix"`
	
	// Database Settings
	DatabaseURL string `json:"database_url"`
	
	// WhatsApp Settings
	PairingNumber string `json:"pairing_number"`
	SessionName   string `json:"session_name"`
}

var Config *BotConfig

// LoadConfig loads configuration from environment variables and defaults
func LoadConfig() *BotConfig {
	config := &BotConfig{
		// Owner Settings - from env or defaults
		Owner:       getStringSlice("OWNER", []string{"6285123865643"}),
		Mods:        getStringSlice("MODS", []string{}),
		Prems:       getStringSlice("PREMS", []string{}),
		MaxWarn:     getIntEnv("MAX_WARN", 2),
		NameOwner:   getEnv("NAME_OWNER", "anggahrm"),
		NumberOwner: getEnv("NUMBER_OWNER", "6285123865643"),
		NumberBot:   getEnv("NUMBER_BOT", "6281253216363"),
		Mail:        getEnv("MAIL", "anggahrm@gmail.com"),
		
		// Payment Info
		Dana:  getEnv("DANA", "6285123865643"),
		Pulsa: getEnv("PULSA", "6285123865643"),
		Gopay: getEnv("GOPAY", "6287842262440"),
		
		// Bot Identity
		NameBot:    getEnv("NAME_BOT", "ZumyNext"),
		ChName:     getEnv("CH_NAME", "arsip raja iblis"),
		Newsletter: getEnv("NEWSLETTER", "120363233010187559@newsletter"),
		SGH:        getEnv("SGH", "https://github.com/Anggahrm"),
		GC:         getEnv("GC", "https://chat.whatsapp.com/GrbvuguKCTPH4mIuRCouV8"),
		
		// Panel Settings
		PanelDomain: getEnv("PANEL_DOMAIN", "https://panel.zumynext.tech"),
		PanelAPI:    getEnv("PANEL_API", "ptla_qf9ZcQGGVK3P5hJA7McoPLzKj25EsXyuMiU1Qpo0u0Z"),
		PanelCAPI:   getEnv("PANEL_CAPI", "ptla_qf9ZcQGGVK3P5hJA7McoPLzKj25EsXyuMiU1Qpo0u0Z"),
		SubZone:     getEnv("SUB_ZONE", "4eebdf1935753d55d75f89c93be43826"),
		SubToken:    getEnv("SUB_TOKEN", "f961b7c159e0d0a6543017434db1a1bd5d490"),
		SubDomain:   getEnv("SUB_DOMAIN", "anggahrm.my.id"),
		Web:         getEnv("WEB", "https://anggahrm.my.id"),
		
		// Media URLs
		Zumy:    getEnv("ZUMY", "https://telegra.ph/file/ab4026bea092a1d640e78.jpg"),
		ZumyGif: getEnv("ZUMY_GIF", "https://telegra.ph/file/b02d8a1ba8a6cca24d120.mp4"),
		Zum:     getEnv("ZUM", "https://telegra.ph/file/df3a9789f4d9362581999.jpg"),
		Iwel:    getEnv("IWEL", "https://telegra.ph/file/a5487661986e69b3ff0cc.jpg"),
		Ilea:    getEnv("ILEA", "https://telegra.ph/file/df46938b10101dfd5b521.jpg"),
		
		// Social Media
		Instagram: getEnv("INSTAGRAM", "https://instagram.com/zumy.xyz"),
		
		// Bot Messages
		Zums:       getEnv("ZUMS", "Powered by ZumyNext"),
		WM:         getEnv("WM", "© ZumyNext"),
		Watermark:  getEnv("WATERMARK", "© ZumyNext"),
		WM2:        getEnv("WM2", "⫹⫺ ZumyNext"),
		Wait:       getEnv("WAIT", "_*Tunggu sedang di proses...*_"),
		EWait:      getEnv("EWAIT", "⏱️"),
		Error:      getEnv("ERROR", "_*Server Error*_"),
		Benar:      getEnv("BENAR", "Benar ✅\n"),
		EDone:      getEnv("EDONE", "✅"),
		Salah:      getEnv("SALAH", "Salah ❌\n"),
		EError:     getEnv("EERROR", "❌"),
		StikerWait: getEnv("STIKER_WAIT", "*⫹⫺ Stiker sedang dibuat...*"),
		
		// Sticker Settings
		PackName: getEnv("PACK_NAME", "🌿"),
		Author:   getEnv("AUTHOR", "anggahrm"),
		
		// API Keys
		AlpisKey: getEnv("ALPIS_KEY", "4d1507f7"),
		FG:       getEnv("FG", "798f2bf0"),
		BTC:      getEnv("BTC", "Mark-HDR"),
		Lann:     getEnv("LANN", "anggagtg"),
		LolKey:   getEnv("LOL_KEY", "SGWN"),
		
		// APIs
		APIs: map[string]string{
			"tio":     "https://api.botcahx.eu.org",
			"lann":    "https://api.betabotz.eu.org",
			"fgmods":  "https://api-fgmods.ddns.net",
			"lol":     "https://api-lolhuman.xyz",
			"alpis":   "https://alpis.eu.org",
			"akuari":  "https://api.akuari.my.id",
		},
		
		// API Keys mapping
		APIKeys: map[string]string{
			"https://api.botcahx.eu.org":    "Mark-HDR",
			"https://api-fgmods.ddns.net":   "798f2bf0",
			"https://api-lolhuman.xyz":      "SGWN",
			"https://api.betabotz.eu.org":   "zumyXD",
		},
		
		// System Settings
		Multiplier: getIntEnv("MULTIPLIER", 45),
		Prefix:     getEnv("PREFIX", "."),
		
		// Database Settings
		DatabaseURL: getEnv("DATABASE_URL", ""),
		
		// WhatsApp Settings
		PairingNumber: getEnv("PAIRING_NUMBER", ""),
		SessionName:   getEnv("SESSION_NAME", "session"),
	}
	
	Config = config
	return config
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// API builds API URL with query parameters
func (c *BotConfig) API(name, path string, query map[string]string) string {
	baseURL, exists := c.APIs[name]
	if !exists {
		baseURL = name
	}
	
	url := baseURL + path
	
	if len(query) > 0 || c.APIKeys[baseURL] != "" {
		url += "?"
		
		// Add API key if exists
		if apiKey := c.APIKeys[baseURL]; apiKey != "" {
			url += "apikey=" + apiKey
			if len(query) > 0 {
				url += "&"
			}
		}
		
		// Add query parameters
		for key, value := range query {
			url += key + "=" + value + "&"
		}
		
		// Remove trailing &
		if strings.HasSuffix(url, "&") {
			url = url[:len(url)-1]
		}
	}
	
	return url
}

// IsOwner checks if the given number is an owner
func (c *BotConfig) IsOwner(number string) bool {
	for _, owner := range c.Owner {
		if owner == number {
			return true
		}
	}
	return false
}

// IsMod checks if the given number is a moderator
func (c *BotConfig) IsMod(number string) bool {
	for _, mod := range c.Mods {
		if mod == number {
			return true
		}
	}
	return false
}

// IsPrem checks if the given number is premium
func (c *BotConfig) IsPrem(number string) bool {
	for _, prem := range c.Prems {
		if prem == number {
			return true
		}
	}
	return false
}

// ToJSON converts config to JSON string
func (c *BotConfig) ToJSON() (string, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}