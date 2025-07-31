package commands

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"zumygo/config"
	"zumygo/database"
	"zumygo/systems"
)

// CommandMessage represents a command message
type CommandMessage struct {
	ID        string
	From      string
	Chat      string
	Text      string
	Command   string
	Args      []string
	IsGroup   bool
	IsOwner   bool
	IsAdmin   bool
	IsPremium bool
	User      *database.User
	ChatData  *database.Chat
	Reply     func(string) error
	React     func(string) error
	Delete    func() error
}

// GeneralCommands handles all general bot commands
type GeneralCommands struct {
	cfg            *config.BotConfig
	db             *database.Database
	miningSystem   *systems.MiningSystem
	healthSystem   *systems.HealthSystem
	economySystem  *systems.EconomySystem
	levelingSystem *systems.LevelingSystem
}

// NewGeneralCommands creates a new general commands handler
func NewGeneralCommands(cfg *config.BotConfig, db *database.Database, ms *systems.MiningSystem, hs *systems.HealthSystem, es *systems.EconomySystem, ls *systems.LevelingSystem) *GeneralCommands {
	return &GeneralCommands{
		cfg:            cfg,
		db:             db,
		miningSystem:   ms,
		healthSystem:   hs,
		economySystem:  es,
		levelingSystem: ls,
	}
}

// HandleCommand processes general commands
func (gc *GeneralCommands) HandleCommand(msg *CommandMessage) bool {
	switch msg.Command {
	// === INFO COMMANDS ===
	case "menu", "help":
		gc.handleMenu(msg)
		return true
		
	case "ping":
		gc.handlePing(msg)
		return true
		
	case "runtime", "uptime":
		gc.handleRuntime(msg)
		return true
		
	case "owner":
		gc.handleOwner(msg)
		return true
		
	case "script", "sc":
		gc.handleScript(msg)
		return true

	// === FUN COMMANDS ===
	case "say":
		gc.handleSay(msg)
		return true
		
	case "truth":
		gc.handleTruth(msg)
		return true
		
	case "dare":
		gc.handleDare(msg)
		return true
		
	case "rate":
		gc.handleRate(msg)
		return true
		
	case "couple":
		gc.handleCouple(msg)
		return true
		
	case "quotes":
		gc.handleQuotes(msg)
		return true
		
	case "motivasi":
		gc.handleMotivasi(msg)
		return true
		
	case "faktaunik":
		gc.handleFaktaUnik(msg)
		return true

	// === RANDOM COMMANDS ===
	case "dadu", "dice":
		gc.handleDadu(msg)
		return true
		
	case "koin", "coin":
		gc.handleKoin(msg)
		return true
		
	case "slot":
		gc.handleSlot(msg)
		return true
		
	case "tebakangka":
		gc.handleTebakAngka(msg)
		return true

	// === TEXT COMMANDS ===
	case "reverse":
		gc.handleReverse(msg)
		return true
		
	case "upper":
		gc.handleUpper(msg)
		return true
		
	case "lower":
		gc.handleLower(msg)
		return true
		
	case "count":
		gc.handleCount(msg)
		return true

	// === CALCULATOR ===
	case "calc", "kalkulator":
		gc.handleCalc(msg)
		return true

	// === TIME COMMANDS ===
	case "waktu", "time":
		gc.handleWaktu(msg)
		return true
		
	case "tanggal", "date":
		gc.handleTanggal(msg)
		return true

	// === PROFILE COMMANDS ===
	case "profile", "profil":
		gc.handleProfile(msg)
		return true
		
	case "setname":
		gc.handleSetName(msg)
		return true
		
	case "setbio":
		gc.handleSetBio(msg)
		return true

	default:
		return false
	}
}

// === INFO COMMAND HANDLERS ===

