//go:build !embed_viewer

package viewer

// viewerBinary is nil when the viewer is not embedded at build time.
// Launch() will fall back to discovering desktopus-viewer on the host.
var viewerBinary []byte
