# desktop-authoring Specification

## Purpose
TBD - created by archiving change baseline-current-specs. Update Purpose after archive.
## Requirements
### Requirement: Desktop definition file
The system SHALL read a desktop definition from `desktopus.yaml`, describing the desktop's identity, base image, Linux user, modules, environment variables, startup scripts, and provisioned files.

#### Scenario: Locating the config
- **WHEN** a command needs the desktop definition and is given a directory (or none, defaulting to the current directory)
- **THEN** the system locates `desktopus.yaml` within that directory and loads it

### Requirement: Name validation
The desktop `name` SHALL be required and DNS-safe: lowercase alphanumeric and hyphens, not starting or ending with a hyphen.

#### Scenario: Invalid name rejected
- **WHEN** `name` is empty, contains characters outside `[a-z0-9-]`, or starts/ends with a hyphen
- **THEN** validation fails with an error identifying the name

### Requirement: Base OS and desktop compatibility
`base.os` and `base.desktop` SHALL both be required, and the pair SHALL be validated against the supported compatibility matrix:
- `ubuntu`, `debian`, `fedora`, `arch`, `alpine` → `i3`, `kde`, `mate`, `xfce`
- `el` → `i3`, `mate`, `xfce` (no `kde`)

#### Scenario: Unsupported OS
- **WHEN** `base.os` is not one of `ubuntu`, `debian`, `fedora`, `el`, `arch`, `alpine`
- **THEN** validation fails, listing the supported OS values

#### Scenario: Incompatible desktop for OS
- **WHEN** `base.desktop` is not available for the chosen `base.os` (e.g. `kde` on `el`)
- **THEN** validation fails, listing the valid desktops for that OS

### Requirement: User resolution
The Linux `user` SHALL default to `desktopus` when unset, MUST NOT be `root`, and MUST be a valid Linux username (matching `[a-z_][a-z0-9_-]*`, at most 32 characters). The special value `abc` selects the built-in linuxserver/webtop user.

#### Scenario: Root rejected
- **WHEN** `user` is `root`
- **THEN** validation fails

#### Scenario: abc user
- **WHEN** `user` is `abc`
- **THEN** the effective home directory resolves to `/config`

### Requirement: Home directory resolution
The effective home directory SHALL be `/config` when the user is `abc`, the explicit `home` value when set (which MUST be an absolute path), otherwise `/home/<effective-user>`.

#### Scenario: Relative home rejected
- **WHEN** `home` is set to a non-absolute path
- **THEN** validation fails

### Requirement: Module references
Each entry in `modules` SHALL accept either a string (the module name) or an object with `name` and optional `vars`. `name` is required.

#### Scenario: String shorthand
- **WHEN** a module entry is the bare string `chrome`
- **THEN** it is treated as module `chrome` with no variable overrides

### Requirement: Startup scripts and provisioned files
`postrun` scripts SHALL each require a `name` and a `script`, with `runas` restricted to `root` or the configured user (default: the configured user). `files` entries SHALL each require `path` and `content`, with an optional octal `mode` (default `0644`), and are provisioned at container startup via envsubst.

#### Scenario: Invalid runas
- **WHEN** a `postrun` entry's `runas` is neither `root` nor the configured user
- **THEN** validation fails

### Requirement: Scaffolding new desktops
The `init` command SHALL generate starter `desktopus.yaml` and `desktopus.runtime.yaml` files, defaulting OS to `ubuntu` and desktop to `xfce`, with a configurable desktop name, output directory, OS, and desktop.

#### Scenario: Default init
- **WHEN** the user runs `init` with no flags
- **THEN** a `desktopus.yaml` (ubuntu/xfce) and a `desktopus.runtime.yaml` are written

