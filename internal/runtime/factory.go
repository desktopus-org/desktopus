package runtime

import (
	"fmt"

	"github.com/moby/moby/client"
)

// NewProvider creates a Provider for the named container runtime.
// An empty name defaults to "docker".
func NewProvider(name string) (Provider, error) {
	switch name {
	case "", "docker":
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return nil, fmt.Errorf("connecting to Docker: %w", err)
		}
		return &DockerProvider{docker: cli}, nil
	default:
		return nil, fmt.Errorf("unknown provider %q (supported: docker)", name)
	}
}
