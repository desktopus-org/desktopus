package config

import "fmt"

// DesktopConfig is the top-level structure for desktopus.yaml
type DesktopConfig struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	User        string            `yaml:"user,omitempty"`
	Home        string            `yaml:"home,omitempty"`
	Base        BaseSpec          `yaml:"base"`
	Modules     []ModuleRef       `yaml:"modules,omitempty"`
	Env         map[string]EnvVar `yaml:"env,omitempty"`
	PostRun     []PostRunScript   `yaml:"postrun,omitempty"`
	Files       []FileSpec        `yaml:"files,omitempty"`
	Runtime     RuntimeSpec       `yaml:"runtime,omitempty"`
}

// EffectiveUser returns the resolved Linux username for this desktop.
// If user is "abc", returns "abc" (the built-in linuxserver/webtop user).
// If user is unset, defaults to "desktopus".
// Otherwise returns the configured user.
func (d *DesktopConfig) EffectiveUser() string {
	if d.User == "" {
		return "desktopus"
	}
	return d.User
}

// EffectiveHome returns the resolved home directory for this desktop.
// If user is "abc", returns "/config" (the built-in linuxserver/webtop home).
// If home is explicitly set, returns that value.
// Otherwise returns "/home/<effective-user>".
func (d *DesktopConfig) EffectiveHome() string {
	if d.User == "abc" {
		return "/config"
	}
	if d.Home != "" {
		return d.Home
	}
	return "/home/" + d.EffectiveUser()
}

// BaseSpec defines the OS and desktop environment
type BaseSpec struct {
	OS      string `yaml:"os"`
	Desktop string `yaml:"desktop"`
	Tag     string `yaml:"tag,omitempty"`
}

// WebtopTag returns the linuxserver/webtop image tag for an OS + desktop pair.
// The alpine-xfce variant is published as "latest" (no alpine-xfce tag exists).
func WebtopTag(os, desktop string) string {
	if os == "alpine" && desktop == "xfce" {
		return "latest"
	}
	return os + "-" + desktop
}

// ImageRef returns the full Docker image reference
func (b BaseSpec) ImageRef() string {
	if b.Tag != "" {
		return "lscr.io/linuxserver/webtop:" + b.Tag
	}
	return "lscr.io/linuxserver/webtop:" + WebtopTag(b.OS, b.Desktop)
}

// ModuleRef references a module to install. Supports both string shorthand
// ("chrome") and object form ({name: "vscode", vars: {...}}).
type ModuleRef struct {
	Name string                 `yaml:"name"`
	Vars map[string]interface{} `yaml:"vars,omitempty"`
}

// UnmarshalYAML allows modules to be either a string or an object
func (m *ModuleRef) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try string first
	var name string
	if err := unmarshal(&name); err == nil {
		m.Name = name
		return nil
	}

	// Fall back to object
	type moduleRefRaw ModuleRef
	var raw moduleRefRaw
	if err := unmarshal(&raw); err != nil {
		return fmt.Errorf("module must be a string or object with 'name' field: %w", err)
	}
	m.Name = raw.Name
	m.Vars = raw.Vars
	return nil
}

// EnvVar declares an environment variable
type EnvVar struct {
	Default     string `yaml:"default,omitempty"`
	Required    bool   `yaml:"required,omitempty"`
	Description string `yaml:"description,omitempty"`
}

// PostRunScript defines a script that runs at container startup via s6
type PostRunScript struct {
	Name   string `yaml:"name"`
	RunAs  string `yaml:"runas,omitempty"` // "root" or the configured user (default: configured user)
	Script string `yaml:"script"`
}

// FileSpec defines a file provisioned at container startup via envsubst
type FileSpec struct {
	Path    string `yaml:"path"`
	Content string `yaml:"content"`
	Mode    string `yaml:"mode,omitempty"` // default "0644"
}

// RuntimeSpec defines container runtime configuration
type RuntimeSpec struct {
	Hostname string            `yaml:"hostname,omitempty"`
	ShmSize  string            `yaml:"shm_size,omitempty"`
	Ports    []string          `yaml:"ports,omitempty"`    // "host:container"
	Volumes  []string          `yaml:"volumes,omitempty"`  // "host:container[:ro]"
	GPU      bool              `yaml:"gpu,omitempty"`
	Memory   string            `yaml:"memory,omitempty"`
	CPUs     int               `yaml:"cpus,omitempty"`
	Restart  string            `yaml:"restart,omitempty"`  // no | always | unless-stopped
	Network  string            `yaml:"network,omitempty"`
	Env      map[string]string `yaml:"env,omitempty"`
}

// ImageTag returns the desktopus image tag for this desktop
func (d *DesktopConfig) ImageTag() string {
	return fmt.Sprintf("desktopus/%s:latest", d.Name)
}
