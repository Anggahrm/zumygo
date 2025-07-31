package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"sync"
	"time"
	"zumygo/config"
	"zumygo/database"
	"zumygo/helpers"
	"zumygo/systems"
)

// Plugin represents a bot plugin
type Plugin interface {
	Name() string
	Description() string
	Commands() []Command
	Initialize(*PluginContext) error
	Cleanup() error
}

// Command represents a plugin command
type Command struct {
	Name        string
	Aliases     []string
	Description string
	Usage       string
	Category    string
	Cooldown    time.Duration
	OwnerOnly   bool
	AdminOnly   bool
	PremiumOnly bool
	GroupOnly   bool
	PrivateOnly bool
	Handler     CommandHandler
}

// CommandHandler is the function signature for command handlers
type CommandHandler func(*PluginContext, *CommandMessage) error

// CommandMessage represents a command message
type CommandMessage struct {
	ID          string
	From        string
	Chat        string
	Text        string
	Command     string
	Args        []string
	IsGroup     bool
	IsOwner     bool
	IsAdmin     bool
	IsPremium   bool
	User        *database.User
	ChatData    *database.Chat
	Reply       func(string) error
	React       func(string) error
	Delete      func() error
}

// PluginContext provides access to bot systems
type PluginContext struct {
	Config        *config.BotConfig
	Database      *database.Database
	MiningSystem  *systems.MiningSystem
	Logger        *helpers.Logger
	Prefix        string
}

// PluginManager manages all loaded plugins
type PluginManager struct {
	plugins     map[string]Plugin
	commands    map[string]*Command
	aliases     map[string]string
	context     *PluginContext
	pluginDir   string
	watchers    map[string]*time.Timer
	mutex       sync.RWMutex
	logger      *helpers.Logger
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(ctx *PluginContext, pluginDir string) *PluginManager {
	return &PluginManager{
		plugins:   make(map[string]Plugin),
		commands:  make(map[string]*Command),
		aliases:   make(map[string]string),
		context:   ctx,
		pluginDir: pluginDir,
		watchers:  make(map[string]*time.Timer),
		logger:    ctx.Logger,
	}
}

// LoadAllPlugins loads all plugins from the plugin directory
func (pm *PluginManager) LoadAllPlugins() error {
	if _, err := os.Stat(pm.pluginDir); os.IsNotExist(err) {
		if err := os.MkdirAll(pm.pluginDir, 0755); err != nil {
			return fmt.Errorf("failed to create plugin directory: %v", err)
		}
		pm.createExamplePlugins()
	}

	return filepath.Walk(pm.pluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(info.Name(), ".so") {
			return pm.LoadPlugin(path)
		}

		return nil
	})
}

// LoadPlugin loads a single plugin
func (pm *PluginManager) LoadPlugin(path string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Load the plugin
	p, err := plugin.Open(path)
	if err != nil {
		pm.logger.Error(fmt.Sprintf("Failed to open plugin %s: %v", path, err))
		return err
	}

	// Get the plugin symbol
	pluginSymbol, err := p.Lookup("Plugin")
	if err != nil {
		pm.logger.Error(fmt.Sprintf("Plugin %s missing Plugin symbol: %v", path, err))
		return err
	}

	// Type assert to Plugin interface
	pluginInstance, ok := pluginSymbol.(Plugin)
	if !ok {
		pm.logger.Error(fmt.Sprintf("Plugin %s does not implement Plugin interface", path))
		return fmt.Errorf("invalid plugin interface")
	}

	// Initialize plugin
	if err := pluginInstance.Initialize(pm.context); err != nil {
		pm.logger.Error(fmt.Sprintf("Failed to initialize plugin %s: %v", path, err))
		return err
	}

	pluginName := pluginInstance.Name()
	
	// Unload existing plugin if it exists
	if existingPlugin, exists := pm.plugins[pluginName]; exists {
		existingPlugin.Cleanup()
		pm.unregisterCommands(pluginName)
	}

	// Register new plugin
	pm.plugins[pluginName] = pluginInstance
	pm.registerCommands(pluginInstance)

	pm.logger.Info(fmt.Sprintf("Successfully loaded plugin: %s", pluginName))
	return nil
}

