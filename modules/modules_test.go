package modules_test

import (
	"io/fs"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/desktopus-org/desktopus/internal/config"
	"github.com/desktopus-org/desktopus/internal/module"
	"github.com/desktopus-org/desktopus/modules"
)

func newRegistry(t *testing.T) *module.Registry {
	t.Helper()
	reg, err := module.NewRegistry(modules.BuiltinFS)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	return reg
}

func TestRegistryLoadsBuiltinModules(t *testing.T) {
	reg := newRegistry(t)
	mods := reg.ListBuiltin()

	if len(mods) < 1 {
		t.Fatal("expected at least 1 built-in module")
	}
	for _, m := range mods {
		if !m.Builtin {
			t.Errorf("module %q should have Builtin=true", m.Name)
		}
	}
}

func TestRegistryResolvesEachBuiltinByName(t *testing.T) {
	reg := newRegistry(t)

	for _, m := range reg.ListBuiltin() {
		t.Run(m.Name, func(t *testing.T) {
			resolved, err := reg.Resolve(m.Name, ".")
			if err != nil {
				t.Fatalf("Resolve(%q): %v", m.Name, err)
			}
			if resolved.Name != m.Name {
				t.Errorf("expected name %q, got %q", m.Name, resolved.Name)
			}
		})
	}
}

func TestRegistryResolveUnknownFails(t *testing.T) {
	reg := newRegistry(t)

	_, err := reg.Resolve("nonexistent", ".")
	if err == nil {
		t.Error("expected error for unknown module")
	}
}

func TestBuiltinModulesHaveRequiredFields(t *testing.T) {
	reg := newRegistry(t)

	for _, m := range reg.ListBuiltin() {
		t.Run(m.Name, func(t *testing.T) {
			if m.Name == "" {
				t.Error("Name is empty")
			}
			if m.Description == "" {
				t.Error("Description is empty")
			}
			if m.Version == "" {
				t.Error("Version is empty")
			}
		})
	}
}

func TestBuiltinModulesCompatibilityUsesValidValues(t *testing.T) {
	reg := newRegistry(t)
	supportedOS := config.SupportedOSList()

	for _, m := range reg.ListBuiltin() {
		t.Run(m.Name, func(t *testing.T) {
			for _, os := range m.Compatibility.OS {
				if !contains(supportedOS, os) {
					t.Errorf("OS %q not in SupportedOSList()", os)
				}
			}
			for _, de := range m.Compatibility.Desktop {
				found := false
				for _, os := range supportedOS {
					if contains(config.SupportedDesktopsForOS(os), de) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("desktop %q not valid for any OS", de)
				}
			}
		})
	}
}

func TestBuiltinModulesHaveOSTaskFiles(t *testing.T) {
	reg := newRegistry(t)

	for _, m := range reg.ListBuiltin() {
		t.Run(m.Name, func(t *testing.T) {
			for _, os := range m.Compatibility.OS {
				taskFile := m.TaskFile(os)
				expected := "tasks/" + os + ".yml"
				if taskFile != expected {
					t.Errorf("OS %q: expected %q, got %q (missing OS-specific task file)", os, expected, taskFile)
				}
			}
		})
	}
}

func TestBuiltinModulesTaskFileFallback(t *testing.T) {
	reg := newRegistry(t)

	for _, m := range reg.ListBuiltin() {
		t.Run(m.Name, func(t *testing.T) {
			// Find an OS that is NOT in the compatibility list
			supported := config.SupportedOSList()
			for _, os := range supported {
				if contains(m.Compatibility.OS, os) {
					continue
				}
				// This OS is not in the module's compatibility list
				taskFile := m.TaskFile(os)
				if taskFile != "tasks/main.yml" {
					t.Errorf("OS %q (not compatible): expected tasks/main.yml fallback, got %q", os, taskFile)
				}
				return // one check is enough
			}
		})
	}
}

func TestBuiltinModulesTaskFilesAreValidYAML(t *testing.T) {
	reg := newRegistry(t)
	builtinFS := reg.BuiltinFS()

	for _, m := range reg.ListBuiltin() {
		t.Run(m.Name, func(t *testing.T) {
			tasksDir := m.Path + "/tasks"
			entries, err := fs.ReadDir(builtinFS, tasksDir)
			if err != nil {
				t.Fatalf("reading tasks dir: %v", err)
			}

			for _, e := range entries {
				if e.IsDir() || !strings.HasSuffix(e.Name(), ".yml") {
					continue
				}
				t.Run(e.Name(), func(t *testing.T) {
					data, err := fs.ReadFile(builtinFS, tasksDir+"/"+e.Name())
					if err != nil {
						t.Fatalf("reading %s: %v", e.Name(), err)
					}
					var parsed interface{}
					if err := yaml.Unmarshal(data, &parsed); err != nil {
						t.Errorf("invalid YAML in %s: %v", e.Name(), err)
					}
				})
			}
		})
	}
}

func TestChromeModule(t *testing.T) {
	reg := newRegistry(t)

	mod, err := reg.Resolve("chrome", ".")
	if err != nil {
		t.Fatalf("Resolve chrome: %v", err)
	}

	// Has chrome_channel var
	if _, ok := mod.Vars["chrome_channel"]; !ok {
		t.Error("expected chrome_channel var")
	}

	// System packages include wget and gnupg
	if !contains(mod.SystemPkgs, "wget") {
		t.Errorf("expected wget in system_packages, got %v", mod.SystemPkgs)
	}
	if !contains(mod.SystemPkgs, "gnupg") {
		t.Errorf("expected gnupg in system_packages, got %v", mod.SystemPkgs)
	}

	// Compatibility lists all supported OSes except Alpine (Chrome unavailable on musl)
	expectedOSes := []string{"ubuntu", "debian", "fedora", "el", "arch"}
	if len(mod.Compatibility.OS) != len(expectedOSes) {
		t.Errorf("expected %d OSes, got %d: %v", len(expectedOSes), len(mod.Compatibility.OS), mod.Compatibility.OS)
	}
	for _, os := range expectedOSes {
		if !contains(mod.Compatibility.OS, os) {
			t.Errorf("missing OS %q in chrome compatibility", os)
		}
	}
	if contains(mod.Compatibility.OS, "alpine") {
		t.Error("alpine should not be in chrome compatibility")
	}

	// Has 5 OS-specific task files (no alpine)
	osCount := 0
	for range mod.OSTaskFiles {
		osCount++
	}
	if osCount != 5 {
		t.Errorf("expected 5 OS task files, got %d: %v", osCount, mod.OSTaskFiles)
	}
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
