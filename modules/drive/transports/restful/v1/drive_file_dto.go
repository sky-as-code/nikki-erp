package v1

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	it "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file"
)

type DriveFileDto struct {
	model.ModelBase     `json:",inline"`
	model.AuditableBase `json:",inline"`

	OwnerRef           model.Id `json:"ownerRef"`
	ParentDriveFileRef model.Id `json:"parentDriveFileRef"`

	Name       string                   `json:"name"`
	MINE       string                   `json:"mime"`
	IsFolder   bool                     `json:"isFolder"`
	Size       uint64                   `json:"size"`
	Path       string                   `json:"-"`
	Storage    enum.DriveFileStorage    `json:"storage"`
	Visibility enum.DriveFileVisibility `json:"visibility"`
	Status     enum.DriveFileStatus     `json:"status"`
	// Children   []*DriveFileDto          `json:"children,omitempty" model:"-"`

	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}

func (this *DriveFileDto) FromDriveFile(f domain.DriveFile) {
	model.MustCopy(f.ModelBase, this)
	model.MustCopy(f.AuditableBase, this)
	model.MustCopy(f, this)
}

type CreateDriveFileRequest = it.CreateDriveFileCommand
type CreateDriveFileResponse = httpserver.RestCreateResponse

type UpdateDriveFileMetadataRequest = it.UpdateDriveFileMetadataCommand
type UpdateDriveFileMetadataResponse = httpserver.RestUpdateResponse

type UpdateDriveFileContentRequest = it.UpdateDriveFileContentCommand
type UpdateDriveFileContentResponse = httpserver.RestUpdateResponse

type GetDriveFileByIdRequest = it.GetDriveFileByIdQuery
type GetDriveFileByIdResponse = DriveFileDto

type MoveDriveFileToTrashRequest = it.MoveDriveFileToTrashCommand
type MoveDriveFileToTrashResponse = httpserver.RestUpdateResponse

type RestoreDriveFileRequest = it.RestoreDriveFileCommand
type RestoreDriveFileResponse = httpserver.RestUpdateResponse

type MoveDriveFileRequest = it.MoveDriveFileCommand
type MoveDriveFileResponse = httpserver.RestUpdateResponse

type DeleteDriveFileRequest = it.DeleteDriveFileCommand
type DeleteDriveFileResponse = httpserver.RestDeleteResponse

type GetDriveFileAncestorsRequest = it.GetDriveFileAncestorsQuery
type GetDriveFileAncestorsResponse = []DriveFileDto

type GetDriveFileByParentRequest = it.GetDriveFileByParentQuery
type GetDriveFileByParentResponse httpserver.RestSearchResponse[DriveFileDto]

func (this *GetDriveFileByParentResponse) FromResult(result *it.GetDriveFileByParentResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(f *domain.DriveFile) DriveFileDto {
		item := DriveFileDto{}
		item.FromDriveFile(*f)
		return item
	})
}

type SearchDriveFileRequest = it.SearchDriveFileQuery
type SearchDriveFileResponse httpserver.RestSearchResponse[DriveFileDto]

func (this *SearchDriveFileResponse) FromResult(result *it.SearchDriveFileResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(f *domain.DriveFile) DriveFileDto {
		item := DriveFileDto{}
		item.FromDriveFile(*f)
		return item
	})
}
