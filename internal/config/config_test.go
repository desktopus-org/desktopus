package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// --- BaseSpec.ImageRef ---

func TestImageRefDefault(t *testing.T) {
	b := BaseSpec{OS: "ubuntu", Desktop: "xfce"}
	want := "lscr.io/linuxserver/webtop:ubuntu-xfce"
	if got := b.ImageRef(); got != want {
		t.Errorf("ImageRef() = %q, want %q", got, want)
	}
}

func TestImageRefAlpineXfceUsesLatest(t *testing.T) {
	b := BaseSpec{OS: "alpine", Desktop: "xfce"}
	want := "lscr.io/linuxserver/webtop:latest"
	if got := b.ImageRef(); got != want {
		t.Errorf("ImageRef() = %q, want %q", got, want)
	}
}

func TestImageRefAlpineNonXfce(t *testing.T) {
	b := BaseSpec{OS: "alpine", Desktop: "kde"}
	want := "lscr.io/linuxserver/webtop:alpine-kde"
	if got := b.ImageRef(); got != want {
		t.Errorf("ImageRef() = %q, want %q", got, want)
	}
}

func TestImageRefCustomTag(t *testing.T) {
	b := BaseSpec{OS: "ubuntu", Desktop: "xfce", Tag: "custom-tag"}
	want := "lscr.io/linuxserver/webtop:custom-tag"
	if got := b.ImageRef(); got != want {
		t.Errorf("ImageRef() = %q, want %q", got, want)
	}
}

// --- DesktopConfig.ImageTag ---

func TestImageTag(t *testing.T) {
	cfg := DesktopConfig{Name: "my-desktop"}
	want := "desktopus/my-desktop:latest"
	if got := cfg.ImageTag(); got != want {
		t.Errorf("ImageTag() = %q, want %q", got, want)
	}
}

// --- ModuleRef UnmarshalYAML ---

func TestModuleRefUnmarshalString(t *testing.T) {
	yaml := `
modules:
  - chrome
  - firefox
`
	cfg := struct {
		Modules []ModuleRef `yaml:"modules"`
	}{}

	if err := yamlUnmarshal([]byte(yaml), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(cfg.Modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(cfg.Modules))
	}
	if cfg.Modules[0].Name != "chrome" {
		t.Errorf("expected chrome, got %s", cfg.Modules[0].Name)
	}
	if cfg.Modules[1].Name != "firefox" {
		t.Errorf("expected firefox, got %s", cfg.Modules[1].Name)
	}
}

func TestModuleRefUnmarshalObject(t *testing.T) {
	yaml := `
modules:
  - name: vscode
    vars:
      extensions: "ms-python.python"
`
	cfg := struct {
		Modules []ModuleRef `yaml:"modules"`
	}{}

	if err := yamlUnmarshal([]byte(yaml), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(cfg.Modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(cfg.Modules))
	}
	if cfg.Modules[0].Name != "vscode" {
		t.Errorf("expected vscode, got %s", cfg.Modules[0].Name)
	}
	if cfg.Modules[0].Vars["extensions"] != "ms-python.python" {
		t.Errorf("expected extensions var, got %v", cfg.Modules[0].Vars)
	}
}

func TestModuleRefUnmarshalMixed(t *testing.T) {
	yaml := `
modules:
  - chrome
  - name: vscode
    vars:
      theme: dark
`
	cfg := struct {
		Modules []ModuleRef `yaml:"modules"`
	}{}

	if err := yamlUnmarshal([]byte(yaml), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(cfg.Modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(cfg.Modules))
	}
	if cfg.Modules[0].Name != "chrome" {
		t.Errorf("expected chrome, got %s", cfg.Modules[0].Name)
	}
	if cfg.Modules[1].Name != "vscode" {
		t.Errorf("expected vscode, got %s", cfg.Modules[1].Name)
	}
}

// --- ValidateDesktop ---

func TestValidateDesktopValid(t *testing.T) {
	cfg := &DesktopConfig{
		Name: "my-desktop",
		Base: BaseSpec{OS: "ubuntu", Desktop: "xfce"},
	}
	if err := ValidateDesktop(cfg); err != nil {
		t.Errorf("expected valid config, got: %v", err)
	}
}

func TestValidateDesktopSingleCharName(t *testing.T) {
	cfg := &DesktopConfig{
		Name: "x",
		Base: BaseSpec{OS: "ubuntu", Desktop: "xfce"},
	}
	if err := ValidateDesktop(cfg); err != nil {
		t.Errorf("single-char name should be valid: %v", err)
	}
}

