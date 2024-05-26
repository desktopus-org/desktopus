package image

import (
	"desktopus/pkg/core/image/base"
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

// Specific error types
type ErrBuildManifestNotSpecified struct {
	BaseError
}

type ErrBuildImageNotSpecified struct {
	BaseError
}

type ErrBuildOSNotSpecified struct {
	BaseError
}

type ErrVersionNotSpecified struct {
	BaseError
}

type ErrBuildUnmarshallManifest struct {
	BaseError
}

type ErrBuildUnmarshallApiVersion struct {
	BaseError
}

type ErrBuildWrongApiVersion struct {
	BaseError
}

type ErrBuildInvalidTemplatesUri struct {
	BaseError
}

type ErrBuildUnsupportedOs struct {
	BaseError
}

type ErrBuildUnsupportedVersion struct {
	BaseError
}

type ErrBuildUnsupportedModule struct {
	BaseError
}

type ErrBuildDownloadingTemplate struct {
	BaseError
}

type ErrBuildCopyingTemplate struct {
	BaseError
}

// Functions to create specific errors with parameters
func NewErrBuildManifestNotSpecified() error {
	return &ErrBuildManifestNotSpecified{
		BaseError: BaseError{
			Message: "manifest not specified",
			Params:  []interface{}{},
		},
	}
}
func NewErrBuildImageNotSpecified() error {
	return &ErrBuildImageNotSpecified{
		BaseError: BaseError{
			Message: "build image not specified",
			Params:  []interface{}{},
		},
	}
}

func NewErrBuildOSNotSpecified() error {
	validOS := "\n"
	for key := range base.BaseImages {
		validOS += fmt.Sprintf("  - %s\n", key)
	}
	return &ErrBuildOSNotSpecified{
		BaseError: BaseError{
			Message: "parameter 'os' not specified in manifest. Valid options: %s",
			Params:  []interface{}{validOS},
		},
	}
}

func NewErrVersionNotSpecified(os string) error {
	validVersions := "\n"
	for key := range base.BaseImages[os] {
		validVersions += fmt.Sprintf("  - %s\n", key)
	}
	return &ErrVersionNotSpecified{
		BaseError: BaseError{
			Message: "parameter 'desktopusVersion' not specified in manifest. Valid options: %s",
			Params:  []interface{}{validVersions},
		},
	}
}

func NewErrBuildUnmarshallManifest(message string) error {
	return &ErrBuildUnmarshallManifest{
		BaseError: BaseError{
			Message: "failed to unmarshal manifest: %s",
			Params:  []interface{}{message},
		},
	}
}

func NewErrBuildUnmarshallApiVersion(message string) error {
	return &ErrBuildUnmarshallApiVersion{
		BaseError: BaseError{
			Message: "failed to unmarshal 'apiVersion' field from manifest: %s.",
			Params:  []interface{}{message},
		},
	}
}

func NewErrBuildWrongApiVersion(version string) error {
	validApiVersions := "\n"
	for key := range base.BaseImages {
		validApiVersions += fmt.Sprintf("  - %s\n", key)
	}
	return &ErrBuildWrongApiVersion{
		BaseError: BaseError{
			Message: "unsupported apiVersion in manifest: %s, valid options: %s",
			Params:  []interface{}{version, validApiVersions},
		},
	}
}

func NewErrBuildInvalidTemplatesUri(uri string) error {
	return &ErrBuildInvalidTemplatesUri{
		BaseError: BaseError{
			Message: "invalid templates URI: %s",
			Params:  []interface{}{uri},
		},
	}
}

func NewErrBuildUnsupportedOs(os string) error {
	validOS := "\n"
	for key := range base.BaseImages {
		validOS += fmt.Sprintf("  - %s\n", key)
	}
	return &ErrBuildUnsupportedOs{
		BaseError: BaseError{
			Message: "unsupported os in manifest: %s, valid options: %s",
			Params:  []interface{}{os, validOS},
		},
	}
}

func NewErrBuildUnsupportedVersion(os, version string) error {
	validVersions := "\n"
	for key := range base.BaseImages[os] {
		validVersions += fmt.Sprintf("  - %s\n", key)
	}
	return &ErrBuildUnsupportedVersion{
		BaseError: BaseError{
			Message: "unsupported version in manifest: %s, valid options: %s",
			Params:  []interface{}{version, validVersions},
		},
	}
}

func NewErrBuildUnsupportedModule(module string) error {
	return &ErrBuildUnsupportedModule{
		BaseError: BaseError{
			Message: "unsupported module in manifest: %s",
			Params:  []interface{}{module},
		},
	}
}

func NewErrBuildCopyingTemplate(message string) error {
	return &ErrBuildCopyingTemplate{
		BaseError: BaseError{
			Message: "failed to copy template: %s",
			Params:  []interface{}{message},
		},
	}
}
