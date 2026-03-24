package drive_file_share

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/drive/domain"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
)

type CreateDriveFileShareCommand struct {
	FileRef    model.Id                `param:"driveFileId"`
	UserRef    model.Id                `json:"userRef"`
	Permission enum.DriveFilePerm `json:"permission"`
}

func (this CreateDriveFileShareCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.FileRef, true),
		model.IdValidateRule(&this.UserRef, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type CreateDriveFileShareResult = crud.OpResult[*domain.DriveFileShare]

type CreateBulkDriveFileShareCommand struct {
	FileRef    model.Id                `param:"driveFileId"`
	UserRefs   []model.Id              `json:"userRefs"`
	Permission enum.DriveFilePerm `json:"permission"`
}

func (this CreateBulkDriveFileShareCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.FileRef, true),
		model.IdValidateRuleMulti(&this.UserRefs, true, 1, model.MODEL_RULE_ID_ARR_MAX),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type CreateBulkDriveFileShareResult = crud.OpResult[[]*domain.DriveFileShare]

type UpdateDriveFileShareCommand struct {
	Id         model.Id                `param:"driveFileShareId" json:"driveFileShareId"`
	Etag       model.Etag              `json:"etag"`
	Permission enum.DriveFilePerm `json:"permission"`
}

func (this UpdateDriveFileShareCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.Id, true),
		model.EtagValidateRule(&this.Etag, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type UpdateDriveFileShareResult = crud.OpResult[*domain.DriveFileShare]

type GetDriveFileShareByIdQuery struct {
	DriveFileShareId model.Id `param:"driveFileShareId" json:"driveFileShareId"`
}

func (this GetDriveFileShareByIdQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.DriveFileShareId, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type GetDriveFileShareByIdResult = crud.OpResult[*domain.DriveFileShare]

type GetDriveFileShareByFileIdQuery struct {
	crud.SearchQuery `json:",inline"`
	DriveFileId      model.Id `param:"driveFileId" json:"driveFileId"`
}

func (this GetDriveFileShareByFileIdQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.DriveFileId, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type GetDriveFileShareByFileIdResultData = crud.PagedResult[*domain.DriveFileShare]
type GetDriveFileShareByFileIdResult = crud.OpResult[*GetDriveFileShareByFileIdResultData]

type GetDriveFileShareByUserQuery struct {
	UserId model.Id `param:"userId" json:"user_id"`
}

func (this GetDriveFileShareByUserQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.UserId, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type GetDriveFileShareByUserResultData = crud.PagedResult[*domain.DriveFileShare]
type GetDriveFileShareByUserResult = crud.OpResult[*GetDriveFileShareByUserResultData]

// ListDriveFileSharesByFileRefsAndUserQuery loads shares for one user on any of the given drive files (e.g. ancestors + target for effective permission).
type ListDriveFileSharesByFileRefsAndUserQuery struct {
	DriveFileIds []model.Id `json:"driveFileIds"`
	UserId       model.Id   `json:"userId"`
}

func (this ListDriveFileSharesByFileRefsAndUserQuery) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.UserId, true),
		model.IdValidateRuleMulti(&this.DriveFileIds, false, 0, model.MODEL_RULE_ID_ARR_MAX),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type ListDriveFileSharesByFileRefsAndUserResult = crud.OpResult[[]*domain.DriveFileShare]

type SearchDriveFileShareQuery struct {
	crud.SearchQuery
}

func (this SearchDriveFileShareQuery) Validate() fault.ValidationErrors {
	rules := this.SearchQuery.ValidationRules()
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

func (this *SearchDriveFileShareQuery) SetDefaults() {
	this.SearchQuery.SetDefaults()
}

type SearchDriveFileShareResultData = crud.PagedResult[*domain.DriveFileShare]
type SearchDriveFileShareResult = crud.OpResult[*SearchDriveFileShareResultData]

type DeleteDriveFileShareCommand struct {
	DriveFileShareId model.Id `param:"driveFileShareId" json:"driveFileShareId"`
}

func (this DeleteDriveFileShareCommand) Validate() fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdValidateRule(&this.DriveFileShareId, true),
	}
	return validator.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteDriveFileShareResult = crud.DeletionResult
