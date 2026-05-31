# Desktopus — Reference

Durable design rationale and external facts that aren't captured elsewhere.

> **Scope of this file.** This is *not* a roadmap or an architecture spec.
> - **What the system does** → `openspec/specs/` (source of truth, built incrementally).
> - **What work is planned / in flight** → `openspec/changes/` (`/opsx:propose`).
> - **Code structure & build commands** → `CLAUDE.md`.
>
> This file keeps only the *why* behind durable decisions and reference facts about
> the base image that would otherwise be rediscovered the hard way. The original
> 6-phase plan (`docs/plan/`) was retired on 2026-05-31 because priorities and goals
> shifted; its history lives in git.

## Design rationale (still in force)

| Decision | Choice | Why |
|---|---|---|
| Language | Go | Single static binary, strong Docker SDK, good for CLI tools |
| CLI framework | Cobra | Nested commands, industry standard (Docker, k8s) |
| Config format | YAML | DevOps-standard (compose, k8s, Ansible), human-friendly |
| Config loading | `gopkg.in/yaml.v3` directly | Simple; Koanf was considered but not adopted |
| Container mgmt | Docker SDK behind a multi-provider runtime abstraction | Full control; not locked to a single backend |
| Build tool | Mage (`magefile.go`) | Go-native build automation |
| Module system | Ansible everywhere | Same mechanism for built-in and custom modules |
| Base image | `lscr.io/linuxserver/webtop` family | Selkies web streaming + s6-overlay, well-maintained |
| Post-run params | Environment variables | Docker-native, Ansible-readable |
| Post-run mechanism | s6 `custom-cont-init.d` | Official linuxserver extension point |
| Runtime file templating | envsubst | POSIX, already present in the base image |
| State / persistence | Docker volumes | Volume-based persistence (the earlier SQLite-store idea was dropped) |
| CLI name | `desktopus` | Full name as command |

### Minimal config format — no apiVersion/kind/metadata/spec nesting
Kubernetes-style `apiVersion`/`kind` is overkill with only a couple of config file
types and no multi-resource manifests. A flat, docker-compose-inspired shape (string
ports/volumes, map-based env vars, string module references) is faster to write and
read. Schema versioning can be added later with a simple `version:` field if a
breaking change is ever needed.

### Ansible runs inside the Docker build, not on the host
Users don't need Ansible installed locally. Provisioning runs in the exact target
environment, so builds are reproducible regardless of host OS.

### s6 `custom-cont-init.d` for post-run scripts
Official linuxserver extension point. Scripts run after container init but before the
desktop services start — the correct moment for user-config provisioning. `runas: abc`
drops privileges via `s6-setuidgid abc`; desktopus auto-adds the
`#!/usr/bin/with-contenv bash` shebang.

### envsubst for runtime file templating
Lightweight POSIX tool already in the base image. Only `${VAR}` substitution, but
that's enough for config files, and it avoids shipping extra binaries in the container.

## Docker-Webtop key facts

External reference about the base image — worth keeping because it's easy to forget
and hard to find:

- Based on **baseimage-selkies** (web desktop streaming via WebSocket + Canvas + VideoDecoder)
- Uses **s6-overlay v3** for process supervision
- Container user: **`abc`**, home dir: **`/config`**
- Ports: **3000** (HTTP), **3001** (HTTPS), **8082** (Selkies WebSocket data)
- Env vars: `PUID`, `PGID`, `TZ`, `DRINODE`, `PIXELFLUX_WAYLAND`
- GPU: mount `/dev/dri`, set `DRINODE` env var
- Extend via `FROM lscr.io/linuxserver/webtop:...`
- **Not** traditional VNC — uses the modern Selkies protocol
