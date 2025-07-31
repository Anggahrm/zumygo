package systems

import (
	"fmt"
	"math/rand"
	"time"
	"zumygo/database"
)

// MiningSystem handles all mining-related operations
type MiningSystem struct {
	db *database.Database
}

// NewMiningSystem creates a new mining system instance
func NewMiningSystem(db *database.Database) *MiningSystem {
	return &MiningSystem{db: db}
}

// PickaxeInfo represents information about a pickaxe type
type PickaxeInfo struct {
	Name        string
	Power       int
	Durability  int
	Price       int64
	Description string
}

// OreInfo represents information about an ore type
type OreInfo struct {
	Name        string
	Rarity      int    // 1-100, higher is rarer
	BasePrice   int64
	Description string
	Emoji       string
}

var (
	// Pickaxe types with their stats
	PickaxeTypes = map[string]PickaxeInfo{
		"wooden": {
			Name:        "Wooden Pickaxe",
			Power:       1,
			Durability:  10,
			Price:       50,
			Description: "Basic wooden pickaxe. Low power but cheap.",
		},
		"stone": {
			Name:        "Stone Pickaxe",
			Power:       2,
			Durability:  25,
			Price:       150,
			Description: "Stone pickaxe with better mining power.",
		},
		"iron": {
			Name:        "Iron Pickaxe",
			Power:       4,
			Durability:  50,
			Price:       500,
			Description: "Iron pickaxe with good mining efficiency.",
		},
		"gold": {
			Name:        "Gold Pickaxe",
			Power:       6,
			Durability:  30,
			Price:       1000,
			Description: "Fast but fragile gold pickaxe.",
		},
		"diamond": {
			Name:        "Diamond Pickaxe",
			Power:       10,
			Durability:  100,
			Price:       5000,
			Description: "The ultimate mining tool with maximum power.",
		},
	}

	// Ore types with their properties
	OreTypes = map[string]OreInfo{
		"coal": {
			Name:        "Coal",
			Rarity:      80,
			BasePrice:   10,
			Description: "Common coal ore, used for basic fuel.",
			Emoji:       "⚫",
		},
		"iron": {
			Name:        "Iron Ore",
			Rarity:      60,
			BasePrice:   25,
			Description: "Iron ore, essential for tools and weapons.",
			Emoji:       "🔩",
		},
		"gold": {
			Name:        "Gold Ore",
			Rarity:      30,
			BasePrice:   100,
			Description: "Precious gold ore, valuable and rare.",
			Emoji:       "🟡",
		},
		"diamond": {
			Name:        "Diamond",
			Rarity:      10,
			BasePrice:   500,
			Description: "Extremely rare and valuable diamond.",
			Emoji:       "💎",
		},
		"emerald": {
			Name:        "Emerald",
			Rarity:      5,
			BasePrice:   1000,
			Description: "Ultra rare emerald, the most valuable ore.",
			Emoji:       "💚",
		},
	}
)

// Mine performs a mining operation for a user
func (ms *MiningSystem) Mine(userJID string) (string, error) {
	user := ms.db.GetUser(userJID)
	
	// Check if user can mine (cooldown)
	now := time.Now().Unix()
	if now-user.Mining.LastMine < 300 { // 5 minute cooldown
		remaining := 300 - (now - user.Mining.LastMine)
		return fmt.Sprintf("⏰ You need to wait %d seconds before mining again!", remaining), nil
	}
	
	// Check if user has a pickaxe
	bestPickaxe := ms.getBestPickaxe(user)
	if bestPickaxe == "" {
		return "⛏️ You don't have any pickaxe! Buy one from the pickaxe shop first.", nil
	}
	
	// Perform mining
	ores := ms.performMining(user, bestPickaxe)
	user.Mining.LastMine = now
	user.Mining.TotalMined++
	
	// Add mining experience
	expGained := int64(len(ores) * 10)
	user.Mining.MiningExp += expGained
	
	// Check for level up
	newLevel := ms.calculateMiningLevel(user.Mining.MiningExp)
	leveledUp := newLevel > user.Mining.MiningLevel
	if leveledUp {
		user.Mining.MiningLevel = newLevel
	}
	
	// Format result message
	result := "⛏️ *Mining Results*\n\n"
	result += fmt.Sprintf("🔧 Used: %s\n", PickaxeTypes[bestPickaxe].Name)
	
	if len(ores) == 0 {
		result += "😔 You didn't find anything this time. Better luck next time!"
	} else {
		result += "🎉 *Ores Found:*\n"
		for ore, amount := range ores {
			oreInfo := OreTypes[ore]
			result += fmt.Sprintf("%s %s x%d\n", oreInfo.Emoji, oreInfo.Name, amount)
		}
	}
	
	result += fmt.Sprintf("\n⭐ Mining EXP: +%d", expGained)
	if leveledUp {
		result += fmt.Sprintf("\n🎊 Level Up! Mining Level: %d", newLevel)
	}
	
	return result, nil
}

