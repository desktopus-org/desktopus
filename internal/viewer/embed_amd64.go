//go:build linux && amd64 && embed_viewer

package viewer

import _ "embed"

//go:embed assets/desktopus-viewer-amd64
var viewerBinary []byte
