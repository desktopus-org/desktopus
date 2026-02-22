package runtime

import (
	"testing"

	"github.com/moby/moby/api/types/container"
)

func TestParseSize(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"", 0},
		{"0", 0},
		{"512", 512},
		{"1k", 1024},
		{"1K", 1024},
		{"512m", 512 * 1024 * 1024},
		{"512M", 512 * 1024 * 1024},
		{"2g", 2 * 1024 * 1024 * 1024},
		{"2G", 2 * 1024 * 1024 * 1024},
		{"  1g  ", 1024 * 1024 * 1024},
		{"abc", 0},
		{"1x", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseSize(tt.input)
			if got != tt.want {
				t.Errorf("parseSize(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildEnvList(t *testing.T) {
	cfg := &DesktopRunConfig{
		Env: map[string]string{
			"TZ":   "UTC",
			"PUID": "1000",
		},
	}
	opts := RunOptions{
		Env: map[string]string{
			"PUID":  "1001", // override
			"EXTRA": "val",
		},
	}

	list := buildEnvList(cfg, opts)

	envMap := make(map[string]string)
	for _, e := range list {
		parts := splitEnv(e)
		envMap[parts[0]] = parts[1]
	}

	if envMap["TZ"] != "UTC" {
		t.Errorf("expected TZ=UTC, got %q", envMap["TZ"])
	}
	if envMap["PUID"] != "1001" {
		t.Errorf("expected PUID=1001 (override), got %q", envMap["PUID"])
	}
	if envMap["EXTRA"] != "val" {
		t.Errorf("expected EXTRA=val, got %q", envMap["EXTRA"])
	}
	if len(list) != 3 {
		t.Errorf("expected 3 env entries, got %d", len(list))
	}
}

func TestBuildEnvListEmpty(t *testing.T) {
	cfg := &DesktopRunConfig{Env: nil}
	opts := RunOptions{}

	list := buildEnvList(cfg, opts)
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d entries", len(list))
	}
}

func splitEnv(s string) [2]string {
	for i := range s {
		if s[i] == '=' {
			return [2]string{s[:i], s[i+1:]}
		}
	}
	return [2]string{s, ""}
}

func TestBuildPortBindings(t *testing.T) {
	cfg := &DesktopRunConfig{
		Ports: []string{"3000:3000", "8080:80"},
	}
	opts := RunOptions{
		Ports: []string{"9090:9090"},
	}

	portMap, portSet, err := buildPortBindings(cfg, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(portMap) != 3 {
		t.Errorf("expected 3 port mappings, got %d", len(portMap))
	}
	if len(portSet) != 3 {
		t.Errorf("expected 3 exposed ports, got %d", len(portSet))
	}
}

func TestBuildPortBindingsInvalid(t *testing.T) {
	cfg := &DesktopRunConfig{
		Ports: []string{"invalid"},
	}
	opts := RunOptions{}

	_, _, err := buildPortBindings(cfg, opts)
	if err == nil {
		t.Error("expected error for invalid port mapping")
	}
}

func TestBuildPortBindingsInvalidPort(t *testing.T) {
	cfg := &DesktopRunConfig{
		Ports: []string{"3000:abc"},
	}
	opts := RunOptions{}

	_, _, err := buildPortBindings(cfg, opts)
	if err == nil {
		t.Error("expected error for invalid container port")
	}
}

func TestBuildVolumes(t *testing.T) {
	cfg := &DesktopRunConfig{
		Volumes: []string{"/data:/data"},
	}
	opts := RunOptions{
		Volumes: []string{"/extra:/extra:ro"},
	}

	binds := buildVolumes(cfg, opts)
	if len(binds) != 2 {
		t.Fatalf("expected 2 volumes, got %d", len(binds))
	}
	if binds[0] != "/data:/data" {
		t.Errorf("expected /data:/data, got %s", binds[0])
	}
	if binds[1] != "/extra:/extra:ro" {
		t.Errorf("expected /extra:/extra:ro, got %s", binds[1])
	}
}

func TestBuildVolumesHomeDirExpansion(t *testing.T) {
	cfg := &DesktopRunConfig{
		Volumes: []string{"~/projects:/config/projects"},
	}
	opts := RunOptions{}

	binds := buildVolumes(cfg, opts)
	if len(binds) != 1 {
		t.Fatalf("expected 1 volume, got %d", len(binds))
	}
	// Should have expanded ~/ to home dir
	if binds[0] == "~/projects:/config/projects" {
		t.Error("expected ~/ to be expanded to home directory")
	}
}

func TestFormatPortsEmpty(t *testing.T) {
	result := formatPorts(nil)
	if result != "-" {
		t.Errorf("expected '-', got %q", result)
	}

	result = formatPorts([]container.PortSummary{})
	if result != "-" {
		t.Errorf("expected '-', got %q", result)
	}
}

func TestFormatPortsNoPublic(t *testing.T) {
	ports := []container.PortSummary{
		{PrivatePort: 3000, PublicPort: 0, Type: "tcp"},
	}
	result := formatPorts(ports)
	if result != "-" {
		t.Errorf("expected '-' for no public ports, got %q", result)
	}
}
