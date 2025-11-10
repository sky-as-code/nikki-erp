package domain

import (
	stdErr "errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules"
)

type ModuleMetadata struct {
	model.ModelBase

	Deps       []string        `json:"deps,omitempty"`
	Label      *model.LangJson `json:"label,omitempty"`
	Name       *string         `json:"name,omitempty"`
	IsOrphaned *bool           `json:"isOrphaned,omitempty"`
	Version    *semver.SemVer  `json:"version,omitempty"`
}

func (this *ModuleMetadata) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.LangJsonPtrValidateRule(&this.Label, true, 1, model.MODEL_RULE_TINY_NAME_LENGTH),
		model.IdPtrValidateRule(&this.Id, forEdit),
		val.Field(&this.Name,
			val.NotNilWhen(!forEdit),
			val.When(this.Name != nil,
				val.NotEmpty,
				model.ModelRuleCodeName,
			),
		),
		val.Field(&this.Version,
			val.NotNilWhen(!forEdit),
			val.When(this.Version != nil,
				val.NotEmpty,
				val.By(func(value any) error {
					s := value.(*semver.SemVer)
					if !s.IsValid() {
						return stdErr.New("invalid semver format")
					}
					return nil
				}),
			),
		),
	}
	return val.ApiBased.ValidateStruct(this, rules...)
}

func (this *ModuleMetadata) ModifiedFields(other modules.InCodeModule) *ModuleMetadata {
	count := 0
	modified := &ModuleMetadata{}
	if this.Label != nil && this.Label.TranslationKey() != other.LabelKey() {
		modified.Label = util.ToPtr(make(model.LangJson))
		modified.Label.SetTranslationKey(other.LabelKey())
		count++
	}
	if this.Version != nil && *this.Version != other.Version() {
		modified.Version = util.ToPtr(other.Version())
		count++
	}
	if count == 0 {
		return nil
	}
	return modified
}
