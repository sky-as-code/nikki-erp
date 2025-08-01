package domain

import (
	"regexp"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	entResource "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/resource"
)

type Resource struct {
	model.ModelBase
	model.AuditableBase

	Name         *string            `json:"name,omitempty"`
	Description  *string            `json:"description,omitempty"`
	ResourceType *ResourceType      `json:"resourceType,omitempty"`
	ResourceRef  *string            `json:"resourceRef,omitempty"`
	ScopeType    *ResourceScopeType `json:"scopeType,omitempty"`

	Actions      []Action      `json:"actions" model:"-"` // TODO: Handle copy
	Entitlements []Entitlement `json:"entitlements" model:"-"`
}

func (this *Resource) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Name,
			val.NotNilWhen(!forEdit),
			val.When(this.Name != nil,
				val.NotEmpty,
				val.RegExp(regexp.MustCompile(`^[a-zA-Z0-9]+$`)), // alphanumeric
				val.Length(1, model.MODEL_RULE_TINY_NAME_LENGTH),
			),
		),
		val.Field(&this.Description,
			val.When(this.Description != nil,
				val.NotEmpty,
				val.Length(1, model.MODEL_RULE_DESC_LENGTH),
			),
		),
		ResourceTypeValidateRule(&this.ResourceType, !forEdit),
		ResourceRefValidateRule(&this.ResourceRef, &this.ResourceType, !forEdit),
		ResourceScopeTypeValidateRule(&this.ScopeType, !forEdit),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)

	return val.ApiBased.ValidateStruct(this, rules...)
}

type ResourceType entResource.ResourceType

const (
	ResourceTypeNikkiApplication = ResourceType(entResource.ResourceTypeNikkiApplication)
	ResourceTypeCustom           = ResourceType(entResource.ResourceTypeCustom)
)

func (this ResourceType) String() string {
	return string(this)
}

func WrapResourceType(s string) *ResourceType {
	st := ResourceType(s)
	return &st
}

func WrapResourceTypeEnt(s entResource.ResourceType) *ResourceType {
	st := ResourceType(s)
	return &st
}

func ResourceTypeValidateRule(field **ResourceType, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.When(*field != nil,
			val.NotEmpty,
			val.OneOf(ResourceTypeNikkiApplication, ResourceTypeCustom),
		),
	)
}

type ResourceScopeType entResource.ScopeType

const (
	ResourceScopeTypeOrg       = ResourceScopeType(entResource.ScopeTypeOrg)
	ResourceScopeTypeHierarchy = ResourceScopeType(entResource.ScopeTypeHierarchy)
	ResourceScopeTypePrivate   = ResourceScopeType(entResource.ScopeTypePrivate)
)

func (this ResourceScopeType) String() string {
	return string(this)
}

func WrapResourceScopeType(s string) *ResourceScopeType {
	st := ResourceScopeType(s)
	return &st
}

func WrapResourceScopeTypeEnt(s entResource.ScopeType) *ResourceScopeType {
	st := ResourceScopeType(s)
	return &st
}

func ResourceScopeTypeValidateRule(field **ResourceScopeType, isRequired bool) *val.FieldRules {
	return val.Field(field,
		val.NotNilWhen(isRequired),
		val.When(*field != nil,
			val.NotEmpty,
			val.OneOf(ResourceScopeTypeOrg, ResourceScopeTypeHierarchy, ResourceScopeTypePrivate),
		),
	)
}

func ResourceRefValidateRule(ref **string, resourceType **ResourceType, isRequired bool) *val.FieldRules {
	return val.Field(ref,
		val.NotNilWhen(isRequired),
		val.When(*ref != nil,
			val.When(resourceType != nil && *resourceType != nil && **resourceType == ResourceTypeNikkiApplication,
				val.NotEmpty,
				val.Length(model.MODEL_RULE_ULID_LENGTH, model.MODEL_RULE_ULID_LENGTH),
			),
		),
	)
}
