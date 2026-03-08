package runtime

import (
	"context"
	"io"
)

// Manager is a thin facade over a Provider.
// It exists as an extension point for future cross-cutting concerns
// (logging, metrics) and to preserve existing call sites.
type Manager struct {
	provider Provider
}

// NewManager creates a new runtime manager backed by the given provider.
func NewManager(provider Provider) *Manager {
	return &Manager{provider: provider}
}

func (m *Manager) Run(ctx context.Context, cfg *DesktopRunConfig, opts RunOptions, output io.Writer) (string, error) {
	return m.provider.Run(ctx, cfg, opts, output)
}

func (m *Manager) Stop(ctx context.Context, nameOrID string, timeout int) error {
	return m.provider.Stop(ctx, nameOrID, timeout)
}

func (m *Manager) Remove(ctx context.Context, nameOrID string, force bool) error {
	return m.provider.Remove(ctx, nameOrID, force)
}

func (m *Manager) List(ctx context.Context, all bool) ([]ContainerInfo, error) {
	return m.provider.List(ctx, all)
}

func (m *Manager) VolumeList(ctx context.Context) ([]VolumeInfo, error) {
	return m.provider.VolumeList(ctx)
}

func (m *Manager) VolumeRemove(ctx context.Context, name string, force bool) error {
	return m.provider.VolumeRemove(ctx, name, force)
}

// DesktopRunConfig is a flattened view of what the runtime needs to create a container.
// This decouples the runtime package from the config package.
type DesktopRunConfig struct {
	Name     string
	ImageTag string
	Hostname string
	ShmSize  string
	Ports    []string
	Volumes  []string
	GPU      string // intel | amd | nvidia
	Memory   string
	CPUs     int
	Restart  string
	Network  string

	// Env is the merged set of all env vars (defaults + runtime.env)
	Env map[string]string

	// PersistenceHome, when non-empty, is the name of a Docker volume to mount
	// at the user's home directory (/config for abc, /home/<user> otherwise).
	PersistenceHome string
}
