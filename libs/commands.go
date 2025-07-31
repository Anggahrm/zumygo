package libs

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
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

// GetPrefixes returns all valid prefixes from environment variable
func GetPrefixes() []string {
	return ParseArrayFromEnv("PREFIX")
}

// ParseArrayFromEnv parses an environment variable as JSON array or comma-separated string
func ParseArrayFromEnv(envVar string) []string {
	envValue := os.Getenv(envVar)
	if envValue == "" {
		return []string{}
	}
	
	// Try to parse as JSON array first
	if strings.HasPrefix(envValue, "[") && strings.HasSuffix(envValue, "]") {
		var result []string
		if err := json.Unmarshal([]byte(envValue), &result); err == nil {
			var validItems []string
			for _, item := range result {
				item = strings.TrimSpace(item)
				if item != "" {
					validItems = append(validItems, item)
				}
			}
			return validItems
		}
	}
	
	// Fallback to comma-separated for backward compatibility
	items := strings.Split(envValue, ",")
	var validItems []string
	
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			validItems = append(validItems, item)
		}
	}
	
	return validItems
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
