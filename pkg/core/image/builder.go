package image

import (
	"desktopus/pkg/core/image/base"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"
)

// Desktopus templates repository
// This is the default repository where to find the templates
var DesktopusTemplatesRepo = "https://raw.githubusercontent.com/desktopus-org/desktopus/base-templates"

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
	// Directory where to find the templates
	// Uris can be local paths or remote URLs
	TemplatesUri string
	// Directory where to generate the build context
	// If empty, a temporary directory will be created
	BuildDir string
}

var ValidApiVersions = []string{
	"desktopus/image/v1alpha1",
}

func imageBuilderFactory(specVersion string, options ImageBuildOptions) (ImageBuilder, error) {
	switch specVersion {
	case "desktopus/image/v1alpha1":
		imageBuilder, err := newV1Alpha1ImageBuilder(options)
		if err != nil {
			return nil, err
		}
		return imageBuilder, nil
	default:
		return nil, fmt.Errorf("unsupported spec version: %s", specVersion)
	}
}

// ----------------
// Utility function to extract version and instantiate the correct ImageSpec
// ----------------
func NewImageBuilder(options ImageBuildOptions) (ImageBuilder, error) {
	var temp struct {
		ApiVersion string `json:"apiVersion" yaml:"apiVersion"`
	}
	rawManifest := options.RawManifest
	if rawManifest == "" {
		return nil, NewErrBuildManifestNotSpecified()
	}

	// Unmarshal the version field
	err := yaml.Unmarshal([]byte(rawManifest), &temp)
	if err != nil {
		return nil, NewErrBuildUnmarshallApiVersion(err.Error())
	}

	// Use the version to create the correct ImageSpec instance
	imageBuilder, err := imageBuilderFactory(temp.ApiVersion, options)
	if err != nil {
		return nil, err
	}
	return imageBuilder, nil

}

// ----------------
// v1alpha1 Image Builder
// ----------------
type v1alpha1ImageBuilder struct {
	// Options for building the image
	Options ImageBuildOptions

	// Parsed manifest
	Manifest v1alpha1ImageManifest
}
type v1alpha1ImageManifest struct {
	DesktopusVersion string   `json:"desktopusVersion" yaml:"desktopusVersion"`
	OS               string   `json:"os" yaml:"os"`
	Modules          []string `json:"modules" yaml:"modules"`
}
type v1alpha1TemplateValues struct {
	Modules []string
}

func newV1Alpha1ImageBuilder(options ImageBuildOptions) (*v1alpha1ImageBuilder, error) {
	// rawManifest is a required field
	if options.RawManifest == "" {
		return nil, NewErrBuildManifestNotSpecified()
	}
	// imageName is a required field
	if options.ImageName == "" {
		return nil, NewErrBuildImageNotSpecified()
	}

	return &v1alpha1ImageBuilder{
		Options: options,
	}, nil
}

