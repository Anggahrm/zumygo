package libs

import (
	"regexp"
	"strings"
	"zumygo/config"
)

var lists []ICommand

func NewCommands(cmd *ICommand) {
	if cmd == nil {
		return
	}
	
	if cmd.Name == "" {
		return
	}
	
	// Check for duplicate commands
	for _, existing := range lists {
		if existing.Name == cmd.Name {
			return // Skip duplicate
		}
	}
	
	lists = append(lists, *cmd)
}

func GetList() []ICommand {
	return lists
}

// GetPrefixes returns all valid prefixes from config
func GetPrefixes() []string {
	if config.Config != nil {
		return config.Config.GetPrefixes()
	}
	return []string{"."}
}



// ExtractPrefix extracts the prefix from a command string
func ExtractPrefix(command string) (string, bool) {
	prefixes := GetPrefixes()
	
	for _, prefix := range prefixes {
		if strings.HasPrefix(command, prefix) {
			return prefix, true
		}
	}
	
	return "", false
}

func HasCommand(name string) bool {
	if name == "" {
		return false
	}
	
	// Check if name already has a prefix
	prefix, hasPrefix := ExtractPrefix(name)
	var commandName string
	
	if hasPrefix {
		// Remove prefix to get the actual command
		commandName = strings.TrimSpace(strings.TrimPrefix(name, prefix))
	} else {
		// No prefix, use name as is
		commandName = name
	}
	
	for _, cmd := range lists {
		if cmd.Name == "" {
			continue
		}
		
		// Check if cmd.Name contains regex patterns (like |, *, +, etc.)
		if strings.ContainsAny(cmd.Name, "|*+?()[]{}") {
			// Use as regex pattern
			re := regexp.MustCompile(`^` + cmd.Name + `$`)
			if valid := len(re.FindAllString(commandName, -1)) > 0; valid {
				return true
			}
		} else {
			// Use exact match with quote meta for safety
			re := regexp.MustCompile(`^` + regexp.QuoteMeta(cmd.Name) + `$`)
			if valid := len(re.FindAllString(commandName, -1)) > 0; valid {
				return true
			}
		}
	}
	return false
}
