package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

//go:embed *.sql
var migrationFS embed.FS

// LoadMigrations loads all migrations from the filesystem
func LoadMigrations(db *sql.DB) (*Migrator, error) {
	migrator := NewMigrator(db)

	// Load embedded migrations first
	if err := loadEmbeddedMigrations(migrator); err != nil {
		log.Printf("Warning: Failed to load embedded migrations: %v", err)
	}

	// Also load from filesystem if running in development
	if _, err := os.Stat("internal/migrations"); err == nil {
		if err := loadFilesystemMigrations(migrator, "internal/migrations"); err != nil {
			log.Printf("Warning: Failed to load filesystem migrations: %v", err)
		}
	}

	return migrator, nil
}

// loadEmbeddedMigrations loads migrations from embedded filesystem
func loadEmbeddedMigrations(m *Migrator) error {
	entries, err := migrationFS.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read embedded migration directory: %w", err)
	}

	migrationGroups := make(map[int][]string)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		version, migrationName, _, err := parseMigrationFilename(name)
		if err != nil {
			log.Printf("Skipping invalid migration filename: %s", name)
			continue
		}

		key := fmt.Sprintf("%d_%s", version, migrationName)
		if migrationGroups[version] == nil {
			migrationGroups[version] = make([]string, 0)
		}
		migrationGroups[version] = append(migrationGroups[version], key)
	}

	// Process migration groups
	for version, files := range migrationGroups {
		if len(files) != 2 {
			log.Printf("Warning: Migration %d has incomplete files: %v", version, files)
			continue
		}

		var upFile, downFile string
		var migrationName string

		for _, file := range files {
			_, name, dir, _ := parseMigrationFilename(file + ".sql")
			if migrationName == "" {
				migrationName = name
			}
			
			if dir == "up" {
				upFile = file + ".sql"
			} else if dir == "down" {
				downFile = file + ".sql"
			}
		}

		if upFile == "" || downFile == "" {
			log.Printf("Warning: Migration %d missing up or down file", version)
			continue
		}

		m.AddMigration(version, migrationName, "", "") // Will be loaded from FS
	}

	return nil
}

// loadFilesystemMigrations loads migrations from the local filesystem
func loadFilesystemMigrations(m *Migrator, migrationDir string) error {
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	migrationGroups := make(map[int]struct{ up, down, name string })

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		version, migrationName, direction, err := parseMigrationFilename(name)
		if err != nil {
			log.Printf("Skipping invalid migration filename: %s", name)
			continue
		}

		group := migrationGroups[version]
		group.name = migrationName

		fullPath := filepath.Join(migrationDir, name)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", name, err)
		}

		if direction == "up" {
			group.up = string(content)
		} else if direction == "down" {
			group.down = string(content)
		}

		migrationGroups[version] = group
	}

	// Add migrations to migrator
	versions := make([]int, 0, len(migrationGroups))
	for version := range migrationGroups {
		versions = append(versions, version)
	}
	sort.Ints(versions)

	for _, version := range versions {
		group := migrationGroups[version]
		if group.up == "" || group.down == "" {
			log.Printf("Warning: Migration %d missing up or down file", version)
			continue
		}

		m.AddMigration(version, group.name, group.up, group.down)
	}

	return nil
}

// parseMigrationFilename parses migration filename and extracts version, name, and direction
func parseMigrationFilename(filename string) (int, string, string, error) {
	// Pattern: 001_create_jobs_table_up.sql or 001_create_jobs_table_down.sql
	re := regexp.MustCompile(`^(\d+)_(.+?)_(up|down)\.sql$`)
	matches := re.FindStringSubmatch(filename)
	
	if len(matches) != 4 {
		return 0, "", "", fmt.Errorf("invalid migration filename format: %s", filename)
	}

	version, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, "", "", fmt.Errorf("invalid version number in filename: %s", filename)
	}

	name := matches[2]
	direction := matches[3]

	return version, name, direction, nil
}

// RunMigrations runs all pending migrations
func RunMigrations(db *sql.DB) error {
	migrator, err := LoadMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	return migrator.Up()
}

// RollbackLastMigration rolls back the last migration
func RollbackLastMigration(db *sql.DB) error {
	migrator, err := LoadMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	return migrator.Down()
}

// ShowMigrationStatus shows the current migration status
func ShowMigrationStatus(db *sql.DB) error {
	migrator, err := LoadMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	return migrator.Status()
}
