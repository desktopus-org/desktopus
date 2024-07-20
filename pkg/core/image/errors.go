package image

import "desktopus/pkg/core/utils"

// Specific error types
type ErrCoreTemplateReadingIndexMeta struct {
	utils.DesktopusError
}

type ErrCoreTemplateReadingCoreMeta struct {
	utils.DesktopusError
}

type ErrCoreTemplateOSNotFound struct {
	utils.DesktopusError
}

// Functions to create specific errors with parameters
func newErrCoreTemplateReadingIndexMeta(errMessage string) error {
	return &utils.DesktopusError{
		Message: "error reading core template index metadata: %s",
		Params:  []interface{}{errMessage},
	}
}

func newErrCoreTemplateReadingCoreMeta(os string, errMessage string) error {
	return &utils.DesktopusError{
		Message: "error reading OS %s metadata: %s",
		Params:  []interface{}{os, errMessage},
	}
}

func newErrCoreTemplateOSNotFound(name string) error {
	return &utils.DesktopusError{
		Message: "error: OS %s not found in core templates",
		Params:  []interface{}{name},
	}
}
