package core

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"time"

	"github.com/Raihanarrasyid/iacctl/internal/store"
	"github.com/Raihanarrasyid/iacctl/internal/terraform"
	"github.com/google/uuid"
)

func ProcessJob(ctx context.Context, db *sql.DB, jobID uuid.UUID) error {
	jobStore := store.NewJobStore(db)

	// 1. Fetch job
	job, err := jobStore.GetJobByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to fetch job: %w", err)
	}

	// 2. Update status to "running"
	if err := jobStore.UpdateJobStatus(ctx, job.ID, "running"); err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// 3. Prepare working dir
	workDir := filepath.Join("/tmp/iacctl/jobs", fmt.Sprintf("job-%s", job.ID))
	logPath := filepath.Join(workDir, "terraform.log")

	runner := terraform.NewRunner(workDir)
	
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomPort := 8000 + rng.Intn(1000) 

	tfData := terraform.TemplateData{
		Name:      job.Name,
		Timestamp: time.Now().Format(time.RFC3339),
		Port:      randomPort,
	}

	// 4. Generate main.tf
	if err := runner.PrepareTerraformFiles(tfData); err != nil {
		_ = jobStore.UpdateJobStatus(ctx, job.ID, "failed")
		return fmt.Errorf("failed to prepare tf files: %w", err)
	}

	// 5. Run terraform (init, plan, apply)
	if err := runner.RunTerraform(ctx, logPath); err != nil {
		_ = jobStore.UpdateJobStatus(ctx, job.ID, "failed")
		return fmt.Errorf("terraform failed: %w", err)
	}

	// 6. Finalize
	if err := jobStore.UpdateJobLogs(ctx, job.ID, logPath); err != nil {
		log.Println("warning: failed to update log path:", err)
	}
	if err := jobStore.UpdateJobStatus(ctx, job.ID, "success"); err != nil {
		return fmt.Errorf("failed to mark job success: %w", err)
	}

	return nil
}