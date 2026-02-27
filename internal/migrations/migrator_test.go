package migrations

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func TestMigrator_AddMigration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	migrator := NewMigrator(db)

	// Add a test migration
	migrator.AddMigration(1, "test_migration", "CREATE TABLE test (id INTEGER);", "DROP TABLE test;")

	if len(migrator.migrations) != 1 {
		t.Errorf("Expected 1 migration, got %d", len(migrator.migrations))
	}

	migration := migrator.migrations[0]
	if migration.Version != 1 {
		t.Errorf("Expected version 1, got %d", migration.Version)
	}

	if migration.Name != "test_migration" {
		t.Errorf("Expected name 'test_migration', got '%s'", migration.Name)
	}
}

func TestMigrator_Up(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	migrator := NewMigrator(db)

	// Add test migrations
	migrator.AddMigration(1, "create_test_table", 
		"CREATE TABLE test_table (id SERIAL PRIMARY KEY, name VARCHAR(100));",
		"DROP TABLE test_table;")

	migrator.AddMigration(2, "add_column",
		"ALTER TABLE test_table ADD COLUMN status VARCHAR(50);",
		"ALTER TABLE test_table DROP COLUMN status;")

	// Run migrations
	err := migrator.Up()
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Verify migrations were executed
	executed, err := migrator.getExecutedMigrations()
	if err != nil {
		t.Fatalf("Failed to get executed migrations: %v", err)
	}

	if len(executed) != 2 {
		t.Errorf("Expected 2 executed migrations, got %d", len(executed))
	}

	// Verify table exists
	var tableName string
	err = db.QueryRow("SELECT table_name FROM information_schema.tables WHERE table_name = 'test_table'").Scan(&tableName)
	if err != nil {
		t.Errorf("Test table should exist: %v", err)
	}

	// Verify column exists
	var columnName string
	err = db.QueryRow("SELECT column_name FROM information_schema.columns WHERE table_name = 'test_table' AND column_name = 'status'").Scan(&columnName)
	if err != nil {
		t.Errorf("Status column should exist: %v", err)
	}
}

func TestMigrator_Down(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	migrator := NewMigrator(db)

	// Add test migration
	migrator.AddMigration(1, "create_test_table", 
		"CREATE TABLE test_table (id SERIAL PRIMARY KEY, name VARCHAR(100));",
		"DROP TABLE test_table;")

	// Run migration up
	err := migrator.Up()
	if err != nil {
		t.Fatalf("Failed to run migration up: %v", err)
	}

	// Verify table exists
	var tableName string
	err = db.QueryRow("SELECT table_name FROM information_schema.tables WHERE table_name = 'test_table'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Test table should exist: %v", err)
	}

	// Run migration down
	err = migrator.Down()
	if err != nil {
		t.Fatalf("Failed to run migration down: %v", err)
	}

	// Verify table was dropped
	err = db.QueryRow("SELECT table_name FROM information_schema.tables WHERE table_name = 'test_table'").Scan(&tableName)
	if err == nil {
		t.Error("Test table should not exist after rollback")
	}
}

func TestMigrator_Status(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	migrator := NewMigrator(db)

	// Add test migrations
	migrator.AddMigration(1, "first_migration", 
		"CREATE TABLE first_table (id INTEGER);",
		"DROP TABLE first_table;")

	migrator.AddMigration(2, "second_migration",
		"CREATE TABLE second_table (id INTEGER);",
		"DROP TABLE second_table;")

	// Run only first migration
	err := migrator.Up()
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Add second migration (this won't be executed yet)
	migrator.AddMigration(3, "third_migration",
		"CREATE TABLE third_table (id INTEGER);",
		"DROP TABLE third_table;")

	// Check status - this should not panic
	err = migrator.Status()
	if err != nil {
		t.Errorf("Status should not fail: %v", err)
	}
}

func TestLoadMigrations(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	migrator, err := LoadMigrations(db)
	if err != nil {
		t.Fatalf("Failed to load migrations: %v", err)
	}

	if migrator == nil {
		t.Fatal("Migrator should not be nil")
	}

	// Should have at least the jobs table migration
	if len(migrator.migrations) == 0 {
		t.Error("Should have at least one migration loaded")
	}
}

func TestParseMigrationFilename(t *testing.T) {
	tests := []struct {
		filename     string
		expectedVer  int
		expectedName string
		expectedDir  string
		expectError  bool
	}{
		{
			filename:     "001_create_jobs_table_up.sql",
			expectedVer:  1,
			expectedName: "create_jobs_table",
			expectedDir:  "up",
			expectError:  false,
		},
		{
			filename:     "002_add_indexes_down.sql",
			expectedVer:  2,
			expectedName: "add_indexes",
			expectedDir:  "down",
			expectError:  false,
		},
		{
			filename:     "invalid.sql",
			expectError:  true,
		},
		{
			filename:     "001_invalid.sql",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			version, name, direction, err := parseMigrationFilename(tt.filename)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if version != tt.expectedVer {
				t.Errorf("Expected version %d, got %d", tt.expectedVer, version)
			}
			
			if name != tt.expectedName {
				t.Errorf("Expected name '%s', got '%s'", tt.expectedName, name)
			}
			
			if direction != tt.expectedDir {
				t.Errorf("Expected direction '%s', got '%s'", tt.expectedDir, direction)
			}
		})
	}
}

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *sql.DB {
	// Use environment variables for test database or fallback to defaults
	dbHost := getEnvOrDefault("TEST_DB_HOST", "localhost")
	dbPort := getEnvOrDefault("TEST_DB_PORT", "5432")
	dbUser := getEnvOrDefault("TEST_DB_USER", "postgres")
	dbPassword := getEnvOrDefault("TEST_DB_PASSWORD", "password")
	dbName := getEnvOrDefault("TEST_DB_NAME", "iacctl_test")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Clean up any existing data
	_, err = db.Exec("DROP TABLE IF EXISTS schema_migrations")
	if err != nil {
		t.Fatalf("Failed to clean up schema_migrations: %v", err)
	}

	_, err = db.Exec("DROP TABLE IF EXISTS test_table")
	if err != nil {
		t.Fatalf("Failed to clean up test_table: %v", err)
	}

	_, err = db.Exec("DROP TABLE IF EXISTS first_table")
	if err != nil {
		t.Fatalf("Failed to clean up first_table: %v", err)
	}

	_, err = db.Exec("DROP TABLE IF EXISTS second_table")
	if err != nil {
		t.Fatalf("Failed to clean up second_table: %v", err)
	}

	_, err = db.Exec("DROP TABLE IF EXISTS third_table")
	if err != nil {
		t.Fatalf("Failed to clean up third_table: %v", err)
	}

	return db
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
