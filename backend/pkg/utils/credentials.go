package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

const (
	// UsernamePrefix for SFTP users
	UsernamePrefix = "mc"
	// PasswordLength for generated passwords
	PasswordLength = 16
)

// GenerateSFTPCredentials generates a random username and password for SFTP access
func GenerateSFTPCredentials(serverName string) (username, password string, err error) {
	// Generate username from server name (sanitized)
	sanitized := sanitizeServerName(serverName)
	username = fmt.Sprintf("%s-%s", UsernamePrefix, sanitized)

	// Generate secure random password
	password, err = generateSecurePassword(PasswordLength)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate password: %w", err)
	}

	return username, password, nil
}

// sanitizeServerName removes invalid characters for usernames
func sanitizeServerName(name string) string {
	// Replace underscores and dots with dashes, remove other special chars
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, ".", "-")
	name = strings.ToLower(name)

	// Keep only alphanumeric and dashes
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	sanitized := result.String()

	// Limit length to 20 characters
	if len(sanitized) > 20 {
		sanitized = sanitized[:20]
	}

	return sanitized
}

// generateSecurePassword generates a cryptographically secure random password
func generateSecurePassword(length int) (string, error) {
	// Generate random bytes
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Encode to base64 and clean up
	password := base64.URLEncoding.EncodeToString(bytes)

	// Remove special characters that might cause issues
	password = strings.ReplaceAll(password, "-", "")
	password = strings.ReplaceAll(password, "_", "")
	password = strings.ReplaceAll(password, "=", "")

	// Ensure it's the right length
	if len(password) > length {
		password = password[:length]
	}

	return password, nil
}
