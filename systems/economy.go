package systems

import (
	"fmt"
	"math/rand"
	"time"
	"zumygo/database"
)

// EconomySystem handles all economy-related operations
type EconomySystem struct {
	db *database.Database
}

// NewEconomySystem creates a new economy system instance
func NewEconomySystem(db *database.Database) *EconomySystem {
	return &EconomySystem{db: db}
}

// WorkInfo represents different types of work
type WorkInfo struct {
	Name        string
	MinReward   int64
	MaxReward   int64
	Cooldown    int64 // in seconds
	Description string
	Emoji       string
}

// ItemInfo represents shop items
type ItemInfo struct {
	Name        string
	Price       int64
	Description string
	Emoji       string
	Category    string
}

var (
	// Work types with their properties
	WorkTypes = []WorkInfo{
		{
			Name:        "Buruh",
			MinReward:   50,
			MaxReward:   150,
			Cooldown:    3600, // 1 hour
			Description: "Kerja sebagai buruh harian",
			Emoji:       "👷",
		},
		{
			Name:        "Ojek Online",
			MinReward:   80,
			MaxReward:   200,
			Cooldown:    2700, // 45 minutes
			Description: "Jadi driver ojek online",
			Emoji:       "🏍️",
		},
		{
			Name:        "Freelancer",
			MinReward:   100,
			MaxReward:   300,
			Cooldown:    5400, // 1.5 hours
			Description: "Kerja freelance programming",
			Emoji:       "💻",
		},
		{
			Name:        "Streamer",
			MinReward:   200,
			MaxReward:   500,
			Cooldown:    7200, // 2 hours
			Description: "Streaming game online",
			Emoji:       "🎮",
		},
		{
			Name:        "YouTuber",
			MinReward:   300,
			MaxReward:   800,
			Cooldown:    10800, // 3 hours
			Description: "Buat konten YouTube",
			Emoji:       "📹",
		},
	}

	// Shop items
	ShopItems = map[string]ItemInfo{
		"phone": {
			Name:        "Smartphone",
			Price:       2000,
			Description: "Smartphone untuk komunikasi",
			Emoji:       "📱",
			Category:    "Electronics",
		},
		"laptop": {
			Name:        "Laptop Gaming",
			Price:       15000,
			Description: "Laptop untuk gaming dan kerja",
			Emoji:       "💻",
			Category:    "Electronics",
		},
		"car": {
			Name:        "Mobil",
			Price:       50000,
			Description: "Mobil untuk transportasi",
			Emoji:       "🚗",
			Category:    "Vehicle",
		},
		"house": {
			Name:        "Rumah",
			Price:       200000,
			Description: "Rumah untuk tempat tinggal",
			Emoji:       "🏠",
			Category:    "Property",
		},
		"ring": {
			Name:        "Cincin Berlian",
			Price:       10000,
			Description: "Cincin berlian mewah",
			Emoji:       "💍",
			Category:    "Jewelry",
		},
		"watch": {
			Name:        "Jam Tangan Mewah",
			Price:       8000,
			Description: "Jam tangan branded",
			Emoji:       "⌚",
			Category:    "Jewelry",
		},
	}
)

// Work allows user to work and earn money
func (es *EconomySystem) Work(userJID string) (string, error) {
	user := es.db.GetUser(userJID)
	
	now := time.Now().Unix()
	
	// Check cooldown (1 hour)
	if now-user.LastWork < 3600 {
		remaining := 3600 - (now - user.LastWork)
		hours := remaining / 3600
		minutes := (remaining % 3600) / 60
		seconds := remaining % 60
		
		var timeStr string
		if hours > 0 {
			timeStr = fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
		} else if minutes > 0 {
			timeStr = fmt.Sprintf("%dm %ds", minutes, seconds)
		} else {
			timeStr = fmt.Sprintf("%ds", seconds)
		}
		
		return fmt.Sprintf("⏰ You need to wait %s before working again!", timeStr), nil
	}
	
	// Random work type
	workType := WorkTypes[rand.Intn(len(WorkTypes))]
	
	// Random reward within range
	reward := workType.MinReward + rand.Int63n(workType.MaxReward-workType.MinReward+1)
	
	// Bonus for premium users
	if user.Premium {
		reward = int64(float64(reward) * 1.5) // 50% bonus
	}
	
	// Add money and update last work time
	user.Money += reward
	user.LastWork = now
	
	// Add some experience
	expGain := reward / 10
	user.Exp += expGain
	
	result := fmt.Sprintf("💼 *Work Complete*\n\n")
	result += fmt.Sprintf("%s **%s**\n", workType.Emoji, workType.Name)
	result += fmt.Sprintf("💰 Earned: %d coins\n", reward)
	result += fmt.Sprintf("⭐ Experience: +%d XP\n", expGain)
	result += fmt.Sprintf("💵 Total Money: %d coins\n", user.Money)
	
	if user.Premium {
		result += "\n🌟 Premium bonus applied! (+50%)"
	}
	
	return result, nil
}

