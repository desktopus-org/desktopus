package modules

import "embed"

//go:embed all:chrome
var BuiltinFS embed.FS
