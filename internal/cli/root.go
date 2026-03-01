package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/desktopus-org/desktopus/internal/config"
)

var (
	cfgFile   string
	logLevel  string
	noColor   bool
	appConfig *config.AppConfig
)

var rootCmd = &cobra.Command{
	Use:   "desktopus",
	Short: "Linux desktop-as-code platform",
	Long:  "Define your Linux desktop as code and deploy it anywhere through containerization.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadApp(cfgFile)
		if err != nil {
			return err
		}
		appConfig = cfg

		if logLevel != "" {
			appConfig.Log.Level = logLevel
		}
		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "~/.desktopus/config.yaml", "config file path")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "", "override log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(listCmd)
}

// Execute runs the root command
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return err
	}
	return nil
}
