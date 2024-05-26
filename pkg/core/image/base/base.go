package base

type BaseImage struct {
	// Path to the root scripts
	// These scripts are root scripts that can be
	// executed by any user at the beginning.
	RootScripts []string

	// Directory with patches to apply to the base os.
	Patches []string

	// Path to the Dockerfile template
	Dockerfile string
	// Map of modules
	// Key: module name
	// Value: module template directory
	Modules map[string]string
}

// Map of valid base images for desktopus
var BaseImages map[string]map[string]BaseImage = map[string]map[string]BaseImage{
	"ubuntu-jammy": UbuntuJammyBaseImage,
}
