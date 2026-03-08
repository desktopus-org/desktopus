package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/desktopus-org/desktopus/internal/runtime"
)

var volumeForce bool

var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Manage desktopus volumes",
}

var volumeLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List desktopus-managed volumes",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := runtime.NewProvider("")
		if err != nil {
			return err
		}
		defer func() { _ = provider.Close() }()

		mgr := runtime.NewManager(provider)
		volumes, err := mgr.VolumeList(context.Background())
		if err != nil {
			return err
		}

		if len(volumes) == 0 {
			fmt.Println("No desktopus volumes found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tDESKTOP\tTYPE")
		for _, v := range volumes {
			t := v.Type
			if t == "" {
				t = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", v.Name, v.Desktop, t)
		}
		return w.Flush()
	},
}

var volumeRmCmd = &cobra.Command{
	Use:   "rm <name...>",
	Short: "Remove desktopus-managed volume(s)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, err := runtime.NewProvider("")
		if err != nil {
			return err
		}
		defer func() { _ = provider.Close() }()

		mgr := runtime.NewManager(provider)
		ctx := context.Background()

		for _, name := range args {
			if err := mgr.VolumeRemove(ctx, name, volumeForce); err != nil {
				fmt.Printf("Warning: failed to remove %s: %v\n", name, err)
			} else {
				fmt.Printf("Removed volume %s\n", name)
			}
		}
		return nil
	},
}

func init() {
	volumeRmCmd.Flags().BoolVar(&volumeForce, "force", false, "force removal even if volume is in use")
	volumeCmd.AddCommand(volumeLsCmd)
	volumeCmd.AddCommand(volumeRmCmd)
}
