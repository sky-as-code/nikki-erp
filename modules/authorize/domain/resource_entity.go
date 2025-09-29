package domain

import (
	"regexp"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/validator"
)

type Resource struct {
	model.ModelBase
	model.AuditableBase

	Name         *string            `json:"name,omitempty"`
	Description  *string            `json:"description,omitempty"`
	ResourceType *ResourceType      `json:"resourceType,omitempty"`
	ResourceRef  *string            `json:"resourceRef,omitempty"`
	ScopeType    *ResourceScopeType `json:"scopeType,omitempty"`

	// HierarchyRef *string `json:"hierarchyRef,omitempty"`

	Actions      []Action      `json:"actions" model:"-"` // TODO: Handle copy
	Entitlements []Entitlement `json:"entitlements" model:"-"`
}

func (this *Resource) Validate(forEdit bool) fault.ValidationErrors {
	rules := []*validator.FieldRules{
		validator.Field(&this.Name,
			validator.NotNilWhen(!forEdit),
			validator.When(this.Name != nil,
				validator.NotEmpty,
				validator.RegExp(regexp.MustCompile(`^[a-zA-Z0-9]+$`)), // alphanumeric
				validator.Length(1, model.MODEL_RULE_TINY_NAME_LENGTH),
			),
		),
		validator.Field(&this.Description,
			validator.When(this.Description != nil,
				validator.NotEmpty,
				validator.Length(1, model.MODEL_RULE_DESC_LENGTH),
			),
		),
		ResourceTypeValidateRule(&this.ResourceType, !forEdit),
		ResourceRefValidateRule(&this.ResourceRef, &this.ResourceType, !forEdit),
		ResourceScopeTypeValidateRule(&this.ScopeType, !forEdit),
	}
	rules = append(rules, this.ModelBase.ValidateRules(forEdit)...)

	return validator.ApiBased.ValidateStruct(this, rules...)
}

type ResourceType string

const (
	ResourceTypeNikkiApplication = ResourceType("nikki_application")
	ResourceTypeCustom           = ResourceType("custom")
)

func (this ResourceType) String() string {
	return string(this)
}

func WrapResourceType(s string) *ResourceType {
	st := ResourceType(s)
	return &st
}

func ResourceTypeValidateRule(field **ResourceType, isRequired bool) *validator.FieldRules {
	return validator.Field(field,
		validator.NotNilWhen(isRequired),
		validator.When(*field != nil,
			validator.NotEmpty,
			validator.OneOf(
				ResourceTypeNikkiApplication,
				ResourceTypeCustom,
			),
		),
	)
}

type ResourceScopeType string

const (
	ResourceScopeTypeDomain    = ResourceScopeType("domain")
	ResourceScopeTypeOrg       = ResourceScopeType("org")
	ResourceScopeTypeHierarchy = ResourceScopeType("hierarchy")
	ResourceScopeTypePrivate   = ResourceScopeType("private")
)

func (this ResourceScopeType) String() string {
	return string(this)
}

func WrapResourceScopeType(s string) *ResourceScopeType {
	st := ResourceScopeType(s)
	return &st
}

func ResourceScopeTypeValidateRule(field **ResourceScopeType, isRequired bool) *validator.FieldRules {
	return validator.Field(field,
		validator.NotNilWhen(isRequired),
		validator.When(*field != nil,
			validator.NotEmpty,
			validator.OneOf(
				ResourceScopeTypeDomain,
				ResourceScopeTypeOrg,
				ResourceScopeTypeHierarchy,
				ResourceScopeTypePrivate,
			),
		),
	)
}

func ResourceRefValidateRule(ref **string, resourceType **ResourceType, isRequired bool) *validator.FieldRules {
	return validator.Field(ref,
		validator.NotNilWhen(isRequired),
		validator.When(*ref != nil,
			validator.When(resourceType != nil && *resourceType != nil && **resourceType == ResourceTypeNikkiApplication,
				validator.NotEmpty,
				validator.Length(model.MODEL_RULE_ULID_LENGTH, model.MODEL_RULE_ULID_LENGTH),
			),
		),
	)
}