func (gc *GeneralCommands) handleMenu(msg *CommandMessage) {
	menu := fmt.Sprintf("🤖 *%s Menu*\n\n", gc.cfg.NameBot)
	menu += "📋 *General Commands:*\n"
	menu += "• menu/help - Show this menu\n"
	menu += "• ping - Check bot response\n"
	menu += "• runtime - Bot uptime\n"
	menu += "• owner - Owner information\n"
	menu += "• script - Bot source code\n\n"
	
	menu += "🎮 *Fun Commands:*\n"
	menu += "• say <text> - Make bot say something\n"
	menu += "• truth - Random truth question\n"
	menu += "• dare - Random dare challenge\n"
	menu += "• rate <text> - Rate something\n"
	menu += "• couple - Random couple match\n"
	menu += "• quotes - Random quotes\n"
	menu += "• motivasi - Motivational quotes\n"
	menu += "• faktaunik - Unique facts\n\n"
	
	menu += "🎲 *Random Commands:*\n"
	menu += "• dadu/dice - Roll dice\n"
	menu += "• koin/coin - Flip coin\n"
	menu += "• slot - Slot machine\n"
	menu += "• tebakangka - Guess number game\n\n"
	
	menu += "📝 *Text Commands:*\n"
	menu += "• reverse <text> - Reverse text\n"
	menu += "• upper <text> - Uppercase text\n"
	menu += "• lower <text> - Lowercase text\n"
	menu += "• count <text> - Count characters\n\n"
	
	menu += "⛏️ *Mining:* mine, mining, pickaxeshop\n"
	menu += "❤️ *Health:* health, usepotion, potionshop\n"
	menu += "💰 *Economy:* work, daily, shop, buy\n"
	menu += "⭐ *Level:* level, leaderboard, roles\n\n"
	
	menu += fmt.Sprintf("🔰 Prefix: %s\n", gc.cfg.Prefix)
	menu += fmt.Sprintf("👑 Owner: %s", gc.cfg.NameOwner)
	
	msg.Reply(menu)
}

func (gc *GeneralCommands) handlePing(msg *CommandMessage) {
	start := time.Now()
	response := "🏓 Pong!"
	duration := time.Since(start)
	
	response += fmt.Sprintf("\n⚡ Response time: %v", duration)
	response += fmt.Sprintf("\n📊 Uptime: %s", gc.getUptimeString())
	
	msg.Reply(response)
}

func (gc *GeneralCommands) handleRuntime(msg *CommandMessage) {
	uptime := gc.db.GetUptime()
	hours := uptime / 3600
	minutes := (uptime % 3600) / 60
	seconds := uptime % 60
	
	response := "⏰ *Bot Runtime*\n\n"
	response += fmt.Sprintf("🕐 Uptime: %dh %dm %ds\n", hours, minutes, seconds)
	response += fmt.Sprintf("📊 Messages: %d\n", gc.db.Stats.TotalMessages)
	response += fmt.Sprintf("👥 Users: %d\n", gc.db.Stats.TotalUsers)
	response += fmt.Sprintf("💬 Chats: %d", gc.db.Stats.TotalChats)
	
	msg.Reply(response)
}

func (gc *GeneralCommands) handleOwner(msg *CommandMessage) {
	response := "👑 *Bot Owner Information*\n\n"
	response += fmt.Sprintf("📛 Name: %s\n", gc.cfg.NameOwner)
	response += fmt.Sprintf("📱 Number: %s\n", gc.cfg.NumberOwner)
	response += fmt.Sprintf("📧 Email: %s\n", gc.cfg.Mail)
	response += fmt.Sprintf("🤖 Bot: %s\n", gc.cfg.NameBot)
	response += fmt.Sprintf("🌐 Website: %s", gc.cfg.Web)
	
	msg.Reply(response)
}

func (gc *GeneralCommands) handleScript(msg *CommandMessage) {
	response := "📜 *Bot Source Code*\n\n"
	response += "🔗 GitHub: https://github.com/YourRepo/ZumyGo\n"
	response += "💻 Language: Go\n"
	response += "📚 Framework: Whatsmeow\n"
	response += "⚡ Version: 2.0\n\n"
	response += "🌟 Features:\n"
	response += "• Mining System\n"
	response += "• Health System\n"
	response += "• Economy System\n"
	response += "• Leveling System\n"
	response += "• Web Dashboard\n\n"
	response += "💝 Free & Open Source!"
	
	msg.Reply(response)
}

