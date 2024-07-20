package image

import (
	"fmt"
)

// BaseError is the base struct for all custom errors
type BaseError struct {
	Message string
	Params  []interface{}
}

// Error implements the error interface for BaseError
func (e *BaseError) Error() string {
	return fmt.Sprintf(e.Message, e.Params...)
}

// // Specific error types
// type ErrBuildManifestNotSpecified struct {
// 	BaseError
// }

// // Functions to create specific errors with parameters

// func NewErrBuildManifestNotSpecified() error {
// 	return &ErrBuildManifestNotSpecified{
// 		BaseError: BaseError{
// 			Message: "manifest not specified",
// 			Params:  []interface{}{},
// 		},
// 	}
// }
