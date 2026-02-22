# Desktopus — Architecture

## Go Module

`github.com/desktopus-org/desktopus`

## Project Structure

```
desktopus/
├── cmd/desktopus/main.go              # Entrypoint
├── internal/
│   ├── cli/                           # Cobra commands
│   │   ├── root.go                    # Root cmd, global flags, koanf loading
│   │   ├── init.go                    # desktopus init
│   │   ├── build.go                   # desktopus build
│   │   ├── run.go                     # desktopus run
│   │   ├── stop.go                    # desktopus stop
│   │   ├── rm.go                      # desktopus rm
│   │   ├── list.go                    # desktopus list
│   │   ├── logs.go                    # desktopus logs
│   │   ├── serve.go                   # desktopus serve (web UI)
│   │   ├── module.go                  # desktopus module list|info|validate
│   │   ├── validate.go               # desktopus validate
│   │   └── version.go                # desktopus version
│   ├── config/                        # Config parsing & validation
│   │   ├── app.go                     # AppConfig (~/.desktopus/config.yaml)
│   │   ├── desktop.go                 # DesktopConfig (desktopus.yaml)
│   │   ├── loader.go                  # Koanf loading logic
│   │   └── validate.go               # Validation
│   ├── build/                         # Image build pipeline
│   │   ├── pipeline.go                # Orchestrates full build
│   │   ├── dockerfile.go              # Generates Dockerfile from template
│   │   ├── context.go                 # Assembles Docker build context (tar)
│   │   ├── ansible.go                 # Generates playbook.yml + ansible.cfg
│   │   └── templates/                 # Embedded Go templates
│   │       ├── Dockerfile.tmpl
│   │       ├── ansible.cfg.tmpl
│   │       └── playbook.yml.tmpl
│   ├── runtime/                       # Container lifecycle (Docker SDK)
│   │   ├── manager.go                 # Create/Start/Stop/Remove/List/Logs
│   │   └── options.go                 # RunOptions, ContainerInfo, etc.
│   ├── module/                        # Module system
│   │   ├── types.go                   # Module, ModuleMeta structs
│   │   ├── registry.go                # Discovery + resolution (builtin + custom)
│   │   ├── loader.go                  # Load from FS or embed.FS
│   │   └── validate.go               # Module validation
│   ├── api/                           # REST API (Chi)
│   │   ├── server.go                  # HTTP server setup
│   │   ├── router.go                  # Route registration
│   │   ├── handlers/                  # desktops, containers, modules, builds, system
│   │   └── middleware/                # logging, errors
│   ├── store/                         # State persistence (SQLite)
│   │   ├── store.go                   # Interface
│   │   ├── sqlite.go                  # Implementation
│   │   └── models.go                  # Desktop, Container, Build records
│   └── postrun/                       # Post-run / s6 integration
│       ├── s6.go                      # Generate s6 init scripts
│       └── envsubst.go                # Runtime file templating
├── modules/                           # Built-in Ansible modules (embedded in binary)
│   ├── chrome/
│   │   ├── module.yaml
│   │   └── tasks/main.yml
│   ├── vscode/
│   │   ├── module.yaml
│   │   └── tasks/main.yml
│   ├── claude-code/
│   │   ├── module.yaml
│   │   └── tasks/main.yml
│   └── firefox/
│       ├── module.yaml
│       └── tasks/main.yml
├── web/                               # React frontend (embedded in binary)
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── index.html
│   └── src/
│       ├── main.tsx
│       ├── App.tsx
│       ├── api/
│       │   ├── client.ts              # Typed API client
│       │   └── types.ts               # TypeScript types
│       ├── pages/
│       │   ├── DashboardPage.tsx
│       │   ├── DesktopsListPage.tsx
│       │   ├── DesktopDetailPage.tsx
│       │   ├── ContainersPage.tsx
│       │   ├── BuildPage.tsx
│       │   ├── ModulesPage.tsx
│       │   └── SettingsPage.tsx
│       └── components/
│           ├── Layout.tsx
│           ├── DesktopCard.tsx
│           ├── ContainerStatus.tsx
│           ├── LogViewer.tsx
│           └── BuildProgress.tsx
├── docs/plan/                         # This documentation
├── Taskfile.yml
├── go.mod
├── go.sum
├── .gitignore
└── LICENSE
```

---

## Configuration Specs

### Desktop Definition (`desktopus.yaml`)

User-authored file that defines what gets built and how the container runs.
Minimal, docker-compose-inspired format — no apiVersion/kind/metadata nesting.

