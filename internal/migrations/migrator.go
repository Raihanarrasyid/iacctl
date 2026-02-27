package migrations

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"
	"time"
)

// Migration represents a single database migration
type Migration struct {
	Version     int
	Name        string
	SQLUp       string
	SQLDown     string
	ExecutedAt  *time.Time
}

// Migrator handles database migrations
type Migrator struct {
	db          *sql.DB
	migrations  []Migration
	schemaTable string
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{
		db:          db,
		migrations:  make([]Migration, 0),
		schemaTable: "schema_migrations",
	}
}

// AddMigration adds a new migration to the list
func (m *Migrator) AddMigration(version int, name, sqlUp, sqlDown string) {
	m.migrations = append(m.migrations, Migration{
		Version: version,
		Name:    name,
		SQLUp:   sqlUp,
		SQLDown: sqlDown,
	})
}

// AddMigrationFromFS adds migrations from embedded filesystem
func (m *Migrator) AddMigrationFromFS(fsys fs.FS, version int, name string) error {
	upPath := fmt.Sprintf("%d_%s_up.sql", version, name)
	downPath := fmt.Sprintf("%d_%s_down.sql", version, name)

	upSQL, err := fs.ReadFile(fsys, upPath)
	if err != nil {
		return fmt.Errorf("failed to read up migration: %w", err)
	}

	downSQL, err := fs.ReadFile(fsys, downPath)
	if err != nil {
		return fmt.Errorf("failed to read down migration: %w", err)
	}

	m.AddMigration(version, name, string(upSQL), string(downSQL))
	return nil
}

// Up runs all pending migrations
func (m *Migrator) Up() error {
	if err := m.createSchemaTable(); err != nil {
		return fmt.Errorf("failed to create schema table: %w", err)
	}

	executed, err := m.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	for _, migration := range m.migrations {
		if _, exists := executed[migration.Version]; !exists {
			log.Printf("Running migration %d: %s", migration.Version, migration.Name)
			
			if err := m.executeMigration(migration, "up"); err != nil {
				return fmt.Errorf("failed to execute migration %d: %w", migration.Version, err)
			}

			if err := m.recordMigration(migration.Version, migration.Name); err != nil {
				return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
			}

			log.Printf("Migration %d: %s completed", migration.Version, migration.Name)
		}
	}

	log.Println("All migrations completed successfully")
	return nil
}

// Down rolls back the last migration
func (m *Migrator) Down() error {
	executed, err := m.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	// Find the latest executed migration
	var latestVersion int
	for version := range executed {
		if version > latestVersion {
			latestVersion = version
		}
	}

	if latestVersion == 0 {
		log.Println("No migrations to rollback")
		return nil
	}

	// Find the migration to rollback
	var migration *Migration
	for i, mig := range m.migrations {
		if mig.Version == latestVersion {
			migration = &m.migrations[i]
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("migration %d not found", latestVersion)
	}

	log.Printf("Rolling back migration %d: %s", migration.Version, migration.Name)

	if err := m.executeMigration(*migration, "down"); err != nil {
		return fmt.Errorf("failed to rollback migration %d: %w", migration.Version, err)
	}

	if err := m.removeMigration(latestVersion); err != nil {
		return fmt.Errorf("failed to remove migration record %d: %w", latestVersion, err)
	}

	log.Printf("Migration %d: %s rolled back", migration.Version, migration.Name)
	return nil
}

// Status shows migration status
func (m *Migrator) Status() error {
	if err := m.createSchemaTable(); err != nil {
		return fmt.Errorf("failed to create schema table: %w", err)
	}

	executed, err := m.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	fmt.Println("\nMigration Status:")
	fmt.Println("================")

	for _, migration := range m.migrations {
		status := "PENDING"
		if execInfo, exists := executed[migration.Version]; exists {
			status = fmt.Sprintf("EXECUTED at %s", execInfo.ExecutedAt.Format("2006-01-02 15:04:05"))
		}
		fmt.Printf("%d: %s - %s\n", migration.Version, migration.Name, status)
	}

	return nil
}

// createSchemaTable creates the schema migrations table if it doesn't exist
func (m *Migrator) createSchemaTable() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`, m.schemaTable)

	_, err := m.db.Exec(query)
	return err
}

// getExecutedMigrations returns a map of executed migrations
func (m *Migrator) getExecutedMigrations() (map[int]Migration, error) {
	query := fmt.Sprintf(`
		SELECT version, name, executed_at 
		FROM %s 
		ORDER BY version
	`, m.schemaTable)

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	executed := make(map[int]Migration)
	for rows.Next() {
		var migration Migration
		if err := rows.Scan(&migration.Version, &migration.Name, &migration.ExecutedAt); err != nil {
			return nil, err
		}
		executed[migration.Version] = migration
	}

	return executed, nil
}

// executeMigration executes a migration in the specified direction
func (m *Migrator) executeMigration(migration Migration, direction string) error {
	var sql string
	switch direction {
	case "up":
		sql = migration.SQLUp
	case "down":
		sql = migration.SQLDown
	default:
		return fmt.Errorf("invalid direction: %s", direction)
	}

	// Split SQL statements by semicolon and execute each one
	statements := strings.Split(sql, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if _, err := m.db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute statement: %s\nError: %w", stmt, err)
		}
	}

	return nil
}

// recordMigration records that a migration has been executed
func (m *Migrator) recordMigration(version int, name string) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (version, name, executed_at) 
		VALUES ($1, $2, NOW())
	`, m.schemaTable)

	_, err := m.db.Exec(query, version, name)
	return err
}

// removeMigration removes a migration record
func (m *Migrator) removeMigration(version int) error {
	query := fmt.Sprintf(`
		DELETE FROM %s WHERE version = $1
	`, m.schemaTable)

	_, err := m.db.Exec(query, version)
	return err
}