// DailyClaim allows user to claim daily rewards
func (es *EconomySystem) DailyClaim(userJID string) (string, error) {
	user := es.db.GetUser(userJID)
	
	now := time.Now().Unix()
	
	// Check if 24 hours have passed
	if now-user.LastClaim < 86400 { // 24 hours = 86400 seconds
		remaining := 86400 - (now - user.LastClaim)
		hours := remaining / 3600
		minutes := (remaining % 3600) / 60
		
		return fmt.Sprintf("⏰ Daily reward already claimed! Next claim in %dh %dm", hours, minutes), nil
	}
	
	// Base daily reward
	baseReward := int64(200)
	zcReward := int64(5)
	
	// Bonus for premium users
	if user.Premium {
		baseReward *= 2
		zcReward *= 2
	}
	
	// Streak bonus (up to 7 days)
	daysSinceLastClaim := (now - user.LastClaim) / 86400
	var streak int64 = 1
	
	if daysSinceLastClaim == 1 {
		// Consecutive day, increase streak
		if user.LastClaim > 0 {
			streak = 2 // Simplified streak system
		}
	}
	
	totalReward := baseReward * streak
	totalZC := zcReward * streak
	
	// Add rewards
	user.Money += totalReward
	user.ZC += totalZC
	user.LastClaim = now
	
	result := fmt.Sprintf("🎁 *Daily Reward Claimed*\n\n")
	result += fmt.Sprintf("💰 Money: +%d coins\n", totalReward)
	result += fmt.Sprintf("🪙 ZumyCoin: +%d ZC\n", totalZC)
	result += fmt.Sprintf("💵 Total Money: %d coins\n", user.Money)
	result += fmt.Sprintf("🪙 Total ZC: %d ZC\n", user.ZC)
	
	if streak > 1 {
		result += fmt.Sprintf("\n🔥 Streak Bonus: x%d", streak)
	}
	
	if user.Premium {
		result += "\n🌟 Premium bonus applied! (x2)"
	}
	
	return result, nil
}

// Transfer allows user to transfer money to another user
func (es *EconomySystem) Transfer(fromJID, toJID string, amount int64) (string, error) {
	fromUser := es.db.GetUser(fromJID)
	toUser := es.db.GetUser(toJID)
	
	// Check if sender has enough money
	if fromUser.Money < amount {
		return fmt.Sprintf("❌ Insufficient funds! You have %d coins, need %d coins.", fromUser.Money, amount), nil
	}
	
	// Check minimum transfer amount
	if amount < 10 {
		return "❌ Minimum transfer amount is 10 coins!", nil
	}
	
	// Transfer fee (5%)
	fee := amount / 20 // 5% fee
	actualAmount := amount - fee
	
	// Perform transfer
	fromUser.Money -= amount
	toUser.Money += actualAmount
	
	result := fmt.Sprintf("💸 *Transfer Complete*\n\n")
	result += fmt.Sprintf("📤 Sent: %d coins\n", amount)
	result += fmt.Sprintf("💰 Received: %d coins\n", actualAmount)
	result += fmt.Sprintf("💳 Fee: %d coins (5%%)\n", fee)
	result += fmt.Sprintf("💵 Your Balance: %d coins\n", fromUser.Money)
	
	return result, nil
}

