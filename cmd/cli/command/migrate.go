package command

import (
	"fmt"

	"github.com/Raihanarrasyid/iacctl/internal/config"
	"github.com/Raihanarrasyid/iacctl/internal/db"
	"github.com/Raihanarrasyid/iacctl/internal/migrations"
	"github.com/spf13/cobra"
)

func NewMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration commands",
		Long:  `Manage database migrations for iacctl`,
	}

	cmd.AddCommand(NewMigrateUpCmd())
	cmd.AddCommand(NewMigrateDownCmd())
	cmd.AddCommand(NewMigrateStatusCmd())

	return cmd
}

func NewMigrateUpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Run all pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			dbConn, err := db.Connect(cfg)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer dbConn.Close()

			fmt.Println("Running database migrations...")
			if err := migrations.RunMigrations(dbConn); err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}

			fmt.Println("Migrations completed successfully!")
			return nil
		},
	}

	return cmd
}

func NewMigrateDownCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Rollback the last migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			dbConn, err := db.Connect(cfg)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer dbConn.Close()

			fmt.Println("Rolling back last migration...")
			if err := migrations.RollbackLastMigration(dbConn); err != nil {
				return fmt.Errorf("rollback failed: %w", err)
			}

			fmt.Println("Rollback completed successfully!")
			return nil
		},
	}

	return cmd
}

func NewMigrateStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			dbConn, err := db.Connect(cfg)
			if err != nil {
				return fmt.Errorf("failed to connect to database: %w", err)
			}
			defer dbConn.Close()

			if err := migrations.ShowMigrationStatus(dbConn); err != nil {
				return fmt.Errorf("failed to get migration status: %w", err)
			}

			return nil
		},
	}

	return cmd
}
