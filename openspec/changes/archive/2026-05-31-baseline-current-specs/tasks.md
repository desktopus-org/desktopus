## 1. Verify capability specs against code

- [x] 1.1 `desktop-authoring`: confirm schema, validation, and `init` defaults match `internal/config/` and `internal/cli/init.go` — confirmed exact
- [x] 1.2 `module-system`: confirm resolution, `module.yaml`, compatibility, and built-ins match `internal/module/` and `modules/` — corrected (arch not enforced; main.yml optional when OS list declared)
- [x] 1.3 `image-build`: confirm pipeline and `build` flags match `internal/build/` and `internal/cli/build.go` — confirmed exact
- [x] 1.4 `container-runtime`: confirm run/stop/list, labels, GPU, ports, and persistence match `internal/runtime/` and `internal/cli/{run,stop,list}.go` — corrected (base-os/user are image labels, not container labels)
- [x] 1.5 `volume-management`: confirm `volume ls`/`rm` match `internal/cli/volume.go` and `internal/runtime/` — confirmed exact
- [x] 1.6 `desktop-viewer`: confirm `connect` and viewer behavior match `internal/viewer/` and `viewer/` — confirmed exact
- [x] 1.7 `app-configuration`: confirm global config, flags, and `version` match `internal/config/app.go` and `internal/cli/root.go` — confirmed exact

## 2. Validate and archive

- [x] 2.1 `openspec validate baseline-current-specs --strict` passes
- [ ] 2.2 Archive the change to populate `openspec/specs/`