// getBestPickaxe returns the best pickaxe the user has
func (ms *MiningSystem) getBestPickaxe(user *database.User) string {
	pickaxes := map[string]int64{
		"diamond": user.Mining.DiamondPickaxe,
		"gold":    user.Mining.GoldPickaxe,
		"iron":    user.Mining.IronPickaxe,
		"stone":   user.Mining.StonePickaxe,
		"wooden":  user.Mining.WoodenPickaxe,
	}
	
	// Return the best pickaxe the user has
	for _, pickaxe := range []string{"diamond", "gold", "iron", "stone", "wooden"} {
		if pickaxes[pickaxe] > 0 {
			return pickaxe
		}
	}
	
	return ""
}

// performMining simulates the mining process and returns found ores
func (ms *MiningSystem) performMining(user *database.User, pickaxeType string) map[string]int64 {
	pickaxe := PickaxeTypes[pickaxeType]
	ores := make(map[string]int64)
	
	// Number of mining attempts based on pickaxe power
	attempts := pickaxe.Power + rand.Intn(3)
	
	for i := 0; i < attempts; i++ {
		// Random chance to find ore
		if rand.Intn(100) < 70 { // 70% chance to find something
			ore := ms.selectRandomOre(pickaxe.Power)
			if ore != "" {
				ores[ore]++
				// Add ore to user's inventory
				ms.addOreToUser(user, ore, 1)
			}
		}
	}
	
	return ores
}

// selectRandomOre selects a random ore based on pickaxe power and ore rarity
func (ms *MiningSystem) selectRandomOre(pickaxePower int) string {
	// Higher power pickaxes can find rarer ores
	totalWeight := 0
	oreWeights := make(map[string]int)
	
	for oreName, oreInfo := range OreTypes {
		// Calculate weight based on rarity and pickaxe power
		weight := oreInfo.Rarity
		
		// Adjust weight based on pickaxe power
		if oreName == "emerald" && pickaxePower < 8 {
			weight = 1 // Very low chance for low-power pickaxes
		} else if oreName == "diamond" && pickaxePower < 6 {
			weight = 5
		} else if oreName == "gold" && pickaxePower < 4 {
			weight = 15
		}
		
		oreWeights[oreName] = weight
		totalWeight += weight
	}
	
	// Select random ore based on weights
	randValue := rand.Intn(totalWeight)
	currentWeight := 0
	
	for oreName, weight := range oreWeights {
		currentWeight += weight
		if randValue < currentWeight {
			return oreName
		}
	}
	
	return "coal" // Fallback
}

// addOreToUser adds ore to user's mining inventory
func (ms *MiningSystem) addOreToUser(user *database.User, oreType string, amount int64) {
	switch oreType {
	case "coal":
		user.Mining.Coal += amount
	case "iron":
		user.Mining.Iron += amount
	case "gold":
		user.Mining.Gold += amount
	case "diamond":
		user.Mining.Diamond += amount
	case "emerald":
		user.Mining.Emerald += amount
	}
}

// calculateMiningLevel calculates mining level based on experience
func (ms *MiningSystem) calculateMiningLevel(exp int64) int {
	// Level formula: level = sqrt(exp / 100)
	level := 1
	requiredExp := int64(100)
	
	for exp >= requiredExp {
		level++
		requiredExp = int64(level * level * 100)
	}
	
	return level
}