// === FUN COMMAND HANDLERS ===

func (gc *GeneralCommands) handleSay(msg *CommandMessage) {
	if len(msg.Args) == 0 {
		msg.Reply("❌ Usage: say <text>")
		return
	}
	
	text := strings.Join(msg.Args, " ")
	msg.Reply(text)
}

func (gc *GeneralCommands) handleTruth(msg *CommandMessage) {
	truths := []string{
		"Apa hal paling memalukan yang pernah kamu lakukan?",
		"Siapa crush kamu saat ini?",
		"Apa rahasia yang belum pernah kamu ceritakan ke siapapun?",
		"Apa hal terburuk yang pernah kamu pikirkan tentang teman kamu?",
		"Pernahkah kamu berbohong kepada orang tua? Tentang apa?",
		"Apa kebiasaan buruk yang kamu sembunyikan?",
		"Siapa orang yang paling kamu benci?",
		"Apa hal paling childish yang masih kamu lakukan?",
		"Pernahkah kamu stalking media sosial mantan?",
		"Apa mimpi paling aneh yang pernah kamu alami?",
	}
	
	truth := truths[rand.Intn(len(truths))]
	response := "🤔 *Truth Question*\n\n" + truth
	msg.Reply(response)
}

func (gc *GeneralCommands) handleDare(msg *CommandMessage) {
	dares := []string{
		"Kirim voice note sambil nyanyi lagu kebangsaan!",
		"Update status WA dengan hal paling memalukan tentang dirimu!",
		"Telepon crush kamu dan bilang kamu kangen!",
		"Minta maaf ke orang yang pernah kamu sakiti!",
		"Post foto jelek kamu di story selama 1 jam!",
		"Chat mantan dan bilang 'hai, apa kabar?'",
		"Teriak dari jendela 'AKU CINTA KAMU SEMUA!'",
		"Makan sesuatu tanpa menggunakan tangan selama 1 menit!",
		"Dance TikTok dan kirim videonya ke grup!",
		"Bilang ke orang tua kamu bahwa kamu sudah punya pacar!",
	}
	
	dare := dares[rand.Intn(len(dares))]
	response := "😈 *Dare Challenge*\n\n" + dare
	msg.Reply(response)
}

func (gc *GeneralCommands) handleRate(msg *CommandMessage) {
	if len(msg.Args) == 0 {
		msg.Reply("❌ Usage: rate <something>")
		return
	}
	
	thing := strings.Join(msg.Args, " ")
	rating := rand.Intn(101)
	
	var emoji string
	if rating >= 80 {
		emoji = "⭐⭐⭐⭐⭐"
	} else if rating >= 60 {
		emoji = "⭐⭐⭐⭐"
	} else if rating >= 40 {
		emoji = "⭐⭐⭐"
	} else if rating >= 20 {
		emoji = "⭐⭐"
	} else {
		emoji = "⭐"
	}
	
	response := fmt.Sprintf("📊 *Rating untuk: %s*\n\n", thing)
	response += fmt.Sprintf("🎯 Score: %d/100\n", rating)
	response += fmt.Sprintf("⭐ Rating: %s", emoji)
	
	msg.Reply(response)
}

func (gc *GeneralCommands) handleCouple(msg *CommandMessage) {
	couples := []string{
		"❤️ Perfect Match! Kalian cocok banget!",
		"💕 Sweet Couple! Relationship goals nih!",
		"💖 Cute Together! Manis banget kalian berdua!",
		"💝 Made for Each Other! Jodoh banget!",
		"💘 Love Birds! Romantis sekali!",
		"💗 Soulmates! Belahan jiwa kalian!",
		"💓 Heartbeat! Bikin baper deh!",
		"💞 True Love! Cinta sejati nih!",
		"💌 Love Letter! Kirim surat cinta yuk!",
		"🌹 Rose for You! Kasih bunga dong!",
	}
	
	couple := couples[rand.Intn(len(couples))]
	percentage := rand.Intn(101)
	
	response := "💕 *Couple Compatibility*\n\n"
	response += fmt.Sprintf("💖 Compatibility: %d%%\n", percentage)
	response += fmt.Sprintf("💌 Result: %s", couple)
	
	msg.Reply(response)
}

