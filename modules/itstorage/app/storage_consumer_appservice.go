package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/storage"
	itFile "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/file"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/itstorage"
)

func NewStorageConsumerAppService(fileSvc itFile.FileDomainService) itstorage.StorageConsumerAppService {
	return &StorageConsumerAppServiceImpl{fileSvc: fileSvc}
}

// StorageConsumerAppServiceImpl implements the core/storage.StorageConsumer port; apptrait
// registers it as that port. It asserts no permission, by the port's contract: the consuming
// module's application service has authorized the write of the record that references the files.
type StorageConsumerAppServiceImpl struct {
	fileSvc itFile.FileDomainService
}

func (this *StorageConsumerAppServiceImpl) AssertFilesUsable(
	ctx corectx.Context, query storage.AssertFilesUsableQuery,
) (ft.ClientErrors, error) {
	return this.fileSvc.AssertFilesUsable(ctx, query)
}

func (this *StorageConsumerAppServiceImpl) MarkFilesInUse(ctx corectx.Context, cmd storage.MarkFilesInUseCommand) error {
	return this.fileSvc.MarkFilesInUse(ctx, cmd)
}

func (this *StorageConsumerAppServiceImpl) ReleaseFiles(ctx corectx.Context, cmd storage.ReleaseFilesCommand) error {
	return this.fileSvc.ReleaseFiles(ctx, cmd)
}

func (this *StorageConsumerAppServiceImpl) DuplicateFile(
	ctx corectx.Context, cmd storage.DuplicateFileCommand,
) (*model.Id, error) {
	return this.fileSvc.DuplicateFile(ctx, cmd)
}
