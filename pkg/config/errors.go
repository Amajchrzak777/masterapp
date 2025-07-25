package config

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidSignalLength     = errors.New("signal length must be greater than 0")
	ErrInvalidSampleRate       = errors.New("sample rate must be greater than 0")
	ErrMismatchedSignalLength  = errors.New("voltage and current signals must have the same length")
	ErrMismatchedSampleRate    = errors.New("voltage and current signals must have the same sample rate")
	ErrDivisionByZero          = errors.New("division by zero in impedance calculation")
	ErrEmptyFrequencies        = errors.New("frequencies array cannot be empty")
	ErrNilSignal               = errors.New("signal cannot be nil")
	ErrInvalidFrequencyRange   = errors.New("invalid frequency range")
	ErrContextCancelled        = errors.New("context was cancelled")
	ErrChannelClosed           = errors.New("channel is closed")
	ErrNetworkTimeout          = errors.New("network request timed out")
	ErrInvalidURL              = errors.New("invalid target URL")
	ErrServerUnavailable       = errors.New("target server is unavailable")
	ErrJSONMarshalFailed       = errors.New("failed to marshal data to JSON")
	ErrInvalidHTTPResponse     = errors.New("invalid HTTP response")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

type ProcessingError struct {
	Operation string
	Cause     error
}

func (e ProcessingError) Error() string {
	return fmt.Sprintf("processing error in operation '%s': %v", e.Operation, e.Cause)
}

func (e ProcessingError) Unwrap() error {
	return e.Cause
}

type NetworkError struct {
	URL    string
	Status int
	Cause  error
}

func (e NetworkError) Error() string {
	return fmt.Sprintf("network error for URL '%s' (status %d): %v", e.URL, e.Status, e.Cause)
}

func (e NetworkError) Unwrap() error {
	return e.Cause
}

func NewValidationError(field, message string) error {
	return ValidationError{Field: field, Message: message}
}

func NewProcessingError(operation string, cause error) error {
	return ProcessingError{Operation: operation, Cause: cause}
}

func NewNetworkError(url string, status int, cause error) error {
	return NetworkError{URL: url, Status: status, Cause: cause}
}