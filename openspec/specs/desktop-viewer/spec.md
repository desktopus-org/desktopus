# desktop-viewer Specification

## Purpose
TBD - created by archiving change baseline-current-specs. Update Purpose after archive.
## Requirements
### Requirement: Connect command
The `connect` command SHALL open the desktopus viewer against a running desktop, resolving the target in priority order: an explicit `http(s)://` URL argument, then lookup by desktop or container name, then `desktopus.runtime.yaml`.

#### Scenario: Connect by name
- **WHEN** `connect` is given a desktop or container name
- **THEN** the viewer opens against that container's resolved web URL

#### Scenario: Connect by URL
- **WHEN** `connect` is given an `http(s)://` URL
- **THEN** the viewer opens that URL directly

### Requirement: Viewer launch and resolution
The viewer binary SHALL be located in priority order: `$DESKTOPUS_VIEWER`, alongside the running desktopus binary, `$PATH`, `~/.desktopus/bin/`, then an embedded copy (when the binary is built with the embed tag).

#### Scenario: Embedded extraction
- **WHEN** no external viewer binary is found and an embedded copy exists
- **THEN** it is extracted to cache (skipped if the cached copy's SHA256 already matches) and launched

### Requirement: Keyboard capture
The viewer SHALL capture keyboard input fully while focused — locking via the Keyboard Lock API in fullscreen, suppressing its own browser shortcuts, and grabbing window-manager-level shortcuts (X11 `XGrabKeyboard` / Wayland inhibit) — and SHALL re-acquire the lock when the window regains focus.

#### Scenario: Refocus re-locks
- **WHEN** the OS releases the keyboard grab (e.g. a VT switch) and the window later regains focus
- **THEN** the viewer re-acquires the keyboard lock

### Requirement: Network and navigation hardening
The viewer SHALL restrict all HTTP and WebSocket traffic to the target host and port, allowing only the `devtools`, `chrome-extension`, `data:`, and `blob:` schemes, and SHALL prevent navigation away from the target URL.

#### Scenario: Cross-host request blocked
- **WHEN** the page attempts a request to a host other than the target host:port
- **THEN** the request is blocked

