package viewer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const binaryName = "desktopus-viewer"

// Find locates the desktopus-viewer binary using the following search order:
//  1. $DESKTOPUS_VIEWER env var
//  2. Alongside the running desktopus binary
//  3. $PATH
//  4. ~/.desktopus/bin/desktopus-viewer
//
// Returns an empty string (no error) when none are found and the caller
// should fall back to the embedded binary.
func findExternal() string {
	if p := os.Getenv("DESKTOPUS_VIEWER"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), binaryName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	if p, err := exec.LookPath(binaryName); err == nil {
		return p
	}

	if home, err := os.UserHomeDir(); err == nil {
		candidate := filepath.Join(home, ".desktopus", "bin", binaryName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

// extract writes the embedded viewerBinary to the cache directory and returns
// its path. On subsequent calls, extraction is skipped when the cached file
// matches the embedded binary's SHA-256 hash.
func extract() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("finding home directory: %w", err)
	}

	cacheDir := filepath.Join(home, ".desktopus", "cache")
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return "", fmt.Errorf("creating cache dir: %w", err)
	}

	dst := filepath.Join(cacheDir, binaryName)
	hashFile := dst + ".sha256"

	sum := sha256.Sum256(viewerBinary)
	hash := hex.EncodeToString(sum[:])

	if stored, err := os.ReadFile(hashFile); err == nil && string(stored) == hash {
		return dst, nil // already up to date
	}

	if err := os.WriteFile(dst, viewerBinary, 0755); err != nil {
		return "", fmt.Errorf("extracting viewer: %w", err)
	}
	if err := os.WriteFile(hashFile, []byte(hash), 0644); err != nil {
		return "", fmt.Errorf("writing viewer hash: %w", err)
	}

	return dst, nil
}

// Launch starts desktopus-viewer pointing at the given URL as a detached
// process. The viewer outlives the CLI invocation.
//
// Resolution order:
//  1. External binary (env var / alongside binary / PATH / ~/.desktopus/bin)
//  2. Embedded binary extracted to ~/.desktopus/cache/desktopus-viewer
func Launch(url string) error {
	bin := findExternal()

	if bin == "" {
		if len(viewerBinary) == 0 {
			return fmt.Errorf(
				"%s not found; install it, set $DESKTOPUS_VIEWER, or rebuild desktopus with -tags embed_viewer",
				binaryName,
			)
		}
		var err error
		bin, err = extract()
		if err != nil {
			return fmt.Errorf("preparing embedded viewer: %w", err)
		}
	}

	cmd := exec.Command(bin, url)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}
