package systems

import (
	"fmt"
	"time"
	"zumygo/database"
)

// HealthSystem handles all health-related operations
type HealthSystem struct {
	db *database.Database
}

// NewHealthSystem creates a new health system instance
func NewHealthSystem(db *database.Database) *HealthSystem {
	return &HealthSystem{db: db}
}

// PotionInfo represents information about health potions
type PotionInfo struct {
	Name        string
	HealAmount  int64
	Price       int64
	Description string
	Emoji       string
}

var (
	// Potion types with their properties
	PotionTypes = map[string]PotionInfo{
		"small": {
			Name:        "Small Health Potion",
			HealAmount:  25,
			Price:       50,
			Description: "Restores 25 HP",
			Emoji:       "🧪",
		},
		"medium": {
			Name:        "Medium Health Potion",
			HealAmount:  50,
			Price:       100,
			Description: "Restores 50 HP",
			Emoji:       "🍶",
		},
		"large": {
			Name:        "Large Health Potion",
			HealAmount:  100,
			Price:       200,
			Description: "Restores 100 HP",
			Emoji:       "⚗️",
		},
		"mega": {
			Name:        "Mega Health Potion",
			HealAmount:  200,
			Price:       500,
			Description: "Restores 200 HP",
			Emoji:       "🧬",
		},
	}
)

// RegenerateHealth regenerates health for a user over time
func (hs *HealthSystem) RegenerateHealth(userJID string) {
	user := hs.db.GetUser(userJID)
	health := user.Health
	
	now := time.Now().Unix()
	timeSinceLastRegen := now - health.LastRegenTime
	
	// Regenerate 1 HP every 5 minutes
	regenAmount := timeSinceLastRegen / 300 // 300 seconds = 5 minutes
	
	if regenAmount > 0 && health.Health < health.MaxHealth {
		newHealth := health.Health + regenAmount
		if newHealth > health.MaxHealth {
			newHealth = health.MaxHealth
		}
		
		health.Health = newHealth
		health.LastRegenTime = now
	}
}

// GetHealthInfo returns user's health information
func (hs *HealthSystem) GetHealthInfo(userJID string) string {
	user := hs.db.GetUser(userJID)
	health := user.Health
	
	// Regenerate health first
	hs.RegenerateHealth(userJID)
	
	result := "❤️ *Your Health Status*\n\n"
	
	// Health bar visualization
	healthPercentage := float64(health.Health) / float64(health.MaxHealth) * 100
	healthBar := hs.generateHealthBar(int(healthPercentage))
	
	result += fmt.Sprintf("❤️ Health: %d/%d HP\n", health.Health, health.MaxHealth)
	result += fmt.Sprintf("%s %.1f%%\n\n", healthBar, healthPercentage)
	
	// Health status
	var status string
	if healthPercentage >= 80 {
		status = "🟢 Excellent"
	} else if healthPercentage >= 60 {
		status = "🟡 Good"
	} else if healthPercentage >= 40 {
		status = "🟠 Fair"
	} else if healthPercentage >= 20 {
		status = "🔴 Poor"
	} else {
		status = "💀 Critical"
	}
	
	result += fmt.Sprintf("📊 Status: %s\n", status)
	result += fmt.Sprintf("🧪 Health Potions: %d\n", health.HealthPotions)
	
	// Regeneration info
	now := time.Now().Unix()
	timeSinceLastRegen := now - health.LastRegenTime
	nextRegenIn := 300 - (timeSinceLastRegen % 300)
	
	if health.Health < health.MaxHealth {
		result += fmt.Sprintf("⏰ Next regen in: %d seconds\n", nextRegenIn)
	} else {
		result += "✅ Health is full!\n"
	}
	
	return result
}

