## Why

Desktopus is adopting OpenSpec for spec-driven development, but `openspec/specs/` is empty — there is no documented source of truth for what the system already does. Before proposing new behavior, we need a baseline that captures the current contract so future changes have something to diff against.

## What Changes

This change is documentation-only. It introduces specs describing the **existing, shipped** behavior of Desktopus — no source code changes. Each capability spec records what the system guarantees today: the CLI commands, config schemas and validation, the build pipeline, module resolution, container lifecycle, volume management, and the viewer client.

## Capabilities

### New Capabilities
- `desktop-authoring`: the `desktopus.yaml` schema, validation rules, and `init` scaffolding
- `module-system`: module resolution (built-in vs filesystem path), `module.yaml` format, compatibility, and built-in modules
- `image-build`: the `build` command and the Docker image build pipeline (Dockerfile + Ansible generation)
- `container-runtime`: the `run`/`stop`/`list` lifecycle, `desktopus.runtime.yaml` input, container labels, GPU passthrough, web ports, and persistence
- `volume-management`: `volume ls` / `volume rm` over desktopus-managed Docker volumes
- `desktop-viewer`: the `connect` command and the hardened Electron viewer client
- `app-configuration`: global `~/.desktopus/config.yaml`, global CLI flags, Docker host, and `version`

### Modified Capabilities
<!-- none — this is the initial baseline -->

## Impact

- Affected: `openspec/specs/` (populated on archive). No application code changes.
- Establishes the contract that subsequent changes will modify via spec deltas.
