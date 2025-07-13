package tag

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	enum "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
)

type Tag struct {
	model.ModelBase

	Label *model.LangJson `json:"label,omitempty"`
	Type  *enum.EnumType  `json:"type,omitempty"`
}

func (this *Tag) Validate(forEdit bool) ft.ValidationErrors {
	enum := enum.Enum{
		Label: this.Label,
		Type:  this.Type,
	}
	return enum.Validate(forEdit)
}

func NewDerivedTag(label *model.LangJson, tagType *enum.EnumType) *DerivedTag {
	return &DerivedTag{
		Label:   label,
		tagType: tagType,
	}
}

type DerivedTag struct {
	model.ModelBase

	Label   *model.LangJson `json:"label,omitempty"`
	tagType *enum.EnumType  `json:"-"`
}

func (this *DerivedTag) Validate(forEdit bool) ft.ValidationErrors {
	enum := Tag{
		Label: this.Label,
		Type:  this.tagType,
	}
	return enum.Validate(forEdit)
}