func (gc *GeneralCommands) handleQuotes(msg *CommandMessage) {
	quotes := []string{
		"\"Hidup ini seperti sepeda, agar tetap seimbang kamu harus terus bergerak.\" - Albert Einstein",
		"\"Masa depan milik mereka yang percaya pada keindahan mimpi mereka.\" - Eleanor Roosevelt",
		"\"Jangan takut gagal, takutlah tidak mencoba.\" - Unknown",
		"\"Kesuksesan adalah kemampuan untuk beralih dari satu kegagalan ke kegagalan lain tanpa kehilangan antusiasme.\" - Winston Churchill",
		"\"Hidup bukan tentang menunggu badai berlalu, tapi belajar menari di tengah hujan.\" - Vivian Greene",
		"\"Cara terbaik untuk memprediksi masa depan adalah menciptakannya.\" - Peter Drucker",
		"\"Jangan biarkan kemarin menggunakan terlalu banyak hari ini.\" - Will Rogers",
		"\"Kebahagiaan bukan sesuatu yang sudah jadi. Itu datang dari tindakan Anda sendiri.\" - Dalai Lama",
		"\"Mimpi tanpa tindakan hanyalah angan-angan. Tindakan tanpa mimpi hanyalah kegiatan.\" - Joel A. Barker",
		"\"Bukan tentang seberapa keras kamu jatuh, tapi seberapa cepat kamu bangkit.\" - Unknown",
	}
	
	quote := quotes[rand.Intn(len(quotes))]
	response := "💭 *Quote of the Day*\n\n" + quote
	msg.Reply(response)
}

func (gc *GeneralCommands) handleMotivasi(msg *CommandMessage) {
	motivations := []string{
		"🔥 Kamu lebih kuat dari yang kamu kira! Terus semangat!",
		"⭐ Setiap hari adalah kesempatan baru untuk menjadi lebih baik!",
		"💪 Jangan menyerah! Kesuksesan ada di depan mata!",
		"🌟 Percaya pada diri sendiri, kamu pasti bisa!",
		"🚀 Mimpi besar dan kerja keras, sukses akan datang!",
		"💎 Kamu adalah berlian yang sedang diasah!",
		"🌈 Setelah hujan pasti ada pelangi!",
		"🏆 Juara tidak dilahirkan, tapi dibentuk!",
		"🔥 Semangat pagi! Hari ini adalah hari mu!",
		"⚡ Energi positif mu menular! Terus berbagi kebaikan!",
	}
	
	motivation := motivations[rand.Intn(len(motivations))]
	response := "💪 *Motivasi Hari Ini*\n\n" + motivation
	msg.Reply(response)
}

func (gc *GeneralCommands) handleFaktaUnik(msg *CommandMessage) {
	facts := []string{
		"🐙 Gurita memiliki 3 jantung dan darah berwarna biru!",
		"🍯 Madu tidak akan pernah basi, bahkan setelah ribuan tahun!",
		"🐘 Gajah adalah satu-satunya mamalia yang tidak bisa melompat!",
		"🌙 Bulan menjauh dari Bumi sekitar 3.8 cm setiap tahunnya!",
		"🐧 Penguin dapat melompat setinggi 6 kaki ke udara!",
		"🧠 Otak manusia menggunakan sekitar 20% dari total energi tubuh!",
		"🦋 Kupu-kupu merasakan dengan kaki mereka!",
		"🐨 Koala tidur hingga 22 jam sehari!",
		"🌍 Bumi berputar dengan kecepatan 1.670 km/jam di khatulistiwa!",
		"🐠 Ikan mas memiliki ingatan lebih dari 3 detik, bisa hingga 3 bulan!",
	}
	
	fact := facts[rand.Intn(len(facts))]
	response := "🤓 *Fakta Unik*\n\n" + fact
	msg.Reply(response)
}

