package external

import (
	"github.com/sky-as-code/nikki-erp/modules/core/storage"
	itStorage "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/itstorage"
)

// NewStorageConsumerImpl points core's storage port at the IT Storage module, which every module
// with `<name>_file_id` fields reaches through it instead of importing IT Storage.
//
// IT Storage lives in this binary, so the port is an in-process call. A build without it leaves
// the port unresolvable, and the dynamic resource engine then stores file ids unchecked; a build
// that split storage into its own service would register an HTTP client here instead.
func NewStorageConsumerImpl(svc itStorage.StorageConsumerAppService) storage.StorageConsumer {
	return svc
}
