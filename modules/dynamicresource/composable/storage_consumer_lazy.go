package composable

import (
	"sync"

	"go.bryk.io/pkg/errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/storage"
)

// lazyStorageConsumer resolves the storage port on first use rather than when the onion is built.
//
// Injecting it through BuildParam would not work: the storage module's own onions take a
// BuildParam, while the port's implementation depends on those onions, so the container would
// see a cycle; and with an optional field, an onion built before the storage module's Init ran
// would silently keep a nil port for the life of the process. By the time a write references a
// file, every module has initialized.
//
// A binary that registers no port stores file fields unchecked, as before the port existed.
type lazyStorageConsumer struct {
	once     sync.Once
	consumer storage.StorageConsumer
}

// NewLazyStorageConsumer is also what a domain service uses when it calls the port itself, such
// as to duplicate a file, for the same init-order reason.
func NewLazyStorageConsumer() storage.StorageConsumer {
	return &lazyStorageConsumer{}
}

func (this *lazyStorageConsumer) resolve() storage.StorageConsumer {
	this.once.Do(func() {
		_ = deps.Invoke(func(consumer storage.StorageConsumer) {
			this.consumer = consumer
		})
	})
	return this.consumer
}

func (this *lazyStorageConsumer) AssertFilesUsable(
	ctx corectx.Context, query storage.AssertFilesUsableQuery,
) (ft.ClientErrors, error) {
	if consumer := this.resolve(); consumer != nil {
		return consumer.AssertFilesUsable(ctx, query)
	}
	return nil, nil
}

func (this *lazyStorageConsumer) MarkFilesInUse(ctx corectx.Context, cmd storage.MarkFilesInUseCommand) error {
	if consumer := this.resolve(); consumer != nil {
		return consumer.MarkFilesInUse(ctx, cmd)
	}
	return nil
}

func (this *lazyStorageConsumer) ReleaseFiles(ctx corectx.Context, cmd storage.ReleaseFilesCommand) error {
	if consumer := this.resolve(); consumer != nil {
		return consumer.ReleaseFiles(ctx, cmd)
	}
	return nil
}

func (this *lazyStorageConsumer) DuplicateFile(
	ctx corectx.Context, cmd storage.DuplicateFileCommand,
) (*model.Id, error) {
	if consumer := this.resolve(); consumer != nil {
		return consumer.DuplicateFile(ctx, cmd)
	}
	return nil, errors.New("no storage module is registered to duplicate a file")
}