// Rob allows user to attempt robbing another user
func (es *EconomySystem) Rob(robberJID, targetJID string) (string, error) {
	robber := es.db.GetUser(robberJID)
	target := es.db.GetUser(targetJID)
	
	now := time.Now().Unix()
	
	// Check cooldown (2 hours)
	if now-robber.LastRob < 7200 {
		remaining := 7200 - (now - robber.LastRob)
		hours := remaining / 3600
		minutes := (remaining % 3600) / 60
		
		return fmt.Sprintf("⏰ You need to wait %dh %dm before robbing again!", hours, minutes), nil
	}
	
	// Check if target has enough money
	if target.Money < 100 {
		return "❌ Target doesn't have enough money to rob!", nil
	}
	
	// Check if robber has minimum money to attempt rob
	if robber.Money < 50 {
		return "❌ You need at least 50 coins to attempt a robbery!", nil
	}
	
	// Success rate (60% base, reduced if target is premium)
	successRate := 60
	if target.Premium {
		successRate = 40 // Premium users are harder to rob
	}
	
	robber.LastRob = now
	
	if rand.Intn(100) < successRate {
		// Successful robbery
		maxSteal := target.Money / 10 // Maximum 10% of target's money
		if maxSteal > 1000 {
			maxSteal = 1000 // Cap at 1000 coins
		}
		
		stolenAmount := rand.Int63n(maxSteal) + 50 // Minimum 50 coins
		
		target.Money -= stolenAmount
		robber.Money += stolenAmount
		
		result := fmt.Sprintf("💰 *Robbery Successful*\n\n")
		result += fmt.Sprintf("🎯 Target robbed successfully!\n")
		result += fmt.Sprintf("💰 Stolen: %d coins\n", stolenAmount)
		result += fmt.Sprintf("💵 Your Balance: %d coins\n", robber.Money)
		
		return result, nil
	} else {
		// Failed robbery - lose money as penalty
		penalty := int64(100)
		robber.Money -= penalty
		
		result := fmt.Sprintf("🚨 *Robbery Failed*\n\n")
		result += fmt.Sprintf("👮 You got caught by police!\n")
		result += fmt.Sprintf("💸 Fine: %d coins\n", penalty)
		result += fmt.Sprintf("💵 Your Balance: %d coins\n", robber.Money)
		
		return result, nil
	}
}

// GetShop returns the shop information
func (es *EconomySystem) GetShop() string {
	result := "🏪 *ZumyNext Shop*\n\n"
	
	categories := make(map[string][]ItemInfo)
	for _, item := range ShopItems {
		categories[item.Category] = append(categories[item.Category], item)
	}
	
	for category, items := range categories {
		result += fmt.Sprintf("**%s**\n", category)
		for _, item := range items {
			result += fmt.Sprintf("%s %s - %d coins\n", item.Emoji, item.Name, item.Price)
			result += fmt.Sprintf("   📝 %s\n", item.Description)
		}
		result += "\n"
	}
	
	result += "💡 Use command: buy <item> to purchase\n"
	result += "💡 Available items: phone, laptop, car, house, ring, watch"
	
	return result
}

// BuyItem allows user to buy items from shop
func (es *EconomySystem) BuyItem(userJID, itemKey string) (string, error) {
	user := es.db.GetUser(userJID)
	
	item, exists := ShopItems[itemKey]
	if !exists {
		return "❌ Item not found! Use 'shop' command to see available items.", nil
	}
	
	// Check if user has enough money
	if user.Money < item.Price {
		return fmt.Sprintf("💰 You need %d coins to buy %s! You have %d coins.", 
			item.Price, item.Name, user.Money), nil
	}
	
	// Check if user already owns this item
	if user.Inventory[itemKey] > 0 {
		return fmt.Sprintf("❌ You already own %s!", item.Name), nil
	}
	
	// Purchase item
	user.Money -= item.Price
	if user.Inventory == nil {
		user.Inventory = make(map[string]int64)
	}
	user.Inventory[itemKey] = 1
	
	result := fmt.Sprintf("✅ *Purchase Successful*\n\n")
	result += fmt.Sprintf("%s **%s**\n", item.Emoji, item.Name)
	result += fmt.Sprintf("💰 Price: %d coins\n", item.Price)
	result += fmt.Sprintf("💵 Remaining Balance: %d coins\n", user.Money)
	result += fmt.Sprintf("📦 Item added to inventory!\n")
	
	return result, nil
}