// === RANDOM COMMAND HANDLERS ===

func (gc *GeneralCommands) handleDadu(msg *CommandMessage) {
	dice1 := rand.Intn(6) + 1
	dice2 := rand.Intn(6) + 1
	total := dice1 + dice2
	
	diceEmojis := []string{"⚀", "⚁", "⚂", "⚃", "⚄", "⚅"}
	
	response := "🎲 *Dice Roll*\n\n"
	response += fmt.Sprintf("🎯 Dice 1: %s (%d)\n", diceEmojis[dice1-1], dice1)
	response += fmt.Sprintf("🎯 Dice 2: %s (%d)\n", diceEmojis[dice2-1], dice2)
	response += fmt.Sprintf("🏆 Total: %d", total)
	
	msg.Reply(response)
}

func (gc *GeneralCommands) handleKoin(msg *CommandMessage) {
	result := rand.Intn(2)
	var coin string
	
	if result == 0 {
		coin = "🪙 HEADS (Gambar)"
	} else {
		coin = "🪙 TAILS (Angka)"
	}
	
	response := "🪙 *Coin Flip*\n\n" + coin
	msg.Reply(response)
}

func (gc *GeneralCommands) handleSlot(msg *CommandMessage) {
	symbols := []string{"🍒", "🍋", "🍊", "🍇", "⭐", "💎", "🔔", "7️⃣"}
	
	slot1 := symbols[rand.Intn(len(symbols))]
	slot2 := symbols[rand.Intn(len(symbols))]
	slot3 := symbols[rand.Intn(len(symbols))]
	
	response := "🎰 *Slot Machine*\n\n"
	response += fmt.Sprintf("[ %s | %s | %s ]\n\n", slot1, slot2, slot3)
	
	if slot1 == slot2 && slot2 == slot3 {
		response += "🎉 JACKPOT! Triple match!"
		// Add reward logic here
		user := msg.User
		reward := int64(1000)
		user.Money += reward
		response += fmt.Sprintf("\n💰 You won %d coins!", reward)
	} else if slot1 == slot2 || slot2 == slot3 || slot1 == slot3 {
		response += "✨ Double match! Nice!"
		user := msg.User
		reward := int64(100)
		user.Money += reward
		response += fmt.Sprintf("\n💰 You won %d coins!", reward)
	} else {
		response += "😔 No match. Try again!"
	}
	
	msg.Reply(response)
}

func (gc *GeneralCommands) handleTebakAngka(msg *CommandMessage) {
	if len(msg.Args) == 0 {
		msg.Reply("❌ Usage: tebakangka <1-100>")
		return
	}
	
	guess, err := strconv.Atoi(msg.Args[0])
	if err != nil || guess < 1 || guess > 100 {
		msg.Reply("❌ Please enter a number between 1-100!")
		return
	}
	
	target := rand.Intn(100) + 1
	difference := abs(guess - target)
	
	response := "🎯 *Number Guessing Game*\n\n"
	response += fmt.Sprintf("🎲 Your guess: %d\n", guess)
	response += fmt.Sprintf("🎯 Target number: %d\n", target)
	response += fmt.Sprintf("📏 Difference: %d\n\n", difference)
	
	if guess == target {
		response += "🎉 PERFECT! Exact match!"
		user := msg.User
		reward := int64(500)
		user.Money += reward
		response += fmt.Sprintf("\n💰 You won %d coins!", reward)
	} else if difference <= 5 {
		response += "🔥 Very close! Great guess!"
		user := msg.User
		reward := int64(100)
		user.Money += reward
		response += fmt.Sprintf("\n💰 You won %d coins!", reward)
	} else if difference <= 10 {
		response += "👍 Close! Good try!"
		user := msg.User
		reward := int64(50)
		user.Money += reward
		response += fmt.Sprintf("\n💰 You won %d coins!", reward)
	} else {
		response += "😅 Not quite! Try again!"
	}
	
	msg.Reply(response)
}

