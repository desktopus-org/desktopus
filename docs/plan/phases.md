# Desktopus — Implementation Phases

## Phase 1: Foundation (Minimal Working Slice)

**Status**: Complete
**Goal**: `desktopus init` + `desktopus build` + `desktopus run` + `desktopus stop` + `desktopus list` working end-to-end.

### Deliverables

1. **Project scaffolding**: `go.mod`, `cmd/desktopus/main.go`, `Taskfile.yml`, `.gitignore`
2. **Config types** (`internal/config/`):
   - `desktop.go` — DesktopConfig, DesktopSpec, BaseSpec, ModuleRef, RuntimeSpec structs
   - `app.go` — AppConfig struct
   - `loader.go` — Koanf-based loading from YAML + env var overrides
   - `validate.go` — Basic validation (required fields, valid OS/desktop combos)
3. **CLI commands** (`internal/cli/`):
   - `root.go` — Root cobra command, global flags (--config, --log-level, --no-color)
   - `init.go` — Generate starter desktopus.yaml
   - `build.go` — Build image from config
   - `run.go` — Create + start container
   - `stop.go` — Stop container
   - `list.go` — List desktops/containers
   - `version.go` — Print version
4. **Module system** (`internal/module/`):
   - `types.go` — Module, ModuleMetadata, ModuleSpec structs
   - `registry.go` — Built-in module discovery from embed.FS
   - `loader.go` — Load and parse module.yaml
   - One built-in module: `modules/chrome/`
5. **Build pipeline** (`internal/build/`):
   - `pipeline.go` — Orchestrate: parse config → resolve modules → generate Dockerfile → build
   - `dockerfile.go` — Go text/template Dockerfile generation
   - `context.go` — Assemble tar build context in memory
   - `ansible.go` — Generate playbook.yml and ansible.cfg
   - `templates/Dockerfile.tmpl`, `templates/playbook.yml.tmpl`, `templates/ansible.cfg.tmpl`
6. **Runtime manager** (`internal/runtime/`):
   - `manager.go` — Docker SDK: Create, Start, Stop, List containers (filtered by desktopus label)
   - `options.go` — RunOptions, ContainerInfo types

### Verification

```bash
mkdir test-desktop && cd test-desktop
desktopus init my-desktop
# desktopus.yaml created with chrome module
desktopus build .
# Image built: desktopus/my-desktop:latest
desktopus run my-desktop
# Container running, desktop at http://localhost:3000
desktopus list
# Shows running container
desktopus stop my-desktop
```

---

## Phase 2: Module Ecosystem + Post-Run

**Status**: Not started
**Goal**: Full module system, custom modules, post-run scripts, runtime files.

### Deliverables

1. **Additional built-in modules**:
   - `modules/vscode/` — VS Code with extension support
   - `modules/firefox/` — Firefox browser
   - `modules/claude-code/` — Claude Code CLI
2. **Custom module support**: filesystem path resolution in `internal/module/loader.go`
3. **Module validation**: `internal/module/validate.go`
4. **CLI subcommands**:
   - `desktopus module list` — List available modules (built-in + discovered)
   - `desktopus module info [name]` — Show module details, vars, compatibility
   - `desktopus module validate [path]` — Validate custom module structure
   - `desktopus validate [path]` — Validate desktopus.yaml
   - `desktopus logs [name]` — Stream container logs
   - `desktopus rm [name...]` — Remove stopped containers
5. **Post-run scripts** (`internal/postrun/`):
   - `s6.go` — Generate `/custom-cont-init.d/` scripts from `spec.postrun`
   - Handle `runas` (root vs abc via `s6-setuidgid`)
   - Numbered script ordering (50-89)
6. **Runtime files**:
   - `envsubst.go` — Generate startup script that processes runtime-phase files
   - Generate `90-desktopus-files.sh` init script
7. **Env var validation** at `desktopus run` — check required vars are provided

### Verification

```bash
desktopus module list
# chrome, vscode, firefox, claude-code
desktopus build . --no-cache
# Build with custom module + post-run scripts
desktopus run my-desktop --env GIT_USER_NAME=Carlos --env GIT_USER_EMAIL=carlos@example.com
# Post-run scripts execute, runtime files templated
desktopus logs my-desktop -f
# Streaming logs
```

---

## Phase 3: State Persistence + Advanced Runtime

**Status**: Not started
**Goal**: SQLite store, GPU passthrough, resource limits, build history.

### Deliverables

1. **SQLite store** (`internal/store/`):
   - `store.go` — Store interface: SaveDesktop, ListDesktops, SaveBuild, etc.
   - `sqlite.go` — SQLite implementation using `modernc.org/sqlite` (pure Go, no CGO)
   - `models.go` — DesktopRecord, BuildRecord, ContainerRecord
   - Auto-initialize `~/.desktopus/desktopus.db` on first use
