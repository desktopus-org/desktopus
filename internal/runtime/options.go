package runtime

import "time"

// RunOptions configures container creation
type RunOptions struct {
	Name    string            // Override container name
	Detach  bool              // Run in background
	GPUType string            // GPU type override (intel|amd|nvidia)
	Ports   []string          // Additional port mappings (host:container)
	Volumes []string          // Additional volume mounts (host:container[:ro])
	Env     map[string]string // Additional env vars
	ShmSize string            // Override shm size
	Remove  bool              // Auto-remove on stop
}

// ContainerInfo represents a running or stopped desktopus container
type ContainerInfo struct {
	ID      string
	Name    string
	Desktop string // from desktopus label
	Image   string
	Status  string // running, exited, etc.
	State   string // human-readable
	Ports   string // formatted port list
	WebPort int    // resolved host port for container port 3000 (0 if not mapped)
	Created time.Time
}

// Labels used to track desktopus-managed containers and volumes
const (
	LabelManagedBy  = "org.desktopus.managed-by"
	LabelDesktop    = "org.desktopus.desktop"
	LabelBaseOS     = "org.desktopus.base-os"
	LabelUser       = "org.desktopus.user"
	LabelVolumeType = "org.desktopus.volume-type"
)

// VolumeInfo represents a desktopus-managed Docker volume
type VolumeInfo struct {
	Name    string
	Desktop string // from org.desktopus.desktop label
	Type    string // from org.desktopus.volume-type label (e.g. "home")
}
