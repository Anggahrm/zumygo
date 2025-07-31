package systems

import (
	"fmt"
	"zumygo/database"
)

// LevelingSystem handles all leveling-related operations
type LevelingSystem struct {
	db *database.Database
}

// NewLevelingSystem creates a new leveling system instance
func NewLevelingSystem(db *database.Database) *LevelingSystem {
	return &LevelingSystem{db: db}
}

// RoleInfo represents information about user roles
type RoleInfo struct {
	Name        string
	MinLevel    int
	Description string
	Emoji       string
	Perks       []string
}

var (
	// Role progression system
	Roles = []RoleInfo{
		{
			Name:        "Newbie ㋡",
			MinLevel:    0,
			Description: "Pemula yang baru bergabung",
			Emoji:       "🌱",
			Perks:       []string{"Basic commands access"},
		},
		{
			Name:        "Beginner ⚡",
			MinLevel:    5,
			Description: "Sudah mulai terbiasa",
			Emoji:       "⚡",
			Perks:       []string{"Basic commands", "Daily bonus +10%"},
		},
		{
			Name:        "Novice 🔰",
			MinLevel:    10,
			Description: "Pengguna aktif",
			Emoji:       "🔰",
			Perks:       []string{"All basic features", "Work bonus +15%"},
		},
		{
			Name:        "Apprentice 📚",
			MinLevel:    20,
			Description: "Sudah berpengalaman",
			Emoji:       "📚",
			Perks:       []string{"Mining bonus +20%", "Reduced cooldowns"},
		},
		{
			Name:        "Skilled 🎯",
			MinLevel:    35,
			Description: "Pengguna berpengalaman",
			Emoji:       "🎯",
			Perks:       []string{"Economy bonus +25%", "Special commands"},
		},
		{
			Name:        "Expert 💎",
			MinLevel:    50,
			Description: "Ahli dalam berbagai bidang",
			Emoji:       "💎",
			Perks:       []string{"All bonuses +30%", "VIP features"},
		},
		{
			Name:        "Master 👑",
			MinLevel:    75,
			Description: "Master pengguna bot",
			Emoji:       "👑",
			Perks:       []string{"Master privileges", "Custom commands"},
		},
		{
			Name:        "Grandmaster 🏆",
			MinLevel:    100,
			Description: "Grandmaster tertinggi",
			Emoji:       "🏆",
			Perks:       []string{"All privileges", "Unlimited access"},
		},
		{
			Name:        "Legend ⭐",
			MinLevel:    150,
			Description: "Legenda hidup",
			Emoji:       "⭐",
			Perks:       []string{"Legendary status", "Special recognition"},
		},
		{
			Name:        "Mythical 🌟",
			MinLevel:    200,
			Description: "Status mitos",
			Emoji:       "🌟",
			Perks:       []string{"Mythical powers", "Ultimate privileges"},
		},
	}
)

// CalculateRequiredExp calculates experience required for a level
func (ls *LevelingSystem) CalculateRequiredExp(level int) int64 {
	// Exponential growth formula: exp = level^2 * 100 + level * 50
	return int64(level*level*100 + level*50)
}

// CalculateLevelFromExp calculates level from total experience
func (ls *LevelingSystem) CalculateLevelFromExp(exp int64) int {
	level := 0
	totalExpNeeded := int64(0)
	
	for {
		expForNextLevel := ls.CalculateRequiredExp(level + 1)
		if totalExpNeeded+expForNextLevel > exp {
			break
		}
		totalExpNeeded += expForNextLevel
		level++
	}
	
	return level
}

// GetExpForCurrentLevel returns experience gained in current level
func (ls *LevelingSystem) GetExpForCurrentLevel(totalExp int64, level int) int64 {
	totalExpForPreviousLevels := int64(0)
	for i := 1; i <= level; i++ {
		totalExpForPreviousLevels += ls.CalculateRequiredExp(i)
	}
	
	return totalExp - totalExpForPreviousLevels
}

// GetRoleFromLevel returns the appropriate role for a level
func (ls *LevelingSystem) GetRoleFromLevel(level int) RoleInfo {
	role := Roles[0] // Default to first role
	
	for _, r := range Roles {
		if level >= r.MinLevel {
			role = r
		} else {
			break
		}
	}
	
	return role
}

