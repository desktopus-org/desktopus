package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// These are set via ldflags at build time
var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("desktopus %s\n", version)
		fmt.Printf("  commit:  %s\n", commit)
		fmt.Printf("  built:   %s\n", buildTime)
	},
}
