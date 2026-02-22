package module

import (
	"testing"
	"testing/fstest"
)

// --- IsCompatible ---

func TestIsCompatibleNoConstraints(t *testing.T) {
	mod := &Module{Name: "test"}
	if !mod.IsCompatible("ubuntu", "xfce") {
		t.Error("module with no constraints should be compatible with anything")
	}
}

func TestIsCompatibleOSMatch(t *testing.T) {
	mod := &Module{
		Name:          "test",
		Compatibility: Compatibility{OS: []string{"ubuntu", "debian"}},
	}
	if !mod.IsCompatible("ubuntu", "xfce") {
		t.Error("should be compatible with ubuntu")
	}
	if !mod.IsCompatible("debian", "kde") {
		t.Error("should be compatible with debian")
	}
	if mod.IsCompatible("fedora", "xfce") {
		t.Error("should NOT be compatible with fedora")
	}
}

func TestIsCompatibleDesktopMatch(t *testing.T) {
	mod := &Module{
		Name:          "test",
		Compatibility: Compatibility{Desktop: []string{"xfce", "kde"}},
	}
	if !mod.IsCompatible("ubuntu", "xfce") {
		t.Error("should be compatible with xfce")
	}
	if mod.IsCompatible("ubuntu", "i3") {
		t.Error("should NOT be compatible with i3")
	}
}

func TestIsCompatibleBothConstraints(t *testing.T) {
	mod := &Module{
		Name: "test",
		Compatibility: Compatibility{
			OS:      []string{"ubuntu"},
			Desktop: []string{"xfce"},
		},
	}
	if !mod.IsCompatible("ubuntu", "xfce") {
		t.Error("should be compatible with ubuntu+xfce")
	}
	if mod.IsCompatible("ubuntu", "kde") {
		t.Error("should NOT be compatible with ubuntu+kde")
	}
	if mod.IsCompatible("fedora", "xfce") {
		t.Error("should NOT be compatible with fedora+xfce")
	}
}

// --- LoadFromFS ---

func TestLoadFromFS(t *testing.T) {
	fsys := fstest.MapFS{
		"mymod/module.yaml": &fstest.MapFile{
			Data: []byte(`
name: mymod
description: Test module
version: "1.0"
vars:
  my_var:
    default: hello
    description: A variable
system_packages:
  - curl
`),
		},
		"mymod/tasks/main.yml": &fstest.MapFile{
			Data: []byte("- name: test\n  debug: msg=hi\n"),
		},
	}

	mod, err := LoadFromFS(fsys, "mymod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mod.Name != "mymod" {
		t.Errorf("expected name 'mymod', got %q", mod.Name)
	}
	if mod.Description != "Test module" {
		t.Errorf("expected description, got %q", mod.Description)
	}
	if mod.Version != "1.0" {
		t.Errorf("expected version 1.0, got %q", mod.Version)
	}
	if len(mod.Vars) != 1 {
		t.Errorf("expected 1 var, got %d", len(mod.Vars))
	}
	if mod.Vars["my_var"].Default != "hello" {
		t.Errorf("expected default 'hello', got %q", mod.Vars["my_var"].Default)
	}
	if len(mod.SystemPkgs) != 1 || mod.SystemPkgs[0] != "curl" {
		t.Errorf("expected system_packages [curl], got %v", mod.SystemPkgs)
	}
}

func TestLoadFromFSMissingModuleYaml(t *testing.T) {
	fsys := fstest.MapFS{
		"mymod/tasks/main.yml": &fstest.MapFile{Data: []byte("")},
	}

	_, err := LoadFromFS(fsys, "mymod")
	if err == nil {
		t.Error("expected error for missing module.yaml")
	}
}

func TestLoadFromFSMissingTasks(t *testing.T) {
	fsys := fstest.MapFS{
		"mymod/module.yaml": &fstest.MapFile{
			Data: []byte("name: mymod\n"),
		},
	}

	_, err := LoadFromFS(fsys, "mymod")
	if err == nil {
		t.Error("expected error for missing tasks/main.yml")
	}
}

func TestLoadFromFSInvalidYAML(t *testing.T) {
	fsys := fstest.MapFS{
		"mymod/module.yaml": &fstest.MapFile{
			Data: []byte("{{invalid"),
		},
		"mymod/tasks/main.yml": &fstest.MapFile{Data: []byte("")},
	}

	_, err := LoadFromFS(fsys, "mymod")
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

// --- Registry ---

func TestNewRegistry(t *testing.T) {
	fsys := fstest.MapFS{
		"chrome/module.yaml": &fstest.MapFile{
			Data: []byte("name: chrome\ndescription: Google Chrome\n"),
		},
		"chrome/tasks/main.yml": &fstest.MapFile{
			Data: []byte("- name: install chrome\n  debug: msg=hi\n"),
		},
		"readme.txt": &fstest.MapFile{
			Data: []byte("not a module"),
		},
	}

	reg, err := NewRegistry(fsys)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mods := reg.ListBuiltin()
	if len(mods) != 1 {
		t.Fatalf("expected 1 built-in module, got %d", len(mods))
	}
	if mods[0].Name != "chrome" {
		t.Errorf("expected chrome, got %s", mods[0].Name)
	}
	if !mods[0].Builtin {
		t.Error("expected module to be marked as builtin")
	}
}

func TestRegistryResolveBuiltin(t *testing.T) {
	fsys := fstest.MapFS{
		"chrome/module.yaml": &fstest.MapFile{
			Data: []byte("name: chrome\n"),
		},
		"chrome/tasks/main.yml": &fstest.MapFile{
			Data: []byte("- name: test\n  debug: msg=hi\n"),
		},
	}

	reg, err := NewRegistry(fsys)
	if err != nil {
		t.Fatal(err)
	}

	mod, err := reg.Resolve("chrome", "/tmp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mod.Name != "chrome" {
		t.Errorf("expected chrome, got %s", mod.Name)
	}
}

func TestRegistryResolveNotFound(t *testing.T) {
	fsys := fstest.MapFS{
		"chrome/module.yaml": &fstest.MapFile{
			Data: []byte("name: chrome\n"),
		},
		"chrome/tasks/main.yml": &fstest.MapFile{
			Data: []byte("- name: test\n  debug: msg=hi\n"),
		},
	}

	reg, err := NewRegistry(fsys)
	if err != nil {
		t.Fatal(err)
	}

	_, err = reg.Resolve("nonexistent", "/tmp")
	if err == nil {
		t.Error("expected error for nonexistent module")
	}
}

// --- isPath ---

func TestIsPath(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"chrome", false},
		{"vscode", false},
		{"./my-module", true},
		{"../other-module", true},
		{"/absolute/path", true},
		{"relative/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isPath(tt.input); got != tt.want {
				t.Errorf("isPath(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestRegistryBuiltinFS(t *testing.T) {
	fsys := fstest.MapFS{
		"chrome/module.yaml": &fstest.MapFile{
			Data: []byte("name: chrome\n"),
		},
		"chrome/tasks/main.yml": &fstest.MapFile{
			Data: []byte("- name: test\n  debug: msg=hi\n"),
		},
	}

	reg, err := NewRegistry(fsys)
	if err != nil {
		t.Fatal(err)
	}

	if reg.BuiltinFS() == nil {
		t.Error("expected non-nil BuiltinFS")
	}
}