// AddExperience adds experience to a user and handles level ups
func (ls *LevelingSystem) AddExperience(userJID string, expGain int64) (string, bool) {
	user := ls.db.GetUser(userJID)
	
	oldLevel := user.Level
	
	// Add experience
	user.Exp += expGain
	
	// Calculate new level
	newLevel := ls.CalculateLevelFromExp(user.Exp)
	leveledUp := newLevel > oldLevel
	
	if leveledUp {
		user.Level = newLevel
		
		// Update role
		newRole := ls.GetRoleFromLevel(newLevel)
		user.Role = newRole.Name
		
		// Auto level up message if enabled
		if user.AutoLevelUp {
			result := fmt.Sprintf("🎉 *LEVEL UP!*\n\n")
			result += fmt.Sprintf("⬆️ Level: %d → %d\n", oldLevel, newLevel)
			result += fmt.Sprintf("⭐ Experience: %d (+%d)\n", user.Exp, expGain)
			result += fmt.Sprintf("🏷️ New Role: %s %s\n", newRole.Emoji, newRole.Name)
			
			// Level up rewards
			coinReward := int64(newLevel * 100)
			zcReward := int64(newLevel * 5)
			
			user.Money += coinReward
			user.ZC += zcReward
			
			result += fmt.Sprintf("\n🎁 *Level Up Rewards:*\n")
			result += fmt.Sprintf("💰 Coins: +%d\n", coinReward)
			result += fmt.Sprintf("🪙 ZumyCoin: +%d ZC\n", zcReward)
			
			// Show role perks
			if len(newRole.Perks) > 0 {
				result += fmt.Sprintf("\n✨ *Role Perks:*\n")
				for _, perk := range newRole.Perks {
					result += fmt.Sprintf("• %s\n", perk)
				}
			}
			
			return result, true
		}
	}
	
	return "", leveledUp
}

// GetLevelInfo returns detailed level information for a user
func (ls *LevelingSystem) GetLevelInfo(userJID string) string {
	user := ls.db.GetUser(userJID)
	
	currentLevel := user.Level
	totalExp := user.Exp
	
	// Experience for current level
	expInCurrentLevel := ls.GetExpForCurrentLevel(totalExp, currentLevel)
	expNeededForNext := ls.CalculateRequiredExp(currentLevel + 1)
	
	// Progress percentage
	progressPercentage := float64(expInCurrentLevel) / float64(expNeededForNext) * 100
	
	// Role information
	currentRole := ls.GetRoleFromLevel(currentLevel)
	nextRole := ls.GetRoleFromLevel(currentLevel + 1)
	
	result := fmt.Sprintf("📊 *Level Information*\n\n")
	result += fmt.Sprintf("👤 **%s**\n", user.Name)
	result += fmt.Sprintf("⭐ Level: **%d**\n", currentLevel)
	result += fmt.Sprintf("🏷️ Role: %s **%s**\n", currentRole.Emoji, currentRole.Name)
	result += fmt.Sprintf("✨ Total Experience: **%d XP**\n\n", totalExp)
	
	// Progress bar
	progressBar := ls.generateProgressBar(int(progressPercentage))
	result += fmt.Sprintf("📈 **Progress to Level %d:**\n", currentLevel+1)
	result += fmt.Sprintf("%s %.1f%%\n", progressBar, progressPercentage)
	result += fmt.Sprintf("🎯 %d / %d XP\n\n", expInCurrentLevel, expNeededForNext)
	
	// Next role info
	if nextRole.Name != currentRole.Name {
		result += fmt.Sprintf("🎯 **Next Role:** %s %s\n", nextRole.Emoji, nextRole.Name)
		result += fmt.Sprintf("📍 Required Level: %d\n", nextRole.MinLevel)
	}
	
	// Current role perks
	if len(currentRole.Perks) > 0 {
		result += fmt.Sprintf("\n✨ **Current Perks:**\n")
		for _, perk := range currentRole.Perks {
			result += fmt.Sprintf("• %s\n", perk)
		}
	}
	
	return result
}

// generateProgressBar creates a visual progress bar
func (ls *LevelingSystem) generateProgressBar(percentage int) string {
	bars := 15
	filledBars := (percentage * bars) / 100
	
	var progressBar string
	for i := 0; i < bars; i++ {
		if i < filledBars {
			progressBar += "█"
		} else {
			progressBar += "░"
		}
	}
	
	return progressBar
}

