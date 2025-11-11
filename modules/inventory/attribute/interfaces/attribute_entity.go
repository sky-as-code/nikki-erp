package interfaces

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type Attribute struct {
	model.ModelBase
	model.AuditableBase

	ProductId     *model.Id       `json:"productId,omitempty"`
	CodeName      *string         `json:"codeName,omitempty"`
	DisplayName   *model.LangJson `json:"displayName,omitempty"`
	SortIndex     *int            `json:"sortIndex,omitempty"`
	DataType      *string         `json:"dataType,omitempty"`
	IsRequired    *bool           `json:"isRequired,omitempty"`
	IsEnum        *bool           `json:"isEnum,omitempty"`
	EnumValue     *model.LangJson `json:"enumValue,omitempty"`
	EnumValueSort *bool           `json:"enumValueSort,omitempty"`
	GroupId       *model.Id       `json:"groupId,omitempty"`
}

func (this *Attribute) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdPtrValidateRule(&this.ProductId, !forEdit),
		val.Field(&this.CodeName,
			val.NotNilWhen(!forEdit),
			val.When(this.CodeName != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_LONG_NAME_LENGTH),
			),
		),
		val.Field(&this.DisplayName,
			val.When(this.DisplayName != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_LONG_NAME_LENGTH),
			),
		),
		val.Field(&this.DataType,
			val.NotNilWhen(!forEdit),
			val.When(this.DataType != nil, val.NotEmpty),
		),

		model.IdPtrValidateRule(&this.GroupId, false),
	}

	return val.ApiBased.ValidateStruct(this, rules...)
}
