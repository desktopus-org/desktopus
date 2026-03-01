package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/desktopus-org/desktopus/internal/config"
	"github.com/desktopus-org/desktopus/internal/runtime"
)

var (
	runFile    string
	runDetach  bool
	runGPU     bool
	runPorts   []string
	runVolumes []string
	runEnvs    []string
	runName    string
	runRemove  bool
)

var runCmd = &cobra.Command{
	Use:   "run [name]",
	Short: "Run a desktop container",
	Long:  "Create and start a container from a built desktop image.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find config file
		file := runFile
		if file == "" {
			file = "."
		}

		configPath, err := config.FindDesktopConfig(file)
		if err != nil {
			return err
		}

		cfg, err := config.LoadDesktop(configPath)
		if err != nil {
			return err
		}

		// Validate required env vars
		if err := validateRequiredEnv(cfg, runEnvs); err != nil {
			return err
		}

		// Create Docker client
		dockerClient, err := newDockerClient()
		if err != nil {
			return err
		}
		defer func() { _ = dockerClient.Close() }()

		mgr := runtime.NewManager(dockerClient)

		// Parse CLI env vars
		envMap := make(map[string]string)
		for _, e := range runEnvs {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}

		opts := runtime.RunOptions{
			Name:    runName,
			Detach:  runDetach,
			GPU:     runGPU,
			Ports:   runPorts,
			Volumes: runVolumes,
			Env:     envMap,
			Remove:  runRemove,
		}

		// Build the runtime config from the desktop config
		runCfg := toDesktopRunConfig(cfg)

		containerID, err := mgr.Run(context.Background(), runCfg, opts, os.Stdout)
		if err != nil {
			return fmt.Errorf("failed to run desktop: %w", err)
		}

		fmt.Printf("Desktop %q running (container: %s)\n", cfg.Name, containerID[:12])

		// Find the web port
		webPort := findWebPort(cfg)
		if webPort != "" {
			fmt.Printf("  Web: http://localhost:%s\n", webPort)
		}

		return nil
	},
}

func init() {
	runCmd.Flags().StringVarP(&runFile, "file", "f", "", "path to desktopus.yaml")
	runCmd.Flags().BoolVarP(&runDetach, "detach", "d", true, "run in background")
	runCmd.Flags().BoolVar(&runGPU, "gpu", false, "enable GPU passthrough")
	runCmd.Flags().StringArrayVar(&runPorts, "port", nil, "additional port mappings (host:container)")
	runCmd.Flags().StringArrayVar(&runVolumes, "volume", nil, "additional volume mounts")
	runCmd.Flags().StringArrayVar(&runEnvs, "env", nil, "set environment variables (KEY=VALUE)")
	runCmd.Flags().StringVar(&runName, "name", "", "override container name")
	runCmd.Flags().BoolVar(&runRemove, "rm", false, "remove container when stopped")
}

func validateRequiredEnv(cfg *config.DesktopConfig, cliEnvs []string) error {
	provided := make(map[string]bool)

	// From defaults
	for name, spec := range cfg.Env {
		if spec.Default != "" {
			provided[name] = true
		}
	}

	// From runtime.env
	for k := range cfg.Runtime.Env {
		provided[k] = true
	}

	// From CLI
	for _, e := range cliEnvs {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) >= 1 {
			provided[parts[0]] = true
		}
	}

	var missing []string
	for name, spec := range cfg.Env {
		if spec.Required && !provided[name] {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("required environment variables not set: %s\nUse --env KEY=VALUE to provide them", strings.Join(missing, ", "))
	}
	return nil
}

func findWebPort(cfg *config.DesktopConfig) string {
	for _, p := range cfg.Runtime.Ports {
		parts := strings.SplitN(p, ":", 2)
		if len(parts) == 2 && parts[1] == "3000" {
			return parts[0]
		}
	}
	return ""
}

// toDesktopRunConfig converts a DesktopConfig into a runtime DesktopRunConfig
func toDesktopRunConfig(cfg *config.DesktopConfig) *runtime.DesktopRunConfig {
	env := make(map[string]string)
	for name, spec := range cfg.Env {
		if spec.Default != "" {
			env[name] = spec.Default
		}
	}
	for k, v := range cfg.Runtime.Env {
		env[k] = v
	}

	return &runtime.DesktopRunConfig{
		Name:     cfg.Name,
		ImageTag: cfg.ImageTag(),
		Hostname: cfg.Runtime.Hostname,
		ShmSize:  cfg.Runtime.ShmSize,
		Ports:    cfg.Runtime.Ports,
		Volumes:  cfg.Runtime.Volumes,
		GPU:      cfg.Runtime.GPU,
		Memory:   cfg.Runtime.Memory,
		CPUs:     cfg.Runtime.CPUs,
		Restart:  cfg.Runtime.Restart,
		Network:  cfg.Runtime.Network,
		Env:      env,
	}
}
