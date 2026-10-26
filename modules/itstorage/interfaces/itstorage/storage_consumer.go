package itstorage

import "github.com/sky-as-code/nikki-erp/modules/core/storage"

// StorageConsumerAppService is this module's implementation of the core/storage port, under its
// own type so the container holds it whether or not apptrait binds the port to it. apptrait
// registers it as storage.StorageConsumer in every binary that loads this module.
type StorageConsumerAppService interface {
	storage.StorageConsumer
}