// GetInventory returns user's inventory
func (es *EconomySystem) GetInventory(userJID string) string {
	user := es.db.GetUser(userJID)
	
	result := "🎒 *Your Inventory*\n\n"
	
	if len(user.Inventory) == 0 {
		result += "📦 Your inventory is empty!\n"
		result += "Visit the shop to buy items."
		return result
	}
	
	totalValue := int64(0)
	
	for itemKey, quantity := range user.Inventory {
		if quantity > 0 {
			if item, exists := ShopItems[itemKey]; exists {
				result += fmt.Sprintf("%s %s x%d\n", item.Emoji, item.Name, quantity)
				result += fmt.Sprintf("   💰 Value: %d coins each\n", item.Price)
				totalValue += item.Price * quantity
			}
		}
	}
	
	result += fmt.Sprintf("\n💎 Total Inventory Value: %d coins", totalValue)
	
	return result
}

// GetEconomyLeaderboard returns top users by money
func (es *EconomySystem) GetEconomyLeaderboard() string {
	type EconomyEntry struct {
		Name  string
		Money int64
		ZC    int64
	}
	
	var entries []EconomyEntry
	
	// Collect economy data
	for _, user := range es.db.Users {
		if user.Name != "" {
			entries = append(entries, EconomyEntry{
				Name:  user.Name,
				Money: user.Money,
				ZC:    user.ZC,
			})
		}
	}
	
	// Sort by money (descending)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Money > entries[i].Money {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	
	result := "🏆 *Economy Leaderboard*\n\n"
	
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
		
		result += fmt.Sprintf("%s %s\n", medal, entry.Name)
		result += fmt.Sprintf("   💰 %d coins | 🪙 %d ZC\n\n", entry.Money, entry.ZC)
	}
	
	return result
}

// ATMDeposit allows user to deposit money to ATM
func (es *EconomySystem) ATMDeposit(userJID string, amount int64) (string, error) {
	user := es.db.GetUser(userJID)
	
	if user.Money < amount {
		return fmt.Sprintf("❌ Insufficient funds! You have %d coins.", user.Money), nil
	}
	
	if amount < 10 {
		return "❌ Minimum deposit amount is 10 coins!", nil
	}
	
	user.Money -= amount
	user.ATM += amount
	
	result := fmt.Sprintf("🏦 *ATM Deposit*\n\n")
	result += fmt.Sprintf("💰 Deposited: %d coins\n", amount)
	result += fmt.Sprintf("🏦 ATM Balance: %d coins\n", user.ATM)
	result += fmt.Sprintf("💵 Cash Balance: %d coins\n", user.Money)
	
	return result, nil
}

// ATMWithdraw allows user to withdraw money from ATM
func (es *EconomySystem) ATMWithdraw(userJID string, amount int64) (string, error) {
	user := es.db.GetUser(userJID)
	
	if user.ATM < amount {
		return fmt.Sprintf("❌ Insufficient ATM balance! You have %d coins in ATM.", user.ATM), nil
	}
	
	if amount < 10 {
		return "❌ Minimum withdrawal amount is 10 coins!", nil
	}
	
	user.ATM -= amount
	user.Money += amount
	
	result := fmt.Sprintf("🏦 *ATM Withdrawal*\n\n")
	result += fmt.Sprintf("💰 Withdrawn: %d coins\n", amount)
	result += fmt.Sprintf("🏦 ATM Balance: %d coins\n", user.ATM)
	result += fmt.Sprintf("💵 Cash Balance: %d coins\n", user.Money)
	
	return result, nil
}

// InitializeEconomySystem initializes the economy system
func InitializeEconomySystem(db *database.Database) *EconomySystem {
	es := NewEconomySystem(db)
	
	// Start periodic economy updates
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // Update every hour
		defer ticker.Stop()
		
		for range ticker.C {
			// Add any periodic economy updates here
			// For example, interest on ATM balances, market fluctuations, etc.
		}
	}()
	
	return es
}