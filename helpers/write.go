package helpers

import (
	"os"
	"strings"
)

func UpdateEnvFile(key string, value string) error {
	file, err := os.ReadFile(".env")
	if err != nil {
		return err
	}

	lines := strings.Split(string(file), "\n")
	found := false

	for i, line := range lines {
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
	return os.WriteFile(".env", []byte(output), 0644)
}
