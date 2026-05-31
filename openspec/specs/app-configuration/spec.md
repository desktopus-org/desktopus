# app-configuration Specification

## Purpose
TBD - created by archiving change baseline-current-specs. Update Purpose after archive.
## Requirements
### Requirement: Global app config
The system SHALL load global configuration from `~/.desktopus/config.yaml` (overridable via `--config`), covering Docker connection, build settings, and logging. Reserved `server` and `store` sections MAY be present for future use without affecting current behavior.

#### Scenario: Default docker host
- **WHEN** no Docker host is configured
- **THEN** it defaults to `unix:///var/run/docker.sock`

### Requirement: Docker connection
The Docker endpoint SHALL be configurable via `docker.host` and used for all build and runtime operations.

#### Scenario: Custom docker host
- **WHEN** `docker.host` is set to a remote endpoint
- **THEN** build and run operate against that endpoint

### Requirement: Global CLI flags
The CLI SHALL expose global flags: `--config` (config file path), `--log-level` (`debug` | `info` | `warn` | `error`), and `--no-color`.

#### Scenario: Log level override
- **WHEN** `--log-level debug` is passed
- **THEN** debug logging is enabled regardless of the level set in the config file

### Requirement: Version reporting
The `version` command SHALL print the build version, commit hash, and build timestamp injected at compile time.

#### Scenario: Version output
- **WHEN** `version` is invoked
- **THEN** the version, commit hash, and build time are printed

