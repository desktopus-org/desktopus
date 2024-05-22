<div align="center">
<img src="https://raw.githubusercontent.com/desktopus-org/.github/main/design/logo_rounded.png" width="200" height="200">
</div>
<h1 align="center">Desktopus</h1>

> [!NOTE]
> This project is in development in early stages.
> Old Proof of concept can be found at [poc-2022](https://github.com/desktopus-org/desktopus/tree/poc-2022)

### What is Desktopus?

- **Define and Deploy Anywhere:** Configure your Linux desktop as code and deploy it across various environments through containerization.
- **Multi-Desktop Support:** Seamlessly use multiple Linux desktops simultaneously, keeping your workflows organized and distinct.
- **Effortless Workspace Switching:** Quickly switch between different Linux workspaces without hassle.
- **Optimized Resource Utilization:** Run multiple Linux desktops concurrently, leveraging all available system resources.
- **Versatile Deployment:** Operate your Linux desktops locally, remotely, or on Kubernetes (k8s) using container technology.
- **Compatible with [Kasm Workspaces](https://kasmweb.com/)**

### Installing Desktopus

Work in progress 👨🏻‍💻

### Desktopus image examples

Work in progress 👨🏻‍💻

### Creating your own Desktopus image

Work in progress 👨🏻‍💻

### Manage your workspaces

Work in progress 👨🏻‍💻

### Roadmap

#### September - December 2024 v0.1.0

- [ ] Desktopus CLI to build and run desktopus workspaces.
- [ ] Desktopus Server used by the CLI to build and run desktopus workspaces.
- [ ] Installer. Only for Linux at the moment.
- [ ] Minimal desktopus yaml specifications:
    - [ ] `desktopus.yaml` specification, which can define how to build images and how to run a desktopus workspace.
    - [ ] `desktopus.profile.yaml` specification, which can define multiple ways to run a desktopus workspace. (with gpu, without gpu, etc)
    - [ ] `desktopus.vars.yaml` specification, which can define variables to be used in desktopus.yaml
- [ ] Support only Ubuntu with XFCE as a base image. (More will be added in the future)
- [ ] Modules to install software which requires custom configurations.
- [ ] Allow users to define custom modules, and add custom config files, userdata at `desktopus.yaml` level.
- [ ] Webpage with documentation and examples.

#### January - March 2025 v0.2.0

Coming soon...
