package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"

	"github.com/desktopus-org/desktopus/internal/progress"
)

// DockerProvider implements Provider using the Docker SDK.
type DockerProvider struct {
	docker *client.Client
}

// Close releases the underlying Docker client connection.
func (d *DockerProvider) Close() error {
	return d.docker.Close()
}

// Run creates and starts a container from a desktop config.
// output receives pull progress if the image is not present locally.
func (d *DockerProvider) Run(ctx context.Context, cfg *DesktopRunConfig, opts RunOptions, output io.Writer) (string, error) {
	if err := d.pullIfMissing(ctx, cfg.ImageTag, output); err != nil {
		return "", err
	}

	gpuType := cfg.GPU
	if opts.GPUType != "" {
		gpuType = opts.GPUType
	}

	if err := d.checkGPUCompatibility(ctx, cfg.ImageTag, gpuType); err != nil {
		return "", err
	}

	containerName := cfg.Name
	if opts.Name != "" {
		containerName = opts.Name
	}

	envMap := buildEnvMap(cfg, opts)
	if gpuType != "" {
		// Enable Wayland compositor with zero-copy GPU encoding.
		// DRINODE/DRI_NODE select the render device for EGL and VAAPI/NVENC.
		// Users can override these via env: in desktopus.runtime.yaml.
		if _, ok := envMap["PIXELFLUX_WAYLAND"]; !ok {
			envMap["PIXELFLUX_WAYLAND"] = "true"
		}
		if _, ok := envMap["DRINODE"]; !ok {
			envMap["DRINODE"] = "/dev/dri/renderD128"
		}
		if _, ok := envMap["DRI_NODE"]; !ok {
			envMap["DRI_NODE"] = "/dev/dri/renderD128"
		}
	}
	envList := envMapToList(envMap)

	portBindings, exposedPorts, err := buildPortBindings(cfg, opts)
	if err != nil {
		return "", fmt.Errorf("configuring ports: %w", err)
	}
	binds := buildVolumes(cfg, opts)
	if cfg.PersistenceHome != "" {
		binds = append(binds, d.persistenceBind(ctx, cfg))
	}
	d.ensureNamedVolumes(ctx, binds, cfg)

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
			hostCfg.Memory = mem
		}
	}
	if cfg.CPUs > 0 {
		hostCfg.NanoCPUs = int64(cfg.CPUs) * 1e9
	}

	switch gpuType {
	case "intel", "amd":
		hostCfg.Devices = append(hostCfg.Devices, container.DeviceMapping{
			PathOnHost:        "/dev/dri",
			PathInContainer:   "/dev/dri",
			CgroupPermissions: "rwm",
		})
	case "nvidia":
		hostCfg.DeviceRequests = append(hostCfg.DeviceRequests, container.DeviceRequest{
			Driver:       "nvidia",
			Count:        -1,
			Capabilities: [][]string{{"compute", "video", "graphics", "utility"}},
		})
	}

	if cfg.Network != "" {
		hostCfg.NetworkMode = container.NetworkMode(cfg.Network)
	}

	resp, err := d.docker.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config:     containerCfg,
		HostConfig: hostCfg,
		Name:       containerName,
	})
	if err != nil {
		return "", fmt.Errorf("creating container: %w", err)
	}

	if _, err := d.docker.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("starting container: %w", err)
	}

	return resp.ID, nil
}

// Stop stops a container by name or ID.
func (d *DockerProvider) Stop(ctx context.Context, nameOrID string, timeout int) error {
	t := timeout
	_, err := d.docker.ContainerStop(ctx, nameOrID, client.ContainerStopOptions{Timeout: &t})
	return err
}

// Remove removes a container.
func (d *DockerProvider) Remove(ctx context.Context, nameOrID string, force bool) error {
	_, err := d.docker.ContainerRemove(ctx, nameOrID, client.ContainerRemoveOptions{Force: force})
	return err
}

