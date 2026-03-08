package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/desktopus-org/desktopus/internal/config"
	"github.com/desktopus-org/desktopus/internal/runtime"
	"github.com/desktopus-org/desktopus/internal/viewer"
)

var (
	runFile     string
	runDetach   bool
	runGPUType  string
	runPorts    []string
	runVolumes  []string
	runEnvs     []string
	runName     string
	runRemove   bool
	runNoClient bool
)

var runCmd = &cobra.Command{
	Use:   "run [image]",
	Short: "Run a desktop container",
	Long:  "Create and start a container from a built desktop image.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var imageOverride string
		if len(args) > 0 {
			imageOverride = args[0]
		}

		// Find desktopus.runtime.yaml
		runtimePath, err := findRuntimeYAML(runFile)
		if err != nil {
			return err
		}

		rt, err := config.LoadRuntime(runtimePath)
		if err != nil {
			return err
		}

		imageTag, err := config.ResolveImageTag(rt.Image, imageOverride)
		if err != nil {
			return err
		}

		provider, err := runtime.NewProvider(rt.Provider)
		if err != nil {
			return err
		}
		defer func() { _ = provider.Close() }()

		mgr := runtime.NewManager(provider)

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
			GPUType: runGPUType,
			Ports:   runPorts,
			Volumes: runVolumes,
			Env:     envMap,
			Remove:  runRemove,
		}

		runCfg := toDesktopRunConfig(rt, imageTag)

		ctx := context.Background()
		containerID, err := mgr.Run(ctx, runCfg, opts, os.Stdout)
		if err != nil {
			return fmt.Errorf("failed to run desktop: %w", err)
		}

		name := runCfg.Name
		fmt.Printf("Desktop %q running (container: %s)\n", name, containerID[:12])

		var webURL string
		if port, err := mgr.WebPort(ctx, containerID); err == nil && port > 0 {
			webURL = fmt.Sprintf("http://localhost:%d", port)
		}

		if webURL != "" {
			fmt.Printf("  Web: %s\n", webURL)
		}

		if !runNoClient && webURL != "" {
			waitForWeb(webURL)
			fmt.Println("  Launching viewer...")
			if err := viewer.Launch(webURL); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not launch viewer: %v\n", err)
			}
		}

		return nil
	},
}

func init() {
	runCmd.Flags().StringVarP(&runFile, "file", "f", "", "path to desktopus.runtime.yaml or its directory")
	runCmd.Flags().BoolVarP(&runDetach, "detach", "d", true, "run in background")
	runCmd.Flags().StringVar(&runGPUType, "gpu", "", "GPU type for passthrough (intel|amd|nvidia)")
	runCmd.Flags().StringArrayVar(&runPorts, "port", nil, "additional port mappings (host:container)")
	runCmd.Flags().StringArrayVar(&runVolumes, "volume", nil, "additional volume mounts")
	runCmd.Flags().StringArrayVar(&runEnvs, "env", nil, "set environment variables (KEY=VALUE)")
	runCmd.Flags().StringVar(&runName, "name", "", "override container name")
	runCmd.Flags().BoolVar(&runRemove, "rm", false, "remove container when stopped")
	runCmd.Flags().BoolVar(&runNoClient, "no-client", false, "do not launch desktopus-viewer after the container starts")
}

// findRuntimeYAML resolves the path to desktopus.runtime.yaml.
// If path is empty or a directory, it looks for desktopus.runtime.yaml inside it.
// If path is a file, it uses it directly.
func findRuntimeYAML(path string) (string, error) {
	if path == "" {
		path = "."
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("accessing path %q: %w", path, err)
	}

	if info.IsDir() {
		return filepath.Join(path, "desktopus.runtime.yaml"), nil
	}

	return path, nil
}

// containerNameFromImage derives a container name from an image tag.
// "registry.example.com/mydesk:v1" → "mydesk"
// "desktopus/mydesk:latest" → "mydesk"
func containerNameFromImage(image string) string {
	name := image
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	if i := strings.Index(name, ":"); i >= 0 {
		name = name[:i]
	}
	return name
}

// toDesktopRunConfig builds a DesktopRunConfig from the runtime config and resolved image tag.
func toDesktopRunConfig(rt *config.RuntimeConfig, imageTag string) *runtime.DesktopRunConfig {
	name := rt.Name
	if name == "" {
		name = containerNameFromImage(imageTag)
	}

	env := make(map[string]string)
	for k, v := range rt.Env {
		env[k] = v
	}

	webHTTPSPort := 0
	if rt.Web != nil {
		webHTTPSPort = rt.Web.HTTPSPort
	}

	webHTTPPort := 0
	if rt.Web != nil && rt.Web.HTTPPort > 0 {
		webHTTPPort = rt.Web.HTTPPort
	}

	return &runtime.DesktopRunConfig{
		Name:            name,
		ImageTag:        imageTag,
		Hostname:        rt.Hostname,
		ShmSize:         rt.ShmSize,
		Ports:           rt.Ports,
		Volumes:         rt.Volumes,
		GPU:             rt.GPU,
		Memory:          rt.Memory,
		CPUs:            rt.CPUs,
		Restart:         rt.Restart,
		Network:         rt.Network,
		Env:             env,
		PersistenceHome: rt.PersistenceHome,
		WebHTTPPort:     webHTTPPort,
		WebHTTPSPort:    webHTTPSPort,
	}
}

// waitForWeb polls the given URL until it returns any HTTP response or the
// 60-second timeout expires. It prints a waiting message on the first probe.
func waitForWeb(url string) {
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(60 * time.Second)
	first := true
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			return
		}
		if first {
			fmt.Printf("  Waiting for web interface to be ready...")
			first = false
		}
		fmt.Print(".")
		time.Sleep(2 * time.Second)
	}
	fmt.Println(" timeout, launching anyway")
}