// registerCommands registers all commands from a plugin
func (pm *PluginManager) registerCommands(plugin Plugin) {
	for _, cmd := range plugin.Commands() {
		// Register main command
		pm.commands[cmd.Name] = &cmd
		
		// Register aliases
		for _, alias := range cmd.Aliases {
			pm.aliases[alias] = cmd.Name
		}
	}
}

// unregisterCommands unregisters all commands from a plugin
func (pm *PluginManager) unregisterCommands(pluginName string) {
	plugin := pm.plugins[pluginName]
	if plugin == nil {
		return
	}

	for _, cmd := range plugin.Commands() {
		delete(pm.commands, cmd.Name)
		
		for _, alias := range cmd.Aliases {
			delete(pm.aliases, alias)
		}
	}
}

// ExecuteCommand executes a command if it exists
func (pm *PluginManager) ExecuteCommand(msg *CommandMessage) error {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	commandName := strings.ToLower(msg.Command)
	
	// Check if it's an alias
	if realCommand, exists := pm.aliases[commandName]; exists {
		commandName = realCommand
	}

	// Find the command
	cmd, exists := pm.commands[commandName]
	if !exists {
		return fmt.Errorf("command not found: %s", commandName)
	}

	// Check permissions
	if err := pm.checkPermissions(cmd, msg); err != nil {
		return err
	}

	// Execute command
	return cmd.Handler(pm.context, msg)
}

// checkPermissions checks if user has permission to execute command
func (pm *PluginManager) checkPermissions(cmd *Command, msg *CommandMessage) error {
	if cmd.OwnerOnly && !msg.IsOwner {
		return fmt.Errorf("this command is owner only")
	}

	if cmd.AdminOnly && !msg.IsAdmin && !msg.IsOwner {
		return fmt.Errorf("this command is admin only")
	}

	if cmd.PremiumOnly && !msg.IsPremium && !msg.IsOwner {
		return fmt.Errorf("this command is premium only")
	}

	if cmd.GroupOnly && !msg.IsGroup {
		return fmt.Errorf("this command can only be used in groups")
	}

	if cmd.PrivateOnly && msg.IsGroup {
		return fmt.Errorf("this command can only be used in private chat")
	}

	return nil
}

// GetCommands returns all registered commands
func (pm *PluginManager) GetCommands() map[string]*Command {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	commands := make(map[string]*Command)
	for name, cmd := range pm.commands {
		commands[name] = cmd
	}
	return commands
}

// GetCommandsByCategory returns commands grouped by category
func (pm *PluginManager) GetCommandsByCategory() map[string][]*Command {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	categories := make(map[string][]*Command)
	for _, cmd := range pm.commands {
		category := cmd.Category
		if category == "" {
			category = "General"
		}
		categories[category] = append(categories[category], cmd)
	}
	return categories
}

// WatchPluginDirectory starts watching the plugin directory for changes
func (pm *PluginManager) WatchPluginDirectory() {
	go func() {
		for {
			time.Sleep(5 * time.Second) // Check every 5 seconds
			
			filepath.Walk(pm.pluginDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if strings.HasSuffix(info.Name(), ".so") {
					// Check if file was modified
					if timer, exists := pm.watchers[path]; exists {
						timer.Stop()
					}

					pm.watchers[path] = time.AfterFunc(1*time.Second, func() {
						pm.logger.Info(fmt.Sprintf("Detected change in plugin: %s", path))
						if err := pm.LoadPlugin(path); err != nil {
							pm.logger.Error(fmt.Sprintf("Failed to reload plugin %s: %v", path, err))
						}
					})
				}

				return nil
			})
		}
	}()
}

