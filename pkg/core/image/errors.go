package image

import "desktopus/pkg/core/utils"

// Specific error types embedding DesktopusError
type ErrCoreTemplateReadingIndexMeta struct {
	*utils.DesktopusError
}

func (e *ErrCoreTemplateReadingIndexMeta) Error() string {
	return e.DesktopusError.Error()
}

type ErrCoreTemplateReadingCoreMeta struct {
	*utils.DesktopusError
}

func (e *ErrCoreTemplateReadingCoreMeta) Error() string {
	return e.DesktopusError.Error()
}

type ErrCoreTemplateOSNotFound struct {
	*utils.DesktopusError
}

func (e *ErrCoreTemplateOSNotFound) Error() string {
	return e.DesktopusError.Error()
}

// Functions to create specific errors with parameters
func newErrCoreTemplateReadingIndexMeta(errMessage string) error {
	return &ErrCoreTemplateReadingIndexMeta{
		DesktopusError: utils.NewDesktopusError("error reading core templates index metadata: %s", errMessage),
	}
}

func newErrCoreTemplateReadingCoreMeta(os string, errMessage string) error {
	return &ErrCoreTemplateReadingCoreMeta{
		DesktopusError: utils.NewDesktopusError("error reading OS %s metadata: %s", os, errMessage),
	}
}

func newErrCoreTemplateOSNotFound(name string) error {
	return &ErrCoreTemplateOSNotFound{
		DesktopusError: utils.NewDesktopusError("error: OS %s not found in core templates", name),
	}
}