// GetMiningInfo returns user's mining information
func (ms *MiningSystem) GetMiningInfo(userJID string) string {
	user := ms.db.GetUser(userJID)
	mining := user.Mining
	
	result := "⛏️ *Your Mining Info*\n\n"
	
	// Mining stats
	result += fmt.Sprintf("📊 *Stats:*\n")
	result += fmt.Sprintf("Level: %d\n", mining.MiningLevel)
	result += fmt.Sprintf("Experience: %d\n", mining.MiningExp)
	result += fmt.Sprintf("Total Mined: %d\n", mining.TotalMined)
	
	// Cooldown info
	now := time.Now().Unix()
	if now-mining.LastMine < 300 {
		remaining := 300 - (now - mining.LastMine)
		result += fmt.Sprintf("⏰ Next mine in: %d seconds\n", remaining)
	} else {
		result += "✅ Ready to mine!\n"
	}
	
	// Pickaxes
	result += "\n🔧 *Pickaxes:*\n"
	if mining.WoodenPickaxe > 0 {
		result += fmt.Sprintf("🪵 Wooden: %d\n", mining.WoodenPickaxe)
	}
	if mining.StonePickaxe > 0 {
		result += fmt.Sprintf("🪨 Stone: %d\n", mining.StonePickaxe)
	}
	if mining.IronPickaxe > 0 {
		result += fmt.Sprintf("⚙️ Iron: %d\n", mining.IronPickaxe)
	}
	if mining.GoldPickaxe > 0 {
		result += fmt.Sprintf("🟨 Gold: %d\n", mining.GoldPickaxe)
	}
	if mining.DiamondPickaxe > 0 {
		result += fmt.Sprintf("💎 Diamond: %d\n", mining.DiamondPickaxe)
	}
	
	// Ores
	result += "\n💎 *Ores:*\n"
	if mining.Coal > 0 {
		result += fmt.Sprintf("⚫ Coal: %d\n", mining.Coal)
	}
	if mining.Iron > 0 {
		result += fmt.Sprintf("🔩 Iron: %d\n", mining.Iron)
	}
	if mining.Gold > 0 {
		result += fmt.Sprintf("🟡 Gold: %d\n", mining.Gold)
	}
	if mining.Diamond > 0 {
		result += fmt.Sprintf("💎 Diamond: %d\n", mining.Diamond)
	}
	if mining.Emerald > 0 {
		result += fmt.Sprintf("💚 Emerald: %d\n", mining.Emerald)
	}
	
	return result
}

// BuyPickaxe allows user to buy a pickaxe
func (ms *MiningSystem) BuyPickaxe(userJID, pickaxeType string) (string, error) {
	user := ms.db.GetUser(userJID)
	
	pickaxeInfo, exists := PickaxeTypes[pickaxeType]
	if !exists {
		return "❌ Invalid pickaxe type!", nil
	}
	
	// Check if user has enough money
	if user.Money < pickaxeInfo.Price {
		return fmt.Sprintf("💰 You need %d coins to buy %s! You have %d coins.", 
			pickaxeInfo.Price, pickaxeInfo.Name, user.Money), nil
	}
	
	// Deduct money and add pickaxe
	user.Money -= pickaxeInfo.Price
	
	switch pickaxeType {
	case "wooden":
		user.Mining.WoodenPickaxe++
	case "stone":
		user.Mining.StonePickaxe++
	case "iron":
		user.Mining.IronPickaxe++
	case "gold":
		user.Mining.GoldPickaxe++
	case "diamond":
		user.Mining.DiamondPickaxe++
	}
	
	return fmt.Sprintf("✅ Successfully bought %s for %d coins!\n💰 Remaining balance: %d coins", 
		pickaxeInfo.Name, pickaxeInfo.Price, user.Money), nil
}

// SellOre allows user to sell ores
func (ms *MiningSystem) SellOre(userJID, oreType string, amount int64) (string, error) {
	user := ms.db.GetUser(userJID)
	
	oreInfo, exists := OreTypes[oreType]
	if !exists {
		return "❌ Invalid ore type!", nil
	}
	
	// Check if user has enough ore
	userOreAmount := ms.getUserOreAmount(user, oreType)
	if userOreAmount < amount {
		return fmt.Sprintf("❌ You don't have enough %s! You have %d.", 
			oreInfo.Name, userOreAmount), nil
	}
	
	// Calculate sale price (with some market fluctuation)
	fluctuation := 0.8 + rand.Float64()*0.4 // 80% to 120% of base price
	salePrice := int64(float64(oreInfo.BasePrice) * fluctuation * float64(amount))
	
	// Remove ore and add money
	ms.removeOreFromUser(user, oreType, amount)
	user.Money += salePrice
	
	return fmt.Sprintf("✅ Sold %d %s for %d coins!\n💰 New balance: %d coins", 
		amount, oreInfo.Name, salePrice, user.Money), nil
}

