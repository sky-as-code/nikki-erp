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
	Permission enum.DriveFilePerm `json:"permission"`

	// User is an optional view populated by application layer when returning API responses.
	// It is not persisted in DriveFileShare storage.
	User *DriveFileShareUser `json:"user,omitempty"`
}

type DriveFileShareUser struct {
	Id          model.Id `json:"id"`
	DisplayName *string  `json:"displayName,omitempty"`
	Email       *string  `json:"email,omitempty"`
	AvatarUrl   *string  `json:"avatarUrl,omitempty"`
}

func (d *DriveFileShare) Validate(forEdit bool) fault.ValidationErrors {
	return validator.ApiBased.ValidateStruct(d)
}
