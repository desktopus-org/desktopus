package base

var UbuntuJammyImageName = "ubuntu-jammy"
var UbuntuJammyBaseImage = map[string]BaseImage{
	"v0.1.0": {
		RootScripts: []string{
			UbuntuJammyImageName + "/docker/root_scripts/permissions_gpu.sh",
		},
		Patches: []string{
			UbuntuJammyImageName + "/docker/patches/vnc_startup.sh.patch",
		},
		Dockerfile: UbuntuJammyImageName + "/docker/Dockerfile.tmpl",
		Modules: map[string]string{
			"chrome": "common/modules/chrome/install.sh",
			"steam":  UbuntuJammyImageName + "/modules/steam/install.sh",
		},
	},
}
