package drive_file

import (
	"mime/multipart"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
)

func init() {
	var req cqrs.Request
	req = (*GetDriveFileByIdQuery)(nil)

	util.Unused(req)
}

type CreateDriveFileCommand struct {
	Name               string                   `json:"name" form:"name"`
	IsFolder           bool                     `json:"isFolder" form:"isFolder"`
	ParentDriveFileRef model.Id                 `json:"parentDriveFileRef" form:"parentDriveFileRef"`
	File               multipart.File           `json:"-" form:"-"`
	FileHeader         multipart.FileHeader     `json:"-" form:"-"`
	Visiblity          enum.DriveFileVisibility `json:"visiblity,omitempty" form:"visiblity"`
	ScopeType          enum.ScopeType           `json:"scope_type,omitempty" form:"scope_type"`
	ScopeRef           model.Id                 `json:"scope_ref,omitempty" form:"scope_ref"`
	OwnerRef           model.Id                 `json:"owner_ref,omitempty" form:"owner_ref"`
}

type CreateDriveFileResult = crud.OpResult[*domain.DriveFile]

type UpdateDriveFileCommand struct {
	Name       string                   `json:"name"`
	File       multipart.File           `json:"file"`
	FileHeader multipart.FileHeader     `json:"file_header"`
	Visiblity  enum.DriveFileVisibility `json:"visiblity,omitempty"`
}

type UpdateDriveFileResult = crud.OpResult[*domain.DriveFile]

var getDriveFileByIdRequestType = cqrs.RequestType{
	Module:    "drive",
	Submodule: "driveFile",
	Action:    "getById",
}

type GetDriveFileByIdQuery struct {
	DriveFileId model.Id `json:"driveFileId" param:"driveFileId"`
}

func (GetDriveFileByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getDriveFileByIdRequestType
}

type GetDriveFileByIdResult = crud.OpResult[*domain.DriveFile]

type GetDriveFileByParentQuery struct {
	crud.PagingOptions `json:",inline"`
	FileParentId       model.Id `json:"fileParentId" param:"fileParentId"`
}

type GetDriveFileByParentResultData = crud.PagedResult[*domain.DriveFile]
type GetDriveFileByParentResult = crud.OpResult[GetDriveFileByParentResultData]

type SearchDriveFileQuery struct {
	crud.SearchQuery
}

type SearchDriveFileResultData = crud.PagedResult[*domain.DriveFile]
type SearchDriveFileResult = crud.OpResult[SearchDriveFileResultData]

type MoveDriveFileToTrashCommand struct {
	DriveFileId model.Id `json:"driveFileId" param:"driveFileId"`
}

type MoveDriveFileToTrashResult = crud.OpResult[*domain.DriveFile]

type DeleteDriveFileCommand struct {
	DriveFileId model.Id `json:"driveFileId" param:"driveFileId"`
}

type DeleteDriveFileResult = crud.DeletionResult
