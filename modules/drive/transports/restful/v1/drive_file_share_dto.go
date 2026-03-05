package v1

import (
	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/httpserver"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
	shareIt "github.com/sky-as-code/nikki-erp/modules/drive/interfaces/drive_file_share"
)

type DriveFileShareDto struct {
	model.ModelBase     `json:",inline"`
	model.AuditableBase `json:",inline"`

	FileRef    model.Id                `json:"file_ref"`
	UserRef    model.Id                `json:"user_ref"`
	Permission enum.DriveFileSharePerm `json:"permission"`
}

func (this *DriveFileShareDto) FromDriveFileShare(s domain.DriveFileShare) {
	model.MustCopy(s.ModelBase, this)
	model.MustCopy(s.AuditableBase, this)
	model.MustCopy(s, this)
}

type CreateDriveFileShareRequest = shareIt.CreateDriveFileShareCommand
type CreateDriveFileShareResponse = httpserver.RestCreateResponse

type CreateBulkDriveFileShareRequest = shareIt.CreateBulkDriveFileShareCommand

type CreateBulkDriveFileShareResponse struct {
	Items []httpserver.RestCreateResponse `json:"items"`
}

type UpdateDriveFileShareRequest = shareIt.UpdateDriveFileShareCommand
type UpdateDriveFileShareResponse = httpserver.RestUpdateResponse

type GetDriveFileShareByIdRequest = shareIt.GetDriveFileShareByIdQuery
type GetDriveFileShareByIdResponse = DriveFileShareDto

type GetDriveFileShareByFileIdRequest = shareIt.GetDriveFileShareByFileIdQuery
type GetDriveFileShareByFileIdResponse httpserver.RestSearchResponse[DriveFileShareDto]

func (this *GetDriveFileShareByFileIdResponse) FromResult(result *shareIt.GetDriveFileShareByFileIdResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(s *domain.DriveFileShare) DriveFileShareDto {
		item := DriveFileShareDto{}
		item.FromDriveFileShare(*s)
		return item
	})
}

type GetDriveFileShareByUserRequest = shareIt.GetDriveFileShareByUserQuery
type GetDriveFileShareByUserResponse httpserver.RestSearchResponse[DriveFileShareDto]

func (this *GetDriveFileShareByUserResponse) FromResult(result *shareIt.GetDriveFileShareByUserResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(s *domain.DriveFileShare) DriveFileShareDto {
		item := DriveFileShareDto{}
		item.FromDriveFileShare(*s)
		return item
	})
}

type SearchDriveFileShareRequest = shareIt.SearchDriveFileShareQuery
type SearchDriveFileShareResponse httpserver.RestSearchResponse[DriveFileShareDto]

func (this *SearchDriveFileShareResponse) FromResult(result *shareIt.SearchDriveFileShareResultData) {
	this.Total = result.Total
	this.Page = result.Page
	this.Size = result.Size
	this.Items = array.Map(result.Items, func(s *domain.DriveFileShare) DriveFileShareDto {
		item := DriveFileShareDto{}
		item.FromDriveFileShare(*s)
		return item
	})
}

type DeleteDriveFileShareRequest = shareIt.DeleteDriveFileShareCommand
type DeleteDriveFileShareResponse = httpserver.RestDeleteResponse
