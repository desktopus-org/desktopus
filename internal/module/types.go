package module

// Module represents a desktopus module definition
type Module struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Version     string            `yaml:"version,omitempty"`
	Author      string            `yaml:"author,omitempty"`
	Tags        []string          `yaml:"tags,omitempty"`

	Compatibility Compatibility   `yaml:"compatibility,omitempty"`
	Vars          map[string]Var  `yaml:"vars,omitempty"`
	Dependencies  []string        `yaml:"dependencies,omitempty"`
	SystemPkgs    []string        `yaml:"system_packages,omitempty"`

	// Path is the resolved filesystem path (not serialized)
	Path    string `yaml:"-"`
	Builtin bool   `yaml:"-"`

	// OSTaskFiles maps OS names to true if tasks/<os>.yml exists
	OSTaskFiles map[string]bool `yaml:"-"`
}

// Compatibility defines which OS/desktop/arch combos a module supports
type Compatibility struct {
	OS      []string `yaml:"os,omitempty"`
	Desktop []string `yaml:"desktop,omitempty"`
	Arch    []string `yaml:"arch,omitempty"`
}

// Var defines a module variable
type Var struct {
	Default     string `yaml:"default,omitempty"`
	Description string `yaml:"description,omitempty"`
}

// TaskFile returns the task file path for the given OS.
// If an OS-specific file exists (tasks/<os>.yml), it is preferred; otherwise tasks/main.yml.
func (m *Module) TaskFile(os string) string {
	if m.OSTaskFiles[os] {
		return "tasks/" + os + ".yml"
	}
	return "tasks/main.yml"
}

// IsCompatible checks if the module supports a given OS and desktop
func (m *Module) IsCompatible(os, desktop string) bool {
	if len(m.Compatibility.OS) > 0 {
		found := false
		for _, o := range m.Compatibility.OS {
			if o == os {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(m.Compatibility.Desktop) > 0 {
		found := false
		for _, d := range m.Compatibility.Desktop {
			if d == desktop {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
