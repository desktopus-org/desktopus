//go:build integration

package modules_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"

	"github.com/desktopus-org/desktopus/internal/build"
	"github.com/desktopus-org/desktopus/internal/config"
	"github.com/desktopus-org/desktopus/internal/module"
	"github.com/desktopus-org/desktopus/modules"
)

func TestBuildModule(t *testing.T) {
	ctx := context.Background()
	buildLog := os.Getenv("BUILD_LOG") != ""

	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Skipf("cannot create Docker client: %v", err)
	}

	if _, err := docker.Ping(ctx, client.PingOptions{}); err != nil {
		t.Skipf("Docker daemon unreachable: %v", err)
	}

	reg, err := module.NewRegistry(modules.BuiltinFS)
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}

	pipeline := build.NewPipeline(docker, reg)

	for _, mod := range reg.ListBuiltin() {
		for _, os := range mod.Compatibility.OS {
			for _, desktop := range config.SupportedDesktopsForOS(os) {
				if !mod.IsCompatible(os, desktop) {
					continue
				}

				name := fmt.Sprintf("%s/%s/%s", mod.Name, os, desktop)
				mod, os, desktop := mod, os, desktop

				t.Run(name, func(t *testing.T) {
					imageTag := fmt.Sprintf("desktopus/test-%s-%s-%s:integration", mod.Name, os, desktop)

					cfg := &config.ImageConfig{
						Name: fmt.Sprintf("test-%s-%s-%s", mod.Name, os, desktop),
						Base: config.BaseSpec{OS: os, Desktop: desktop},
						Modules: []config.ModuleRef{
							{Name: mod.Name},
						},
					}

					t.Cleanup(func() {
						_, _ = docker.ImageRemove(context.Background(), imageTag, client.ImageRemoveOptions{Force: true})
					})

					var output bytes.Buffer
					opts := build.Options{Tag: imageTag}

					if err := pipeline.Build(ctx, cfg, ".", opts, &output); err != nil {
						t.Logf("Build output:\n%s", output.String())
						t.Fatalf("Build failed: %v", err)
					}
					if buildLog {
						t.Logf("Build output:\n%s", output.String())
					}

					for i, cmd := range mod.SmokeCmds(os) {
						i, cmd := i, cmd
						t.Run(fmt.Sprintf("smoke/%d", i), func(t *testing.T) {
							runSmokeTest(t, ctx, docker, imageTag, cmd)
						})
					}
				})
			}
		}
	}
}

// runSmokeTest runs a one-shot container overriding the entrypoint with cmd,
// waits for it to exit, and fails the test if the exit code is non-zero.
func runSmokeTest(t *testing.T, ctx context.Context, docker *client.Client, imageTag string, cmd []string) {
	t.Helper()

	resp, err := docker.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Image:      imageTag,
			Entrypoint: cmd[:1],
			Cmd:        cmd[1:],
		},
	})
	if err != nil {
		t.Fatalf("ContainerCreate: %v", err)
	}
	defer docker.ContainerRemove(ctx, resp.ID, client.ContainerRemoveOptions{Force: true}) //nolint:errcheck

	if _, err := docker.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		t.Fatalf("ContainerStart: %v", err)
	}

	result := docker.ContainerWait(ctx, resp.ID, client.ContainerWaitOptions{Condition: container.WaitConditionNotRunning})
	select {
	case err := <-result.Error:
		t.Fatalf("ContainerWait: %v", err)
	case status := <-result.Result:
		if status.StatusCode != 0 {
			logsResult, _ := docker.ContainerLogs(ctx, resp.ID, client.ContainerLogsOptions{
				ShowStdout: true,
				ShowStderr: true,
			})
			var buf bytes.Buffer
			if logsResult != nil {
				_, _ = io.Copy(&buf, logsResult)
			}
			t.Fatalf("smoke test exited %d:\n%s", status.StatusCode, buf.String())
		}
	}
}
