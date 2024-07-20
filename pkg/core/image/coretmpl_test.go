package image

import (
	"fmt"
	"testing"
)

func BasicTestCoreTemplateOSValidImage(t *testing.T) {
	validCoreTemplateOS, err := newCoreTemplateOS("ubuntu-jammy")
	if err != nil {
		t.Errorf("Error creating core template OS: %s", err)
	}
	fmt.Println(validCoreTemplateOS)
}
