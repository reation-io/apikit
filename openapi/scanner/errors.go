package scanner

import "fmt"

// ScanError represents an error that occurred during scanning
type ScanError struct {
	Message  string
	Pattern  string
	FilePath string
	Cause    error
}

func (e *ScanError) Error() string {
	if e.FilePath != "" {
		return fmt.Sprintf("%s: %s (file: %s)", e.Message, e.Cause, e.FilePath)
	}
	if e.Pattern != "" {
		return fmt.Sprintf("%s: %s (pattern: %s)", e.Message, e.Cause, e.Pattern)
	}
	return fmt.Sprintf("%s: %s", e.Message, e.Cause)
}

func (e *ScanError) Unwrap() error {
	return e.Cause
}