// GetLeaderboard returns top users by level
func (ls *LevelingSystem) GetLeaderboard() string {
	type LevelEntry struct {
		Name  string
		Level int
		Exp   int64
		Role  string
	}
	
	var entries []LevelEntry
	
	// Collect level data
	for _, user := range ls.db.Users {
		if user.Name != "" {
			entries = append(entries, LevelEntry{
				Name:  user.Name,
				Level: user.Level,
				Exp:   user.Exp,
				Role:  user.Role,
			})
		}
	}
	
	// Sort by level (descending), then by exp
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Level > entries[i].Level || 
			   (entries[j].Level == entries[i].Level && entries[j].Exp > entries[i].Exp) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	
	result := "🏆 *Level Leaderboard*\n\n"
	
	for i, entry := range entries {
		if i >= 10 { // Top 10 only
			break
		}
		
		rank := i + 1
		var medal string
		switch rank {
		case 1:
			medal = "🥇"
		case 2:
			medal = "🥈"
		case 3:
			medal = "🥉"
		default:
			medal = fmt.Sprintf("%d.", rank)
		}
		
		// Get role emoji
		roleInfo := ls.GetRoleFromLevel(entry.Level)
		
		result += fmt.Sprintf("%s **%s**\n", medal, entry.Name)
		result += fmt.Sprintf("   ⭐ Level %d | ✨ %d XP\n", entry.Level, entry.Exp)
		result += fmt.Sprintf("   %s %s\n\n", roleInfo.Emoji, entry.Role)
	}
	
	return result
}

// ToggleAutoLevelUp toggles auto level up notifications for a user
func (ls *LevelingSystem) ToggleAutoLevelUp(userJID string) string {
	user := ls.db.GetUser(userJID)
	
	user.AutoLevelUp = !user.AutoLevelUp
	
	status := "disabled"
	if user.AutoLevelUp {
		status = "enabled"
	}
	
	return fmt.Sprintf("🔔 Auto level up notifications %s!", status)
}

// GetRoleList returns information about all available roles
func (ls *LevelingSystem) GetRoleList() string {
	result := "🏷️ *Available Roles*\n\n"
	
	for _, role := range Roles {
		result += fmt.Sprintf("%s **%s**\n", role.Emoji, role.Name)
		result += fmt.Sprintf("📍 Required Level: %d\n", role.MinLevel)
		result += fmt.Sprintf("📝 %s\n", role.Description)
		
		if len(role.Perks) > 0 {
			result += fmt.Sprintf("✨ Perks:\n")
			for _, perk := range role.Perks {
				result += fmt.Sprintf("  • %s\n", perk)
			}
		}
		result += "\n"
	}
	
	return result
}

// GetExpMultiplier returns experience multiplier based on user level and role
func (ls *LevelingSystem) GetExpMultiplier(userJID string) float64 {
	user := ls.db.GetUser(userJID)
	role := ls.GetRoleFromLevel(user.Level)
	
	// Base multiplier
	multiplier := 1.0
	
	// Role-based bonuses
	switch role.MinLevel {
	case 0: // Newbie
		multiplier = 1.0
	case 5: // Beginner
		multiplier = 1.1
	case 10: // Novice
		multiplier = 1.15
	case 20: // Apprentice
		multiplier = 1.2
	case 35: // Skilled
		multiplier = 1.25
	case 50: // Expert
		multiplier = 1.3
	case 75: // Master
		multiplier = 1.4
	case 100: // Grandmaster
		multiplier = 1.5
	case 150: // Legend
		multiplier = 1.75
	case 200: // Mythical
		multiplier = 2.0
	}
	
	// Premium bonus
	if user.Premium {
		multiplier *= 1.5
	}
	
	return multiplier
}

// CalculateExpGain calculates experience gain with bonuses applied
func (ls *LevelingSystem) CalculateExpGain(userJID string, baseExp int64) int64 {
	multiplier := ls.GetExpMultiplier(userJID)
	return int64(float64(baseExp) * multiplier)
}

// InitializeLevelingSystem initializes the leveling system
func InitializeLevelingSystem(db *database.Database) *LevelingSystem {
	ls := NewLevelingSystem(db)
	
	// Update all users' levels and roles based on their current experience
	for _, user := range db.Users {
		correctLevel := ls.CalculateLevelFromExp(user.Exp)
		if correctLevel != user.Level {
			user.Level = correctLevel
			user.Role = ls.GetRoleFromLevel(correctLevel).Name
		}
	}
	
	return ls
}