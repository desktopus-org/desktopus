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

		imagePath := filepath.Join(dir, "desktopus.yaml")
		if _, err := os.Stat(imagePath); err == nil {
			return fmt.Errorf("desktopus.yaml already exists in %s", dir)
		}

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory: %w", err)
		}

		if err := os.WriteFile(imagePath, []byte(generateImageYAML(name, initOS, initDesktop)), 0644); err != nil {
			return fmt.Errorf("writing desktopus.yaml: %w", err)
		}

		runtimePath := filepath.Join(dir, "desktopus.runtime.yaml")
		if err := os.WriteFile(runtimePath, []byte(generateRuntimeYAML(name)), 0644); err != nil {
			return fmt.Errorf("writing desktopus.runtime.yaml: %w", err)
		}

		fmt.Printf("Created %s\n", imagePath)
		fmt.Printf("Created %s\n", runtimePath)
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. Edit desktopus.yaml to customize your desktop\n")
		fmt.Printf("  2. Edit desktopus.runtime.yaml for machine-local settings (ports, volumes, GPU)\n")
		fmt.Printf("  3. Run: desktopus build .\n")
		fmt.Printf("  4. Run: desktopus run\n")
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initOS, "os", "ubuntu", "base OS")
	initCmd.Flags().StringVar(&initDesktop, "desktop", "xfce", "desktop environment")
	initCmd.Flags().StringVar(&initDir, "dir", "", "output directory (default: current)")
}

func generateImageYAML(name, osName, desktop string) string {
	return fmt.Sprintf(`name: %s
description: "My desktop environment"
image: %s:latest

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
`, name, name, osName, desktop)
}

func generateRuntimeYAML(name string) string {
	return fmt.Sprintf(`name: %s
shm_size: 2g
ports:
  - "3000:3000"
  - "3001:3001"
  - "8082:8082"
# volumes:
#   - ~/projects:/config/projects
env:
  PUID: "1000"
  PGID: "1000"
  TZ: UTC
`, name)
}