func (s *v1alpha1ImageBuilder) generateDockerFiles() error {
	var manifest v1alpha1ImageManifest
	rawManifest := s.Options.RawManifest

	// Try to unmarshal as JSON first, then as YAML if JSON fails
	if errJson := json.Unmarshal([]byte(rawManifest), &manifest); errJson != nil {
		if errYaml := yaml.Unmarshal([]byte(rawManifest), &manifest); errYaml != nil {
			// Return an error that includes both JSON and YAML unmarshalling errors
			return NewErrBuildUnmarshallManifest(
				fmt.Sprintf("JSON error: %s, YAML error: %s", errJson.Error(), errYaml.Error()),
			)
		}
	}
	s.Manifest = manifest

	baseOS := s.Manifest.OS
	if baseOS == "" {
		return NewErrBuildOSNotSpecified()
	}

	// Check if the OS is supported
	if _, ok := base.BaseImages[baseOS]; !ok {
		return NewErrBuildUnsupportedOs(baseOS)
	}

	version := s.Manifest.DesktopusVersion
	if version == "" {
		return NewErrVersionNotSpecified(baseOS)
	}

	// Check if the version is supported
	if _, ok := base.BaseImages[baseOS][version]; !ok {
		return NewErrBuildUnsupportedVersion(version, baseOS)
	}

	imageName := s.Options.ImageName

	// Check if buildDir is specified
	// If not, create a temporary directory
	if s.Options.BuildDir == "" {
		// Create a temporary directory
		var err error
		s.Options.BuildDir, err = os.MkdirTemp("", fmt.Sprintf("desktopus-%s", imageName))
		if err != nil {
			return err
		}
	}

	// Check if templatesUri is specified
	if s.Options.TemplatesUri == "" {
		s.Options.TemplatesUri = fmt.Sprintf("%s/%s/", DesktopusTemplatesRepo, version)
	} else {
		// Only valid URIs are http(s) and file
		templateUri := s.Options.TemplatesUri
		if !validTemplatesUri(templateUri) {
			return NewErrBuildInvalidTemplatesUri(templateUri)
		}
		s.Options.TemplatesUri = fmt.Sprintf("%s/%s/", s.Options.TemplatesUri, version)
	}

	// Prepare the template values
	templateValues := v1alpha1TemplateValues{
		Modules: []string{},
	}

	// Read the modules
	selectedModules := map[string]string{}
	for _, module := range s.Manifest.Modules {
		// Check if the module is supported
		baseImage := base.BaseImages[baseOS][version]
		moduleDir, ok := baseImage.Modules[module]
		if !ok {
			return NewErrBuildUnsupportedModule(module)
		}
		fileName := filepath.Base(moduleDir)
		finalModuleDir := fmt.Sprintf("/modules/%s/%s", module, fileName)
		templateValues.Modules = append(templateValues.Modules, finalModuleDir)
		selectedModules[module] = moduleDir
	}
	// Move files from the templatesUri to the buildDir
	templatesUri := s.Options.TemplatesUri
	buildDir := s.Options.BuildDir

	isHTTP := strings.HasPrefix(templatesUri, "http://") || strings.HasPrefix(templatesUri, "https://")
	isFile := strings.HasPrefix(templatesUri, "file://")

	if isHTTP || isFile {
		var err error

		// Dockerfile
		baseDockerfile := base.BaseImages[baseOS][version].Dockerfile
		dockerfileSrc := fmt.Sprintf("%s/%s", templatesUri, baseDockerfile)
		dockerfileDest := fmt.Sprintf("%s/Dockerfile.tmpl", buildDir)

		if isFile {
			dockerfileSrc = fmt.Sprintf("%s/%s", strings.TrimPrefix(templatesUri, "file://"), baseDockerfile)
		}

		if err = CopyFile(dockerfileSrc, dockerfileDest, isHTTP); err != nil {
			return NewErrBuildCopyingTemplate(err.Error())
		}

		dockerfileTemplate, err := template.ParseFiles(dockerfileDest)
		if err != nil {
			return err
		}
		parsedDockerfile, err := os.Create(fmt.Sprintf("%s/Dockerfile", buildDir))
		if err != nil {
			return err
		}
		err = dockerfileTemplate.Execute(parsedDockerfile, templateValues)
		if err != nil {
			return err
		}
		// Remove the template file
		err = os.Remove(dockerfileDest)
		if err != nil {
			return err
		}

		// Root scripts
		for _, rootScript := range base.BaseImages[baseOS][version].RootScripts {
			rootScriptSrc := fmt.Sprintf("%s/%s", templatesUri, rootScript)
			fileName := filepath.Base(rootScript)
			rootScriptDest := fmt.Sprintf("%s/root_scripts/%s", buildDir, fileName)

			if isFile {
				rootScriptSrc = fmt.Sprintf("%s/%s", strings.TrimPrefix(templatesUri, "file://"), rootScript)
			}

			if err = CopyFile(rootScriptSrc, rootScriptDest, isHTTP); err != nil {
				return NewErrBuildCopyingTemplate(err.Error())
			}
		}

		// Patches
		for _, patch := range base.BaseImages[baseOS][version].Patches {
			patchSrc := fmt.Sprintf("%s/%s", templatesUri, patch)
			fileName := filepath.Base(patch)
			patchDest := fmt.Sprintf("%s/patches/%s", buildDir, fileName)

			if isFile {
				patchSrc = fmt.Sprintf("%s/%s", strings.TrimPrefix(templatesUri, "file://"), patch)
			}

			if err = CopyFile(patchSrc, patchDest, isHTTP); err != nil {
				return NewErrBuildCopyingTemplate(err.Error())
			}
		}

		// Modules
		for module, moduleDir := range selectedModules {
			moduleSrc := fmt.Sprintf("%s/%s", templatesUri, moduleDir)
			moduleDest := fmt.Sprintf("%s/modules/%s/install.sh", buildDir, module)

			if isFile {
				moduleSrc = fmt.Sprintf("%s/%s", strings.TrimPrefix(templatesUri, "file://"), moduleDir)
			}

			if err = CopyFile(moduleSrc, moduleDest, isHTTP); err != nil {
				return NewErrBuildCopyingTemplate(err.Error())
			}
		}
	} else {
		return NewErrBuildInvalidTemplatesUri(templatesUri)
	}

	return nil

}

func (s *v1alpha1ImageBuilder) Build() error {
	return s.generateDockerFiles()
}
