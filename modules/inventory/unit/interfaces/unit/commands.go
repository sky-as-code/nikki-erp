package unit

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/inventory/unit/domain"
)

// Create

var createUnitCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit",
	Action:    "create",
}

type CreateUnitCommand struct {
	BaseUnit   *string        `json:"baseUnit,omitempty"`
	CategoryId *model.Id      `json:"categoryId,omitempty"`
	Multiplier *int           `json:"multiplier,omitempty"`
	Name       model.LangJson `json:"name"`
	OrgId      *model.Id      `param:"orgId" json:"orgId"`
	Status     *string        `json:"status,omitempty"`
	Symbol     string         `json:"symbol"`
}

func (CreateUnitCommand) CqrsRequestType() cqrs.RequestType {
	return createUnitCommandType
}

type CreateUnitResult = GetUnitByIdResult

// Update

var updateUnitCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit",
	Action:    "update",
}

type UpdateUnitCommand struct {
	Id         model.Id        `param:"id" json:"id"`
	BaseUnit   *string         `json:"baseUnit,omitempty"`
	CategoryId *model.Id       `json:"categoryId,omitempty"`
	Etag       model.Etag      `json:"etag"`
	Multiplier *int            `json:"multiplier,omitempty"`
	Name       *model.LangJson `json:"name,omitempty"`
	OrgId      *model.Id       `json:"orgId,omitempty"`
	Status     *string         `json:"status,omitempty"`
	Symbol     *string         `json:"symbol,omitempty"`
}

func (UpdateUnitCommand) CqrsRequestType() cqrs.RequestType {
	return updateUnitCommandType
}

type UpdateUnitResult = GetUnitByIdResult

// Delete

var deleteUnitCommandType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit",
	Action:    "delete",
}

type DeleteUnitCommand struct {
	Id    model.Id `json:"id" param:"id"`
	OrgId model.Id `json:"orgId"`
}

func (this DeleteUnitCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

func (DeleteUnitCommand) CqrsRequestType() cqrs.RequestType {
	return deleteUnitCommandType
}

type DeleteUnitResult = crud.DeletionResult

// Get by ID

var getUnitByIdQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit",
	Action:    "getById",
}

type GetUnitByIdQuery struct {
	Id model.Id `param:"id" json:"id"`
}

func (this GetUnitByIdQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}
	return val.ApiBased.ValidateStruct(&this, rules...)
}

func (GetUnitByIdQuery) CqrsRequestType() cqrs.RequestType {
	return getUnitByIdQueryType
}

type GetUnitByIdResult = crud.OpResult[*domain.Unit]

var searchUnitsQueryType = cqrs.RequestType{
	Module:    "inventory",
	Submodule: "unit",
	Action:    "search",
}

// Search (advanced)

type SearchUnitsQuery struct {
	// Filled by service from Graph
	crud.SearchQuery
}

func (this SearchUnitsQuery) CqrsRequestType() cqrs.RequestType {
	return searchUnitsQueryType
}

func (this SearchUnitsQuery) Validate() ft.ValidationErrors {
	rules := this.SearchQuery.ValidationRules()

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type SearchUnitsResultData = crud.PagedResult[domain.Unit]
type SearchUnitsResult = crud.OpResult[*SearchUnitsResultData]
