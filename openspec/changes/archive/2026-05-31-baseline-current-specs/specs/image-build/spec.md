## ADDED Requirements

### Requirement: Build command
The `build` command SHALL build a Docker image from a desktop definition, reading the image tag from the `image` field of `desktopus.yaml` unless overridden by `--tag`, and SHALL fail if no tag is available.

#### Scenario: No tag available
- **WHEN** neither `image` is set in `desktopus.yaml` nor `--tag` is provided
- **THEN** the build fails with an error requesting an image tag

#### Scenario: Tag override
- **WHEN** `--tag` is provided
- **THEN** it takes precedence over the `image` field in the config

### Requirement: Build pipeline
The build SHALL resolve modules, generate a Dockerfile, an Ansible playbook, and an `ansible.cfg` from embedded templates, assemble an in-memory tar build context, and build the image via the Docker SDK while streaming build output.

#### Scenario: Module provisioning
- **WHEN** the desktop lists modules
- **THEN** their tasks and variable overrides are included in the generated playbook and executed during the image build

### Requirement: Base image and user setup
The generated image SHALL be based on the resolved `lscr.io/linuxserver/webtop` tag for the OS/desktop pair (or an explicit `base.tag` override), and SHALL create the configured Linux user, rewriting webtop's default `abc` user references unless the configured user is `abc`.

#### Scenario: alpine-xfce tag
- **WHEN** the base is `alpine` + `xfce`
- **THEN** the webtop image tag resolves to `latest` (no `alpine-xfce` tag exists)

### Requirement: Startup hooks and runtime files
The build SHALL install `postrun` scripts as s6 `custom-cont-init.d` hooks and stage `files` entries for envsubst-based provisioning at container startup.

#### Scenario: Postrun included
- **WHEN** a `postrun` script is defined
- **THEN** it is added to the image as an s6 init hook that runs as the specified user

### Requirement: Image labels
Built images SHALL be labeled with `org.desktopus.*` metadata (name, description, base OS, user) for later introspection.

#### Scenario: Base OS label
- **WHEN** an image is built
- **THEN** it carries an `org.desktopus.base-os` label that is read at runtime for compatibility checks

### Requirement: Build cache control
The build SHALL support disabling the Docker layer cache via `--no-cache`.

#### Scenario: no-cache build
- **WHEN** `--no-cache` is set
- **THEN** the image is built without reusing cached layers
