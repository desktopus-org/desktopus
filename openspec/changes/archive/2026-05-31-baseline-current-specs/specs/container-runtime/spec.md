## ADDED Requirements

### Requirement: Runtime configuration file
Container runtime options SHALL be read from `desktopus.runtime.yaml` (located alongside the desktop definition, or via `--file`). The file is optional and defaults to empty. It MAY declare `name`, `image` override, `hostname`, `shm_size`, `ports`, `volumes`, `gpu`, `memory`, `cpus`, `restart`, `network`, `env`, `provider`, `persistence_home`, and `web` port bindings.

#### Scenario: Optional runtime config
- **WHEN** no `desktopus.runtime.yaml` exists
- **THEN** `run` proceeds with default runtime settings

### Requirement: Image resolution for run
The `run` command SHALL resolve the image tag in priority order: the CLI argument, else the `image` field of `desktopus.runtime.yaml`, else an error. (This is independent of `build`, which reads `image` only from `desktopus.yaml`.)

#### Scenario: No image for run
- **WHEN** `run` is invoked with no image argument and `desktopus.runtime.yaml` has no `image`
- **THEN** the command fails asking for an image

### Requirement: Runtime validation
The runtime config SHALL validate `restart` (`no` | `always` | `unless-stopped` | `on-failure`), `provider` (`docker` only), and `gpu` (`intel` | `amd` | `nvidia`).

#### Scenario: Invalid restart policy
- **WHEN** `restart` is not one of the allowed values
- **THEN** validation fails

### Requirement: Run command
The `run` command SHALL create and start a container from a built image, detaching by default, applying ports, volumes, env, GPU, name, and `--rm` from both the runtime config and CLI flags, and SHALL launch the viewer client unless `--no-client` is set.

#### Scenario: Detached run
- **WHEN** `run` is invoked without overriding detach
- **THEN** the container runs in the background and its ID is returned

#### Scenario: Skipping the client
- **WHEN** `--no-client` is set
- **THEN** the container starts but the viewer is not launched

### Requirement: Web port mapping
The container's web interface SHALL be exposed by mapping container port 3000 (HTTP) and optionally 3001 (HTTPS) to host ports. A port value of `0` means a random host port (HTTP) or disabled (HTTPS). The resolved host HTTP port SHALL be discoverable so the viewer can connect.

#### Scenario: Random HTTP port
- **WHEN** the HTTP host port is `0`
- **THEN** Docker assigns a random host port and the runtime resolves it for the client

### Requirement: GPU passthrough
The runtime SHALL support GPU passthrough: `intel`/`amd` mount `/dev/dri` and set the relevant rendering env vars; `nvidia` uses Docker device requests with compute/video/graphics capabilities. Incompatible OS/GPU combinations SHALL be rejected.

#### Scenario: nvidia on alpine
- **WHEN** GPU is `nvidia` and the image's base OS is `alpine`
- **THEN** the run is rejected as incompatible

### Requirement: Container labels
Created containers SHALL be labeled `org.desktopus.managed-by=desktopus` and `org.desktopus.desktop=<name>`, so desktopus can filter the containers it manages. Base OS and user are NOT set as container labels; they are recorded as image labels (`org.desktopus.base-os` / `org.desktopus.user`) at build time and read back via image inspection at runtime (e.g. for GPU compatibility checks and persistence path resolution).

#### Scenario: Listing only managed containers
- **WHEN** `list` is invoked
- **THEN** only containers labeled `org.desktopus.managed-by=desktopus` are shown

#### Scenario: Base OS resolved from image
- **WHEN** the runtime needs the base OS (e.g. to validate GPU compatibility)
- **THEN** it reads the `org.desktopus.base-os` label from the image, not from the container

### Requirement: Persistent home volume
When `persistence_home` names a Docker volume, the runtime SHALL mount it at the desktop's effective home directory so user data survives container removal, labeling it as a desktopus home volume.

#### Scenario: Persistence configured
- **WHEN** `persistence_home` is set
- **THEN** the named volume is mounted at the effective home and tagged as a desktopus home volume

### Requirement: Lifecycle commands
The system SHALL support stopping containers (`stop`, with a timeout before force-kill and `--all` to stop every managed container) and listing them (`list`/`ls`, with `--all` to include stopped containers), scoped to desktopus-managed containers.

#### Scenario: Stop all
- **WHEN** `stop --all` is invoked
- **THEN** all desktopus-managed containers are stopped

#### Scenario: List including stopped
- **WHEN** `list --all` is invoked
- **THEN** stopped managed containers are included in the output