// List returns all desktopus-managed containers.
func (d *DockerProvider) List(ctx context.Context, all bool) ([]ContainerInfo, error) {
	f := make(client.Filters)
	f.Add("label", LabelManagedBy+"=desktopus")

	result, err := d.docker.ContainerList(ctx, client.ContainerListOptions{
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

// checkGPUCompatibility inspects the image's desktopus labels and rejects
// combinations that are known to be unsupported (e.g. nvidia on alpine).
// If the image has no desktopus labels (non-desktopus image), the check is skipped.
func (d *DockerProvider) checkGPUCompatibility(ctx context.Context, image, gpuType string) error {
	if gpuType != "nvidia" {
		return nil
	}
	info, err := d.docker.ImageInspect(ctx, image)
	if err != nil || info.Config == nil {
		return nil // image not inspectable or no config — skip
	}
	if baseOS := info.Config.Labels[LabelBaseOS]; baseOS == "alpine" {
		return fmt.Errorf("gpu \"nvidia\" is not supported on alpine: musl libc is incompatible with Nvidia proprietary drivers; use ubuntu, debian, fedora, arch, or el as the base OS")
	}
	return nil
}

// pullIfMissing pulls the image if it is not already present locally.
func (d *DockerProvider) pullIfMissing(ctx context.Context, image string, output io.Writer) error {
	_, err := d.docker.ImageInspect(ctx, image)
	if err == nil {
		return nil
	}

	fmt.Fprintf(output, "Pulling %s...\n", image)
	rc, err := d.docker.ImagePull(ctx, image, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("pulling image %s: %w", image, err)
	}
	defer func() { _ = rc.Close() }()

	return streamPullOutput(rc, output)
}

func streamPullOutput(reader io.Reader, output io.Writer) error {
	decoder := json.NewDecoder(reader)
	pr := progress.New(output)
	for {
		var event struct {
			Status   string `json:"status"`
			Progress string `json:"progress"`
			ID       string `json:"id"`
			Error    string `json:"error"`
		}
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				pr.Flush()
				return nil
			}
			return err
		}
		if event.Error != "" {
			pr.Clear()
			return fmt.Errorf("%s", event.Error)
		}
		if event.Status == "" {
			continue
		}
		if event.ID != "" {
			pr.Update(event.ID, event.Status, event.Progress)
		} else {
			pr.Print(event.Status)
		}
	}
}

func buildEnvMap(cfg *DesktopRunConfig, opts RunOptions) map[string]string {
	env := make(map[string]string)
	for k, v := range cfg.Env {
		env[k] = v
	}
	for k, v := range opts.Env {
		env[k] = v
	}
	return env
}

func envMapToList(env map[string]string) []string {
	list := make([]string, 0, len(env))
	for k, v := range env {
		list = append(list, k+"="+v)
	}
	return list
}

func buildEnvList(cfg *DesktopRunConfig, opts RunOptions) []string {
	return envMapToList(buildEnvMap(cfg, opts))
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

// persistenceBind returns a named-volume bind string for the desktop's persistent
// data directory. It inspects the image for org.desktopus.user to determine the
// correct mount path: /config for the abc user, /home/<user> otherwise.
// The volume is named <container-name>-data and is auto-created by Docker if absent.
func (d *DockerProvider) persistenceBind(ctx context.Context, cfg *DesktopRunConfig) string {
	var userLabel string
	info, err := d.docker.ImageInspect(ctx, cfg.ImageTag)
	if err == nil && info.Config != nil {
		userLabel = info.Config.Labels[LabelUser]
	}
	return cfg.PersistenceHome + ":" + persistenceMountPath(userLabel)
}

// persistenceMountPath returns the container path to mount for persistence.
// /config is used for the abc user (linuxserver/webtop default), /home/<user> otherwise.
func persistenceMountPath(userLabel string) string {
	if userLabel == "" || userLabel == "abc" {
		return "/config"
	}
	return "/home/" + userLabel
}

// ensureNamedVolumes creates any named Docker volumes in binds that don't yet
// exist, labelling them so desktopus can track them. Bind mounts (sources
// starting with "/") are skipped. Errors are silently ignored — Docker will
// handle missing volumes at container creation time.
func (d *DockerProvider) ensureNamedVolumes(ctx context.Context, binds []string, cfg *DesktopRunConfig) {
	for _, bind := range binds {
		source := strings.SplitN(bind, ":", 2)[0]
		if strings.HasPrefix(source, "/") {
			continue // bind mount
		}
		labels := map[string]string{
			LabelManagedBy: "desktopus",
			LabelDesktop:   cfg.Name,
		}
		if source == cfg.PersistenceHome {
			labels[LabelVolumeType] = "home"
		}
		_, _ = d.docker.VolumeCreate(ctx, client.VolumeCreateOptions{
			Name:   source,
			Labels: labels,
		})
	}
}

// VolumeList returns all desktopus-managed Docker volumes.
func (d *DockerProvider) VolumeList(ctx context.Context) ([]VolumeInfo, error) {
	f := make(client.Filters)
	f.Add("label", LabelManagedBy+"=desktopus")

	result, err := d.docker.VolumeList(ctx, client.VolumeListOptions{Filters: f})
	if err != nil {
		return nil, fmt.Errorf("listing volumes: %w", err)
	}

	infos := make([]VolumeInfo, len(result.Items))
	for i, v := range result.Items {
		infos[i] = VolumeInfo{
			Name:    v.Name,
			Desktop: v.Labels[LabelDesktop],
			Type:    v.Labels[LabelVolumeType],
		}
	}
	return infos, nil
}

// VolumeRemove removes a desktopus-managed Docker volume by name.
func (d *DockerProvider) VolumeRemove(ctx context.Context, name string, force bool) error {
	_, err := d.docker.VolumeRemove(ctx, name, client.VolumeRemoveOptions{Force: force})
	return err
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
