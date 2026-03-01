<div align="center">
<img src="https://raw.githubusercontent.com/desktopus-org/.github/main/design/logo_rounded.png" width="200" height="200">
</div>
<h1 align="center">Desktopus</h1>

### What is Desktopus?

- **Define and Deploy Anywhere:** Configure your Linux desktop as code and deploy it across various environments through containerization.
- **Multi-Desktop Support:** Seamlessly use multiple Linux desktops simultaneously, keeping your workflows organized and distinct.
- **Effortless Workspace Switching:** Quickly switch between different Linux workspaces without hassle.
- **Optimized Resource Utilization:** Run multiple Linux desktops concurrently, leveraging all available system resources.
- **Versatile Deployment:** Operate your Linux desktops locally, remotely, or on Kubernetes (k8s) using container technology.

> [!WARNING]
> This project is in early stages of development.

## Prerequisites

- [Go](https://go.dev/) 1.24+
- [Docker](https://docs.docker.com/get-docker/) running locally
- [Mage](https://magefile.org/) (optional, for build automation)

## Install

### From source

```bash
# Clone the repo
git clone https://github.com/desktopus-org/desktopus.git
cd desktopus

# Build with Mage
mage build

# Or build directly with Go
go build -o bin/desktopus ./cmd/desktopus
```

The binary will be at `bin/desktopus`. Add it to your `PATH` or move it somewhere convenient.

## Quick start

```bash
# 1. Scaffold a new desktop project
desktopus init my-desktop

# 2. Build the image (pulls base image on first run, takes a few minutes)
desktopus build .

# 3. Run it
desktopus run

# 4. Open http://localhost:3000 in your browser

# 5. When done
desktopus stop my-desktop
```

## Configuration

Desktopus uses a `desktopus.yaml` file to define your desktop:

```yaml
name: my-desktop
description: "Dev desktop with Chrome"

base:
  os: ubuntu          # ubuntu | debian | fedora | arch | alpine | el
  desktop: xfce       # xfce | kde | i3 | mate (availability varies by OS)

modules:
  - chrome
  - name: vscode
    vars:
      extensions: "ms-python.python,golang.go"

env:
  GIT_USER_NAME:
    required: true
    description: "Git user name"
  GIT_USER_EMAIL:
    required: true
    description: "Git email"

postrun:
  - name: configure-git
    runas: abc
    script: |
      git config --global user.name "${GIT_USER_NAME}"
      git config --global user.email "${GIT_USER_EMAIL}"

files:
  - path: /config/.config/settings.json
    content: '{"fontSize": 14}'
    mode: "0644"

runtime:
  hostname: dev-desktop
  shm_size: 2g
  ports:
    - "3000:3000"
  volumes:
    - ~/projects:/config/projects
  gpu: false
  memory: 8g
  cpus: 4
  env:
    PUID: "1000"
    PGID: "1000"
    TZ: UTC
```

## CLI Reference

```
desktopus init [name]       Scaffold a new desktopus.yaml
desktopus build [path]      Build a Docker image from config
desktopus run [name]        Run a desktop container
desktopus stop [name...]    Stop container(s)
desktopus list              List running desktopus containers
desktopus version           Print version info
```

### Build flags

```
--tag string     Override the image tag
--no-cache       Build without Docker layer cache
```

### Run flags

```
-f, --file string    Path to desktopus.yaml
-d, --detach         Run in background (default: true)
    --gpu            Enable GPU passthrough (/dev/dri)
    --port strings   Additional port mappings (host:container)
    --volume strings Additional volume mounts
    --env strings    Set environment variables (KEY=VALUE)
    --name string    Override container name
    --rm             Remove container when stopped
```

## Modules

Modules are Ansible roles that run inside the Docker build. Each module has:

```
module-name/
  module.yaml        # Metadata, vars, compatibility
  tasks/main.yml     # Ansible tasks
```

### Built-in modules

| Module | Description |
|--------|-------------|
| chrome | Google Chrome browser |

### Custom modules

Reference a local directory in your config:

```yaml
modules:
  - ./my-modules/custom-tool
```

A custom module needs a `module.yaml` and `tasks/main.yml`:

```yaml
# module.yaml
name: custom-tool
description: "My custom tool"
compatibility:
  os: [ubuntu, debian]
vars:
  version:
    default: "latest"
    description: "Version to install"
system_packages:
  - curl
```

```yaml
# tasks/main.yml
- name: Install custom tool
  ansible.builtin.shell: |
    curl -fsSL https://example.com/install.sh | bash
  args:
    creates: /usr/local/bin/custom-tool
```

## How it works

1. **Parse** — reads `desktopus.yaml` and resolves all modules
2. **Generate** — creates a Dockerfile with an Ansible provisioning layer, plus a playbook that includes each module's tasks
3. **Build** — assembles a tar build context and calls the Docker API to build the image
4. **Run** — creates and starts a container with the configured ports, volumes, env vars, and resource limits

Post-run scripts land in `/custom-cont-init.d/` (the linuxserver s6-overlay extension point) and run at container startup. Runtime files are templated with `envsubst` so environment variables get substituted at boot.

## Development

```bash
# Run tests
mage test

# Run tests with verbose output
mage testV

# Build and run in dev mode
mage dev

# Tidy modules
mage tidy

# Lint (requires golangci-lint)
mage lint

# Clean build artifacts
mage clean

# List all targets
mage -l
```

## License

TBD
