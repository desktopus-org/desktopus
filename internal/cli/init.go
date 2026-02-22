package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	initOS      string
	initDesktop string
	initDir     string
)

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Create a new desktopus.yaml",
	Long:  "Scaffold a new desktop definition file in the current or specified directory.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := "my-desktop"
		if len(args) > 0 {
			name = args[0]
		}

		dir := initDir
		if dir == "" {
			dir = "."
		}

		outputPath := filepath.Join(dir, "desktopus.yaml")
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("desktopus.yaml already exists in %s", dir)
		}

		content := generateInitYAML(name, initOS, initDesktop)

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory: %w", err)
		}

		if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing desktopus.yaml: %w", err)
		}

		fmt.Printf("Created %s\n", outputPath)
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. Edit desktopus.yaml to customize your desktop\n")
		fmt.Printf("  2. Run: desktopus build .\n")
		fmt.Printf("  3. Run: desktopus run %s\n", name)
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initOS, "os", "ubuntu", "base OS")
	initCmd.Flags().StringVar(&initDesktop, "desktop", "xfce", "desktop environment")
	initCmd.Flags().StringVar(&initDir, "dir", "", "output directory (default: current)")
}

func generateInitYAML(name, osName, desktop string) string {
	return fmt.Sprintf(`name: %s
description: "My desktop environment"

base:
  os: %s
  desktop: %s

modules:
  - chrome

# env:
#   MY_VAR:
#     default: "value"
#     description: "Example variable"

# postrun:
#   - name: setup
#     runas: abc
#     script: |
#       echo "Hello from post-run script"

runtime:
  shm_size: 2g
  ports:
    - "3000:3000"
    - "3001:3001"
  # volumes:
  #   - ~/projects:/config/projects
  env:
    PUID: "1000"
    PGID: "1000"
    TZ: UTC
`, name, osName, desktop)
}
