package module

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// Registry discovers and resolves modules
type Registry struct {
	builtins  map[string]*Module
	builtinFS fs.FS
}

// NewRegistry creates a registry from the embedded built-in modules filesystem.
// The FS should have module directories at the top level (e.g., "chrome/", "vscode/").
func NewRegistry(builtinFS fs.FS) (*Registry, error) {
	r := &Registry{
		builtins:  make(map[string]*Module),
		builtinFS: builtinFS,
	}

	entries, err := fs.ReadDir(builtinFS, ".")
	if err != nil {
		return nil, fmt.Errorf("reading built-in modules: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		mod, err := LoadFromFS(builtinFS, entry.Name())
		if err != nil {
			return nil, fmt.Errorf("loading built-in module %q: %w", entry.Name(), err)
		}
		mod.Path = entry.Name()
		mod.Builtin = true
		r.builtins[mod.Name] = mod
	}

	return r, nil
}

// Resolve looks up a module by name. If the name is a path (starts with
// ./ or ../ or /), it loads from disk. Otherwise, looks up built-ins.
func (r *Registry) Resolve(name string, basePath string) (*Module, error) {
	if isPath(name) {
		absPath := name
		if !filepath.IsAbs(name) {
			absPath = filepath.Join(basePath, name)
		}
		return LoadFromDisk(absPath)
	}

	mod, ok := r.builtins[name]
	if !ok {
		return nil, fmt.Errorf("module %q not found (available: %s)", name, r.listBuiltinNames())
	}
	return mod, nil
}

// BuiltinFS returns the embedded filesystem for built-in modules
func (r *Registry) BuiltinFS() fs.FS {
	return r.builtinFS
}

// ListBuiltin returns all built-in modules
func (r *Registry) ListBuiltin() []*Module {
	mods := make([]*Module, 0, len(r.builtins))
	for _, m := range r.builtins {
		mods = append(mods, m)
	}
	return mods
}

func (r *Registry) listBuiltinNames() string {
	names := make([]string, 0, len(r.builtins))
	for name := range r.builtins {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}

func isPath(name string) bool {
	return strings.HasPrefix(name, "./") ||
		strings.HasPrefix(name, "../") ||
		strings.HasPrefix(name, "/")
}
