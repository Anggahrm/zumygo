package commands

import (
	"fmt"
	"strings"
	"time"
	"zumygo/config"
	"zumygo/libs"

	"go.mau.fi/whatsmeow"
)

// Global variables for bio system
var (
	lastBioUpdate time.Time
	bioTicker     *time.Ticker
)

// BioData holds dynamic data for bio template
type BioData struct {
	Time     string
	Status   string
	Web      string
	Uptime   string
	Commands int
	Users    int
	Groups   int
}

func init() {
	libs.NewCommands(&libs.ICommand{
		Before: func(conn *libs.IClient, m *libs.IMessage) {
			// Check if auto update bio is enabled
			cfg := config.Config
			if !cfg.AutoUpdateBio {
				return
			}

			// Check if it's time to update bio (every BioInterval minutes)
			now := time.Now()
			if now.Sub(lastBioUpdate) < time.Duration(cfg.BioInterval)*time.Minute {
				return
			}

			// Update bio
			updateBio(conn, cfg)
			lastBioUpdate = now
		},
	})
}

// updateBio performs the actual bio update
func updateBio(conn *libs.IClient, cfg *config.BotConfig) {
	// Generate bio data
	bioData := generateBioData(cfg)

	// Process template
	bioText := processTemplate(cfg.BioTemplate, bioData)

	// Update profile
	err := conn.WA.SetStatusMessage(bioText)
	if err != nil {
		// Log error but don't panic
		return
	}
}

// generateBioData generates dynamic data for bio template
func generateBioData(cfg *config.BotConfig) *BioData {
	now := time.Now()
	
	return &BioData{
		Time:     now.Format("15:04"),
		Status:   "🟢 Online",
		Web:      cfg.Web,
		Uptime:   getUptime(),
		Commands: getCommandCount(),
		Users:    getUserCount(),
		Groups:   getGroupCount(),
	}
}

// processTemplate processes the bio template with dynamic data
func processTemplate(template string, data *BioData) string {
	result := template

	// Replace placeholders with actual data
	replacements := map[string]string{
		"{time}":     data.Time,
		"{status}":   data.Status,
		"{web}":      data.Web,
		"{uptime}":   data.Uptime,
		"{commands}": fmt.Sprintf("%d", data.Commands),
		"{users}":    fmt.Sprintf("%d", data.Users),
		"{groups}":   fmt.Sprintf("%d", data.Groups),
	}

	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// getUptime returns bot uptime in a readable format
func getUptime() string {
	// This would need to be implemented with actual uptime tracking
	// For now, return a placeholder
	return "24h"
}

// getCommandCount returns the number of available commands
func getCommandCount() int {
	// This would need to be implemented with actual command counting
	// For now, return a placeholder
	return 50
}

// getUserCount returns the number of users in database
func getUserCount() int {
	// This would need to be implemented with actual database query
	// For now, return a placeholder
	return 100
}

// getGroupCount returns the number of groups in database
func getGroupCount() int {
	// This would need to be implemented with actual database query
	// For now, return a placeholder
	return 25
}

// Global bio system instance for commands
var globalBioSystem *BioSystem

// BioSystem for command control (simplified)
type BioSystem struct {
	cfg *config.BotConfig
}

// InitializeBioSystem creates a new bio system
func InitializeBioSystem(cfg *config.BotConfig, logger interface{}) *BioSystem {
	return &BioSystem{
		cfg: cfg,
	}
}

// SetGlobalBioSystem sets the global bio system instance
func SetGlobalBioSystem(bs *BioSystem) {
	globalBioSystem = bs
}

// GetGlobalBioSystem returns the global bio system instance
func GetGlobalBioSystem() *BioSystem {
	return globalBioSystem
}

// SetClient sets the WhatsApp client for bio updates (not needed in this approach)
func (bs *BioSystem) SetClient(client *whatsmeow.Client) {
	// Not needed in Before hook approach
}

// Start starts the auto bio update system (not needed in this approach)
func (bs *BioSystem) Start() error {
	// Not needed in Before hook approach
	return nil
}

// Stop stops the auto bio update system (not needed in this approach)
func (bs *BioSystem) Stop() error {
	// Not needed in Before hook approach
	return nil
}

// IsRunning checks if the bio system is running (not needed in this approach)
func (bs *BioSystem) IsRunning() bool {
	// Not needed in Before hook approach
	return true
}

// UpdateBioNow forces an immediate bio update
func (bs *BioSystem) UpdateBioNow() error {
	// This would need to be implemented to force update
	return nil
}

// SetBioTemplate sets a new bio template
func (bs *BioSystem) SetBioTemplate(template string) {
	bs.cfg.BioTemplate = template
}

// SetBioInterval sets a new bio update interval
func (bs *BioSystem) SetBioInterval(minutes int) error {
	if minutes < 1 {
		return fmt.Errorf("interval must be at least 1 minute")
	}
	bs.cfg.BioInterval = minutes
	return nil
}

// ToggleAutoUpdate toggles the auto update bio setting
func (bs *BioSystem) ToggleAutoUpdate() bool {
	bs.cfg.AutoUpdateBio = !bs.cfg.AutoUpdateBio
	return bs.cfg.AutoUpdateBio
}

// GetStatus returns the current status of the bio system
func (bs *BioSystem) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"enabled":  bs.cfg.AutoUpdateBio,
		"running":  true, // Always running in Before hook approach
		"template": bs.cfg.BioTemplate,
		"interval": bs.cfg.BioInterval,
	}
} 