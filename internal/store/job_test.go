package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// TestJobStore_CreateJob tests creating a new job
func TestJobStore_CreateJob(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer db.Close()

	store := NewJobStore(db)

	// Test data
	job := &Job{
		Name:     "test-job",
		Status:   JobPending,
		TfModule: "docker",
		TfVars:   json.RawMessage(`{"image": "nginx:latest"}`),
	}

	// Execute
	jobID, err := store.CreateJob(context.Background(), job)

	// Verify
	if err != nil {
		t.Fatalf("Failed to create job: %v", err)
	}

	if jobID == uuid.Nil {
		t.Fatal("Job ID should not be nil")
	}

	// Verify job was actually created
	retrievedJob, err := store.GetJobByID(context.Background(), jobID)
	if err != nil {
		t.Fatalf("Failed to retrieve created job: %v", err)
	}

	if retrievedJob.Name != job.Name {
		t.Errorf("Expected job name %s, got %s", job.Name, retrievedJob.Name)
	}

	if retrievedJob.Status != JobPending {
		t.Errorf("Expected job status %s, got %s", JobPending, retrievedJob.Status)
	}
}

// TestJobStore_GetJobByID tests retrieving a job by ID
func TestJobStore_GetJobByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewJobStore(db)

	// Create a job first
	job := &Job{
		Name:     "test-retrieve",
		Status:   JobPending,
		TfModule: "docker",
	}
	jobID, _ := store.CreateJob(context.Background(), job)

	// Test successful retrieval
	retrievedJob, err := store.GetJobByID(context.Background(), jobID)
	if err != nil {
		t.Fatalf("Failed to retrieve job: %v", err)
	}

	if retrievedJob.ID != jobID {
		t.Errorf("Expected job ID %s, got %s", jobID, retrievedJob.ID)
	}

	// Test non-existent job
	nonExistentID := uuid.New()
	_, err = store.GetJobByID(context.Background(), nonExistentID)
	if err == nil {
		t.Error("Expected error when retrieving non-existent job")
	}
}

// TestJobStore_UpdateJobStatus tests updating job status
func TestJobStore_UpdateJobStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewJobStore(db)

	// Create a job
	job := &Job{
		Name:     "test-update",
		Status:   JobPending,
		TfModule: "docker",
	}
	jobID, _ := store.CreateJob(context.Background(), job)

	// Update status
	err := store.UpdateJobStatus(context.Background(), jobID, JobRunning)
	if err != nil {
		t.Fatalf("Failed to update job status: %v", err)
	}

	// Verify update
	retrievedJob, err := store.GetJobByID(context.Background(), jobID)
	if err != nil {
		t.Fatalf("Failed to retrieve updated job: %v", err)
	}

	if retrievedJob.Status != JobRunning {
		t.Errorf("Expected job status %s, got %s", JobRunning, retrievedJob.Status)
	}

	// Verify updated_at timestamp
	if retrievedJob.UpdatedAt.Before(retrievedJob.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt")
	}
}

// TestJobStore_UpdateJobLogs tests updating job logs
func TestJobStore_UpdateJobLogs(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewJobStore(db)

	// Create a job
	job := &Job{
		Name:     "test-logs",
		Status:   JobPending,
		TfModule: "docker",
	}
	jobID, _ := store.CreateJob(context.Background(), job)

	// Update logs
	logContent := "terraform apply completed successfully"
	err := store.UpdateJobLogs(context.Background(), jobID, logContent)
	if err != nil {
		t.Fatalf("Failed to update job logs: %v", err)
	}

	// Verify update
	retrievedJob, err := store.GetJobByID(context.Background(), jobID)
	if err != nil {
		t.Fatalf("Failed to retrieve job with updated logs: %v", err)
	}

	if retrievedJob.Logs != logContent {
		t.Errorf("Expected logs %s, got %s", logContent, retrievedJob.Logs)
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
	_, err = db.Exec("DELETE FROM jobs")
	if err != nil {
		t.Fatalf("Failed to clean up test database: %v", err)
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

// BenchmarkJobStore_CreateJob benchmarks job creation
func BenchmarkJobStore_CreateJob(b *testing.B) {
	db := setupBenchmarkDB(b)
	defer db.Close()

	store := NewJobStore(db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		job := &Job{
			Name:     fmt.Sprintf("bench-job-%d", i),
			Status:   JobPending,
			TfModule: "docker",
		}
		_, err := store.CreateJob(context.Background(), job)
		if err != nil {
			b.Fatalf("Failed to create job: %v", err)
		}
	}
}

// setupBenchmarkDB creates a database for benchmarking
func setupBenchmarkDB(b *testing.B) *sql.DB {
	dbHost := getEnvOrDefault("TEST_DB_HOST", "localhost")
	dbPort := getEnvOrDefault("TEST_DB_PORT", "5432")
	dbUser := getEnvOrDefault("TEST_DB_USER", "postgres")
	dbPassword := getEnvOrDefault("TEST_DB_PASSWORD", "password")
	dbName := getEnvOrDefault("TEST_DB_NAME", "iacctl_test")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		b.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		b.Fatalf("Failed to ping test database: %v", err)
	}

	return db
}
