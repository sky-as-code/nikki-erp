package drive_file_share

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain"
)

type CreateDriveFileShareCommand struct {
	FileRef    model.Id            `json:"-" param:"drive_file_id"`
	UserRef    model.Id            `json:"user_ref"`
	Permission domain.DriveFilePerm `json:"permission"`
	UserId     model.Id            `json:"-"`
}

func (this CreateDriveFileShareCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.FileRef, true),
		model.IdValidateRule(&this.UserRef, true),
		model.IdValidateRule(&this.UserId, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type CreateDriveFileShareResult = dyn.OpResult[domain.DriveFileShare]

type CreateBulkDriveFileShareCommand struct {
	FileRef    model.Id            `json:"-" param:"drive_file_id"`
	UserRefs   []model.Id          `json:"user_refs"`
	Permission domain.DriveFilePerm `json:"permission"`
	UserId     model.Id            `json:"-"`
}

func (this CreateBulkDriveFileShareCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.FileRef, true),
		model.IdValidateRuleMulti(&this.UserRefs, true, 1, model.MODEL_RULE_ID_ARR_MAX),
		model.IdValidateRule(&this.UserId, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type CreateBulkDriveFileShareResult = dyn.OpResult[[]domain.DriveFileShare]

type UpdateDriveFileShareCommand struct {
	DriveFileId model.Id `json:"-" param:"drive_file_id"`
	Id          model.Id `json:"drive_file_share_id" param:"drive_file_share_id"`
	Etag        model.Etag              `json:"etag"`
	Permission  domain.DriveFilePerm `json:"permission"`
	UserId      model.Id              `json:"-"`
}

func (this UpdateDriveFileShareCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.DriveFileId, true),
		model.IdValidateRule(&this.Id, true),
		model.EtagValidateRule(&this.Etag, true),
		model.IdValidateRule(&this.UserId, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type UpdateDriveFileShareResult = dyn.OpResult[dyn.MutateResultData]

type GetDriveFileShareByIdQuery struct {
	DriveFileShareId model.Id `json:"drive_file_share_id" param:"drive_file_share_id"`
}

func (this GetDriveFileShareByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.DriveFileShareId, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetDriveFileShareByIdResult = dyn.OpResult[domain.DriveFileShare]

type GetDriveFileShareByFileIdQuery struct {
	crud.SearchQuery `json:",inline"`
	DriveFileId      model.Id `json:"drive_file_id" param:"drive_file_id"`
}

func (this GetDriveFileShareByFileIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.DriveFileId, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetDriveFileShareByFileIdResultData = dyn.PagedResultData[domain.DriveFileShare]
type GetDriveFileShareByFileIdResult = dyn.OpResult[GetDriveFileShareByFileIdResultData]

type GetDriveFileAncestorOwnersByFileIdQuery struct {
	DriveFileId model.Id `json:"drive_file_id" param:"drive_file_id"`
}

func (this GetDriveFileAncestorOwnersByFileIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.DriveFileId, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetDriveFileAncestorOwnersByFileIdResult = dyn.OpResult[[]domain.DriveFileShare]

type GetDriveFileResolvedSharesByFileIdQuery struct {
	crud.SearchQuery `json:",inline"`
	DriveFileId      model.Id `json:"drive_file_id" param:"drive_file_id"`
}

func (this GetDriveFileResolvedSharesByFileIdQuery) Validate() ft.ValidationErrors {
	rules := this.SearchQuery.ValidationRules()
	rules = append(rules, model.IdValidateRule(&this.DriveFileId, true))
	return val.ApiBased.ValidateStruct(&this, rules...)
}

func (this *GetDriveFileResolvedSharesByFileIdQuery) SetDefaults() {
	this.SearchQuery.SetDefaults()
}

type GetDriveFileResolvedSharesByFileIdResultData = dyn.PagedResultData[domain.DriveFileShare]
type GetDriveFileResolvedSharesByFileIdResult = dyn.OpResult[GetDriveFileResolvedSharesByFileIdResultData]

type GetDriveFileUserShareDetailsQuery struct {
	DriveFileId model.Id `json:"drive_file_id" param:"drive_file_id"`
	UserId      model.Id `json:"user_id" param:"user_id"`
}

func (this GetDriveFileUserShareDetailsQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.DriveFileId, true),
		model.IdValidateRule(&this.UserId, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type DriveFileUserShareDetail struct {
	Id          model.Id                   `json:"id"`
	Etag        model.Etag                 `json:"etag,omitempty"`
	CreatedAt   *model.ModelDateTime       `json:"created_at,omitempty"`
	UpdatedAt   *model.ModelDateTime       `json:"updated_at,omitempty"`
	FileRef     model.Id                   `json:"file_ref"`
	UserRef     model.Id                   `json:"user_ref"`
	Permission  domain.DriveFilePerm       `json:"permission"`
	User        *domain.DriveFileShareUser  `json:"user,omitempty"`
	File        *domain.DriveFileShareFile `json:"file,omitempty"`
}

type GetDriveFileUserShareDetailsResult = dyn.OpResult[[]DriveFileUserShareDetail]

type GetDriveFileShareByUserQuery struct {
	UserId model.Id `json:"user_id" param:"user_id"`
}

func (this GetDriveFileShareByUserQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.UserId, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetDriveFileShareByUserResultData = dyn.PagedResultData[domain.DriveFileShare]
type GetDriveFileShareByUserResult = dyn.OpResult[GetDriveFileShareByUserResultData]

type ListDriveFileSharesByFileRefsAndUserQuery struct {
	DriveFileIds []model.Id `json:"drive_file_ids"`
	UserId       model.Id   `json:"user_id"`
}

func (this ListDriveFileSharesByFileRefsAndUserQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.UserId, true),
		model.IdValidateRuleMulti(&this.DriveFileIds, false, 0, model.MODEL_RULE_ID_ARR_MAX),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type ListDriveFileSharesByFileRefsAndUserResult = dyn.OpResult[[]domain.DriveFileShare]

type SearchDriveFileShareQuery struct {
	crud.SearchQuery
}

func (this SearchDriveFileShareQuery) Validate() ft.ValidationErrors {
	rules := this.SearchQuery.ValidationRules()
	return val.ApiBased.ValidateStruct(&this, rules...)
}

func (this *SearchDriveFileShareQuery) SetDefaults() {
	this.SearchQuery.SetDefaults()
}

type SearchDriveFileShareResultData = dyn.PagedResultData[domain.DriveFileShare]
type SearchDriveFileShareResult = dyn.OpResult[SearchDriveFileShareResultData]

type DeleteDriveFileShareCommand struct {
	DriveFileId      model.Id `json:"-" param:"drive_file_id"`
	DriveFileShareId model.Id `json:"drive_file_share_id" param:"drive_file_share_id"`
	UserId           model.Id `json:"-"`
}

func (this DeleteDriveFileShareCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.DriveFileId, true),
		model.IdValidateRule(&this.DriveFileShareId, true),
		model.IdValidateRule(&this.UserId, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteDriveFileShareResult = dyn.OpResult[dyn.MutateResultData]
