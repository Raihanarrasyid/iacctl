package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	JobPending JobStatus = "pending"
	JobRunning JobStatus = "running"
	JobSuccess JobStatus = "success"
	JobFailed  JobStatus = "failed"
)

type Job struct {
    ID        uuid.UUID       `json:"id"`
    Name      string          `json:"name"`
    Status    JobStatus       `json:"status"`
    TfModule  string          `json:"tf_module"`
    TfVars    json.RawMessage `json:"tf_vars"` // raw JSON for flexibility
    Logs      string          `json:"logs"`
    CreatedAt time.Time       `json:"created_at"`
    UpdatedAt time.Time       `json:"updated_at"`
}

type JobStore struct {
	DB *sql.DB
}

func NewJobStore(db *sql.DB) *JobStore {
	return &JobStore{
		DB: db,
	}
}

func (s *JobStore) CreateJob(ctx context.Context, job *Job) (uuid.UUID, error)  {
	job.ID = uuid.New()
	job.Status = JobPending
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	query := `
		INSERT INTO jobs (id, name, status, tf_module, tf_vars, logs, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := s.DB.ExecContext(ctx, query,
		job.ID, job.Name, job.Status, job.TfModule, job.TfVars, job.Logs, job.CreatedAt, job.UpdatedAt,
	)

	return job.ID, err
}

func (s *JobStore) GetJobByID(ctx context.Context, id uuid.UUID) (*Job, error) {
	query := `
		SELECT id, name, status, tf_module, tf_vars, logs, created_at, updated_at
		FROM jobs
		WHERE id = $1
	`

	row := s.DB.QueryRowContext(ctx, query, id)
	
	var job Job
	err := row.Scan(&job.ID, &job.Name, &job.Status, &job.TfModule, &job.TfVars, &job.Logs, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *JobStore) UpdateJobStatus(ctx context.Context, id uuid.UUID, status JobStatus) error {
    query := `
    UPDATE jobs
    SET status = $1, updated_at = $2
    WHERE id = $3
    `
    _, err := s.DB.ExecContext(ctx, query, status, time.Now(), id)
    return err
}

func (s *JobStore) UpdateJobLogs(ctx context.Context, id uuid.UUID, logs string) error {
    query := `
    UPDATE jobs
    SET logs = $1, updated_at = $2
    WHERE id = $3
    `
    _, err := s.DB.ExecContext(ctx, query, logs, time.Now(), id)
    return err
}