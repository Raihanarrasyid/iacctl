package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Value   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s: %s (value: %s)", e.Field, e.Message, e.Value)
}

// ValidationResult contains validation results
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

// Validate performs comprehensive configuration validation
func (c *Config) Validate() *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
	}

	// Database validation
	result.merge(c.validateDatabase())
	
	// Optional: Terraform validation
	result.merge(c.validateTerraform())
	
	// Optional: Docker validation
	result.merge(c.validateDocker())

	return result
}

// validateDatabase validates database configuration
func (c *Config) validateDatabase() *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
	}

	// Required fields
	if c.DBHost == "" {
		result.addError("DB_HOST", c.DBHost, "database host is required")
	}
	
	if c.DBUser == "" {
		result.addError("DB_USER", c.DBUser, "database user is required")
	}
	
	if c.DBName == "" {
		result.addError("DB_NAME", c.DBName, "database name is required")
	}

	// Port validation
	if c.DBPort != "" {
		if port, err := strconv.Atoi(c.DBPort); err != nil {
			result.addError("DB_PORT", c.DBPort, "must be a valid port number")
		} else if port < 1 || port > 65535 {
			result.addError("DB_PORT", c.DBPort, "must be between 1 and 65535")
		}
	} else {
		result.addError("DB_PORT", c.DBPort, "database port is required")
	}

	// Host validation (if not localhost)
	if c.DBHost != "" && c.DBHost != "localhost" && c.DBHost != "127.0.0.1" {
		if net.ParseIP(c.DBHost) == nil {
			// Not an IP, try to resolve as hostname
			if _, err := net.LookupHost(c.DBHost); err != nil {
				result.addError("DB_HOST", c.DBHost, "must be a valid IP address or hostname")
			}
		}
	}

	// SSL Mode validation
	if c.DBSSLMode != "" {
		validModes := []string{"disable", "allow", "prefer", "require", "verify-ca", "verify-full"}
		valid := false
		for _, mode := range validModes {
			if c.DBSSLMode == mode {
				valid = true
				break
			}
		}
		if !valid {
			result.addError("SSL_MODE", c.DBSSLMode, 
				fmt.Sprintf("must be one of: %s", strings.Join(validModes, ", ")))
		}
	}

	return result
}

// validateTerraform validates Terraform configuration
func (c *Config) validateTerraform() *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
	}

	// Check if Terraform binary exists
	if tfPath := os.Getenv("TERRAFORM_PATH"); tfPath != "" {
		if _, err := os.Stat(tfPath); err != nil {
			result.addError("TERRAFORM_PATH", tfPath, "Terraform binary not found at specified path")
		}
	}

	return result
}

// validateDocker validates Docker configuration
func (c *Config) validateDocker() *ValidationResult {
	result := &ValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
	}

	// Check Docker socket if running on Unix
	if dockerHost := os.Getenv("DOCKER_HOST"); dockerHost == "" {
		if _, err := os.Stat("/var/run/docker.sock"); os.IsNotExist(err) {
			// Docker socket not found, but this might be okay if using remote Docker
			result.addWarning("DOCKER_HOST", "", "Docker socket not found, ensure Docker is running or DOCKER_HOST is set")
		}
	}

	return result
}

// addError adds a validation error
func (r *ValidationResult) addError(field, value, message string) {
	r.Errors = append(r.Errors, ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	})
	r.Valid = false
}

// addWarning adds a validation warning (doesn't make validation invalid)
func (r *ValidationResult) addWarning(field, value, message string) {
	r.Errors = append(r.Errors, ValidationError{
		Field:   field,
		Value:   value,
		Message: "WARNING: " + message,
	})
}

// merge merges another validation result
func (r *ValidationResult) merge(other *ValidationResult) {
	if !other.Valid {
		r.Valid = false
	}
	r.Errors = append(r.Errors, other.Errors...)
}

// ValidateOrFatal validates configuration and exits on error
func (c *Config) ValidateOrFatal() {
	result := c.Validate()
	
	if !result.Valid {
		fmt.Println("Configuration validation failed:")
		fmt.Println("===================================")
		
		for _, err := range result.Errors {
			fmt.Printf("❌ %s\n", err.Error())
		}
		
		fmt.Println("\nPlease check your environment variables and try again.")
		os.Exit(1)
	}
	
	// Show warnings
	warnings := make([]ValidationError, 0)
	for _, err := range result.Errors {
		if strings.Contains(err.Message, "WARNING") {
			warnings = append(warnings, err)
		}
	}
	
	if len(warnings) > 0 {
		fmt.Println("Configuration warnings:")
		fmt.Println("=======================")
		for _, warning := range warnings {
			fmt.Printf("⚠️  %s\n", warning.Error())
		}
		fmt.Println()
	}
}

// ValidateWithDefaults validates and sets default values
func (c *Config) ValidateWithDefaults() error {
	result := c.Validate()
	
	if !result.Valid {
		return fmt.Errorf("configuration validation failed: %d errors found", len(result.Errors))
	}
	
	// Set defaults for optional fields
	if c.DBPort == "" {
		c.DBPort = "5432"
	}
	
	if c.DBSSLMode == "" {
		c.DBSSLMode = "prefer"
	}
	
	return nil
}
