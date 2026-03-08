package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/moby/moby/client"
	"github.com/spf13/cobra"

	"github.com/desktopus-org/desktopus/internal/build"
	"github.com/desktopus-org/desktopus/internal/config"
	"github.com/desktopus-org/desktopus/internal/module"
	"github.com/desktopus-org/desktopus/modules"
)

var (
	buildTag     string
	buildNoCache bool
)

var buildCmd = &cobra.Command{
	Use:   "build [path]",
	Short: "Build a desktop image",
	Long:  "Build a Docker image from a desktopus.yaml configuration file.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		// Find and load config
		configPath, err := config.FindImageConfig(path)
		if err != nil {
			return err
		}

		cfg, err := config.LoadImage(configPath)
		if err != nil {
			return err
		}

		imageTag, err := config.ResolveImageTag(cfg.Image, buildTag)
		if err != nil {
			return err
		}

		configDir := filepath.Dir(configPath)

		// Create Docker client
		dockerClient, err := newDockerClient()
		if err != nil {
			return err
		}
		defer func() { _ = dockerClient.Close() }()

		// Create module registry
		registry, err := module.NewRegistry(modules.BuiltinFS)
		if err != nil {
			return fmt.Errorf("initializing module registry: %w", err)
		}

		// Run build
		pipeline := build.NewPipeline(dockerClient, registry)
		opts := build.Options{
			Tag:              imageTag,
			NoCache:          buildNoCache,
			AnsibleVerbosity: appConfig.Build.AnsibleVerbosity,
		}

		fmt.Printf("Building image for %q...\n", cfg.Name)
		fmt.Printf("  Base: %s\n", cfg.Base.ImageRef())
		fmt.Printf("  Modules: ")
		for i, m := range cfg.Modules {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(m.Name)
		}
		fmt.Println()
		fmt.Println()

		if err := pipeline.Build(context.Background(), cfg, configDir, opts, os.Stdout); err != nil {
			return fmt.Errorf("build failed: %w", err)
		}

		fmt.Printf("\nBuild complete: %s\n", imageTag)
		return nil
	},
}

func init() {
	buildCmd.Flags().StringVarP(&buildTag, "tag", "t", "", "override image tag")
	buildCmd.Flags().BoolVar(&buildNoCache, "no-cache", false, "build without Docker cache")
}

func newDockerClient() (*client.Client, error) {
	opts := []client.Opt{
		client.FromEnv,
	}

	if appConfig != nil && appConfig.Docker.Host != "" {
		opts = append(opts, client.WithHost(appConfig.Docker.Host))
	}

	cli, err := client.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("connecting to Docker: %w", err)
	}
	return cli, nil
}
