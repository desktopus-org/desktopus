package main

import (
	"os"
	"time"

	"github.com/charmbracelet/log"
)

var Version = "v0.1.0"

var logger = log.NewWithOptions(os.Stderr, log.Options{
	ReportCaller:    true,
	ReportTimestamp: true,
	TimeFormat:      time.Kitchen,
})

func main() {
	// 	manifest := `apiVersion: desktopus/image/v1alpha1
	// desktopusVersion: v0.1.0
	// os: ubuntu-jammy
	// modules:
	//   - chrome
	//   - steam
	// `

	//	buildOptions := image.ImageBuildOptions{
	//		RawManifest: manifest,
	//		ImageName:   "ubuntu-test",
	//		BuildDir:    "build",
	//	}
	//
	// imageBuilder, err := image.NewImageBuilder(buildOptions)
	//
	//	if err != nil {
	//		logger.Error(fmt.Sprintf("Error creating image builder: %s", err))
	//		os.Exit(1)
	//	}
	//
	// err = imageBuilder.Build()
	//
	//	if err != nil {
	//		logger.Error(fmt.Sprintf("Error building image: %s", err))
	//		os.Exit(1)
	//	}
}
