package app

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/infra/storage/filestorage"
	itFile "github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/file"
	"github.com/sky-as-code/nikki-erp/modules/itstorage/interfaces/itstorage"
)

func NewFileExtService(fileSvc itFile.FileDomainService) itstorage.FileExtService {
	return &FileExtServiceImpl{fileSvc: fileSvc}
}

// FileExtServiceImpl is the cross-module port for serving files to a principal the consuming
// module authenticated itself, such as a kiosk on the mTLS desktop API. It asserts nothing.
type FileExtServiceImpl struct {
	fileSvc itFile.FileDomainService
}

func (this *FileExtServiceImpl) OpenFile(
	ctx corectx.Context, query itstorage.OpenFileQuery,
) (*itstorage.OpenFileResult, error) {
	file, err := this.fileSvc.FindFile(ctx, query.FileId)
	if err != nil {
		return nil, err
	}
	if file == nil || !visibleToOrg(file.GetOrgId(), query.OrgId) {
		return &itstorage.OpenFileResult{Found: false}, nil
	}
	stream, err := this.fileSvc.OpenObject(ctx, file, query.RangeHeader)
	if err != nil {
		return nil, err
	}
	return &itstorage.OpenFileResult{
		Found:    true,
		Name:     stream.Name,
		MimeType: stream.MimeType,
		Stream: &filestorage.StreamObjectResult{
			Body: stream.Body, ContentLength: stream.ContentLength, ContentRange: stream.ContentRange,
		},
	}, nil
}

// visibleToOrg admits the org's own files and the tenant-wide ones, which belong to no org.
func visibleToOrg(fileOrgId *model.Id, orgId *model.Id) bool {
	return fileOrgId == nil || (orgId != nil && *fileOrgId == *orgId)
}