func TestValidateDesktopMissingName(t *testing.T) {
	cfg := &DesktopConfig{
		Base: BaseSpec{OS: "ubuntu", Desktop: "xfce"},
	}
	err := ValidateDesktop(cfg)
	if err == nil {
		t.Error("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Errorf("expected 'name is required', got: %v", err)
	}
}

func TestValidateDesktopInvalidName(t *testing.T) {
	tests := []string{
		"My-Desktop",  // uppercase
		"-bad",        // starts with hyphen
		"bad-",        // ends with hyphen
		"has space",   // space
		"under_score", // underscore
	}
	for _, name := range tests {
		cfg := &DesktopConfig{
			Name: name,
			Base: BaseSpec{OS: "ubuntu", Desktop: "xfce"},
		}
		if err := ValidateDesktop(cfg); err == nil {
			t.Errorf("expected error for name %q", name)
		}
	}
}

func TestValidateDesktopInvalidOS(t *testing.T) {
	cfg := &DesktopConfig{
		Name: "test",
		Base: BaseSpec{OS: "windows", Desktop: "xfce"},
	}
	err := ValidateDesktop(cfg)
	if err == nil {
		t.Error("expected error for invalid OS")
	}
	if !strings.Contains(err.Error(), "not supported") {
		t.Errorf("expected 'not supported', got: %v", err)
	}
}

func TestValidateDesktopInvalidDesktopEnv(t *testing.T) {
	cfg := &DesktopConfig{
		Name: "test",
		Base: BaseSpec{OS: "ubuntu", Desktop: "gnome"},
	}
	err := ValidateDesktop(cfg)
	if err == nil {
		t.Error("expected error for invalid desktop")
	}
}

func TestValidateDesktopAllValidOS(t *testing.T) {
	for _, os := range []string{"ubuntu", "debian", "fedora", "arch", "alpine", "el"} {
		cfg := &DesktopConfig{
			Name: "test",
			Base: BaseSpec{OS: os, Desktop: "xfce"},
		}
		if err := ValidateDesktop(cfg); err != nil {
			t.Errorf("OS %q should be valid: %v", os, err)
		}
	}
}

func TestValidateDesktopAllValidDesktops(t *testing.T) {
	for _, de := range []string{"xfce", "kde", "i3", "mate"} {
		cfg := &DesktopConfig{
			Name: "test",
			Base: BaseSpec{OS: "ubuntu", Desktop: de},
		}
		if err := ValidateDesktop(cfg); err != nil {
			t.Errorf("desktop %q should be valid: %v", de, err)
		}
	}
}

func TestValidateDesktopIncompatibleCombo(t *testing.T) {
	cfg := &DesktopConfig{
		Name: "test",
		Base: BaseSpec{OS: "el", Desktop: "kde"},
	}
	err := ValidateDesktop(cfg)
	if err == nil {
		t.Error("expected error for el + kde (incompatible)")
	}
	if !strings.Contains(err.Error(), "not available for os") {
		t.Errorf("expected 'not available for os' error, got: %v", err)
	}
}

func TestValidateDesktopELValidCombos(t *testing.T) {
	for _, de := range []string{"i3", "mate", "xfce"} {
		cfg := &DesktopConfig{
			Name: "test",
			Base: BaseSpec{OS: "el", Desktop: de},
		}
		if err := ValidateDesktop(cfg); err != nil {
			t.Errorf("el + %q should be valid: %v", de, err)
		}
	}
}

func TestValidateDesktopRemovedDesktops(t *testing.T) {
	for _, de := range []string{"openbox", "icewm"} {
		cfg := &DesktopConfig{
			Name: "test",
			Base: BaseSpec{OS: "ubuntu", Desktop: de},
		}
		if err := ValidateDesktop(cfg); err == nil {
			t.Errorf("desktop %q should be rejected", de)
		}
	}
}

func TestValidateDesktopPostRunInvalidRunAs(t *testing.T) {
	cfg := &DesktopConfig{
		Name: "test",
		Base: BaseSpec{OS: "ubuntu", Desktop: "xfce"},
		PostRun: []PostRunScript{
			{Name: "setup", Script: "echo hi", RunAs: "nobody"},
		},
	}
	err := ValidateDesktop(cfg)
	if err == nil {
		t.Error("expected error for invalid runas")
	}
	if !strings.Contains(err.Error(), "runas must be") {
		t.Errorf("expected runas error, got: %v", err)
	}
}

func TestValidateDesktopPostRunValidRunAs(t *testing.T) {
	for _, runas := range []string{"", "root", "abc"} {
		cfg := &DesktopConfig{
			Name: "test",
			Base: BaseSpec{OS: "ubuntu", Desktop: "xfce"},
			PostRun: []PostRunScript{
				{Name: "setup", Script: "echo hi", RunAs: runas},
			},
		}
		if err := ValidateDesktop(cfg); err != nil {
			t.Errorf("runas %q should be valid: %v", runas, err)
		}
	}
}

func TestValidateDesktopMultipleErrors(t *testing.T) {
	cfg := &DesktopConfig{} // missing everything
	err := ValidateDesktop(cfg)
	if err == nil {
		t.Fatal("expected errors")
	}
	// Should have at least name + os + desktop errors
	if strings.Count(err.Error(), "\n") < 2 {
		t.Errorf("expected multiple errors, got: %v", err)
	}
}

func TestValidateDesktopModuleMissingName(t *testing.T) {
	cfg := &DesktopConfig{
		Name:    "test",
		Base:    BaseSpec{OS: "ubuntu", Desktop: "xfce"},
		Modules: []ModuleRef{{Name: ""}},
	}
	err := ValidateDesktop(cfg)
	if err == nil {
		t.Error("expected error for module without name")
	}
}

func TestValidateDesktopFilesMissingPath(t *testing.T) {
	cfg := &DesktopConfig{
		Name: "test",
		Base: BaseSpec{OS: "ubuntu", Desktop: "xfce"},
		Files: []FileSpec{{Content: "hello"}},
	}
	err := ValidateDesktop(cfg)
	if err == nil {
		t.Error("expected error for file without path")
	}
}

// --- FindDesktopConfig ---

func TestFindDesktopConfigFromDir(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "desktopus.yaml")
	if err := os.WriteFile(configFile, []byte("name: test"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := FindDesktopConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != configFile {
		t.Errorf("expected %s, got %s", configFile, got)
	}
}

func TestFindDesktopConfigFromFile(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "custom.yaml")
	if err := os.WriteFile(configFile, []byte("name: test"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := FindDesktopConfig(configFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != configFile {
		t.Errorf("expected %s, got %s", configFile, got)
	}
}

func TestFindDesktopConfigMissing(t *testing.T) {
	dir := t.TempDir()

	_, err := FindDesktopConfig(dir)
	if err == nil {
		t.Error("expected error when no desktopus.yaml in directory")
	}
}

func TestFindDesktopConfigBadPath(t *testing.T) {
	_, err := FindDesktopConfig("/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

// --- LoadDesktop ---

func TestLoadDesktop(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "desktopus.yaml")

	yaml := `
name: test-desktop
base:
  os: ubuntu
  desktop: xfce
modules:
  - chrome
  - name: vscode
    vars:
      theme: dark
runtime:
  shm_size: 2g
  ports:
    - "3000:3000"
  env:
    TZ: UTC
`
	if err := os.WriteFile(configFile, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadDesktop(configFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Name != "test-desktop" {
		t.Errorf("expected name 'test-desktop', got %q", cfg.Name)
	}
	if cfg.Base.OS != "ubuntu" {
		t.Errorf("expected OS ubuntu, got %q", cfg.Base.OS)
	}
	if len(cfg.Modules) != 2 {
		t.Errorf("expected 2 modules, got %d", len(cfg.Modules))
	}
	if cfg.Runtime.ShmSize != "2g" {
		t.Errorf("expected shm_size 2g, got %q", cfg.Runtime.ShmSize)
	}
	if cfg.Runtime.Env["TZ"] != "UTC" {
		t.Errorf("expected TZ=UTC, got %q", cfg.Runtime.Env["TZ"])
	}
}

func TestLoadDesktopInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "desktopus.yaml")
	if err := os.WriteFile(configFile, []byte("{{invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadDesktop(configFile)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadDesktopValidationFails(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "desktopus.yaml")

	yaml := `
name: INVALID
base:
  os: ubuntu
  desktop: xfce
`
	if err := os.WriteFile(configFile, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadDesktop(configFile)
	if err == nil {
		t.Error("expected validation error")
	}
}

// --- LoadApp ---

func TestLoadAppDefaults(t *testing.T) {
	cfg, err := LoadApp("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Server.Port != 7575 {
		t.Errorf("expected default port 7575, got %d", cfg.Server.Port)
	}
}

func TestLoadAppMissingFile(t *testing.T) {
	cfg, err := LoadApp("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return defaults
	if cfg == nil {
		t.Fatal("expected defaults for missing file")
	}
}

func TestLoadAppFromFile(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")

	yaml := `
docker:
  host: "tcp://localhost:2375"
server:
  port: 9999
`
	if err := os.WriteFile(configFile, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadApp(configFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Docker.Host != "tcp://localhost:2375" {
		t.Errorf("expected custom docker host, got %q", cfg.Docker.Host)
	}
	if cfg.Server.Port != 9999 {
		t.Errorf("expected port 9999, got %d", cfg.Server.Port)
	}
}

// helper wrapping yaml.Unmarshal
func yamlUnmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}
