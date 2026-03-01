//go:build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var Default = Build

const (
	binaryName = "desktopus"
	buildDir   = "bin"
	ldPkg      = "github.com/desktopus-org/desktopus/internal/cli"
)

func ldflags() string {
	version := cmdOutput("git", "describe", "--tags", "--always", "--dirty")
	if version == "" {
		version = "dev"
	}
	commit := cmdOutput("git", "rev-parse", "--short", "HEAD")
	if commit == "" {
		commit = "unknown"
	}
	buildTime := time.Now().UTC().Format(time.RFC3339)

	return strings.Join([]string{
		"-s", "-w",
		fmt.Sprintf("-X %s.version=%s", ldPkg, version),
		fmt.Sprintf("-X %s.commit=%s", ldPkg, commit),
		fmt.Sprintf("-X %s.buildTime=%s", ldPkg, buildTime),
	}, " ")
}

func cmdOutput(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Build compiles the desktopus binary.
func Build() error {
	fmt.Println("Building", binaryName+"...")
	os.MkdirAll(buildDir, 0755)
	return run("go", "build", "-ldflags", ldflags(), "-o", buildDir+"/"+binaryName, "./cmd/desktopus")
}

// Dev runs desktopus in development mode (go run).
func Dev() error {
	return run("go", "run", "-ldflags", ldflags(), "./cmd/desktopus")
}

// Test runs all tests.
func Test() error {
	return run("go", "test", "./...")
}

// TestV runs all tests with verbose output.
func TestV() error {
	return run("go", "test", "-v", "./...")
}

// Lint runs golangci-lint.
func Lint() error {
	return run("golangci-lint", "run", "./...")
}

// Tidy tidies Go modules.
func Tidy() error {
	return run("go", "mod", "tidy")
}

// Integration runs all integration tests (requires Docker).
//
//	mage integration
func Integration() error {
	return run("go", "test", "-tags", "integration", "-v", "-timeout", "30m", "./modules/")
}

// IntegrationModule runs integration tests for a single module across all OS/desktop combos.
//
//	mage integrationmodule chrome
func IntegrationModule(module string) error {
	return run("go", "test", "-tags", "integration", "-v", "-timeout", "30m",
		"-run", "TestBuildModule/"+module, "./modules/")
}

// IntegrationSpecific runs an integration test for one module + OS + desktop.
//
//	mage integrationspecific chrome ubuntu xfce
func IntegrationSpecific(module, os, desktop string) error {
	return run("go", "test", "-tags", "integration", "-v", "-timeout", "30m",
		"-run", fmt.Sprintf("TestBuildModule/%s/%s/%s", module, os, desktop), "./modules/")
}

// Clean removes build artifacts.
func Clean() error {
	fmt.Println("Cleaning", buildDir+"...")
	return os.RemoveAll(buildDir)
}
