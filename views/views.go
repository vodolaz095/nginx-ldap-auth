package views

import (
	"embed"
)

// Views are templates being used
//
//go:embed *.tpl
var Views embed.FS
