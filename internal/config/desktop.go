package config

import "fmt"

// ImageConfig is the top-level structure for desktopus.yaml
type ImageConfig struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Image       string            `yaml:"image,omitempty"`
	User        string            `yaml:"user,omitempty"`
	Home        string            `yaml:"home,omitempty"`
	Base        BaseSpec          `yaml:"base"`
	Modules     []ModuleRef       `yaml:"modules,omitempty"`
	Env         map[string]EnvVar `yaml:"env,omitempty"`
	PostRun     []PostRunScript   `yaml:"postrun,omitempty"`
	Files       []FileSpec        `yaml:"files,omitempty"`
}

// EffectiveUser returns the resolved Linux username for this desktop.
// If user is "abc", returns "abc" (the built-in linuxserver/webtop user).
// If user is unset, defaults to "desktopus".
// Otherwise returns the configured user.
func (d *ImageConfig) EffectiveUser() string {
	if d.User == "" {
		return "desktopus"
	}
	return d.User
}

// EffectiveHome returns the resolved home directory for this desktop.
// If user is "abc", returns "/config" (the built-in linuxserver/webtop home).
// If home is explicitly set, returns that value.
// Otherwise returns "/home/<effective-user>".
func (d *ImageConfig) EffectiveHome() string {
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

// WebConfig defines the host port bindings for the desktop's web interface.
// A value of 0 for a port means Docker assigns a random host port.
type WebConfig struct {
	HTTPPort  int `yaml:"http_port,omitempty"`  // host port for container port 3000; 0 = random
	HTTPSPort int `yaml:"https_port,omitempty"` // host port for container port 3001; 0 = disabled
}

// RuntimeConfig defines container runtime configuration (desktopus.runtime.yaml)
type RuntimeConfig struct {
	Name     string            `yaml:"name,omitempty"`
	Image    string            `yaml:"image,omitempty"` // overrides desktopus.yaml image for this machine
	Hostname string            `yaml:"hostname,omitempty"`
	ShmSize  string            `yaml:"shm_size,omitempty"`
	Ports    []string          `yaml:"ports,omitempty"`   // "host:container"
	Volumes  []string          `yaml:"volumes,omitempty"` // "host:container[:ro]"
	GPU      string            `yaml:"gpu,omitempty"` // intel | amd | nvidia
	Memory   string            `yaml:"memory,omitempty"`
	CPUs     int               `yaml:"cpus,omitempty"`
	Restart  string            `yaml:"restart,omitempty"` // no | always | unless-stopped
	Network  string            `yaml:"network,omitempty"`
	Env      map[string]string `yaml:"env,omitempty"`
	Provider        string            `yaml:"provider,omitempty"`         // container runtime provider (default: docker)
	PersistenceHome string            `yaml:"persistence_home,omitempty"` // named Docker volume to mount at home/config
	Web             *WebConfig        `yaml:"web,omitempty"`
}

// ResolveImageTag resolves the Docker image tag in priority order:
// override (CLI flag) > image (from config file) > error
func ResolveImageTag(image, override string) (string, error) {
	if override != "" {
		return override, nil
	}
	if image != "" {
		return image, nil
	}
	return "", fmt.Errorf("no image defined: set image in the config file, or specify it explicitly")
}
