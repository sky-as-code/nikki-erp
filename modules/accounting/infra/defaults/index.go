// Package defaults holds the records every new organization starts with, one `<schema>.json`
// file per resource. See orglifecycle/defaults for the format.
package defaults

import "embed"

//go:embed *.json
var Files embed.FS
