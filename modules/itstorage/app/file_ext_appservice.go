package app

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/domain/models"
	itFile "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/file"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/itstorage"
)

// filePathField names the offending value in the client errors this port returns. The owning
// module's field is "<name>_file_path"; the port does not know <name>, so it reports the suffix.
const filePathField = "file_path"

// errPathNotUsable is shared with the storagefield rule, which owns its translation in common.json.
const errPathNotUsable = "file.err_path_not_usable"

func NewFileExtService(fileSvc itFile.FileDomainService) itstorage.FileExtService {
	return &FileExtServiceImpl{fileSvc: fileSvc}
}

// FileExtServiceImpl is the cross-module port. It asserts nothing: the consuming module's
// application service already authorized the write that carries the path.
//
// A client error is returned as a composable.ClientErrorSignal, so a consumer running inside
// composable.InRefusableTransaction rolls back and answers 400 without unwrapping it itself; one
// outside can recover the errors with composable.AsClientErrorSignal.
type FileExtServiceImpl struct {
	fileSvc itFile.FileDomainService
}

func (this *FileExtServiceImpl) AssertFileUsable(
	ctx corectx.Context, path string, orgId *model.Id,
) (*model.Id, error) {
	resolved, err := this.fileSvc.FindByPath(ctx, path)
	if err != nil {
		return nil, err
	}
	if resolved == nil {
		return nil, composable.ClientErrorSignal{Errors: ft.ClientErrors{*ft.NewValidationError(filePathField,
			ft.ErrorKey(errPathNotUsable), "no stored file has this path")}}
	}
	if !sameOrg(resolved.File.GetOrgId(), orgId) {
		return nil, composable.ClientErrorSignal{Errors: ft.ClientErrors{*ft.NewValidationError(filePathField,
			ft.ErrorKey(errPathNotUsable),
			"the file belongs to another organization than the record referencing it")}}
	}
	return resolved.File.GetId(), nil
}

// DeleteFileByPath is best-effort by contract: a path naming no file is already in the desired
// state.
func (this *FileExtServiceImpl) DeleteFileByPath(ctx corectx.Context, path string) error {
	resolved, err := this.fileSvc.FindByPath(ctx, path)
	if err != nil || resolved == nil {
		return err
	}
	result, err := this.fileSvc.Delete(ctx, composable.DeleteCommand{models.FileFieldId: *resolved.File.GetId()})
	if err != nil {
		return err
	}
	if result.ClientErrors.Count() > 0 {
		return composable.ClientErrorSignal{Errors: result.ClientErrors}
	}
	return nil
}

func sameOrg(fileOrgId *model.Id, wantOrgId *model.Id) bool {
	if fileOrgId == nil || wantOrgId == nil {
		return fileOrgId == nil && wantOrgId == nil
	}
	return *fileOrgId == *wantOrgId
}