// createExamplePlugins creates example plugins for reference
func (pm *PluginManager) createExamplePlugins() {
	// Create example plugin source files
	examplePluginGo := `package main

import (
	"fmt"
	"time"
	"zumygo/plugins"
)

type ExamplePlugin struct{}

func (p *ExamplePlugin) Name() string {
	return "example"
}

func (p *ExamplePlugin) Description() string {
	return "Example plugin demonstrating basic functionality"
}

func (p *ExamplePlugin) Commands() []plugins.Command {
	return []plugins.Command{
		{
			Name:        "ping",
			Aliases:     []string{"p"},
			Description: "Check bot responsiveness",
			Usage:       "ping",
			Category:    "General",
			Cooldown:    5 * time.Second,
			Handler:     p.handlePing,
		},
		{
			Name:        "info",
			Description: "Show user information",
			Usage:       "info [@user]",
			Category:    "General",
			Handler:     p.handleInfo,
		},
	}
}

func (p *ExamplePlugin) Initialize(ctx *plugins.PluginContext) error {
	ctx.Logger.Info("Example plugin initialized")
	return nil
}

func (p *ExamplePlugin) Cleanup() error {
	return nil
}

func (p *ExamplePlugin) handlePing(ctx *plugins.PluginContext, msg *plugins.CommandMessage) error {
	return msg.Reply("🏓 Pong! Bot is responsive.")
}

func (p *ExamplePlugin) handleInfo(ctx *plugins.PluginContext, msg *plugins.CommandMessage) error {
	user := msg.User
	info := fmt.Sprintf("👤 *User Information*\n\n")
	info += fmt.Sprintf("Name: %s\n", user.Name)
	info += fmt.Sprintf("Level: %d\n", user.Level)
	info += fmt.Sprintf("Experience: %d\n", user.Exp)
	info += fmt.Sprintf("Money: %d coins\n", user.Money)
	info += fmt.Sprintf("ZumyCoin: %d ZC\n", user.ZC)
	info += fmt.Sprintf("Role: %s\n", user.Role)
	info += fmt.Sprintf("Premium: %t\n", user.Premium)
	
	return msg.Reply(info)
}

// Plugin is the exported symbol that the plugin manager looks for
var Plugin ExamplePlugin
`

	// Write example plugin
	examplePath := filepath.Join(pm.pluginDir, "example.go")
	if err := os.WriteFile(examplePath, []byte(examplePluginGo), 0644); err != nil {
		pm.logger.Error(fmt.Sprintf("Failed to create example plugin: %v", err))
	}

	// Create README
	readme := `# ZumyGo Plugins

This directory contains plugins for the ZumyGo bot.

## Creating a Plugin

1. Create a new Go file (e.g., myplugin.go)
2. Implement the Plugin interface
3. Build as a shared library: go build -buildmode=plugin myplugin.go
4. The plugin will be automatically loaded

## Plugin Interface

Your plugin must implement:
- Name() string
- Description() string  
- Commands() []Command
- Initialize(*PluginContext) error
- Cleanup() error

## Example

See example.go for a basic plugin implementation.

## Building Plugins

To build a plugin:
` + "```bash" + `
go build -buildmode=plugin example.go
` + "```" + `

This will create example.so which will be automatically loaded.
`

	readmePath := filepath.Join(pm.pluginDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
		pm.logger.Error(fmt.Sprintf("Failed to create plugin README: %v", err))
	}

	pm.logger.Info("Created example plugin files in " + pm.pluginDir)
}

// GetPluginInfo returns information about loaded plugins
func (pm *PluginManager) GetPluginInfo() string {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	if len(pm.plugins) == 0 {
		return "No plugins loaded."
	}

	info := "🔌 *Loaded Plugins*\n\n"
	for name, plugin := range pm.plugins {
		info += fmt.Sprintf("**%s**\n", name)
		info += fmt.Sprintf("Description: %s\n", plugin.Description())
		info += fmt.Sprintf("Commands: %d\n\n", len(plugin.Commands()))
	}

	return info
}

// ReloadPlugin reloads a specific plugin
func (pm *PluginManager) ReloadPlugin(name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Find plugin file
	var pluginPath string
	filepath.Walk(pm.pluginDir, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, name) && strings.HasSuffix(path, ".so") {
			pluginPath = path
		}
		return nil
	})

	if pluginPath == "" {
		return fmt.Errorf("plugin file for %s not found", name)
	}

	// Cleanup old plugin
	plugin.Cleanup()
	pm.unregisterCommands(name)
	delete(pm.plugins, name)

	// Load new plugin
	return pm.LoadPlugin(pluginPath)
}

// UnloadPlugin unloads a specific plugin
func (pm *PluginManager) UnloadPlugin(name string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	plugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Cleanup and remove
	plugin.Cleanup()
	pm.unregisterCommands(name)
	delete(pm.plugins, name)

	pm.logger.Info(fmt.Sprintf("Unloaded plugin: %s", name))
	return nil
}