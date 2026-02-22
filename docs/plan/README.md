# Desktopus — Project Plan

Linux desktop-as-code platform. Define desktops in YAML, build as Docker images (on top of [linuxserver/docker-webtop](https://github.com/linuxserver/docker-webtop)), and manage via CLI + web UI.

## Documents

| Document | Description |
|---|---|
| [decisions.md](decisions.md) | Technology choices and rationale |
| [architecture.md](architecture.md) | Project structure, config specs, CLI commands, build pipeline, API |
| [phases.md](phases.md) | 6-phase implementation plan with deliverables and verification |

## Tech Stack

- **Backend**: Go, Cobra (CLI), Chi (API), Koanf (config), Docker SDK
- **Frontend**: React 19 + TypeScript + Vite, embedded in Go binary
- **Modules**: Ansible playbooks (built-in embedded + custom from filesystem)
- **Base image**: `lscr.io/linuxserver/webtop:ubuntu-xfce`
- **Storage**: SQLite for state persistence
- **Build tool**: Taskfile.yml
- **Post-run**: s6-overlay `custom-cont-init.d`, envsubst for runtime files

## Phase Summary

| Phase | Goal | Key Output |
|---|---|---|
| **1: Foundation** | Minimal working slice | `init` + `build` + `run` + `stop` + `list` with chrome module |
| **2: Modules + Post-Run** | Full module ecosystem | vscode/firefox/claude-code, custom modules, post-run scripts |
| **3: State + Runtime** | Persistence + advanced features | SQLite store, GPU passthrough, resource limits |
| **4: REST API** | HTTP API for web UI | Chi server, all endpoints, SSE streaming |
| **5: Web UI** | React dashboard | Full management UI embedded in binary |
| **6: Polish** | Release v0.1.0 | CI/CD, goreleaser, tests, docs |
