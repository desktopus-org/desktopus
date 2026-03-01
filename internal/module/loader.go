package module

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// validOSNames is the set of recognized OS names for task file discovery.
var validOSNames = map[string]bool{
	"alpine": true,
	"arch":   true,
	"debian": true,
	"el":     true,
	"fedora": true,
	"ubuntu": true,
}

// LoadFromFS loads a module from a filesystem (real or embedded)
func LoadFromFS(fsys fs.FS, dir string) (*Module, error) {
	metaPath := dir + "/module.yaml"
	data, err := fs.ReadFile(fsys, metaPath)
	if err != nil {
		return nil, fmt.Errorf("reading module.yaml from %s: %w", dir, err)
	}

	var mod Module
	if err := yaml.Unmarshal(data, &mod); err != nil {
		return nil, fmt.Errorf("parsing module.yaml from %s: %w", dir, err)
	}

	// Require tasks/main.yml only when no compatible OSes are declared.
	// When specific OSes are listed, each is expected to have its own task file.
	if len(mod.Compatibility.OS) == 0 {
		tasksPath := dir + "/tasks/main.yml"
		if _, err := fs.Stat(fsys, tasksPath); err != nil {
			return nil, fmt.Errorf("module %q missing tasks/main.yml", mod.Name)
		}
	}

	mod.OSTaskFiles = discoverOSTaskFiles(fsys, dir)
	mod.Tests = loadModuleTests(fsys, dir)

	return &mod, nil
}

// loadModuleTests optionally loads module_test.yaml from the module directory.
// Returns nil if the file does not exist or cannot be parsed.
func loadModuleTests(fsys fs.FS, dir string) *ModuleTests {
	data, err := fs.ReadFile(fsys, dir+"/module_test.yaml")
	if err != nil {
		return nil
	}
	var tests ModuleTests
	if err := yaml.Unmarshal(data, &tests); err != nil {
		return nil
	}
	return &tests
}

// discoverOSTaskFiles scans the tasks/ directory for OS-specific task files.
// It returns a map of OS names that have a corresponding <os>.yml file.
func discoverOSTaskFiles(fsys fs.FS, dir string) map[string]bool {
	tasksDir := dir + "/tasks"
	entries, err := fs.ReadDir(fsys, tasksDir)
	if err != nil {
		return nil
	}

	result := make(map[string]bool)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yml") {
			continue
		}
		osName := strings.TrimSuffix(name, ".yml")
		if validOSNames[osName] {
			result[osName] = true
		}
	}
	return result
}

// LoadFromDisk loads a module from a directory on disk
func LoadFromDisk(path string) (*Module, error) {
	mod, err := LoadFromFS(os.DirFS(path), ".")
	if err != nil {
		return nil, fmt.Errorf("loading module from %s: %w", path, err)
	}
	mod.Path = path
	return mod, nil
}
