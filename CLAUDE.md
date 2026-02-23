# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

This project uses [Mage](https://magefile.org/) as its build tool (`magefile.go`).

```bash
mage build      # Compile to bin/desktopus (default target, injects version via ldflags)
mage dev        # Run without compiling (go run)
mage test       # Run all tests
mage testv      # Run tests with verbose output
mage lint       # Run golangci-lint
mage tidy       # go mod tidy
mage clean      # Remove bin/
```

Run a single test:
```bash
go test ./internal/config/ -run TestValidateDesktop -v
```

Integration tests (require Docker):
```bash
mage integration                              # All modules × all OS × all desktops
mage integrationmodule chrome                  # One module, all OS/desktop combos
mage integrationspecific chrome alpine xfce    # One module + OS + desktop
BUILD_LOG=1 mage integrationspecific chrome alpine xfce  # With Docker build output
```

## Architecture

Desktopus is a Linux desktop-as-code CLI. Users define desktops in `desktopus.yaml`, which gets built into a Docker image (on linuxserver/webtop) and run as a container.

### Package Layout

- **`cmd/desktopus/`** — Entry point. Calls `cli.Execute()`.
- **`internal/cli/`** — Cobra commands. Each file is one command. `root.go` sets up global flags and initializes AppConfig.
- **`internal/config/`** — Two config types: `DesktopConfig` (per-project `desktopus.yaml`) and `AppConfig` (global `~/.desktopus/config.yaml`). Loader uses `gopkg.in/yaml.v3` directly (not Koanf yet).
- **`internal/build/`** — Build pipeline. `pipeline.go` orchestrates: resolve modules → generate Dockerfile/playbook/ansible.cfg from Go templates → assemble tar context in memory → call Docker SDK `ImageBuild`. Templates live in `templates/` and are embedded.
- **`internal/module/`** — Module system. `registry.go` discovers built-in modules from `modules.BuiltinFS` (embed.FS). `loader.go` parses `module.yaml`. Modules can be built-in (by name) or custom (by filesystem path starting with `./`, `../`, or `/`).
- **`internal/runtime/`** — Container lifecycle via Docker SDK. `manager.go` handles create/start/stop/remove/list. Containers are labeled with `org.desktopus.*` for filtering.
- **`modules/`** — Built-in module directory. `embed.go` exports `BuiltinFS` via `//go:embed`. Each module has `module.yaml` + `tasks/main.yml` (Ansible).

### Build Pipeline Flow

```
desktopus.yaml → resolve modules → check compatibility
  → generate Dockerfile (from template, includes ansible-playbook)
  → generate playbook.yml (includes all module tasks)
  → generate ansible.cfg
  → add post-run scripts (s6 custom-cont-init.d)
  → add runtime files (envsubst templates)
  → assemble tar context → Docker SDK ImageBuild → stream output
```

### Key Conventions

- **Module resolution**: string `"chrome"` = built-in lookup; path `"./my-module"` = filesystem lookup.
- **Container labels**: `org.desktopus.managed-by=desktopus` and `org.desktopus.desktop=<name>` used to filter desktopus-managed containers.
- **Config validation**: Uses a compatibility matrix — each OS (`ubuntu|debian|fedora|arch|alpine|el`) maps to its valid desktops (`i3|kde|mate|xfce`; EL has no KDE).
- **Version injection**: `main.go` has `var version, commit, buildTime` set via ldflags in `magefile.go`.
- **Image tagging**: `desktopus/<name>:<tag>` (default tag: `latest`).

## Plan Documents

All planning and design docs are in `docs/plan/`. See `docs/plan/README.md` for the index. The project follows a 6-phase implementation plan.

## Commit Style

Do not add `Co-Authored-By` lines or any author attribution for AI in commit messages.
