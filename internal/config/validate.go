package config

import (
	"fmt"
	"regexp"
	"strings"
)

var validName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`)

var validOS = map[string]bool{
	"ubuntu": true,
	"debian": true,
	"fedora": true,
	"arch":   true,
	"alpine": true,
}

var validDesktop = map[string]bool{
	"xfce":    true,
	"kde":     true,
	"i3":      true,
	"mate":    true,
	"openbox": true,
	"icewm":   true,
}

// ValidateDesktop checks a DesktopConfig for errors
func ValidateDesktop(cfg *DesktopConfig) error {
	var errs []string

	if cfg.Name == "" {
		errs = append(errs, "name is required")
	} else if !validName.MatchString(cfg.Name) {
		errs = append(errs, fmt.Sprintf("name %q must be DNS-safe: lowercase alphanumeric and hyphens", cfg.Name))
	}

	if cfg.Base.OS == "" {
		errs = append(errs, "base.os is required")
	} else if !validOS[cfg.Base.OS] {
		errs = append(errs, fmt.Sprintf("base.os %q is not supported (valid: ubuntu, debian, fedora, arch, alpine)", cfg.Base.OS))
	}

	if cfg.Base.Desktop == "" {
		errs = append(errs, "base.desktop is required")
	} else if !validDesktop[cfg.Base.Desktop] {
		errs = append(errs, fmt.Sprintf("base.desktop %q is not supported (valid: xfce, kde, i3, mate, openbox, icewm)", cfg.Base.Desktop))
	}

	for i, m := range cfg.Modules {
		if m.Name == "" {
			errs = append(errs, fmt.Sprintf("modules[%d]: name is required", i))
		}
	}

	for i, pr := range cfg.PostRun {
		if pr.Name == "" {
			errs = append(errs, fmt.Sprintf("postrun[%d]: name is required", i))
		}
		if pr.Script == "" {
			errs = append(errs, fmt.Sprintf("postrun[%d]: script is required", i))
		}
		if pr.RunAs != "" && pr.RunAs != "root" && pr.RunAs != "abc" {
			errs = append(errs, fmt.Sprintf("postrun[%d]: runas must be 'root' or 'abc'", i))
		}
	}

	for i, f := range cfg.Files {
		if f.Path == "" {
			errs = append(errs, fmt.Sprintf("files[%d]: path is required", i))
		}
		if f.Content == "" {
			errs = append(errs, fmt.Sprintf("files[%d]: content is required", i))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}
