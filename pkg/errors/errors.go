package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents different types of errors in the application
type ErrorCode string

const (
	// Database errors
	ErrCodeDatabaseConnection ErrorCode = "DATABASE_CONNECTION"
	ErrCodeDatabaseQuery       ErrorCode = "DATABASE_QUERY"
	ErrCodeDatabaseTransaction ErrorCode = "DATABASE_TRANSACTION"
	ErrCodeRecordNotFound       ErrorCode = "RECORD_NOT_FOUND"
	ErrCodeDuplicateRecord      ErrorCode = "DUPLICATE_RECORD"

	// Validation errors
	ErrCodeValidation    ErrorCode = "VALIDATION"
	ErrCodeInvalidInput  ErrorCode = "INVALID_INPUT"
	ErrCodeMissingField  ErrorCode = "MISSING_FIELD"
	ErrCodeInvalidFormat ErrorCode = "INVALID_FORMAT"

	// Business logic errors
	ErrCodeJobNotFound     ErrorCode = "JOB_NOT_FOUND"
	ErrCodeJobAlreadyRun   ErrorCode = "JOB_ALREADY_RUN"
	ErrCodeJobFailed       ErrorCode = "JOB_FAILED"
	ErrCodeInvalidJobState ErrorCode = "INVALID_JOB_STATE"

	// Terraform errors
	ErrCodeTerraformNotFound ErrorCode = "TERRAFORM_NOT_FOUND"
	ErrCodeTerraformInit     ErrorCode = "TERRAFORM_INIT"
	ErrCodeTerraformPlan     ErrorCode = "TERRAFORM_PLAN"
	ErrCodeTerraformApply    ErrorCode = "TERRAFORM_APPLY"
	ErrCodeTerraformDestroy  ErrorCode = "TERRAFORM_DESTROY"

	// System errors
	ErrCodeInternalServer ErrorCode = "INTERNAL_SERVER"
	ErrCodeTimeout        ErrorCode = "TIMEOUT"
	ErrCodeUnauthorized   ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden      ErrorCode = "FORBIDDEN"
	ErrCodeNotFound       ErrorCode = "NOT_FOUND"
)

// AppError represents a structured application error
type AppError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
	Cause      error                  `json:"-"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithCause sets the underlying cause
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: getDefaultHTTPStatus(code),
	}
}

// NewAppErrorWithDetails creates a new application error with details
func NewAppErrorWithDetails(code ErrorCode, message, details string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		HTTPStatus: getDefaultHTTPStatus(code),
	}
}

// Database error constructors
func NewDatabaseConnectionError(cause error) *AppError {
	return NewAppError(ErrCodeDatabaseConnection, "Failed to connect to database").
		WithCause(cause)
}

func NewDatabaseQueryError(query string, cause error) *AppError {
	return NewAppError(ErrCodeDatabaseQuery, "Database query failed").
		WithCause(cause).
		WithContext("query", query)
}

func NewRecordNotFoundError(resource string, id interface{}) *AppError {
	return NewAppErrorWithDetails(ErrCodeRecordNotFound, 
		fmt.Sprintf("%s not found", resource), 
		fmt.Sprintf("ID: %v", id))
}

// Validation error constructors
func NewValidationError(message string) *AppError {
	return NewAppError(ErrCodeValidation, message)
}

func NewInvalidInputError(field, value string) *AppError {
	return NewAppErrorWithDetails(ErrCodeInvalidInput, 
		fmt.Sprintf("Invalid input for %s", field), 
		fmt.Sprintf("Value: %s", value))
}

func NewMissingFieldError(field string) *AppError {
	return NewAppErrorWithDetails(ErrCodeMissingField, 
		"Required field is missing", 
		field)
}

// Job error constructors
func NewJobNotFoundError(jobID interface{}) *AppError {
	return NewRecordNotFoundError("Job", jobID)
}

func NewJobAlreadyRunError(jobID interface{}) *AppError {
	return NewAppErrorWithDetails(ErrCodeJobAlreadyRun, 
		"Job has already been run", 
		fmt.Sprintf("Job ID: %v", jobID))
}

func NewJobFailedError(jobID interface{}, cause error) *AppError {
	return NewAppErrorWithDetails(ErrCodeJobFailed, 
		"Job execution failed", 
		fmt.Sprintf("Job ID: %v", jobID)).
		WithCause(cause)
}

// Terraform error constructors
func NewTerraformNotFoundError() *AppError {
	return NewAppError(ErrCodeTerraformNotFound, 
		"Terraform binary not found in PATH")
}

func NewTerraformError(operation string, cause error) *AppError {
	code := ErrCodeInternalServer
	switch operation {
	case "init":
		code = ErrCodeTerraformInit
	case "plan":
		code = ErrCodeTerraformPlan
	case "apply":
		code = ErrCodeTerraformApply
	case "destroy":
		code = ErrCodeTerraformDestroy
	}

	return NewAppError(code, fmt.Sprintf("Terraform %s failed", operation)).
		WithCause(cause)
}

// System error constructors
func NewInternalServerError(message string) *AppError {
	return NewAppError(ErrCodeInternalServer, message)
}

func NewTimeoutError(operation string) *AppError {
	return NewAppErrorWithDetails(ErrCodeTimeout, 
		"Operation timed out", 
		fmt.Sprintf("Operation: %s", operation))
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError converts an error to AppError if possible
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return NewInternalServerError(err.Error()).WithCause(err)
}

// WrapError wraps any error as an AppError
func WrapError(err error, code ErrorCode, message string) *AppError {
	return NewAppError(code, message).WithCause(err)
}

// getDefaultHTTPStatus returns the appropriate HTTP status for an error code
func getDefaultHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrCodeValidation, ErrCodeInvalidInput, ErrCodeMissingField, ErrCodeInvalidFormat:
		return http.StatusBadRequest
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeRecordNotFound, ErrCodeJobNotFound, ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeDuplicateRecord:
		return http.StatusConflict
	case ErrCodeTimeout:
		return http.StatusRequestTimeout
	case ErrCodeDatabaseConnection, ErrCodeDatabaseQuery, ErrCodeDatabaseTransaction,
		 ErrCodeJobAlreadyRun, ErrCodeJobFailed, ErrCodeInvalidJobState,
		 ErrCodeTerraformNotFound, ErrCodeTerraformInit, ErrCodeTerraformPlan, 
		 ErrCodeTerraformApply, ErrCodeTerraformDestroy, ErrCodeInternalServer:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
