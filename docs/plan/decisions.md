# Desktopus — Decision Log

## Technology Choices

| Decision | Choice | Rationale |
|---|---|---|
| Language | Go | Fast, single binary, strong Docker SDK, good for CLI tools |
| CLI framework | Cobra | Industry standard (used by Docker, K8s), nested commands, Viper integration |
| Config format | YAML | Human-friendly, DevOps-standard (docker-compose, k8s, Ansible) |
| Config library | Koanf | Lighter than Viper, respects YAML spec, modular design |
| API framework | Chi | Minimal, stdlib-compatible, zero external deps, idiomatic Go |
| Frontend | React 19 + TypeScript + Vite | Large ecosystem, type safety, fast builds |
| Web architecture | Embedded in Go binary (embed.FS) | Single binary deployment, no CORS issues |
| Container mgmt | Docker SDK for Go (`github.com/docker/docker/client`) | Official, full control, no docker-compose dependency |
| Build tool | Taskfile.yml | YAML syntax, cross-platform, coordinates frontend + backend builds |
| Module system | Ansible everywhere | Consistent for built-in and custom modules. Runs inside Docker build (users don't need Ansible locally) |
| Base image | `lscr.io/linuxserver/webtop:ubuntu-xfce` | Selkies-based web streaming, s6-overlay, well-maintained |
| Post-run params | Environment variables | Docker-native, simple, Ansible-readable via env lookup |
| Post-run mechanism | s6 `custom-cont-init.d` | Official linuxserver extension point |
| Runtime file templating | envsubst | Lightweight, POSIX, already in base image |
| State persistence | SQLite | Zero-dependency, single-file, local-first |
| Licensing | Open source | Fully open source, all phases |
| CLI name | `desktopus` | Full name as command |

## Architectural Decisions

### Minimal config format — no apiVersion/kind/metadata/spec nesting
Kubernetes-style `apiVersion`/`kind` is overkill when there are only two config file types and no multi-resource manifests. Flat structure with docker-compose-inspired shorthand (string ports/volumes, map-based env vars, string module references) is faster to write and easier to read. Schema versioning can be added later with a simple `version: 2` field if breaking changes are needed.

### Ansible runs inside Docker build, not on the host
Users don't need Ansible installed locally. Provisioning runs in the exact target environment, and builds are reproducible regardless of host OS.

### s6 `custom-cont-init.d` for post-run scripts
Official linuxserver extension point. Scripts run after container init but before desktop services start — correct time for user config provisioning.

### envsubst for runtime file templating
Lightweight POSIX tool already present in base image. Only supports `${VAR}` substitution, but that's sufficient for configuration files. Avoids shipping custom binaries inside the container.

### SQLite for state persistence
Single-file database, zero external dependencies. Suitable for local-first usage. Stores desktop definitions, build history, and container records.

### Single binary with embedded web UI
`desktopus serve` is one command with zero dependencies beyond Docker. Eliminates deployment complexity.

## Docker-Webtop Key Facts

- Based on baseimage-selkies (web desktop streaming via WebSocket + Canvas + VideoDecoder)
- Uses s6-overlay v3 for process supervision
- Container user: `abc`, home dir: `/config`
- Ports: 3000 (HTTP), 3001 (HTTPS)
- Env vars: PUID, PGID, TZ, DRINODE, PIXELFLUX_WAYLAND
- GPU: mount `/dev/dri`, set DRINODE env var
- Extend via `FROM lscr.io/linuxserver/webtop:...`
- NOT traditional VNC — uses modern Selkies protocol