// getUserOreAmount returns the amount of ore a user has
func (ms *MiningSystem) getUserOreAmount(user *database.User, oreType string) int64 {
	switch oreType {
	case "coal":
		return user.Mining.Coal
	case "iron":
		return user.Mining.Iron
	case "gold":
		return user.Mining.Gold
	case "diamond":
		return user.Mining.Diamond
	case "emerald":
		return user.Mining.Emerald
	default:
		return 0
	}
}

// removeOreFromUser removes ore from user's inventory
func (ms *MiningSystem) removeOreFromUser(user *database.User, oreType string, amount int64) {
	switch oreType {
	case "coal":
		user.Mining.Coal -= amount
	case "iron":
		user.Mining.Iron -= amount
	case "gold":
		user.Mining.Gold -= amount
	case "diamond":
		user.Mining.Diamond -= amount
	case "emerald":
		user.Mining.Emerald -= amount
	}
}

// GetPickaxeShop returns the pickaxe shop information
func (ms *MiningSystem) GetPickaxeShop() string {
	result := "🏪 *Pickaxe Shop*\n\n"
	
	for i, pickaxeType := range []string{"wooden", "stone", "iron", "gold", "diamond"} {
		info := PickaxeTypes[pickaxeType]
		result += fmt.Sprintf("%d. %s\n", i+1, info.Name)
		result += fmt.Sprintf("   💰 Price: %d coins\n", info.Price)
		result += fmt.Sprintf("   ⚡ Power: %d\n", info.Power)
		result += fmt.Sprintf("   🛡️ Durability: %d\n", info.Durability)
		result += fmt.Sprintf("   📝 %s\n\n", info.Description)
	}
	
	result += "💡 Use command: buyPickaxe <type> to purchase"
	
	return result
}

// InitializeMiningSystem initializes the mining system with periodic updates
func InitializeMiningSystem(db *database.Database) *MiningSystem {
	ms := NewMiningSystem(db)
	
	// Start periodic ore stock updates
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // Update every hour
		defer ticker.Stop()
		
		for range ticker.C {
			ms.updateOreStock()
		}
	}()
	
	return ms
}

// updateOreStock updates the global ore stock and prices
func (ms *MiningSystem) updateOreStock() {
	if len(ms.db.OreStock) > 0 {
		stock := &ms.db.OreStock[0]
		
		// Update stock amounts (simulate market dynamics)
		stock.Coal += int64(rand.Intn(200) - 100)     // ±100
		stock.Iron += int64(rand.Intn(100) - 50)      // ±50
		stock.Gold += int64(rand.Intn(40) - 20)       // ±20
		stock.Diamond += int64(rand.Intn(20) - 10)    // ±10
		stock.Emerald += int64(rand.Intn(10) - 5)     // ±5
		
		// Ensure minimum stock
		if stock.Coal < 100 { stock.Coal = 100 }
		if stock.Iron < 50 { stock.Iron = 50 }
		if stock.Gold < 20 { stock.Gold = 20 }
		if stock.Diamond < 10 { stock.Diamond = 10 }
		if stock.Emerald < 5 { stock.Emerald = 5 }
		
		// Update prices based on stock (supply and demand)
		for ore, basePrice := range map[string]int64{
			"coal": 10, "iron": 25, "gold": 100, "diamond": 500, "emerald": 1000,
		} {
			var currentStock int64
			switch ore {
			case "coal": currentStock = stock.Coal
			case "iron": currentStock = stock.Iron
			case "gold": currentStock = stock.Gold
			case "diamond": currentStock = stock.Diamond
			case "emerald": currentStock = stock.Emerald
			}
			
			// Price inversely related to stock
			priceMultiplier := 1.0
			if currentStock < 50 {
				priceMultiplier = 1.5 // High demand, low supply
			} else if currentStock > 200 {
				priceMultiplier = 0.7 // Low demand, high supply
			}
			
			stock.Prices[ore] = int64(float64(basePrice) * priceMultiplier)
		}
		
		stock.LastUpdate = time.Now().Unix()
	}
}