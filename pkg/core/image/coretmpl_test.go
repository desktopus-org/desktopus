package image

import (
	"testing"
)

func TestCoreTemplateUbuntuJammy(t *testing.T) {
	validCoreTemplateOS, err := newCoreTemplateOS("ubuntu-jammy")
	if err != nil {
		t.Errorf("Error creating core template OS: %s", err)
	}
	if validCoreTemplateOS.name != "ubuntu-jammy" {
		t.Errorf("Expected core template OS name to be ubuntu-jammy, got %s", validCoreTemplateOS.name)
	}
	if len(validCoreTemplateOS.rootScripts) == 0 {
		t.Errorf("Expected core template OS rootScripts to be non-empty, got empty")
	}
	// Check that all predefined root scripts exist in the FS
	for _, fileDir := range validCoreTemplateOS.rootScripts {
		_, err := coreTemplatesFS.ReadFile(fileDir)
		if err != nil {
			t.Errorf("Error reading root script %s: %s", fileDir, err)
		}
	}

	if len(validCoreTemplateOS.patches) == 0 {
		t.Errorf("Expected core template OS patches to be non-empty, got empty")
	}

	if validCoreTemplateOS.dockerfile == "" {
		t.Errorf("Expected core template OS dockerfile to be non-empty, got empty")
	}
	_, err = coreTemplatesFS.ReadFile(validCoreTemplateOS.dockerfile)
	if err != nil {
		t.Errorf("Error reading Dockerfile %s: %s", validCoreTemplateOS.dockerfile, err)
	}

	if len(validCoreTemplateOS.commonModules) == 0 {
		t.Errorf("Expected core template OS commonModules to be non-empty, got empty")
	}

	for module, fileDir := range validCoreTemplateOS.commonModules {
		_, err := coreTemplatesFS.ReadFile(fileDir)
		if err != nil {
			t.Errorf("Error reading common module %s: %s", module, err)
		}
	}

	if len(validCoreTemplateOS.modules) == 0 {
		t.Errorf("Expected core template OS modules to be non-empty, got empty")
	}

	for module, fileDir := range validCoreTemplateOS.modules {
		_, err := coreTemplatesFS.ReadFile(fileDir)
		if err != nil {
			t.Errorf("Error reading module %s: %s", module, err)
		}
	}

}

func TestCoreTemplateInvalidImage(t *testing.T) {
	_, err := newCoreTemplateOS("nonexistent-os")
	if err == nil {
		t.Errorf("Expected error creating core template OS, got nil")
	}
	if _, ok := err.(*ErrCoreTemplateOSNotFound); !ok {
		t.Errorf("Expected error type ErrCoreTemplateOSNotFound, got %T", err)
	}
}
