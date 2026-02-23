//go:build integration

package modules_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/moby/moby/client"

	"github.com/desktopus-org/desktopus/internal/build"
	"github.com/desktopus-org/desktopus/internal/config"
	"github.com/desktopus-org/desktopus/internal/module"
	"github.com/desktopus-org/desktopus/modules"
)

func TestBuildModule(t *testing.T) {
	ctx := context.Background()

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

					cfg := &config.DesktopConfig{
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
						t.Log(output.String())
						t.Fatalf("Build failed: %v", err)
					}
				})
			}
		}
	}
}
