package runtime

import (
	"context"
	"fmt"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

// Manager handles container lifecycle via the Docker SDK
type Manager struct {
	docker *client.Client
}

// NewManager creates a new runtime manager
func NewManager(docker *client.Client) *Manager {
	return &Manager{docker: docker}
}

// Run creates and starts a container from a desktop config
func (m *Manager) Run(ctx context.Context, cfg *DesktopRunConfig, opts RunOptions) (string, error) {
	containerName := cfg.Name
	if opts.Name != "" {
		containerName = opts.Name
	}

	envList := buildEnvList(cfg, opts)
	portBindings, exposedPorts, err := buildPortBindings(cfg, opts)
	if err != nil {
		return "", fmt.Errorf("configuring ports: %w", err)
	}
	binds := buildVolumes(cfg, opts)

	shmSize := parseSize(cfg.ShmSize)
	if opts.ShmSize != "" {
		shmSize = parseSize(opts.ShmSize)
	}

	containerCfg := &container.Config{
		Image:        cfg.ImageTag,
		Hostname:     cfg.Hostname,
		Env:          envList,
		ExposedPorts: exposedPorts,
		Labels: map[string]string{
			LabelManagedBy: "desktopus",
			LabelDesktop:   cfg.Name,
		},
	}

	hostCfg := &container.HostConfig{
		PortBindings: portBindings,
		Binds:        binds,
		ShmSize:      shmSize,
		AutoRemove:   opts.Remove,
	}

	if cfg.Restart != "" && cfg.Restart != "no" {
		hostCfg.RestartPolicy = container.RestartPolicy{Name: container.RestartPolicyMode(cfg.Restart)}
	}

	if cfg.Memory != "" {
		if mem := parseSize(cfg.Memory); mem > 0 {
			hostCfg.Resources.Memory = mem
		}
	}
	if cfg.CPUs > 0 {
		hostCfg.Resources.NanoCPUs = int64(cfg.CPUs) * 1e9
	}

	if cfg.GPU || opts.GPU {
		hostCfg.Devices = append(hostCfg.Devices, container.DeviceMapping{
			PathOnHost:        "/dev/dri",
			PathInContainer:   "/dev/dri",
			CgroupPermissions: "rwm",
		})
	}

	if cfg.Network != "" {
		hostCfg.NetworkMode = container.NetworkMode(cfg.Network)
	}

	resp, err := m.docker.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config:     containerCfg,
		HostConfig: hostCfg,
		Name:       containerName,
	})
	if err != nil {
		return "", fmt.Errorf("creating container: %w", err)
	}

	if _, err := m.docker.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("starting container: %w", err)
	}

	return resp.ID, nil
}

// Stop stops a container by name or ID
func (m *Manager) Stop(ctx context.Context, nameOrID string, timeout int) error {
	t := timeout
	_, err := m.docker.ContainerStop(ctx, nameOrID, client.ContainerStopOptions{Timeout: &t})
	return err
}

// Remove removes a container
func (m *Manager) Remove(ctx context.Context, nameOrID string, force bool) error {
	_, err := m.docker.ContainerRemove(ctx, nameOrID, client.ContainerRemoveOptions{Force: force})
	return err
}

// List returns all desktopus-managed containers
func (m *Manager) List(ctx context.Context, all bool) ([]ContainerInfo, error) {
	f := make(client.Filters)
	f.Add("label", LabelManagedBy+"=desktopus")

	result, err := m.docker.ContainerList(ctx, client.ContainerListOptions{
		All:     all,
		Filters: f,
	})
	if err != nil {
		return nil, fmt.Errorf("listing containers: %w", err)
	}

	infos := make([]ContainerInfo, len(result.Items))
	for i, c := range result.Items {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		ports := formatPorts(c.Ports)

		infos[i] = ContainerInfo{
			ID:      c.ID[:12],
			Name:    name,
			Desktop: c.Labels[LabelDesktop],
			Image:   c.Image,
			Status:  string(c.State),
			State:   c.Status,
			Ports:   ports,
			Created: time.Unix(c.Created, 0),
		}
	}

	return infos, nil
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
	GPU      bool
	Memory   string
	CPUs     int
	Restart  string
	Network  string

	// Env is the merged set of all env vars (defaults + runtime.env)
	Env map[string]string
}

func buildEnvList(cfg *DesktopRunConfig, opts RunOptions) []string {
	env := make(map[string]string)
	for k, v := range cfg.Env {
		env[k] = v
	}
	for k, v := range opts.Env {
		env[k] = v
	}

	list := make([]string, 0, len(env))
	for k, v := range env {
		list = append(list, k+"="+v)
	}
	return list
}

func buildPortBindings(cfg *DesktopRunConfig, opts RunOptions) (network.PortMap, network.PortSet, error) {
	portMap := network.PortMap{}
	portSet := network.PortSet{}

	allPorts := append(cfg.Ports, opts.Ports...)
	for _, p := range allPorts {
		parts := strings.SplitN(p, ":", 2)
		if len(parts) != 2 {
			return nil, nil, fmt.Errorf("invalid port mapping %q (expected host:container)", p)
		}

		containerPort, err := network.ParsePort(parts[1] + "/tcp")
		if err != nil {
			return nil, nil, fmt.Errorf("invalid container port %q: %w", parts[1], err)
		}
		portSet[containerPort] = struct{}{}
		portMap[containerPort] = []network.PortBinding{
			{HostIP: netip.MustParseAddr("0.0.0.0"), HostPort: parts[0]},
		}
	}

	return portMap, portSet, nil
}

func buildVolumes(cfg *DesktopRunConfig, opts RunOptions) []string {
	allVolumes := append(cfg.Volumes, opts.Volumes...)

	binds := make([]string, 0, len(allVolumes))
	for _, v := range allVolumes {
		if strings.HasPrefix(v, "~/") {
			if home, err := os.UserHomeDir(); err == nil {
				v = home + v[1:]
			}
		}
		binds = append(binds, v)
	}
	return binds
}

func parseSize(s string) int64 {
	if s == "" {
		return 0
	}
	s = strings.TrimSpace(strings.ToLower(s))
	multiplier := int64(1)

	if strings.HasSuffix(s, "g") {
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "g")
	} else if strings.HasSuffix(s, "m") {
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "m")
	} else if strings.HasSuffix(s, "k") {
		multiplier = 1024
		s = strings.TrimSuffix(s, "k")
	}

	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return n * multiplier
}

func formatPorts(ports []container.PortSummary) string {
	if len(ports) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(ports))
	for _, p := range ports {
		if p.PublicPort > 0 {
			parts = append(parts, fmt.Sprintf("%s:%d->%d/%s", p.IP.String(), p.PublicPort, p.PrivatePort, p.Type))
		}
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, ", ")
}
