# Zumygo - WhatsApp Bot

A powerful and secure WhatsApp bot built with Go using the [whatsmeow](https://github.com/tulir/whatsmeow) library.

## 🚀 Features

- **Multi-prefix Support** - Use `.`, `!`, or any custom prefix
- **Command System** - Easy to add new commands
- **Owner Protection** - Secure owner-only commands
- **Group & Private Chat Support** - Works in both environments
- **Media Support** - Handle images, videos, documents
- **Clean Logging** - Professional output without spam
- **Error Recovery** - Graceful error handling and recovery

## 📋 Prerequisites

- Go 1.19 or higher
- Git
- WhatsApp account

## 🛠️ Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd zumygo
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Setup environment variables**
   ```bash
   # Copy the example environment file
   cp .env.example .env 
   # Edit the .env file with your configuration
   ```

## ⚙️ Configuration

### Required Environment Variables

#### `OWNER`
Your WhatsApp phone number (without + or country code)
```
OWNER=6281234567890
```

#### `PREFIX`
Command prefixes (comma-separated or JSON array)
```
# Comma-separated
PREFIX=.,!

# JSON array
PREFIX=[".", "!", "/"]
```

### Optional Environment Variables

#### `PUBLIC`
Set to "true" to allow all users to use commands
```
PUBLIC=false
```

#### `PAIRING_NUMBER`
Phone number for pairing (leave empty for QR code)
```
PAIRING_NUMBER=6281234567890
```

## 🚀 Running the Bot

1. **Run the bot**
   ```bash
   go run .
   ```

2. **Login**
   - If `PAIRING_NUMBER` is set: Enter the pairing code
   - If not set: Scan the QR code with WhatsApp

## 📚 Available Commands

### Main Commands
- `.ping` / `!ping` - Test bot response time
- `.menu` / `!menu` - Show command list

### Owner Commands
- `.mode` / `!mode` - Mode settings (owner only)

## 📁 Project Structure

```
zumygo/
├── commands/          # Command handlers
│   ├── main/         # Main commands
│   │   ├── menu.go   # Menu command
│   │   └── ping.go   # Ping command
│   └── owner/        # Owner commands
│       └── mode.go   # Mode settings
├── handlers/          # Message handlers
│   └── message.go    # Message processing
├── helpers/           # Helper functions
├── libs/             # Core libraries
│   ├── client.go     # Client wrapper
│   ├── commands.go   # Command management
│   ├── message.go    # Message serialization
│   └── types.go      # Type definitions
├── .env              # Environment variables (create from env.example)
├── .gitignore        # Git ignore rules
├── env.example       # Environment variables template
├── go.mod            # Go modules
├── go.sum            # Go modules checksum
├── main.go           # Entry point
├── README.md         # This file
├── SETUP.md          # Setup documentation
└── zumygo.go         # Main bot logic
```

## 🔧 Development

### Adding New Commands

1. Create a new file in `commands/` directory
2. Define your command structure
3. Register the command in `init()` function
4. Rebuild the bot

### Example Command

```go
package commands

import "zumygo/libs"

func init() {
    libs.NewCommands(&libs.ICommand{
        Name:     "hello",
        As:       []string{"hello"},
        Tags:     "main",
        IsPrefix: true,
        Execute: func(conn *libs.IClient, m *libs.IMessage) bool {
            m.Reply("Hello, World!")
            return true
        },
    })
}
```

### Command Properties

| Property | Description |
|----------|-------------|
| `Name` | Command name (supports regex) |
| `As` | Alternative names |
| `Tags` | Category for menu grouping |
| `IsPrefix` | Whether command requires prefix |
| `IsOwner` | Owner-only command |
| `IsGroup` | Group-only command |
| `IsPrivate` | Private-only command |
| `IsMedia` | Requires media attachment |
| `IsQuery` | Requires text input |
| `IsWait` | Shows loading indicator |
| `Execute` | Command handler function |

## 🔒 Security Features

- **Owner Protection** - Commands can be restricted to owner only
- **Prefix Validation** - Commands only work with valid prefixes
- **Input Sanitization** - Safe command processing
- **Error Recovery** - Graceful error handling
- **Timeout Protection** - Prevents hanging commands

## 📝 Logging

The bot provides clean, professional logging:

```
From : User Name 1234567890
Command : ping
Message : .ping
```

### Log Levels
- **Info** - Bot status and connection
- **Error** - Error messages and recovery
- **User** - Message and command activity

## 🐛 Troubleshooting

### Common Issues

1. **"OWNER environment variable is required"**
   - Make sure you have created `.env` file from `env.example`
   - Check that `OWNER` variable is set correctly

2. **"PREFIX environment variable is required"**
   - Ensure `PREFIX` variable is set in `.env` file

3. **Connection issues**
   - Check your internet connection
   - Try using QR code login instead of pairing

4. **Command not working**
   - Verify the command prefix matches your configuration
   - Check if you're the owner (if `PUBLIC=false`)

### Logs

The bot generates logs in the console. Common log messages:
- `Configuration validation passed` - Environment setup is correct
- `Starting WhatsApp bot...` - Bot is starting
- `Connecting Socket` - Connecting to WhatsApp
- `Connected Socket` - Successfully connected
- `Qr Required` - QR code login required

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [whatsmeow](https://github.com/tulir/whatsmeow) - WhatsApp Web API library
- [otto](https://github.com/robertkrimen/otto) - JavaScript VM (if used)
en an issue on GitHub

---

**Note**: This bot is for educational and personal use. Please respect WhatsApp's terms of service and use responsibly. 