// === TEXT COMMAND HANDLERS ===

func (gc *GeneralCommands) handleReverse(msg *CommandMessage) {
	if len(msg.Args) == 0 {
		msg.Reply("❌ Usage: reverse <text>")
		return
	}
	
	text := strings.Join(msg.Args, " ")
	reversed := reverseString(text)
	
	response := "🔄 *Text Reverser*\n\n"
	response += fmt.Sprintf("📝 Original: %s\n", text)
	response += fmt.Sprintf("🔄 Reversed: %s", reversed)
	
	msg.Reply(response)
}

func (gc *GeneralCommands) handleUpper(msg *CommandMessage) {
	if len(msg.Args) == 0 {
		msg.Reply("❌ Usage: upper <text>")
		return
	}
	
	text := strings.Join(msg.Args, " ")
	upper := strings.ToUpper(text)
	
	response := "🔠 *Text Uppercase*\n\n"
	response += fmt.Sprintf("📝 Original: %s\n", text)
	response += fmt.Sprintf("🔠 Uppercase: %s", upper)
	
	msg.Reply(response)
}

func (gc *GeneralCommands) handleLower(msg *CommandMessage) {
	if len(msg.Args) == 0 {
		msg.Reply("❌ Usage: lower <text>")
		return
	}
	
	text := strings.Join(msg.Args, " ")
	lower := strings.ToLower(text)
	
	response := "🔡 *Text Lowercase*\n\n"
	response += fmt.Sprintf("📝 Original: %s\n", text)
	response += fmt.Sprintf("🔡 Lowercase: %s", lower)
	
	msg.Reply(response)
}

func (gc *GeneralCommands) handleCount(msg *CommandMessage) {
	if len(msg.Args) == 0 {
		msg.Reply("❌ Usage: count <text>")
		return
	}
	
	text := strings.Join(msg.Args, " ")
	chars := len(text)
	words := len(strings.Fields(text))
	
	response := "📊 *Text Counter*\n\n"
	response += fmt.Sprintf("📝 Text: %s\n", text)
	response += fmt.Sprintf("🔢 Characters: %d\n", chars)
	response += fmt.Sprintf("📝 Words: %d", words)
	
	msg.Reply(response)
}

// === CALCULATOR ===

func (gc *GeneralCommands) handleCalc(msg *CommandMessage) {
	if len(msg.Args) == 0 {
		msg.Reply("❌ Usage: calc <expression>\nExample: calc 5 + 3")
		return
	}
	
	expression := strings.Join(msg.Args, " ")
	
	// Simple calculator (basic operations only)
	result, err := gc.evaluateExpression(expression)
	if err != nil {
		msg.Reply("❌ Invalid expression! Use +, -, *, / operators")
		return
	}
	
	response := "🧮 *Calculator*\n\n"
	response += fmt.Sprintf("📝 Expression: %s\n", expression)
	response += fmt.Sprintf("🔢 Result: %.2f", result)
	
	msg.Reply(response)
}

// === TIME COMMANDS ===

func (gc *GeneralCommands) handleWaktu(msg *CommandMessage) {
	now := time.Now()
	
	response := "🕐 *Current Time*\n\n"
	response += fmt.Sprintf("⏰ Time: %s\n", now.Format("15:04:05"))
	response += fmt.Sprintf("📅 Date: %s\n", now.Format("2006-01-02"))
	response += fmt.Sprintf("📆 Day: %s\n", now.Format("Monday"))
	response += fmt.Sprintf("🌍 Timezone: %s", now.Format("MST"))
	
	msg.Reply(response)
}

