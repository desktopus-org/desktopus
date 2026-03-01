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

	SmokeTest *SmokeTest   `yaml:"smoke_test,omitempty"`
	Tests     *ModuleTests `yaml:"tests,omitempty"`
}

// Compatibility defines which OS/desktop/arch combos a module supports
type Compatibility struct {
	OS      []string `yaml:"os,omitempty"`
	Desktop []string `yaml:"desktop,omitempty"`
	Arch    []string `yaml:"arch,omitempty"`
}

// ModuleTests defines declarative contract assertions evaluated by unit tests.
type ModuleTests struct {
	RequiredVars           []string `yaml:"required_vars,omitempty"`
	RequiredSystemPackages []string `yaml:"required_system_packages,omitempty"`
	ExcludedOS             []string `yaml:"excluded_os,omitempty"`
	OSSpecificTaskFiles    bool     `yaml:"os_specific_task_files,omitempty"`
}

// SmokeTest defines commands to run inside a built image to verify the module works.
// Each entry is a separate command. Default is used for all OSes unless an OS-specific
// override is present.
type SmokeTest struct {
	Default [][]string            `yaml:"default,omitempty"`
	OS      map[string][][]string `yaml:"os,omitempty"`
}

// SmokeCmds returns the smoke test commands for the given OS, or nil if none are defined.
func (m *Module) SmokeCmds(targetOS string) [][]string {
	if m.SmokeTest == nil {
		return nil
	}
	if cmds, ok := m.SmokeTest.OS[targetOS]; ok {
		return cmds
	}
	return m.SmokeTest.Default
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
