# volume-management Specification

## Purpose
TBD - created by archiving change baseline-current-specs. Update Purpose after archive.
## Requirements
### Requirement: List volumes
The `volume ls` command SHALL list Docker volumes managed by desktopus, showing the volume name, the owning desktop, and the volume type.

#### Scenario: Listing managed volumes
- **WHEN** `volume ls` is invoked
- **THEN** only volumes labeled `org.desktopus.managed-by=desktopus` are listed, with name, desktop, and type

### Requirement: Remove volumes
The `volume rm` command SHALL remove named desktopus-managed volumes, supporting `--force`.

#### Scenario: Remove in-use volume
- **WHEN** `volume rm` targets a volume that is in use and `--force` is not set
- **THEN** removal fails

#### Scenario: Force remove
- **WHEN** `volume rm --force` is invoked for a managed volume
- **THEN** the volume is removed

