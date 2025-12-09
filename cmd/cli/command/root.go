package command

import (
	"fmt"
	"os"

	"database/sql"

	"github.com/Raihanarrasyid/iacctl/internal/config"
	"github.com/Raihanarrasyid/iacctl/internal/db"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
)

func NewRootCommand(db *sql.DB) *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "iacctl",
		Short: "iacctl is an IaC automation CLI tool",
		Long:  `iacctl is a CLI tool to provision infrastructure using Terraform.`,
	}
	
	rootCmd.AddCommand(NewJobCreateCmd(db))
	return rootCmd
}

func Execute() {
	cfg := config.Load()
	dbConn, err := db.Connect(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	rootCmd := NewRootCommand(dbConn) 
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}