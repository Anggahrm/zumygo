package helpers

import (
	"fmt"
	"os"
	"strings"
)

func UpdateEnvFile(key string, value string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	
	if strings.Contains(key, "=") {
		return fmt.Errorf("key cannot contain '=' character")
	}
	
	// Check if .env file exists, create if not
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		// Create new .env file
		content := key + "=" + value + "\n"
		return os.WriteFile(".env", []byte(content), 0644)
	}
	
	file, err := os.ReadFile(".env")
	if err != nil {
		return fmt.Errorf("failed to read .env file: %v", err)
	}

	lines := strings.Split(string(file), "\n")
	found := false

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, key+"=") {
			lines[i] = key + "=" + value
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, key+"="+value)
	}

	output := strings.Join(lines, "\n")
	err = os.WriteFile(".env", []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("failed to write .env file: %v", err)
	}
	
	return nil
}
