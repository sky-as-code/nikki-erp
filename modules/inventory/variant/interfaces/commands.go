package interfaces

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

// Create

var createVariantCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "variant",
	Action:    "create",
}

type CreateVariantCommand struct {
	ProductId     model.Id `param:"product_id" json:"productId"`
	Sku           string   `json:"sku"`
	Barcode       *string  `json:"barcode,omitempty"`
	ProposedPrice *int     `json:"proposedPrice,omitempty"`
	Status        *string  `json:"status,omitempty"`
}

func (CreateVariantCommand) CqrsRequestType() cqrs.RequestType {
	return createVariantCommandType
}

type CreateVariantResult = GetVariantByIdResult

// Update

var updateVariantCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "variant",
	Action:    "update",
}

type UpdateVariantCommand struct {
	Id            model.Id   `param:"id" json:"id"`
	Etag          model.Etag `json:"etag"`
	Barcode       *string    `json:"barcode,omitempty"`
	ProposedPrice *int       `json:"proposedPrice,omitempty"`
	Status        *string    `json:"status,omitempty"`
}

func (UpdateVariantCommand) CqrsRequestType() cqrs.RequestType {
	return updateVariantCommandType
}

type UpdateVariantResult = GetVariantByIdResult

// Delete

var deleteVariantCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "variant",
	Action:    "delete",
}

type DeleteVariantCommand struct {
	Id model.Id `json:"id" param:"id"`
}

func (this DeleteVariantCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return val.ApiBased.ValidateStruct(this, rules...)
}

func (DeleteVariantCommand) CqrsRequestType() cqrs.RequestType {
	return deleteVariantCommandType
}

type DeleteVariantResult = crud.DeletionResult

// Get by ID

var getVariantByIdQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "variant",
	Action:    "getById",
}

type GetVariantByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (this GetVariantByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

func (GetVariantByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getVariantByIdQueryType
}

type GetVariantByIdResult = crud.OpResult[*Variant]

var searchVariantsQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "variant",
	Action:    "search",
}

// Search (advanced)

type SearchVariantsQuery struct {
	// Filled by service from Graph
	crud.SearchQuery
}

func (this SearchVariantsQuery) CqrsRequestType() cqrs.RequestType {
	return searchVariantsQueryType
}

func (this SearchVariantsQuery) Validate() ft.ValidationErrors {
	rules := this.SearchQuery.ValidationRules()

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type SearchVariantsResultData = crud.PagedResult[Variant]
type SearchVariantsResult = crud.OpResult[*SearchVariantsResultData]
