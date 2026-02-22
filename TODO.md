# Desktopus — TODO

## Current State

**Phase 1: Foundation** is complete. The core CLI works end-to-end: `init` → `build` → `run` → `stop` → `list` with the Chrome module.

### What's done (Phase 1)

- [x] Project scaffolding (go.mod, cmd/desktopus, magefile.go, .gitignore)
- [x] Config types: DesktopConfig, AppConfig, loader, validation
- [x] CLI commands: root, init, build, run, stop, list, version
- [x] Module system: types, registry (embed.FS), loader, Chrome module
- [x] Build pipeline: Dockerfile generation, tar context, playbook/ansible.cfg generation
- [x] Runtime manager: Docker SDK container lifecycle (create, start, stop, list)
- [x] Unit tests for config, build, module, and runtime packages

## Up Next: Phase 2 — Module Ecosystem + Post-Run

- [ ] Built-in modules: vscode, firefox, claude-code
- [ ] Custom module filesystem path resolution
- [ ] Module validation (`internal/module/validate.go`)
- [ ] CLI subcommands: `module list`, `module info`, `module validate`
- [ ] CLI commands: `validate`, `logs`, `rm`
- [ ] Post-run scripts: s6 `custom-cont-init.d` generation with `runas` support
- [ ] Runtime files: envsubst templating at container startup
- [ ] Env var validation at `desktopus run`

## Phase 3: State + Runtime

- [ ] SQLite store (modernc.org/sqlite, pure Go)
- [ ] GPU passthrough (DRI device mounting, DRINODE env)
- [ ] Resource limits (memory, CPU, SHM)
- [ ] Build history tracking
- [ ] Container state reconciliation

## Phase 4: REST API

- [ ] Chi HTTP server with middleware
- [ ] All endpoints under `/api/v1/`
- [ ] SSE streaming for build/container logs
- [ ] `desktopus serve` command

## Phase 5: Web UI

- [ ] React 19 + TypeScript + Vite + Tailwind + shadcn/ui
- [ ] Full management dashboard
- [ ] Embed in Go binary via `//go:embed`

## Phase 6: Polish & Release

- [ ] Shell completion (bash, zsh, fish)
- [ ] Colored CLI output (lipgloss/tablewriter)
- [ ] goreleaser multi-platform builds
- [ ] GitHub Actions CI/CD
- [ ] Integration tests (testcontainers-go)
- [ ] README quick start guide
- [ ] v0.1.0 release