2. **GPU passthrough**:
   - DRI device mounting in runtime manager
   - DRINODE + DRI_NODE env var injection
   - `--gpu` flag on `desktopus run`
3. **Resource limits**:
   - Memory limit via Docker SDK `container.Resources.Memory`
   - CPU limit via `container.Resources.NanoCPUs`
   - SHM size via `HostConfig.ShmSize`
4. **Build history**:
   - Record each build (timestamp, status, duration, image tag, logs) in store
   - `desktopus list --images` shows build history
5. **Container tracking**:
   - Record container state changes in store
   - Reconcile store state with Docker on startup

### Verification

```bash
desktopus build .
desktopus build .  # Second build
desktopus list --images
# Shows both builds with timestamps
desktopus run my-desktop --gpu --env TZ=Europe/Madrid
# GPU devices mounted, resource limits applied
desktopus list
# Shows container with GPU badge, memory/CPU info
```

---

## Phase 4: REST API

**Status**: Not started
**Goal**: Full REST API serving all functionality over HTTP.

### Deliverables

1. **HTTP server** (`internal/api/server.go`):
   - Chi router with middleware (RequestID, Logger, Recoverer, CORS, Compress)
   - Graceful shutdown
   - Configurable listen address + port from `~/.desktopus/config.yaml`
2. **Router** (`internal/api/router.go`):
   - All routes under `/api/v1/`
   - SPA fallback: serve `index.html` for non-API, non-static routes
3. **Handlers** (`internal/api/handlers/`):
   - `desktops.go` — CRUD + build trigger + run trigger
   - `containers.go` — List, get, start, stop, remove
   - `builds.go` — Get build status, stream build logs (SSE)
   - `modules.go` — List, get module details
   - `system.go` — Health, version, Docker info
4. **SSE streaming**:
   - Build logs: `GET /api/v1/builds/{id}/logs` (text/event-stream)
   - Container logs: `GET /api/v1/containers/{id}/logs` (text/event-stream)
5. **Consistent response envelope**: `{ data, error, meta: { timestamp, request_id } }`
6. **CLI command**: `desktopus serve` — starts server, auto-opens browser

### Verification

```bash
desktopus serve
# Server started on http://127.0.0.1:7575
curl http://localhost:7575/api/v1/health
# {"data":{"status":"ok"},...}
curl http://localhost:7575/api/v1/desktops
curl -X POST http://localhost:7575/api/v1/desktops/my-desktop/build
# SSE: curl -N http://localhost:7575/api/v1/builds/{id}/logs
```

---

## Phase 5: Web UI

**Status**: Not started
**Goal**: React SPA with full desktop management.

### Deliverables

1. **Project setup** (`web/`):
   - Vite + React 19 + TypeScript
   - Tailwind CSS v4 + shadcn/ui
   - TanStack Query + TanStack Router
2. **API client** (`web/src/api/`):
   - Typed fetch wrapper matching all API endpoints
   - SSE helper for log streaming
3. **Pages**:
   - `DashboardPage` — Overview: running containers, recent builds, quick actions
   - `DesktopsListPage` — All desktop definitions with status badges
   - `DesktopDetailPage` — Config view, build history, running instances
   - `ContainersPage` — All containers with status, actions (stop/start/remove)
   - `BuildPage` — Build progress with streaming log viewer
   - `ModulesPage` — Browse modules with search/filter
   - `SettingsPage` — App config, Docker connection status
4. **Components**:
   - `Layout` — Sidebar nav + top bar with server status
   - `DesktopCard` — Name, OS/DE badge, modules, actions
   - `ContainerStatus` — Real-time status badge with uptime
   - `LogViewer` — Terminal-like SSE log viewer (xterm.js or ansi-to-html)
   - `BuildProgress` — Step indicator + streaming output
5. **Embed in Go binary**:
   - `//go:embed all:web/dist` in Go source
   - Taskfile: `build:web` runs before `build:go`
   - SPA routing: return index.html for 404s on non-API paths

### Verification

```bash
task build        # Builds web + Go
desktopus serve   # Opens browser
# Full workflow in browser: create desktop, build, run, view logs, stop
```

---

## Phase 6: Polish & Release

**Status**: Not started
**Goal**: Production readiness for v0.1.0.

### Deliverables

1. Shell completion: `desktopus completion bash|zsh|fish`
2. Colored CLI output: tables (lipgloss/tablewriter), progress bars, status badges
3. `goreleaser` config for multi-platform builds (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64)
4. GitHub Actions CI/CD: lint (golangci-lint), test, build, release
5. Integration tests using testcontainers-go
6. README with quick start guide
7. `~/.desktopus/` auto-initialization on first run
8. Error messages with actionable suggestions
