//go:build linux && arm64 && embed_viewer

package viewer

import _ "embed"

//go:embed assets/desktopus-viewer-arm64
var viewerBinary []byte
