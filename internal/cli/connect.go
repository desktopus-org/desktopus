package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/desktopus-org/desktopus/internal/config"
	"github.com/desktopus-org/desktopus/internal/runtime"
	"github.com/desktopus-org/desktopus/internal/viewer"
)

var connectFile string

var connectCmd = &cobra.Command{
	Use:   "connect [name | url]",
	Short: "Open the desktop viewer",
	Long: `Launch desktopus-viewer and connect to a running desktop.

Resolution order:
  1. Argument starts with http:// or https:// → use as URL directly
  2. Argument is a desktop name → look up the running container's web port
  3. No argument → read desktopus.runtime.yaml for the container name, then look it up`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var targetURL string

		if len(args) > 0 {
			arg := args[0]
			if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
				targetURL = arg
			} else {
				targetURL = connectByName(arg)
				if targetURL == "" {
					return fmt.Errorf("desktop %q not found or has no web port mapped", arg)
				}
			}
		} else {
			// No arg — derive container name from the runtime YAML then look it up.
			runtimePath, err := findRuntimeYAML(connectFile)
			if err != nil {
				return err
			}
			rt, err := config.LoadRuntime(runtimePath)
			if err != nil {
				return err
			}
			name := rt.Name
			if name == "" {
				return fmt.Errorf("no name set in %s; pass a desktop name or URL as an argument", runtimePath)
			}
			targetURL = connectByName(name)
			if targetURL == "" {
				return fmt.Errorf("desktop %q not found or has no web port mapped", name)
			}
		}

		fmt.Printf("Opening viewer → %s\n", targetURL)
		return viewer.Launch(targetURL)
	},
}

func init() {
	connectCmd.Flags().StringVarP(&connectFile, "file", "f", "", "path to desktopus.runtime.yaml or its directory")
}

// connectByName looks up a running container by desktop name or container name
// and returns "http://localhost:<port>", or "" if not found.
func connectByName(name string) string {
	provider, err := runtime.NewProvider("")
	if err != nil {
		return ""
	}
	defer func() { _ = provider.Close() }()

	mgr := runtime.NewManager(provider)
	containers, err := mgr.List(context.Background(), false)
	if err != nil {
		return ""
	}
	for _, c := range containers {
		if c.Desktop == name || c.Name == name {
			if c.WebPort > 0 {
				return fmt.Sprintf("http://localhost:%d", c.WebPort)
			}
			return ""
		}
	}
	return ""
}
