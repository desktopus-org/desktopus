package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/desktopus-org/desktopus/internal/config"
	"github.com/desktopus-org/desktopus/internal/runtime"
)

// --- containerNameFromImage ---

func TestContainerNameFromImage(t *testing.T) {
	tests := []struct {
		image string
		want  string
	}{
		{"desktopus/mydesk:latest", "mydesk"},
		{"mydesk:latest", "mydesk"},
		{"mydesk", "mydesk"},
		{"registry.example.com/mydesk:v1", "mydesk"},
		{"registry.example.com/org/mydesk:v1", "mydesk"},
	}
	for _, tt := range tests {
		got := containerNameFromImage(tt.image)
		if got != tt.want {
			t.Errorf("containerNameFromImage(%q) = %q, want %q", tt.image, got, tt.want)
		}
	}
}

// --- findRuntimeYAML ---

func TestFindRuntimeYAMLFromDir(t *testing.T) {
	dir := t.TempDir()
	got, err := findRuntimeYAML(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(dir, "desktopus.runtime.yaml")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFindRuntimeYAMLFromFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "custom.runtime.yaml")
	if err := os.WriteFile(file, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := findRuntimeYAML(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != file {
		t.Errorf("got %q, want %q", got, file)
	}
}

func TestFindRuntimeYAMLEmpty(t *testing.T) {
	dir := t.TempDir()
	got, err := findRuntimeYAML("")
	if err != nil {
		t.Fatalf("empty path should use '.', got error: %v", err)
	}
	// Should resolve to desktopus.runtime.yaml in the working directory
	want := "desktopus.runtime.yaml"
	if filepath.Base(got) != want {
		t.Errorf("got base %q, want %q (full: %q, dir: %q)", filepath.Base(got), want, got, dir)
	}
}

func TestFindRuntimeYAMLBadPath(t *testing.T) {
	_, err := findRuntimeYAML("/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

// --- toDesktopRunConfig ---

func TestToDesktopRunConfigUsesRuntimeName(t *testing.T) {
	rt := &config.RuntimeConfig{
		Name:     "mydesk",
		ShmSize:  "2g",
		Env:      map[string]string{"TZ": "UTC"},
	}
	cfg := toDesktopRunConfig(rt, "desktopus/mydesk:latest")
	if cfg.Name != "mydesk" {
		t.Errorf("expected name 'mydesk', got %q", cfg.Name)
	}
}

func TestToDesktopRunConfigDerivesNameFromImage(t *testing.T) {
	rt := &config.RuntimeConfig{ShmSize: "2g"}
	cfg := toDesktopRunConfig(rt, "desktopus/mydesk:latest")
	if cfg.Name != "mydesk" {
		t.Errorf("expected name derived from image 'mydesk', got %q", cfg.Name)
	}
}

func TestToDesktopRunConfigSetsImageTag(t *testing.T) {
	rt := &config.RuntimeConfig{}
	cfg := toDesktopRunConfig(rt, "registry.example.com/desk:v1")
	if cfg.ImageTag != "registry.example.com/desk:v1" {
		t.Errorf("expected ImageTag 'registry.example.com/desk:v1', got %q", cfg.ImageTag)
	}
}

func TestToDesktopRunConfigMapsRuntimeFields(t *testing.T) {
	rt := &config.RuntimeConfig{
		Hostname: "myhost",
		ShmSize:  "4g",
		Ports:    []string{"3000:3000"},
		Volumes:  []string{"~/projects:/config/projects"},
		GPU:      "intel",
		Memory:   "8g",
		CPUs:     4,
		Restart:  "unless-stopped",
		Network:  "host",
		Env:      map[string]string{"TZ": "UTC", "PUID": "1000"},
	}
	cfg := toDesktopRunConfig(rt, "mydesk:latest")

	checks := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Hostname", cfg.Hostname, "myhost"},
		{"ShmSize", cfg.ShmSize, "4g"},
		{"GPU", cfg.GPU, "intel"},
		{"Memory", cfg.Memory, "8g"},
		{"CPUs", cfg.CPUs, 4},
		{"Restart", cfg.Restart, "unless-stopped"},
		{"Network", cfg.Network, "host"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, c.got, c.want)
		}
	}
	if len(cfg.Ports) != 1 || cfg.Ports[0] != "3000:3000" {
		t.Errorf("Ports: got %v, want [3000:3000]", cfg.Ports)
	}
	if cfg.Env["TZ"] != "UTC" || cfg.Env["PUID"] != "1000" {
		t.Errorf("Env: got %v", cfg.Env)
	}
}

// verify toDesktopRunConfig returns the right type
var _ *runtime.DesktopRunConfig = (*runtime.DesktopRunConfig)(nil)
