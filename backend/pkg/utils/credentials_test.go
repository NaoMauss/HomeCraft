package utils

import (
	"strings"
	"testing"
)

func TestGenerateSFTPCredentials(t *testing.T) {
	tests := []struct {
		name               string
		serverName         string
		wantUsernamePrefix string
		wantErr            bool
	}{
		{
			name:               "simple server name",
			serverName:         "test-server",
			wantUsernamePrefix: "mc-test-server",
			wantErr:            false,
		},
		{
			name:               "server name with underscores",
			serverName:         "my_awesome_server",
			wantUsernamePrefix: "mc-my-awesome-server",
			wantErr:            false,
		},
		{
			name:               "server name with dots",
			serverName:         "server.prod.v1",
			wantUsernamePrefix: "mc-server-prod-v1",
			wantErr:            false,
		},
		{
			name:               "server name with mixed case",
			serverName:         "MyServer",
			wantUsernamePrefix: "mc-myserver",
			wantErr:            false,
		},
		{
			name:               "server name with special chars",
			serverName:         "test@server#123",
			wantUsernamePrefix: "mc-testserver123",
			wantErr:            false,
		},
		{
			name:               "very long server name",
			serverName:         "this-is-a-very-long-server-name-that-exceeds-twenty-characters",
			wantUsernamePrefix: "mc-this-is-a-very-lon", // Should be truncated
			wantErr:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username, password, err := GenerateSFTPCredentials(tt.serverName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSFTPCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check username format
				if !strings.HasPrefix(username, tt.wantUsernamePrefix) {
					t.Errorf("GenerateSFTPCredentials() username = %v, want prefix %v", username, tt.wantUsernamePrefix)
				}

				// Check password length
				if len(password) != PasswordLength {
					t.Errorf("GenerateSFTPCredentials() password length = %v, want %v", len(password), PasswordLength)
				}

				// Check password only contains alphanumeric characters (no special chars)
				for _, char := range password {
					if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
						t.Errorf("GenerateSFTPCredentials() password contains special character: %c", char)
					}
				}
			}
		})
	}
}

func TestGenerateSFTPCredentials_Uniqueness(t *testing.T) {
	// Generate multiple credentials and ensure passwords are unique
	passwords := make(map[string]bool)
	serverName := "test-server"

	for i := 0; i < 100; i++ {
		_, password, err := GenerateSFTPCredentials(serverName)
		if err != nil {
			t.Fatalf("GenerateSFTPCredentials() unexpected error: %v", err)
		}

		if passwords[password] {
			t.Errorf("GenerateSFTPCredentials() generated duplicate password: %s", password)
		}
		passwords[password] = true
	}

	// Should have 100 unique passwords
	if len(passwords) != 100 {
		t.Errorf("GenerateSFTPCredentials() uniqueness test: got %d unique passwords, want 100", len(passwords))
	}
}

func TestSanitizeServerName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "test",
			expected: "test",
		},
		{
			name:     "name with dashes",
			input:    "test-server",
			expected: "test-server",
		},
		{
			name:     "name with underscores",
			input:    "test_server",
			expected: "test-server",
		},
		{
			name:     "name with dots",
			input:    "test.server",
			expected: "test-server",
		},
		{
			name:     "uppercase name",
			input:    "TEST",
			expected: "test",
		},
		{
			name:     "name with special chars",
			input:    "test@server#123!",
			expected: "testserver123",
		},
		{
			name:     "name with spaces",
			input:    "test server",
			expected: "testserver",
		},
		{
			name:     "very long name",
			input:    "this-is-a-very-long-server-name-that-should-be-truncated",
			expected: "this-is-a-very-long-", // Max 20 chars
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeServerName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeServerName(%q) = %q, want %q", tt.input, result, tt.expected)
			}

			// Check length
			if len(result) > 20 {
				t.Errorf("sanitizeServerName(%q) result too long: %d chars (max 20)", tt.input, len(result))
			}
		})
	}
}

func TestGenerateSecurePassword(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{
			name:    "standard length",
			length:  16,
			wantErr: false,
		},
		{
			name:    "short password",
			length:  8,
			wantErr: false,
		},
		{
			name:    "long password",
			length:  32,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := generateSecurePassword(tt.length)

			if (err != nil) != tt.wantErr {
				t.Errorf("generateSecurePassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Password length should be exactly as requested
				if len(password) != tt.length {
					t.Errorf("generateSecurePassword(%d) length = %d, want %d", tt.length, len(password), tt.length)
				}

				// Should not contain special characters
				if strings.Contains(password, "-") || strings.Contains(password, "_") || strings.Contains(password, "=") {
					t.Errorf("generateSecurePassword() contains unwanted special characters: %s", password)
				}
			}
		})
	}
}

func BenchmarkGenerateSFTPCredentials(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = GenerateSFTPCredentials("test-server")
	}
}

func BenchmarkGenerateSecurePassword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = generateSecurePassword(16)
	}
}
