package attributevalue

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/product/domain"
)

// Create

var createAttributeValueCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute_value",
	Action:    "create",
}

type CreateAttributeValueCommand struct {
	VariantId    model.Id        `json:"variantId,omitempty"`
	AttributeId  model.Id        `param:"attribute_id" json:"attributeId"`
	ValueText    *model.LangJson `json:"valueText,omitempty"`
	ValueNumber  *float64        `json:"valueNumber,omitempty"`
	ValueBool    *bool           `json:"valueBool,omitempty"`
	ValueRef     *string         `json:"valueRef,omitempty"`
	VariantCount *int            `json:"variantCount,omitempty"`
}

func (CreateAttributeValueCommand) CqrsRequestType() cqrs.RequestType {
	return createAttributeValueCommandType
}

func (this CreateAttributeValueCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.AttributeId, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

type CreateAttributeValueResult = GetAttributeValueByIdResult

// Update

var updateAttributeValueCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute_value",
	Action:    "update",
}

type UpdateAttributeValueCommand struct {
	Id           model.Id        `param:"id" json:"id"`
	Etag         model.Etag      `json:"etag"`
	ValueText    *model.LangJson `json:"valueText,omitempty"`
	ValueNumber  *float64        `json:"valueNumber,omitempty"`
	ValueBool    *bool           `json:"valueBool,omitempty"`
	ValueRef     *string         `json:"valueRef,omitempty"`
	VariantCount *int            `json:"variantCount,omitempty"`
}

func (UpdateAttributeValueCommand) CqrsRequestType() cqrs.RequestType {
	return updateAttributeValueCommandType
}

type UpdateAttributeValueResult = GetAttributeValueByIdResult

// Delete

var deleteAttributeValueCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute_value",
	Action:    "delete",
}

type DeleteAttributeValueCommand struct {
	Id model.Id `json:"id" param:"id"`
}

func (this DeleteAttributeValueCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return val.ApiBased.ValidateStruct(this, rules...)
}

func (DeleteAttributeValueCommand) CqrsRequestType() cqrs.RequestType {
	return deleteAttributeValueCommandType
}

type DeleteAttributeValueResult = crud.DeletionResult

// Get by ID

var getAttributeValueByIdQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute_value",
	Action:    "getById",
}

type GetAttributeValueByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (this GetAttributeValueByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

func (GetAttributeValueByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getAttributeValueByIdQueryType
}

type GetAttributeValueByIdResult = crud.OpResult[*domain.AttributeValue]

var searchAttributeValuesQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "attribute_value",
	Action:    "search",
}

// Search (advanced)

type SearchAttributeValuesQuery struct {
	// Filled by service from Graph
	crud.SearchQuery
}

func (this SearchAttributeValuesQuery) CqrsRequestType() cqrs.RequestType {
	return searchAttributeValuesQueryType
}

func (this SearchAttributeValuesQuery) Validate() ft.ValidationErrors {
	rules := this.SearchQuery.ValidationRules()

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type SearchAttributeValuesResultData = crud.PagedResult[domain.AttributeValue]
type SearchAttributeValuesResult = crud.OpResult[*SearchAttributeValuesResultData]
