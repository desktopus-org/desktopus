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

// --- ResolveImageTag ---

func TestResolveImageTagCLIOverride(t *testing.T) {
	rt := &RuntimeConfig{DefaultImage: "desktopus/desk:latest"}
	got, err := ResolveImageTag(rt, "myregistry.io/desk:override")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "myregistry.io/desk:override" {
		t.Errorf("expected CLI override to win, got %q", got)
	}
}

func TestResolveImageTagRuntimeDefault(t *testing.T) {
	rt := &RuntimeConfig{DefaultImage: "registry.example.com/desk:v1"}
	got, err := ResolveImageTag(rt, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "registry.example.com/desk:v1" {
		t.Errorf("expected runtime default_image, got %q", got)
	}
}

func TestResolveImageTagNilRuntime(t *testing.T) {
	_, err := ResolveImageTag(nil, "")
	if err == nil {
		t.Fatal("expected error for nil runtime with no override")
	}
	if !strings.Contains(err.Error(), "no image tag defined") {
		t.Errorf("expected 'no image tag defined' error, got: %v", err)
	}
}

func TestResolveImageTagError(t *testing.T) {
	rt := &RuntimeConfig{}
	_, err := ResolveImageTag(rt, "")
	if err == nil {
		t.Fatal("expected error when no image tag is defined")
	}
	if !strings.Contains(err.Error(), "no image tag defined") {
		t.Errorf("expected 'no image tag defined' error, got: %v", err)
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

// --- ValidateImage ---

func TestValidateImageValid(t *testing.T) {
	cfg := &ImageConfig{
		Name: "my-desktop",
		Base: BaseSpec{OS: "ubuntu", Desktop: "xfce"},
	}
	if err := ValidateImage(cfg); err != nil {
		t.Errorf("expected valid config, got: %v", err)
	}
}

func TestValidateImageSingleCharName(t *testing.T) {
	cfg := &ImageConfig{
		Name: "x",
		Base: BaseSpec{OS: "ubuntu", Desktop: "xfce"},
	}
	if err := ValidateImage(cfg); err != nil {
		t.Errorf("single-char name should be valid: %v", err)
	}
}

func TestValidateImageMissingName(t *testing.T) {
	cfg := &ImageConfig{
		Base: BaseSpec{OS: "ubuntu", Desktop: "xfce"},
	}
	err := ValidateImage(cfg)
	if err == nil {
		t.Error("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Errorf("expected 'name is required', got: %v", err)
	}
}

func TestValidateImageInvalidName(t *testing.T) {
	tests := []string{
		"My-Desktop",  // uppercase
		"-bad",        // starts with hyphen
		"bad-",        // ends with hyphen
		"has space",   // space
		"under_score", // underscore
	}
	for _, name := range tests {
		cfg := &ImageConfig{
			Name: name,
			Base: BaseSpec{OS: "ubuntu", Desktop: "xfce"},
		}
		if err := ValidateImage(cfg); err == nil {
			t.Errorf("expected error for name %q", name)
		}
	}
}

func TestValidateImageInvalidOS(t *testing.T) {
	cfg := &ImageConfig{
		Name: "test",
		Base: BaseSpec{OS: "windows", Desktop: "xfce"},
	}
	err := ValidateImage(cfg)
	if err == nil {
		t.Error("expected error for invalid OS")
	}
	if !strings.Contains(err.Error(), "not supported") {
		t.Errorf("expected 'not supported', got: %v", err)
	}
}

func TestValidateImageInvalidDesktopEnv(t *testing.T) {
	cfg := &ImageConfig{
		Name: "test",
		Base: BaseSpec{OS: "ubuntu", Desktop: "gnome"},
	}
	err := ValidateImage(cfg)
	if err == nil {
		t.Error("expected error for invalid desktop")
	}
}

func TestValidateImageAllValidOS(t *testing.T) {
	for _, os := range []string{"ubuntu", "debian", "fedora", "arch", "alpine", "el"} {
		cfg := &ImageConfig{
			Name: "test",
			Base: BaseSpec{OS: os, Desktop: "xfce"},
		}
		if err := ValidateImage(cfg); err != nil {
			t.Errorf("OS %q should be valid: %v", os, err)
		}
	}
}

func TestValidateImageAllValidDesktops(t *testing.T) {
	for _, de := range []string{"xfce", "kde", "i3", "mate"} {
		cfg := &ImageConfig{
			Name: "test",
			Base: BaseSpec{OS: "ubuntu", Desktop: de},
		}
		if err := ValidateImage(cfg); err != nil {
			t.Errorf("desktop %q should be valid: %v", de, err)
		}
	}
}

func TestValidateImageIncompatibleCombo(t *testing.T) {
	cfg := &ImageConfig{
		Name: "test",
		Base: BaseSpec{OS: "el", Desktop: "kde"},
	}
	err := ValidateImage(cfg)
	if err == nil {
		t.Error("expected error for el + kde (incompatible)")
	}
	if !strings.Contains(err.Error(), "not available for os") {
		t.Errorf("expected 'not available for os' error, got: %v", err)
	}
}

func TestValidateImageELValidCombos(t *testing.T) {
	for _, de := range []string{"i3", "mate", "xfce"} {
		cfg := &ImageConfig{
			Name: "test",
			Base: BaseSpec{OS: "el", Desktop: de},
		}
		if err := ValidateImage(cfg); err != nil {
			t.Errorf("el + %q should be valid: %v", de, err)
		}
	}
}

func TestValidateImageRemovedDesktops(t *testing.T) {
	for _, de := range []string{"openbox", "icewm"} {
		cfg := &ImageConfig{
			Name: "test",
			Base: BaseSpec{OS: "ubuntu", Desktop: de},
		}
		if err := ValidateImage(cfg); err == nil {
			t.Errorf("desktop %q should be rejected", de)
		}
	}
}

func TestValidateImagePostRunInvalidRunAs(t *testing.T) {
	cfg := &ImageConfig{
		Name: "test",
		Base: BaseSpec{OS: "ubuntu", Desktop: "xfce"},
		PostRun: []PostRunScript{
			{Name: "setup", Script: "echo hi", RunAs: "nobody"},
		},
	}
	err := ValidateImage(cfg)
	if err == nil {
		t.Error("expected error for invalid runas")
	}
	if !strings.Contains(err.Error(), "runas must be") {
		t.Errorf("expected runas error, got: %v", err)
	}
}

func TestValidateImagePostRunValidRunAs(t *testing.T) {
	// Default user is "desktopus" when no user is set
	for _, runas := range []string{"", "root", "desktopus"} {
		cfg := &ImageConfig{
			Name: "test",
			Base: BaseSpec{OS: "ubuntu", Desktop: "xfce"},
			PostRun: []PostRunScript{
				{Name: "setup", Script: "echo hi", RunAs: runas},
			},
		}
		if err := ValidateImage(cfg); err != nil {
			t.Errorf("runas %q should be valid: %v", runas, err)
		}
	}
	// Custom user: runas must match the configured user
	cfg := &ImageConfig{
		Name: "test",
		Base: BaseSpec{OS: "ubuntu", Desktop: "xfce"},
		User: "carlos",
		PostRun: []PostRunScript{
			{Name: "setup", Script: "echo hi", RunAs: "carlos"},
		},
	}
	if err := ValidateImage(cfg); err != nil {
		t.Errorf("runas 'carlos' should be valid for user 'carlos': %v", err)
	}
}

func TestValidateImageMultipleErrors(t *testing.T) {
	cfg := &ImageConfig{} // missing everything
	err := ValidateImage(cfg)
	if err == nil {
		t.Fatal("expected errors")
	}
	// Should have at least name + os + desktop errors
	if strings.Count(err.Error(), "\n") < 2 {
		t.Errorf("expected multiple errors, got: %v", err)
	}
}

func TestValidateImageModuleMissingName(t *testing.T) {
	cfg := &ImageConfig{
		Name:    "test",
		Base:    BaseSpec{OS: "ubuntu", Desktop: "xfce"},
		Modules: []ModuleRef{{Name: ""}},
	}
	err := ValidateImage(cfg)
	if err == nil {
		t.Error("expected error for module without name")
	}
}

func TestValidateImageFilesMissingPath(t *testing.T) {
	cfg := &ImageConfig{
		Name: "test",
		Base: BaseSpec{OS: "ubuntu", Desktop: "xfce"},
		Files: []FileSpec{{Content: "hello"}},
	}
	err := ValidateImage(cfg)
	if err == nil {
		t.Error("expected error for file without path")
	}
}

// --- FindImageConfig ---

func TestFindImageConfigFromDir(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "desktopus.yaml")
	if err := os.WriteFile(configFile, []byte("name: test"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := FindImageConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != configFile {
		t.Errorf("expected %s, got %s", configFile, got)
	}
}

func TestFindImageConfigFromFile(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "custom.yaml")
	if err := os.WriteFile(configFile, []byte("name: test"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := FindImageConfig(configFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != configFile {
		t.Errorf("expected %s, got %s", configFile, got)
	}
}

func TestFindImageConfigMissing(t *testing.T) {
	dir := t.TempDir()

	_, err := FindImageConfig(dir)
	if err == nil {
		t.Error("expected error when no desktopus.yaml in directory")
	}
}

func TestFindImageConfigBadPath(t *testing.T) {
	_, err := FindImageConfig("/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

// --- LoadImage ---

func TestLoadImage(t *testing.T) {
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
`
	if err := os.WriteFile(configFile, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadImage(configFile)
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
}

func TestLoadImageInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "desktopus.yaml")
	if err := os.WriteFile(configFile, []byte("{{invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadImage(configFile)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadImageValidationFails(t *testing.T) {
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

	_, err := LoadImage(configFile)
	if err == nil {
		t.Error("expected validation error")
	}
}

// --- LoadRuntime ---

func TestLoadRuntimeMissingFile(t *testing.T) {
	cfg, err := LoadRuntime("/nonexistent/desktopus.runtime.yaml")
	if err != nil {
		t.Fatalf("expected no error for missing runtime file, got: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil zero-value RuntimeConfig")
	}
}

func TestLoadRuntime(t *testing.T) {
	dir := t.TempDir()
	runtimeFile := filepath.Join(dir, "desktopus.runtime.yaml")

	content := `
shm_size: 2g
ports:
  - "3000:3000"
env:
  TZ: UTC
`
	if err := os.WriteFile(runtimeFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadRuntime(runtimeFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ShmSize != "2g" {
		t.Errorf("expected shm_size 2g, got %q", cfg.ShmSize)
	}
	if cfg.Env["TZ"] != "UTC" {
		t.Errorf("expected TZ=UTC, got %q", cfg.Env["TZ"])
	}
}

func TestLoadRuntimeInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	runtimeFile := filepath.Join(dir, "desktopus.runtime.yaml")
	if err := os.WriteFile(runtimeFile, []byte("{{invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadRuntime(runtimeFile)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

// --- FindRuntimeConfig ---

func TestFindRuntimeConfig(t *testing.T) {
	got := FindRuntimeConfig("/some/dir/desktopus.yaml")
	want := "/some/dir/desktopus.runtime.yaml"
	if got != want {
		t.Errorf("FindRuntimeConfig() = %q, want %q", got, want)
	}
}

// --- ValidateRuntime ---

func TestValidateRuntimeValidRestart(t *testing.T) {
	for _, r := range []string{"", "no", "always", "unless-stopped", "on-failure"} {
		cfg := &RuntimeConfig{Restart: r}
		if err := ValidateRuntime(cfg); err != nil {
			t.Errorf("restart %q should be valid: %v", r, err)
		}
	}
}

func TestValidateRuntimeInvalidRestart(t *testing.T) {
	cfg := &RuntimeConfig{Restart: "sometimes"}
	if err := ValidateRuntime(cfg); err == nil {
		t.Error("expected error for invalid restart policy")
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

// --- Custom user validation ---

func TestValidateImageCustomUser(t *testing.T) {
	for _, user := range []string{"carlos", "my_user", "x"} {
		cfg := &ImageConfig{
			Name: "test",
			Base: BaseSpec{OS: "ubuntu", Desktop: "xfce"},
			User: user,
		}
		if err := ValidateImage(cfg); err != nil {
			t.Errorf("user %q should be valid: %v", user, err)
		}
	}
}

func TestValidateImageInvalidUser(t *testing.T) {
	tests := []struct {
		user string
		desc string
	}{
		{"root", "root is reserved"},
		{"Root", "uppercase not allowed"},
		{"my user", "spaces not allowed"},
		{strings.Repeat("a", 33), "exceeds 32 chars"},
	}
	for _, tt := range tests {
		cfg := &ImageConfig{
			Name: "test",
			Base: BaseSpec{OS: "ubuntu", Desktop: "xfce"},
			User: tt.user,
		}
		if err := ValidateImage(cfg); err == nil {
			t.Errorf("user %q (%s) should be invalid", tt.user, tt.desc)
		}
	}
}

func TestValidateImageCustomHome(t *testing.T) {
	// Valid absolute paths
	for _, home := range []string{"/home/carlos", "/workspace", "/"} {
		cfg := &ImageConfig{
			Name: "test",
			Base: BaseSpec{OS: "ubuntu", Desktop: "xfce"},
			Home: home,
		}
		if err := ValidateImage(cfg); err != nil {
			t.Errorf("home %q should be valid: %v", home, err)
		}
	}
	// Invalid: relative path
	cfg := &ImageConfig{
		Name: "test",
		Base: BaseSpec{OS: "ubuntu", Desktop: "xfce"},
		Home: "home/carlos",
	}
	if err := ValidateImage(cfg); err == nil {
		t.Error("relative home path should be invalid")
	}
}

// --- EffectiveUser / EffectiveHome ---

func TestEffectiveUserHome(t *testing.T) {
	tests := []struct {
		user     string
		home     string
		wantUser string
		wantHome string
	}{
		{"", "", "desktopus", "/home/desktopus"},
		{"abc", "", "abc", "/config"},
		{"carlos", "", "carlos", "/home/carlos"},
		{"carlos", "/workspace", "carlos", "/workspace"},
	}
	for _, tt := range tests {
		cfg := &ImageConfig{User: tt.user, Home: tt.home}
		if got := cfg.EffectiveUser(); got != tt.wantUser {
			t.Errorf("user=%q home=%q: EffectiveUser()=%q, want %q", tt.user, tt.home, got, tt.wantUser)
		}
		if got := cfg.EffectiveHome(); got != tt.wantHome {
			t.Errorf("user=%q home=%q: EffectiveHome()=%q, want %q", tt.user, tt.home, got, tt.wantHome)
		}
	}
}

// helper wrapping yaml.Unmarshal
func yamlUnmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}
