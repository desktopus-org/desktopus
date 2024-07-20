package utils

import (
	"fmt"
)

// BaseError is the base struct for all custom errors
type DesktopusError struct {
	Message string
	Params  []interface{}
}

// Error implements the error interface for BaseError
func (e *DesktopusError) Error() string {
	return fmt.Sprintf(e.Message, e.Params...)
}
