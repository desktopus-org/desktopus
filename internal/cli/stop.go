package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/desktopus-org/desktopus/internal/runtime"
)

var (
	stopTimeout int
	stopAll     bool
)

var stopCmd = &cobra.Command{
	Use:   "stop [name...]",
	Short: "Stop desktop container(s)",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 && !stopAll {
			return fmt.Errorf("specify container name(s) or use --all")
		}

		provider, err := runtime.NewProvider("")
		if err != nil {
			return err
		}
		defer func() { _ = provider.Close() }()

		mgr := runtime.NewManager(provider)
		ctx := context.Background()

		if stopAll {
			containers, err := mgr.List(ctx, false)
			if err != nil {
				return err
			}
			for _, c := range containers {
				if err := mgr.Stop(ctx, c.Name, stopTimeout); err != nil {
					fmt.Printf("Warning: failed to stop %s: %v\n", c.Name, err)
				} else {
					fmt.Printf("Stopped %s\n", c.Name)
				}
			}
			return nil
		}

		for _, name := range args {
			if err := mgr.Stop(ctx, name, stopTimeout); err != nil {
				fmt.Printf("Warning: failed to stop %s: %v\n", name, err)
			} else {
				fmt.Printf("Stopped %s\n", name)
			}
		}
		return nil
	},
}

func init() {
	stopCmd.Flags().IntVarP(&stopTimeout, "timeout", "t", 10, "seconds before force kill")
	stopCmd.Flags().BoolVar(&stopAll, "all", false, "stop all desktopus containers")
}
