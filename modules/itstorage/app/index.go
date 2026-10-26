package app

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

// InitApplicationServices registers the services this module publishes beyond its onions: the
// FileExtService port other modules inject for their `*_file_path` fields.
func InitApplicationServices() error {
	return deps.Register(NewFileExtService)
}