func (gc *GeneralCommands) handleTanggal(msg *CommandMessage) {
	now := time.Now()
	
	response := "📅 *Current Date*\n\n"
	response += fmt.Sprintf("📅 Date: %s\n", now.Format("Monday, 2 January 2006"))
	response += fmt.Sprintf("📆 Short: %s\n", now.Format("02/01/2006"))
	response += fmt.Sprintf("🗓️ ISO: %s\n", now.Format("2006-01-02"))
	response += fmt.Sprintf("⏰ Time: %s", now.Format("15:04:05"))
	
	msg.Reply(response)
}

// === PROFILE COMMANDS ===

func (gc *GeneralCommands) handleProfile(msg *CommandMessage) {
	user := msg.User
	
	response := "👤 *Your Profile*\n\n"
	response += fmt.Sprintf("📛 Name: %s\n", user.Name)
	response += fmt.Sprintf("📝 Bio: %s\n", user.Bio)
	response += fmt.Sprintf("⭐ Level: %d\n", user.Level)
	response += fmt.Sprintf("🏷️ Role: %s\n", user.Role)
	response += fmt.Sprintf("💰 Money: %d coins\n", user.Money)
	response += fmt.Sprintf("🪙 ZumyCoin: %d ZC\n", user.ZC)
	response += fmt.Sprintf("❤️ Health: %d/%d HP\n", user.Health.Health, user.Health.MaxHealth)
	response += fmt.Sprintf("⛏️ Mining Level: %d\n", user.Mining.Level)
	response += fmt.Sprintf("📅 Joined: %s", time.Unix(user.RegisterTime, 0).Format("2006-01-02"))
	
	msg.Reply(response)
}

func (gc *GeneralCommands) handleSetName(msg *CommandMessage) {
	if len(msg.Args) == 0 {
		msg.Reply("❌ Usage: setname <name>")
		return
	}
	
	name := strings.Join(msg.Args, " ")
	if len(name) > 50 {
		msg.Reply("❌ Name too long! Maximum 50 characters.")
		return
	}
	
	user := msg.User
	oldName := user.Name
	user.Name = name
	
	response := "✅ *Name Updated*\n\n"
	response += fmt.Sprintf("📛 Old name: %s\n", oldName)
	response += fmt.Sprintf("📛 New name: %s", name)
	
	msg.Reply(response)
}

func (gc *GeneralCommands) handleSetBio(msg *CommandMessage) {
	if len(msg.Args) == 0 {
		msg.Reply("❌ Usage: setbio <bio>")
		return
	}
	
	bio := strings.Join(msg.Args, " ")
	if len(bio) > 100 {
		msg.Reply("❌ Bio too long! Maximum 100 characters.")
		return
	}
	
	user := msg.User
	user.Bio = bio
	
	response := "✅ *Bio Updated*\n\n"
	response += fmt.Sprintf("📝 New bio: %s", bio)
	
	msg.Reply(response)
}

// === HELPER FUNCTIONS ===

func (gc *GeneralCommands) getUptimeString() string {
	uptime := gc.db.GetUptime()
	hours := uptime / 3600
	minutes := (uptime % 3600) / 60
	seconds := uptime % 60
	
	return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (gc *GeneralCommands) evaluateExpression(expr string) (float64, error) {
	// Simple expression evaluator for basic math
	expr = strings.ReplaceAll(expr, " ", "")
	
	// Handle basic operations
	if strings.Contains(expr, "+") {
		parts := strings.Split(expr, "+")
		if len(parts) == 2 {
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				return a + b, nil
			}
		}
	} else if strings.Contains(expr, "-") {
		parts := strings.Split(expr, "-")
		if len(parts) == 2 {
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				return a - b, nil
			}
		}
	} else if strings.Contains(expr, "*") {
		parts := strings.Split(expr, "*")
		if len(parts) == 2 {
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil {
				return a * b, nil
			}
		}
	} else if strings.Contains(expr, "/") {
		parts := strings.Split(expr, "/")
		if len(parts) == 2 {
			a, err1 := strconv.ParseFloat(parts[0], 64)
			b, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil && b != 0 {
				return a / b, nil
			}
		}
	}
	
	return 0, fmt.Errorf("unsupported expression")
}