// generateHealthBar creates a visual health bar
func (hs *HealthSystem) generateHealthBar(percentage int) string {
	bars := 10
	filledBars := (percentage * bars) / 100
	
	var healthBar string
	for i := 0; i < bars; i++ {
		if i < filledBars {
			healthBar += "█"
		} else {
			healthBar += "░"
		}
	}
	
	return healthBar
}

// UseHealthPotion uses a health potion to restore HP
func (hs *HealthSystem) UseHealthPotion(userJID string) (string, error) {
	user := hs.db.GetUser(userJID)
	health := user.Health
	
	// Check if user has potions
	if health.HealthPotions <= 0 {
		return "❌ You don't have any health potions! Buy some from the shop.", nil
	}
	
	// Check if health is already full
	if health.Health >= health.MaxHealth {
		return "❤️ Your health is already full!", nil
	}
	
	// Use potion
	healAmount := int64(50) // Standard potion heals 50 HP
	oldHealth := health.Health
	
	health.Health += healAmount
	if health.Health > health.MaxHealth {
		health.Health = health.MaxHealth
	}
	
	health.HealthPotions--
	actualHeal := health.Health - oldHealth
	
	result := fmt.Sprintf("🧪 *Health Potion Used*\n\n")
	result += fmt.Sprintf("❤️ Healed: +%d HP\n", actualHeal)
	result += fmt.Sprintf("❤️ Current Health: %d/%d HP\n", health.Health, health.MaxHealth)
	result += fmt.Sprintf("🧪 Remaining Potions: %d\n", health.HealthPotions)
	
	return result, nil
}

// BuyHealthPotion allows user to buy health potions
func (hs *HealthSystem) BuyHealthPotion(userJID, potionType string, quantity int64) (string, error) {
	user := hs.db.GetUser(userJID)
	
	potionInfo, exists := PotionTypes[potionType]
	if !exists {
		return "❌ Invalid potion type! Available: small, medium, large, mega", nil
	}
	
	totalCost := potionInfo.Price * quantity
	
	// Check if user has enough money
	if user.Money < totalCost {
		return fmt.Sprintf("💰 You need %d coins to buy %d %s! You have %d coins.", 
			totalCost, quantity, potionInfo.Name, user.Money), nil
	}
	
	// Purchase potions
	user.Money -= totalCost
	user.Health.HealthPotions += quantity
	
	return fmt.Sprintf("✅ Successfully bought %d %s for %d coins!\n💰 Remaining balance: %d coins\n🧪 Total potions: %d", 
		quantity, potionInfo.Name, totalCost, user.Money, user.Health.HealthPotions), nil
}

// TakeDamage deals damage to a user
func (hs *HealthSystem) TakeDamage(userJID string, damage int64, source string) (string, error) {
	user := hs.db.GetUser(userJID)
	health := user.Health
	
	oldHealth := health.Health
	health.Health -= damage
	if health.Health < 0 {
		health.Health = 0
	}
	
	health.LastDamage = time.Now().Unix()
	actualDamage := oldHealth - health.Health
	
	result := fmt.Sprintf("💥 *Damage Taken*\n\n")
	result += fmt.Sprintf("⚔️ Source: %s\n", source)
	result += fmt.Sprintf("💔 Damage: -%d HP\n", actualDamage)
	result += fmt.Sprintf("❤️ Current Health: %d/%d HP\n", health.Health, health.MaxHealth)
	
	if health.Health == 0 {
		result += "\n💀 You have been defeated! Your health will regenerate over time."
	} else if health.Health < health.MaxHealth/4 {
		result += "\n⚠️ Critical health! Consider using a health potion."
	}
	
	return result, nil
}

// HealUser heals a user (admin function)
func (hs *HealthSystem) HealUser(userJID string, amount int64) (string, error) {
	user := hs.db.GetUser(userJID)
	health := user.Health
	
	oldHealth := health.Health
	health.Health += amount
	if health.Health > health.MaxHealth {
		health.Health = health.MaxHealth
	}
	
	actualHeal := health.Health - oldHealth
	
	result := fmt.Sprintf("✨ *Divine Healing*\n\n")
	result += fmt.Sprintf("❤️ Healed: +%d HP\n", actualHeal)
	result += fmt.Sprintf("❤️ Current Health: %d/%d HP\n", health.Health, health.MaxHealth)
	
	return result, nil
}

