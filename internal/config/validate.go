package config

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var validName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`)
var validLinuxUsername = regexp.MustCompile(`^[a-z_][a-z0-9_-]*$`)

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

// ValidateImage checks an ImageConfig for errors
func ValidateImage(cfg *ImageConfig) error {
	var errs []string

	if cfg.Name == "" {
		errs = append(errs, "name is required")
	} else if !validName.MatchString(cfg.Name) {
		errs = append(errs, fmt.Sprintf("name %q must be DNS-safe: lowercase alphanumeric and hyphens", cfg.Name))
	}

	if cfg.User != "" {
		if cfg.User == "root" {
			errs = append(errs, "user must not be 'root'")
		} else if !validLinuxUsername.MatchString(cfg.User) {
			errs = append(errs, fmt.Sprintf("user %q is not a valid Linux username (must match [a-z_][a-z0-9_-]*)", cfg.User))
		} else if len(cfg.User) > 32 {
			errs = append(errs, fmt.Sprintf("user %q exceeds maximum length of 32 characters", cfg.User))
		}
	}

	if cfg.Home != "" && !strings.HasPrefix(cfg.Home, "/") {
		errs = append(errs, fmt.Sprintf("home %q must be an absolute path", cfg.Home))
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
		if pr.RunAs != "" && pr.RunAs != "root" && pr.RunAs != cfg.EffectiveUser() {
			errs = append(errs, fmt.Sprintf("postrun[%d]: runas must be 'root' or '%s'", i, cfg.EffectiveUser()))
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

var validRestartPolicies = []string{"no", "always", "unless-stopped", "on-failure"}
var validProviders = []string{"docker"}

// ValidateRuntime checks a RuntimeConfig for errors
func ValidateRuntime(cfg *RuntimeConfig) error {
	var errs []string

	if cfg.Restart != "" && !sliceContains(validRestartPolicies, cfg.Restart) {
		errs = append(errs, fmt.Sprintf("restart %q is not valid (valid: %s)",
			cfg.Restart, strings.Join(validRestartPolicies, ", ")))
	}

	if cfg.Provider != "" && !sliceContains(validProviders, cfg.Provider) {
		errs = append(errs, fmt.Sprintf("provider %q is not supported (valid: %s)",
			cfg.Provider, strings.Join(validProviders, ", ")))
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}
