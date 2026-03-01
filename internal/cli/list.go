package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/desktopus-org/desktopus/internal/runtime"
)

var (
	listAll bool
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List desktop containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		dockerClient, err := newDockerClient()
		if err != nil {
			return err
		}
		defer func() { _ = dockerClient.Close() }()

		mgr := runtime.NewManager(dockerClient)
		containers, err := mgr.List(context.Background(), listAll)
		if err != nil {
			return err
		}

		if len(containers) == 0 {
			fmt.Println("No desktopus containers found.")
			if !listAll {
				fmt.Println("Use --all to include stopped containers.")
			}
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "NAME\tIMAGE\tSTATUS\tPORTS")
		for _, c := range containers {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.Name, c.Image, c.State, c.Ports)
		}
		return w.Flush()
	},
}

func init() {
	listCmd.Flags().BoolVar(&listAll, "all", false, "include stopped containers")
}
