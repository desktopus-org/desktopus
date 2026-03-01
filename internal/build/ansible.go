package build

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/desktopus-org/desktopus/internal/module"
)

// moduleEntry is the template data for a single module in the playbook
type moduleEntry struct {
	Name     string
	Dir      string // directory name inside the build context modules/ dir
	TaskFile string // e.g. "tasks/main.yml" or "tasks/ubuntu.yml"
	Vars     map[string]interface{}
}

type playbookData struct {
	Modules []moduleEntry
	User    string
	Home    string
}

// generatePlaybook renders the Ansible playbook from resolved modules
func generatePlaybook(tmpl *template.Template, modules []*module.Module, varsOverrides []map[string]interface{}, targetOS, desktopusUser, desktopusHome string) ([]byte, error) {
	entries := make([]moduleEntry, len(modules))
	for i, mod := range modules {
		vars := make(map[string]interface{})
		// Set defaults from module definition
		for k, v := range mod.Vars {
			if v.Default != "" {
				vars[k] = v.Default
			}
		}
		// Apply user overrides
		if i < len(varsOverrides) && varsOverrides[i] != nil {
			for k, v := range varsOverrides[i] {
				vars[k] = v
			}
		}

		dir := mod.Name
		if mod.Builtin {
			dir = mod.Path // for built-in, Path is the directory name in embed.FS
		}

		entries[i] = moduleEntry{
			Name:     mod.Name,
			Dir:      dir,
			TaskFile: mod.TaskFile(targetOS),
			Vars:     vars,
		}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, playbookData{Modules: entries, User: desktopusUser, Home: desktopusHome}); err != nil {
		return nil, fmt.Errorf("rendering playbook: %w", err)
	}
	return buf.Bytes(), nil
}

// generateAnsibleCfg renders the ansible.cfg
func generateAnsibleCfg(tmpl *template.Template) ([]byte, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return nil, fmt.Errorf("rendering ansible.cfg: %w", err)
	}
	return buf.Bytes(), nil
}
