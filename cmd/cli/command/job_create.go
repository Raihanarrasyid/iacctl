package command

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Raihanarrasyid/iacctl/internal/core"
	"github.com/Raihanarrasyid/iacctl/internal/store"
	"github.com/spf13/cobra"
)

func NewJobCreateCmd(db *sql.DB) *cobra.Command {
	var jobName string

	cmd := &cobra.Command{
		Use:   "job:create",
		Short: "Create a new Terraform job",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Creating job:", jobName)

			jobStore := store.NewJobStore(db)

			// Buat job baru dan ambil ID-nya
			job := &store.Job{Name: jobName, TfVars: json.RawMessage(`{}`)}
			jobID, err := jobStore.CreateJob(cmd.Context(), job)
			if err != nil {
				return fmt.Errorf("failed to create job: %w", err)
			}

			fmt.Println("Job created with ID:", jobID)

			// Jalankan proses Terraform
			fmt.Println("Running job...")
			if err := core.ProcessJob(cmd.Context(), db, jobID); err != nil {
				log.Println("Job failed:", err)
				return err
			}

			fmt.Println("Job completed successfully")
			return nil
		},
	}

	cmd.Flags().StringVarP(&jobName, "name", "n", "", "Job name (required)")
	cmd.MarkFlagRequired("name")
	return cmd
}
