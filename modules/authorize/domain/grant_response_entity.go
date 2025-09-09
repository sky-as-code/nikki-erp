package domain

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/validator"
)

type GrantResponse struct {
	model.ModelBase
	model.AuditableBase

	RequestId   *model.Id  `json:"requestId,omitempty"`
	IsApproved  *bool      `json:"isApproved,omitempty"`
	Reason      *string    `json:"reason,omitempty"`
	ResponderId *model.Id  `json:"responderId,omitempty"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`

	// Related entities
	GrantRequest *GrantRequest `json:"grantRequest,omitempty" model:"-"`
}

func (this *GrantResponse) Validate(forEdit bool) fault.ValidationErrors {
	rules := []*validator.FieldRules{
		model.IdPtrValidateRule(&this.RequestId, !forEdit),
		validator.Field(&this.IsApproved,
			validator.NotNilWhen(!forEdit),
		),
		validator.Field(&this.Reason,
			validator.When(this.Reason != nil,
				validator.NotEmpty,
				validator.Length(1, model.MODEL_RULE_DESC_LENGTH),
			),
		),
		model.IdPtrValidateRule(&this.ResponderId, !forEdit),
		validator.Field(&this.CreatedAt,
			validator.NotNilWhen(!forEdit),
		),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)
	rules = append(rules, this.AuditableBase.ValidateRules(forEdit)...)

	return validator.ApiBased.ValidateStruct(this, rules...)
}

func (this *GrantResponse) SetDefaults() {
	this.ModelBase.SetDefaults()
}
