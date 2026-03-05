package domain

import (
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
)

type DriveFileShare struct {
	model.ModelBase     `json:",inline"`
	model.AuditableBase `json:",inline"`

	FileRef    model.Id                `json:"file_ref"`
	UserRef    model.Id                `json:"user_ref"`
	Permission enum.DriveFileSharePerm `json:"permission"`
}

func (d *DriveFileShare) Validate(forEdit bool) fault.ValidationErrors {
	return validator.ApiBased.ValidateStruct(d)
}
