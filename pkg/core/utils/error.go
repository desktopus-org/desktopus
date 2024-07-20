package utils

import (
	"fmt"
)

// DesktopusError is the base struct for all custom errors
type DesktopusError struct {
	Message string
	Params  []interface{}
}

// Error implements the error interface for DesktopusError
func (e *DesktopusError) Error() string {
	return fmt.Sprintf(e.Message, e.Params...)
}

// NewDesktopusError creates a new DesktopusError with the given message and parameters
func NewDesktopusError(message string, params ...interface{}) *DesktopusError {
	return &DesktopusError{
		Message: message,
		Params:  params,
	}
}
