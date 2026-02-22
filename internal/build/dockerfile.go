package build

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/desktopus-org/desktopus/internal/config"
	"github.com/desktopus-org/desktopus/internal/module"
)

type dockerfileData struct {
	Name             string
	Description      string
	BaseImage        string
	SystemPackages   []string
	AnsibleVerbosity int
	PostRunScripts   bool
	RuntimeFiles     bool
}

// generateDockerfile renders the Dockerfile from the template and config
func generateDockerfile(tmpl *template.Template, cfg *config.DesktopConfig, modules []*module.Module, ansibleVerbosity int) ([]byte, error) {
	// Collect all system packages from modules
	pkgSet := make(map[string]bool)
	for _, mod := range modules {
		for _, pkg := range mod.SystemPkgs {
			pkgSet[pkg] = true
		}
	}
	pkgs := make([]string, 0, len(pkgSet))
	for pkg := range pkgSet {
		pkgs = append(pkgs, pkg)
	}

	data := dockerfileData{
		Name:             cfg.Name,
		Description:      cfg.Description,
		BaseImage:        cfg.Base.ImageRef(),
		SystemPackages:   pkgs,
		AnsibleVerbosity: ansibleVerbosity,
		PostRunScripts:   len(cfg.PostRun) > 0,
		RuntimeFiles:     len(cfg.Files) > 0,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("rendering Dockerfile: %w", err)
	}
	return buf.Bytes(), nil
}

// templateFuncs provides custom template functions
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"repeat": func(count int, s string) string {
			return strings.Repeat(s, count)
		},
	}
}
