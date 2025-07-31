package config

import (
	"encoding/json"
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
	Multiplier int      `json:"multiplier"`
	Prefix     string   `json:"prefix"`
	Prefixes   []string `json:"prefixes"`
	
	// Bot Mode Settings
	PublicMode  bool `json:"public_mode"`
	ReadStatus  bool `json:"read_status"`
	ReactStatus bool `json:"react_status"`
	
	// Database Settings
	DatabaseURL string `json:"database_url"`
	
	// WhatsApp Settings
	PairingNumber string `json:"pairing_number"`
	SessionName   string `json:"session_name"`
}

var Config *BotConfig

// LoadConfig loads configuration with static defaults
func LoadConfig() *BotConfig {
	config := &BotConfig{
		// Owner Settings - static defaults
		Owner:       []string{"6285123865643"},
		Mods:        []string{},
		Prems:       []string{},
		MaxWarn:     2,
		NameOwner:   "anggahrm",
		NumberOwner: "6285123865643",
		NumberBot:   "6281253216363",
		Mail:        "anggahrm@gmail.com",
		
		// Payment Info
		Dana:  "6285123865643",
		Pulsa: "6285123865643",
		Gopay: "6287842262440",
		
		// Bot Identity
		NameBot:    "ZumyNext",
		ChName:     "arsip raja iblis",
		Newsletter: "120363233010187559@newsletter",
		SGH:        "https://github.com/Anggahrm",
		GC:         "https://chat.whatsapp.com/GrbvuguKCTPH4mIuRCouV8",
		
		// Panel Settings
		PanelDomain: "https://panel.zumynext.tech",
		PanelAPI:    "ptla_qf9ZcQGGVK3P5hJA7McoPLzKj25EsXyuMiU1Qpo0u0Z",
		PanelCAPI:   "ptla_qf9ZcQGGVK3P5hJA7McoPLzKj25EsXyuMiU1Qpo0u0Z",
		SubZone:     "4eebdf1935753d55d75f89c93be43826",
		SubToken:    "f961b7c159e0d0a6543017434db1a1bd5d490",
		SubDomain:   "anggahrm.my.id",
		Web:         "https://anggahrm.my.id",
		
		// Media URLs
		Zumy:    "https://telegra.ph/file/ab4026bea092a1d640e78.jpg",
		ZumyGif: "https://telegra.ph/file/b02d8a1ba8a6cca24d120.mp4",
		Zum:     "https://telegra.ph/file/df3a9789f4d9362581999.jpg",
		Iwel:    "https://telegra.ph/file/a5487661986e69b3ff0cc.jpg",
		Ilea:    "https://telegra.ph/file/df46938b10101dfd5b521.jpg",
		
		// Social Media
		Instagram: "https://instagram.com/zumy.xyz",
		
		// Bot Messages
		Zums:       "Powered by ZumyNext",
		WM:         "© ZumyNext",
		Watermark:  "© ZumyNext",
		WM2:        "⫹⫺ ZumyNext",
		Wait:       "_*Tunggu sedang di proses...*_",
		EWait:      "⏱️",
		Error:      "_*Server Error*_",
		Benar:      "Benar ✅\n",
		EDone:      "✅",
		Salah:      "Salah ❌\n",
		EError:     "❌",
		StikerWait: "*⫹⫺ Stiker sedang dibuat...*",
		
		// Sticker Settings
		PackName: "🌿",
		Author:   "anggahrm",
		
		// API Keys
		AlpisKey: "4d1507f7",
		FG:       "798f2bf0",
		BTC:      "Mark-HDR",
		Lann:     "anggagtg",
		LolKey:   "SGWN",
		
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
		Multiplier: 45,
		Prefix:     ".",
		Prefixes:   []string{".", "!", "#", "$", "&", "?"},
		
		// Bot Mode Settings
		PublicMode:  false, // Private mode by default
		ReadStatus:  true,  // Auto-read status enabled by default
		ReactStatus: true,  // Auto-react status enabled by default
		
		// Database Settings
		DatabaseURL: "",
		
		// WhatsApp Settings
		PairingNumber: "",
		SessionName:   "session",
	}
	
	Config = config
	return config
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

// TogglePublicMode toggles the public mode setting
func (c *BotConfig) TogglePublicMode() {
	c.PublicMode = !c.PublicMode
}

// GetPrefixes returns all valid prefixes
func (c *BotConfig) GetPrefixes() []string {
	if len(c.Prefixes) > 0 {
		return c.Prefixes
	}
	return []string{c.Prefix}
}

// ToJSON converts config to JSON string
func (c *BotConfig) ToJSON() (string, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}