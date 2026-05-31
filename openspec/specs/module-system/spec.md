# module-system Specification

## Purpose
TBD - created by archiving change baseline-current-specs. Update Purpose after archive.
## Requirements
### Requirement: Module resolution by name or path
The system SHALL resolve a module reference as a built-in module when it is a bare name, and as a custom filesystem module when it is a path (starting with `./`, `../`, or `/`).

#### Scenario: Built-in lookup
- **WHEN** a module is referenced by the bare name `chrome`
- **THEN** the system loads the built-in `chrome` module from the embedded module set

#### Scenario: Custom path lookup
- **WHEN** a module is referenced by a path such as `./my-module`
- **THEN** the system loads the module from that filesystem location

### Requirement: Module metadata
Each module SHALL provide a `module.yaml` declaring at least its `name`, and optionally description, version, author, tags, compatibility (OS / desktop / architecture lists), variables with defaults, dependencies, system packages, and tests.

#### Scenario: Variable defaults and overrides
- **WHEN** a module declares a variable with a default and the desktop config overrides it via `vars`
- **THEN** the override value is used during provisioning; otherwise the default is used

### Requirement: Module compatibility enforcement
When a module declares OS or desktop compatibility constraints, the build SHALL verify the target OS and desktop are supported before provisioning. Empty constraints mean the module supports all targets. An `architecture` constraint MAY be declared in `module.yaml` but is NOT currently enforced at resolution time.

#### Scenario: Incompatible OS
- **WHEN** a module declares it supports only `ubuntu` and the build targets `alpine`
- **THEN** module resolution fails with a compatibility error

#### Scenario: Architecture not enforced
- **WHEN** a module declares amd64-only compatibility and the build targets a different architecture
- **THEN** resolution does not fail on that basis (the constraint is recorded but unenforced; provisioning may still fail later)

### Requirement: Module task layout
A module SHALL provide its Ansible tasks either as `tasks/main.yml` or as per-OS files `tasks/<os>.yml` for any of `alpine`, `arch`, `debian`, `el`, `fedora`, `ubuntu`. `tasks/main.yml` is required only when the module declares no OS compatibility list; when an OS list is declared, per-OS task files alone are sufficient. When both `tasks/main.yml` and a matching `tasks/<os>.yml` exist, the per-OS file takes precedence. Optional `handlers/`, `files/`, and `templates/` directories MAY be present.

#### Scenario: OS-specific tasks take precedence
- **WHEN** a module has both `tasks/main.yml` and `tasks/alpine.yml` and the build targets alpine
- **THEN** the alpine-specific tasks are used for that build

#### Scenario: Per-OS tasks without main.yml
- **WHEN** a module declares an OS compatibility list and ships only per-OS task files (no `tasks/main.yml`)
- **THEN** the module loads successfully and the matching per-OS task file is used

### Requirement: Built-in modules
The system SHALL ship a set of built-in modules embedded in the binary, available by name. The current built-in set is `chrome` (Google Chrome; declares compatibility with `ubuntu`, `debian`, `fedora`, `el`, `arch`, and architecture `amd64`, and ships per-OS task files with no `tasks/main.yml`).

#### Scenario: Chrome module
- **WHEN** the desktop config lists module `chrome` and the target is one of its supported OSes
- **THEN** Chrome is installed during the image build

