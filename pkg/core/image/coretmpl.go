package image

import (
	"embed"
	"fmt"

	"gopkg.in/yaml.v2"
)

//go:embed all:core_templates/*
var coreTemplatesFS embed.FS

// CoreTemplateIndex defines all the metadata related with
// all available core template OSs.
type metaCoreTemplateIndex struct {
	// Type of the metadata
	MetaType string `json:"type" yaml:"type"`
	// List of OS availables
	OsImages []string `json:"os_images" yaml:"os_images"`
}

// CoreTemplateOS defines the structure of the metadata
// for each core template OS
type metaCoreTemplateOS struct {
	// Type of the metadata
	MetaType string `json:"type" yaml:"type"`
	// Name of the OS
	Name string `json:"name" yaml:"name"`

	// List of scripts to run as root at the beggining at startup time
	RootScripts []string `json:"root_scripts" yaml:"root_scripts"`

	// Directory with patches to apply to the OS files
	Patches []string `json:"patches" yaml:"patches"`

	// Directory of the Dockerfile to use
	// All the templatings is applied to this file at build time
	Dockerfile string `json:"dockerfile" yaml:"dockerfile"`

	// List of common modules to include in the image
	// These modules can be shared between different OS
	CommonModules []string `json:"common_modules" yaml:"common_modules"`

	// List of modules to include in the image
	// These modules are specific to the OS
	Modules []string `json:"modules" yaml:"modules"`
}

type coreTemplateOS struct {
	name          string
	rootScripts   []string
	patches       []string
	dockerfile    string
	commonModules map[string]string
	modules       map[string]string
}

func newCoreTemplateOS(osName string) (*coreTemplateOS, error) {
	// Check if meta.yaml of the OS exists
	var metaIndex metaCoreTemplateIndex
	metaIndexFile, err := coreTemplatesFS.ReadFile("core_templates/meta.yaml")
	if err != nil {
		return nil, newErrCoreTemplateReadingIndexMeta(err.Error())
	}
	yaml.Unmarshal(metaIndexFile, &metaIndex)
	if metaIndex.MetaType != "core_index" {
		return nil, newErrCoreTemplateReadingIndexMeta("invalid type")
	}

	// Check if the OS exists in the index
	for index, os := range metaIndex.OsImages {
		if os == osName {
			break
		}
		if index == len(metaIndex.OsImages)-1 {
			return nil, newErrCoreTemplateOSNotFound(osName)
		}
	}

	// Check if the OS metadata exists
	var metaOS metaCoreTemplateOS
	metaOSFile, err := coreTemplatesFS.ReadFile(fmt.Sprintf("core_templates/os/%s/meta.yaml", osName))
	if err != nil {
		return nil, newErrCoreTemplateReadingCoreMeta(osName, err.Error())
	}
	yaml.Unmarshal(metaOSFile, &metaOS)
	if metaOS.MetaType != "core_os" {
		return nil, newErrCoreTemplateReadingCoreMeta(osName, "invalid type: "+metaOS.MetaType)
	}

	// Create the coreTemplateOS struct
	commonModulesBaseDir := "core_templates/common/modules/"
	baseDir := "core_templates/os/" + osName + "/"
	coreOS := coreTemplateOS{
		name: metaOS.Name,
	}
	for _, script := range metaOS.RootScripts {
		coreOS.rootScripts = append(coreOS.rootScripts, baseDir+script)
	}
	for _, patch := range metaOS.Patches {
		coreOS.patches = append(coreOS.patches, baseDir+patch)
	}
	coreOS.dockerfile = baseDir + metaOS.Dockerfile
	coreOS.commonModules = map[string]string{}
	for _, commonModule := range metaOS.CommonModules {
		coreOS.commonModules[commonModule] = commonModulesBaseDir + commonModule + "/install.sh"
	}
	coreOS.modules = map[string]string{}
	for _, module := range metaOS.Modules {
		coreOS.modules[module] = baseDir + "modules/" + module + "/install.sh"
	}

	return &coreOS, nil
}