```yaml
# desktopus.yaml

name: my-dev-desktop                    # DNS-safe: [a-z0-9-]
description: "Dev desktop with VS Code and Chrome"

base:
  os: ubuntu                            # ubuntu | debian | fedora | arch | alpine
  desktop: xfce                         # xfce | kde | i3 | mate
  tag: ""                               # Optional: pin specific webtop image version

modules:
  - chrome                              # Simple string = built-in module, no vars
  - name: vscode                        # Object form for modules with vars
    vars:
      extensions:
        - ms-python.python
        - golang.go
  - name: ./my-modules/custom           # Relative path = custom module

env:                                    # Map of env var declarations
  GIT_USER_NAME:
    required: true
  GIT_USER_EMAIL:
    required: true
    description: "Git commit email"
  FONT_SIZE:
    default: "14"

postrun:                                # Scripts that run at container startup (s6)
  - name: configure-git
    runas: abc                          # "root" or "abc" (default: abc)
    script: |
      git config --global user.name "${GIT_USER_NAME}"
      git config --global user.email "${GIT_USER_EMAIL}"

files:                                  # Files provisioned at container startup (envsubst)
  - path: /config/.config/settings.json
    content: '{"fontSize": ${FONT_SIZE}}'
    mode: "0644"

runtime:                                # Container runtime configuration
  hostname: dev-desktop
  shm_size: 2g
  ports:                                # Docker-compose style "host:container"
    - "3000:3000"
    - "3001:3001"
  volumes:                              # Docker-compose style "host:container[:ro]"
    - ~/projects:/config/projects
    - ~/.ssh:/config/.ssh:ro
  gpu: false                            # Simple boolean (or object for advanced config)
  memory: 8g
  cpus: 4
  restart: unless-stopped               # no | always | unless-stopped
  network: ""                           # Optional Docker network name
  env:                                  # Static env vars passed at runtime
    TZ: Europe/Madrid
    PUID: "1000"
    PGID: "1000"
```

**Design notes:**
- `modules` supports both string shorthand (`- chrome`) and object form (`- name: vscode`)
- `env` is a map (not array) — key is the var name, value is options. Simpler to read and write.
- `postrun[].script` — no need for the `#!/usr/bin/with-contenv bash` shebang, desktopus adds it.
- `files` are always runtime (envsubst at startup). No `phase` field — simplifies the mental model.
- `runtime.ports` and `runtime.volumes` use docker-compose string notation.
- `runtime.gpu` is a simple boolean. Advanced GPU config (specific devices, driver node) can be added later if needed.
- `runtime.memory` and `runtime.cpus` are top-level under runtime (no `resources` nesting).
- `runtime.restart` and `runtime.network` added for common container needs.

### App Config (`~/.desktopus/config.yaml`)

Application-level configuration. Also minimal — no apiVersion/kind.

```yaml
# ~/.desktopus/config.yaml

docker:
  host: "unix:///var/run/docker.sock"

server:
  listen: "127.0.0.1"
  port: 7575

build:
  cache_dir: "~/.desktopus/cache"
  parallel: 2
  ansible_verbosity: 0

log:
  level: "info"                         # debug | info | warn | error
  format: "text"                        # text | json

store:
  path: "~/.desktopus/desktopus.db"
```

---

## CLI Commands

```
desktopus init [name]                   # Scaffold desktopus.yaml
  --os string        Base OS (default "ubuntu")
  --desktop string   Desktop environment (default "xfce")
  --dir string       Output directory (default ".")

desktopus build [path]                  # Build image from desktopus.yaml
  --tag string       Override image tag
  --no-cache         Build without Docker cache
  --progress string  Progress output: auto, plain, tty

desktopus run [name]                    # Run desktop container
  -f, --file string      Path to desktopus.yaml
  -d, --detach            Run in background
  --gpu                   Enable GPU passthrough
  --port stringArray      Additional ports (host:container)
  --volume stringArray    Additional volumes (host:container)
  --env stringArray       Set env vars (KEY=VALUE)
  --rm                    Remove on stop

desktopus stop [name...]                # Stop container(s)
  -t, --timeout int   Seconds before force kill (default 10)
  --all               Stop all desktopus containers

desktopus rm [name...]                  # Remove container(s)
  -f, --force   Force remove running containers
  --all         Remove all desktopus containers

desktopus list                          # List desktops + containers
  --images      List built images only
  --containers  List containers only
  --all         Include stopped containers
  -o, --output  Output format: table, json, yaml

desktopus logs [name]                   # Stream container logs
  -f, --follow       Follow output
  --tail string      Lines from end (default "100")
  --since string     Show logs since timestamp
  --timestamps       Show timestamps

desktopus serve                         # Start API + web UI
  --listen string   Listen address
  --port int        Listen port
  --no-open         Don't auto-open browser

desktopus module list                   # List available modules
desktopus module info [name]            # Module details
desktopus module validate [path]        # Validate custom module

desktopus validate [path]               # Validate desktopus.yaml
desktopus version                       # Version info
desktopus completion [bash|zsh|fish]    # Shell completions

# Global flags
  --config string       Config file (default "~/.desktopus/config.yaml")
  --docker-host string  Override Docker host
  --log-level string    Override log level
  --no-color            Disable colored output
```