// GetPotionShop returns the potion shop information
func (hs *HealthSystem) GetPotionShop() string {
	result := "🏪 *Health Potion Shop*\n\n"
	
	for i, potionType := range []string{"small", "medium", "large", "mega"} {
		info := PotionTypes[potionType]
		result += fmt.Sprintf("%d. %s %s\n", i+1, info.Emoji, info.Name)
		result += fmt.Sprintf("   💰 Price: %d coins\n", info.Price)
		result += fmt.Sprintf("   ❤️ Healing: %d HP\n", info.HealAmount)
		result += fmt.Sprintf("   📝 %s\n\n", info.Description)
	}
	
	result += "💡 Use command: buypotion <type> [quantity] to purchase\n"
	result += "💡 Available types: small, medium, large, mega"
	
	return result
}

// UpgradeMaxHealth increases user's maximum health
func (hs *HealthSystem) UpgradeMaxHealth(userJID string) (string, error) {
	user := hs.db.GetUser(userJID)
	health := user.Health
	
	// Calculate upgrade cost based on current max health
	upgradeCost := (health.MaxHealth - 100) * 10 + 500 // Base cost 500, increases by 10 per HP
	
	if user.Money < upgradeCost {
		return fmt.Sprintf("💰 You need %d coins to upgrade your max health! You have %d coins.", 
			upgradeCost, user.Money), nil
	}
	
	// Upgrade max health
	user.Money -= upgradeCost
	oldMaxHealth := health.MaxHealth
	health.MaxHealth += 10 // Increase by 10 HP
	
	// Also heal the user by the upgrade amount
	health.Health += 10
	if health.Health > health.MaxHealth {
		health.Health = health.MaxHealth
	}
	
	return fmt.Sprintf("⬆️ *Max Health Upgraded*\n\n"+
		"❤️ Max Health: %d → %d HP\n"+
		"❤️ Current Health: %d HP\n"+
		"💰 Cost: %d coins\n"+
		"💰 Remaining balance: %d coins", 
		oldMaxHealth, health.MaxHealth, health.Health, upgradeCost, user.Money), nil
}

// InitializeHealthSystem initializes the health system with periodic regeneration
func InitializeHealthSystem(db *database.Database) *HealthSystem {
	hs := NewHealthSystem(db)
	
	// Start periodic health regeneration for all users
	go func() {
		ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
		defer ticker.Stop()
		
		for range ticker.C {
			// Regenerate health for all users
			for userJID := range db.Users {
				hs.RegenerateHealth(userJID)
			}
		}
	}()
	
	return hs
}

// GetHealthLeaderboard returns top users by health
func (hs *HealthSystem) GetHealthLeaderboard() string {
	type HealthEntry struct {
		Name      string
		Health    int64
		MaxHealth int64
	}
	
	var entries []HealthEntry
	
	// Collect health data
	for _, user := range hs.db.Users {
		if user.Name != "" {
			entries = append(entries, HealthEntry{
				Name:      user.Name,
				Health:    user.Health.Health,
				MaxHealth: user.Health.MaxHealth,
			})
		}
	}
	
	// Sort by max health (descending)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].MaxHealth > entries[i].MaxHealth {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	
	result := "🏆 *Health Leaderboard*\n\n"
	
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
		
		percentage := float64(entry.Health) / float64(entry.MaxHealth) * 100
		result += fmt.Sprintf("%s %s\n", medal, entry.Name)
		result += fmt.Sprintf("   ❤️ %d/%d HP (%.1f%%)\n\n", entry.Health, entry.MaxHealth, percentage)
	}
	
	return result
}