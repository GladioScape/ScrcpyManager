package main

import (
	"errors"
	"fmt"
)

// Custom error types for better error handling
var (
	ErrNoDeviceSelected  = errors.New("no device selected")
	ErrDeviceNotFound    = errors.New("device not found")
	ErrADBNotFound       = errors.New("adb binary not found")
	ErrScrcpyNotFound    = errors.New("scrcpy binary not found")
	ErrAppAlreadyRunning = errors.New("app is already running")
	ErrInvalidBitrate    = errors.New("invalid bitrate value")
	ErrInvalidFPS        = errors.New("invalid fps value")
	ErrInvalidDisplayID  = errors.New("invalid display ID")
	ErrProcessNotFound   = errors.New("process not found")
)

// AppError provides detailed error information
type AppError struct {
	Op      string // Operation that failed
	Err     error  // Underlying error
	Message string // User-friendly message
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s (%v)", e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(op string, err error, message string) *AppError {
	return &AppError{
		Op:      op,
		Err:     err,
		Message: message,
	}
}
