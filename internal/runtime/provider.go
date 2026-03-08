package runtime

import (
	"context"
	"io"
)

// Provider abstracts a container runtime backend.
type Provider interface {
	Run(ctx context.Context, cfg *DesktopRunConfig, opts RunOptions, output io.Writer) (string, error)
	Stop(ctx context.Context, nameOrID string, timeout int) error
	Remove(ctx context.Context, nameOrID string, force bool) error
	List(ctx context.Context, all bool) ([]ContainerInfo, error)
	VolumeList(ctx context.Context) ([]VolumeInfo, error)
	VolumeRemove(ctx context.Context, name string, force bool) error
	Close() error
}
