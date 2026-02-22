package module

import (
	"fmt"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

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

	// Verify tasks/main.yml exists
	tasksPath := dir + "/tasks/main.yml"
	if _, err := fs.Stat(fsys, tasksPath); err != nil {
		return nil, fmt.Errorf("module %q missing tasks/main.yml", mod.Name)
	}

	return &mod, nil
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
