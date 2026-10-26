package app

import (
	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
)

// InitApplicationServices registers the services this module publishes beyond its onions: the
// FileExtService port, and the implementation apptrait binds to core/storage.StorageConsumer.
func InitApplicationServices() error {
	return deps.Register(NewFileExtService, NewStorageConsumerAppService)
}
