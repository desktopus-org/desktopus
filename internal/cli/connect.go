package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/desktopus-org/desktopus/internal/config"
	"github.com/desktopus-org/desktopus/internal/viewer"
)

var connectFile string

var connectCmd = &cobra.Command{
	Use:   "connect [url]",
	Short: "Open the desktop viewer",
	Long: `Launch desktopus-viewer and connect to a running desktop.

Without a URL argument, reads desktopus.runtime.yaml to find the web port.
With a URL argument, connects directly to that address.

Toggle kiosk capture:  Ctrl+Alt+M
Quit viewer:           Ctrl+Alt+Q`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var targetURL string

		if len(args) > 0 {
			targetURL = args[0]
		} else {
			runtimePath, err := findRuntimeYAML(connectFile)
			if err != nil {
				return err
			}
			rt, err := config.LoadRuntime(runtimePath)
			if err != nil {
				return err
			}
			port := findWebPort(rt)
			if port == "" {
				return fmt.Errorf("no web port mapping found in %s; pass a URL directly", runtimePath)
			}
			targetURL = "http://localhost:" + port
		}

		fmt.Printf("Opening viewer → %s\n", targetURL)
		return viewer.Launch(targetURL)
	},
}

func init() {
	connectCmd.Flags().StringVarP(&connectFile, "file", "f", "", "path to desktopus.runtime.yaml or its directory")
}
