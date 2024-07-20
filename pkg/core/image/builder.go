package image

// ImageBuilder defines a common interface for all image specifications
type ImageBuilder interface {
	Build() error
}

type ImageBuildOptions struct {
	// Manifest raw content
	// Can be JSON or YAML
	RawManifest string

	// Name of the image to build
	ImageName string

	// Directory where to generate the build context
	// If empty, a temporary directory will be created
	BuildDir string
}

type ImageBuilderImpl struct {
	// Options for building the image
	options ImageBuildOptions

	// Parsed manifest
	// manifest imageManifest
}
