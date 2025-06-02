package domain

import (
	"regexp"

	"go.bryk.io/pkg/errors"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	entResource "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/resource"
)

type Resource struct {
	model.ModelBase

	Name         *string            `json:"name,omitempty"`
	Description  *string            `json:"description,omitempty"`
	ResourceType *ResourceType      `json:"resourceType,omitempty"`
	ResourceRef  *string            `json:"resourceRef,omitempty"`
	ScopeType    *ResourceScopeType `json:"scopeType,omitempty"`

	Actions []Action `json:"actions,omitempty"`
}

func (this *Resource) Validate(forEdit bool) ft.ValidationErrors {
	rules := []*val.FieldRules{
		val.Field(&this.Name,
			val.Required,
			val.RegExp(regexp.MustCompile(`^[a-zA-Z0-9_\-\s]+$`)), // alphanumeric, underscore, dash and space
			val.Length(1, 50),
		),
		ResourceScopeTypeValidateRule(&this.ScopeType),
	}

	return val.ApiBased.ValidateStruct(this, rules...)
}

type ResourceType entResource.ResourceType

const (
	ResourceTypeNikkiApplication = ResourceType(entResource.ResourceTypeNikkiApplication)
	ResourceTypeCustom           = ResourceType(entResource.ResourceTypeCustom)
)

func (this ResourceType) Validate() error {
	switch this {
	case ResourceTypeNikkiApplication, ResourceTypeCustom:
		return nil
	default:
		return errors.Errorf("invalid resource type value: %s", this)
	}
}

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

func ResourceTypeValidateRule(field any) *val.FieldRules {
	return val.Field(field,
		val.Required,
		val.OneOf(ResourceTypeNikkiApplication, ResourceTypeCustom),
	)
}

type ResourceScopeType entResource.ScopeType

const (
	ResourceScopeTypeOrg       = ResourceScopeType(entResource.ScopeTypeOrg)
	ResourceScopeTypeHierarchy = ResourceScopeType(entResource.ScopeTypeHierarchy)
	ResourceScopeTypePrivate   = ResourceScopeType(entResource.ScopeTypePrivate)
)

func (this ResourceScopeType) Validate() error {
	switch this {
	case ResourceScopeTypeOrg, ResourceScopeTypeHierarchy, ResourceScopeTypePrivate:
		return nil
	default:
		return errors.Errorf("invalid scope type value: %s", this)
	}
}

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

func ResourceScopeTypeValidateRule(field any) *val.FieldRules {
	return val.Field(field,
		val.Required,
		val.OneOf(ResourceScopeTypeOrg, ResourceScopeTypeHierarchy, ResourceScopeTypePrivate),
	)
}