---

## Image Build Pipeline

```
desktopus.yaml
      │
      ▼
[1. Parse & Validate Config]
      │
      ▼
[2. Resolve Modules]  ◄── Registry lookup (builtin embed.FS or filesystem path)
      │
      ▼
[3. Generate Build Context]
      ├── Dockerfile         (from Dockerfile.tmpl + config)
      ├── modules/           (resolved module directories)
      ├── playbook.yml       (generated: includes each module's tasks)
      ├── ansible.cfg        (generated)
      ├── files/             (runtime file templates from files)
      └── postrun/           (s6 init scripts from postrun)
      │
      ▼
[4. Docker SDK ImageBuild()]  ◄── Tar archive of build context
      │
      ▼
[5. Stream build output to terminal/SSE]
      │
      ▼
[6. Tag: desktopus/<name>:latest]
      │
      ▼
[7. Record build in SQLite store]
```

### Generated Dockerfile Layers

1. `FROM lscr.io/linuxserver/webtop:ubuntu-xfce`
2. Install Ansible + system packages
3. Copy modules + playbook → `ansible-playbook --connection=local`
4. Remove Ansible (reduce image size)
5. Copy post-run scripts to `/custom-cont-init.d/`
6. Copy runtime file templates + envsubst provisioner script
7. Add desktopus labels

---

## Module System

### Module Directory Structure

```
module-name/
├── module.yaml          # Required: metadata, vars, compatibility
├── tasks/
│   └── main.yml         # Required: Ansible tasks
├── files/               # Optional: static files to copy
├── templates/           # Optional: Jinja2 templates
├── handlers/            # Optional: Ansible handlers
└── defaults/            # Optional: default variable values
```

### module.yaml Spec

```yaml
# module.yaml — also minimal, no apiVersion/kind

name: chrome
description: "Google Chrome browser"
version: "1.0.0"
author: desktopus
tags: [browser, web]

compatibility:
  os: [ubuntu, debian]
  desktop: []                # Empty = all desktops
  arch: [amd64]

vars:
  chrome_version:
    default: "latest"
    description: "Chrome version to install"

dependencies: []
system_packages: [wget, gnupg]
```

### Resolution

- Name starts with `./`, `/`, or `../` → filesystem path (relative to desktopus.yaml)
- Otherwise → lookup in built-in registry (embedded via `embed.FS`)

---

## REST API

All endpoints under `/api/v1/`. Consistent response envelope: `{ data, error, meta }`.

| Method | Path | Description |
|---|---|---|
| GET | /health | Health check |
| GET | /version | Version info |
| GET | /system/info | Docker info, system stats |
| GET | /desktops | List desktops |
| POST | /desktops | Register desktop |
| GET | /desktops/{name} | Desktop details |
| PUT | /desktops/{name} | Update desktop |
| DELETE | /desktops/{name} | Remove desktop |
| POST | /desktops/{name}/build | Trigger build |
| GET | /desktops/{name}/builds | List builds |
| POST | /desktops/{name}/run | Run container |
| GET | /containers | List containers |
| GET | /containers/{id} | Container details |
| POST | /containers/{id}/start | Start |
| POST | /containers/{id}/stop | Stop |
| DELETE | /containers/{id} | Remove |
| GET | /containers/{id}/logs | Stream logs (SSE) |
| GET | /builds/{id} | Build status |
| GET | /builds/{id}/logs | Stream build logs (SSE) |
| GET | /modules | List modules |
| GET | /modules/{name} | Module details |
| GET | /images | List images |
| DELETE | /images/{id} | Remove image |

---

## Post-Run & s6 Integration

### Post-run scripts
- `postrun` entries → files in `/custom-cont-init.d/` baked into image at build time
- Numbered 50-89 for ordering: `50-desktopus-<name>.sh`
- `runas: abc` uses `s6-setuidgid abc` to drop privileges
- Desktopus auto-adds `#!/usr/bin/with-contenv bash` shebang — users just write the script body

### Runtime files
- `files` entries → templates stored at `/tmp/desktopus-runtime-files/` in image
- All files are processed at container startup by `90-desktopus-files.sh` using `envsubst`
- Supports `${VAR}` substitution from container env vars

### Environment variable flow
1. `env` defaults + `runtime.env` → merged at `desktopus run`
2. Passed as `container.Config.Env` to Docker SDK
3. Available inside container via s6 `with-contenv`
4. Required vars validated before container creation
