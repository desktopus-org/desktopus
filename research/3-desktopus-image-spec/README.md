# Desktopus Image File Specification

This document provides a detailed description of the Desktopus configuration file format to define a Desktopus image. This is used to specify the software modules to be installed, files to be created or modified, and startup scripts to be executed.

Startup scripts, file content and parameters are always interpreted when running the Desktopus image, to allow for dynamic and sensitive information to be used.

## File Structure

### Header

- **type**: Specifies the type of the configuration file. For Desktopus image files, this should be set to `desktopus-image`.
- **specVersion**: Specifies the version of the Desktopus image file specification used in the file.
- **desktopusVersion**: Specifies the version of the Desktopus version used to build the Desktopus image.
  - Example: `v0.1.0`
- **os**: Indicates the target operating system for the desktopus image.
  - Example: `ubuntu-jammy`

### Parameters

The `parameters` section defines a list of configurable parameters that can be used throughout the configuration file. Each parameter includes:

- **name**: The name of the parameter.
- **type**: The data type of the parameter (e.g., `string`, `number`).
- **default**: The default value for the parameter.
- **description**: A description of the parameter (optional).
- **required**: Indicates whether the parameter is required (default is `false`).
- **RegExp**: A regular expression to validate the parameter value (optional).

#### Example Parameters:

```yaml
envs:
  - name: USERNAME
    type: string
    default: user
  - name: PASSWORD
    type: string
    default: password
  - name: VSCODE_FONT
    type: string
    default: Fira Code
  - name: VSCODE_FONT_SIZE
    type: number
    default: 14
```

The environment variables defined in the `envs` section are resolved on startup and checked, so the container can't start if they are not defined or values are not valid.

### Modules

The `modules` section lists the software modules to be installed. Each module is specified by its name.

#### Example Modules:

```yaml
modules:
  - chrome
  - firefox
  - vscode
```

### Meta

The `meta` section contains metadata and file configuration details.

#### Files

The `files` subsection specifies files to be created or modified. Each file entry includes:

- **content**: The content to be written to the file, which can include placeholders for parameters.
- **mode**: File permissions (e.g., `0644`).
- **owner**: Owner of the file.
- **group**: Group of the file.

##### Example File Entries:

```yaml
files:
  '/home/user/.config/Code/User/settings.json':
    content: |
      {
        "editor.fontSize": ${fontSize},
        "editor.fontFamily": ${font},
      }
    mode: '0644'
    owner: ${username}
    group: ${username}

  '/home/user/.config/Code/User/keybindings.json':
    content: |
      [
        {
          "key": "ctrl+shift+alt+down",
          "command": "editor.action.copyLinesDownAction",
          "when": "editorTextFocus && !editorReadonly"
        }
      ]
    mode: '0644'
    owner: ${username}
    group: ${username}
```

#### Startup Script

The `startup_script` subsection specifies a script to be executed at startup. The script can include commands to install software or configure the environment. Executes by default as root to allow for system-wide changes.

- **content**: The script content.

##### Example Startup Script:

```yaml
startup_script:
  content: |
    #!/bin/sh

    # Install vscode extensions
    code --install-extension ms-python.python
    code --install-extension ms-vscode.cpptools
  mode: '0755'
  owner: ${username}
  group: ${username}
```

## Example Desktopus Image file

Below is a complete example of a Desktopus configuration file:

```yaml
# Desktopus file YAML spec
type: desktopus-image
specVersion: v1alpha1
os: ubuntu-jammy
envs:
  - name: USERNAME
    type: string
    default: user
  - name: PASSWORD
    type: string
    default: password
  - name: VSCODE_FONT
    type: string
    default: Fira Code
  - name: VSCODE_FONT_SIZE
    type: number
    default: 14
modules:
  - chrome
  - firefox
  - vscode
meta:
  files:
    '/home/user/.config/Code/User/settings.json':
      content: |
        {
          "editor.fontSize": ${fontSize},
          "editor.fontFamily": ${font},
        }
      mode: '0644'
      owner: ${username}
      group: ${username}
    '/home/user/.config/Code/User/keybindings.json':
      content: |
        [
          {
            "key": "ctrl+shift+alt+down",
            "command": "editor.action.copyLinesDownAction",
            "when": "editorTextFocus && !editorReadonly"
          }
        ]
      mode: '0644'
      owner: ${username}
      group: ${username}
    startup_script:
      content: |
        #!/bin/sh

        # Install vscode extensions
        code --install-extension ms-python.python
        code --install-extension ms-vscode.cpptools
```
