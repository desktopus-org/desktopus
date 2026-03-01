package config

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var validName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`)

// compatMatrix maps each supported OS to its available desktop environments.
// This reflects the actual linuxserver/webtop image tags.
var compatMatrix = map[string][]string{
	"alpine": {"i3", "kde", "mate", "xfce"},
	"arch":   {"i3", "kde", "mate", "xfce"},
	"debian": {"i3", "kde", "mate", "xfce"},
	"el":     {"i3", "mate", "xfce"},
	"fedora": {"i3", "kde", "mate", "xfce"},
	"ubuntu": {"i3", "kde", "mate", "xfce"},
}

// SupportedOSList returns all supported OS names in sorted order.
func SupportedOSList() []string {
	list := make([]string, 0, len(compatMatrix))
	for os := range compatMatrix {
		list = append(list, os)
	}
	sort.Strings(list)
	return list
}

// SupportedDesktopsForOS returns the valid desktop environments for a given OS.
func SupportedDesktopsForOS(os string) []string {
	return compatMatrix[os]
}

func sliceContains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

// ValidateDesktop checks a DesktopConfig for errors
func ValidateDesktop(cfg *DesktopConfig) error {
	var errs []string

	if cfg.Name == "" {
		errs = append(errs, "name is required")
	} else if !validName.MatchString(cfg.Name) {
		errs = append(errs, fmt.Sprintf("name %q must be DNS-safe: lowercase alphanumeric and hyphens", cfg.Name))
	}

	desktops := compatMatrix[cfg.Base.OS]

	if cfg.Base.OS == "" {
		errs = append(errs, "base.os is required")
	} else if desktops == nil {
		errs = append(errs, fmt.Sprintf("base.os %q is not supported (valid: %s)", cfg.Base.OS, strings.Join(SupportedOSList(), ", ")))
	}

	if cfg.Base.Desktop == "" {
		errs = append(errs, "base.desktop is required")
	} else if desktops != nil && !sliceContains(desktops, cfg.Base.Desktop) {
		errs = append(errs, fmt.Sprintf("base.desktop %q is not available for os %q (valid: %s)", cfg.Base.Desktop, cfg.Base.OS, strings.Join(desktops, ", ")))
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
