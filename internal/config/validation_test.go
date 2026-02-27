package config

import (
	"os"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected bool
		errors   int
	}{
		{
			name: "Valid config",
			config: Config{
				DBHost:     "localhost",
				DBPort:     "5432",
				DBUser:     "postgres",
				DBPassword: "password",
				DBName:     "testdb",
				DBSSLMode:  "require",
			},
			expected: true,
			errors:   0,
		},
		{
			name: "Missing required fields",
			config: Config{
				DBHost:     "",
				DBPort:     "5432",
				DBUser:     "",
				DBPassword: "password",
				DBName:     "testdb",
				DBSSLMode:  "require",
			},
			expected: false,
			errors:   2, // DBHost and DBUser missing
		},
		{
			name: "Invalid port",
			config: Config{
				DBHost:     "localhost",
				DBPort:     "invalid",
				DBUser:     "postgres",
				DBPassword: "password",
				DBName:     "testdb",
				DBSSLMode:  "require",
			},
			expected: false,
			errors:   1, // Invalid port
		},
		{
			name: "Invalid SSL mode",
			config: Config{
				DBHost:     "localhost",
				DBPort:     "5432",
				DBUser:     "postgres",
				DBPassword: "password",
				DBName:     "testdb",
				DBSSLMode:  "invalid",
			},
			expected: false,
			errors:   1, // Invalid SSL mode
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.Validate()
			
			if result.Valid != tt.expected {
				t.Errorf("Expected valid=%t, got valid=%t", tt.expected, result.Valid)
			}
			
			if len(result.Errors) != tt.errors {
				t.Errorf("Expected %d errors, got %d errors", tt.errors, len(result.Errors))
			}
		})
	}
}

func TestConfig_ValidateDatabase(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected bool
	}{
		{
			name: "Valid database config",
			config: Config{
				DBHost:     "localhost",
				DBPort:     "5432",
				DBUser:     "postgres",
				DBPassword: "password",
				DBName:     "testdb",
				DBSSLMode:  "require",
			},
			expected: true,
		},
		{
			name: "Valid IP address",
			config: Config{
				DBHost:     "192.168.1.100",
				DBPort:     "5432",
				DBUser:     "postgres",
				DBPassword: "password",
				DBName:     "testdb",
				DBSSLMode:  "prefer",
			},
			expected: true,
		},
		{
			name: "Port out of range",
			config: Config{
				DBHost:     "localhost",
				DBPort:     "99999",
				DBUser:     "postgres",
				DBPassword: "password",
				DBName:     "testdb",
				DBSSLMode:  "prefer",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.validateDatabase()
			if result.Valid != tt.expected {
				t.Errorf("Expected valid=%t, got valid=%t", tt.expected, result.Valid)
			}
		})
	}
}

func TestConfig_ValidateWithDefaults(t *testing.T) {
	tests := []struct {
		name           string
		config         Config
		expectedPort   string
		expectedSSLMode string
		expectError    bool
	}{
		{
			name: "Valid config with defaults",
			config: Config{
				DBHost:     "localhost",
				DBUser:     "postgres",
				DBPassword: "password",
				DBName:     "testdb",
			},
			expectedPort:   "5432",
			expectedSSLMode: "prefer",
			expectError:    false,
		},
		{
			name: "Invalid config",
			config: Config{
				DBHost:     "",
				DBUser:     "postgres",
				DBPassword: "password",
				DBName:     "testdb",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.ValidateWithDefaults()
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.expectError {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				
				if tt.config.DBPort != tt.expectedPort {
					t.Errorf("Expected port %s, got %s", tt.expectedPort, tt.config.DBPort)
				}
				
				if tt.config.DBSSLMode != tt.expectedSSLMode {
					t.Errorf("Expected SSL mode %s, got %s", tt.expectedSSLMode, tt.config.DBSSLMode)
				}
			}
		})
	}
}

func TestLoad(t *testing.T) {
	// Save original environment variables
	originalVars := map[string]string{
		"DB_HOST":     os.Getenv("DB_HOST"),
		"DB_PORT":     os.Getenv("DB_PORT"),
		"DB_USER":     os.Getenv("DB_USER"),
		"DB_PASSWORD": os.Getenv("DB_PASSWORD"),
		"DB_NAME":     os.Getenv("DB_NAME"),
		"SSL_MODE":    os.Getenv("SSL_MODE"),
	}

	// Cleanup function
	defer func() {
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// Test with valid environment
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "password")
	os.Setenv("DB_NAME", "testdb")

	// This should not panic
	cfg := Load()
	
	if cfg.DBHost != "localhost" {
		t.Errorf("Expected DB_HOST=localhost, got %s", cfg.DBHost)
	}
	
	if cfg.DBUser != "postgres" {
		t.Errorf("Expected DB_USER=postgres, got %s", cfg.DBUser)
	}
}

func TestConfig_ValidateOrFatal(t *testing.T) {
	// Capture stdout to prevent actual fatal from stopping tests
	// Note: This test is limited since os.Exit is hard to test in Go
	
	tests := []struct {
		name   string
		config Config
		shouldPanic bool
	}{
		{
			name: "Valid config should not panic",
			config: Config{
				DBHost:     "localhost",
				DBPort:     "5432",
				DBUser:     "postgres",
				DBPassword: "password",
				DBName:     "testdb",
				DBSSLMode:  "prefer",
			},
			shouldPanic: false,
		},
		{
			name: "Invalid config should panic",
			config: Config{
				DBHost:     "",
				DBPort:     "5432",
				DBUser:     "",
				DBPassword: "password",
				DBName:     "testdb",
				DBSSLMode:  "prefer",
			},
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("Unexpected panic: %v", r)
					}
				} else {
					if tt.shouldPanic {
						t.Error("Expected panic but none occurred")
					}
				}
			}()
			
			tt.config.ValidateOrFatal()
		})
	}
}
