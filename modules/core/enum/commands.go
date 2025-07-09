package enum

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
)

func init() {
	// Assert interface implementation
	var req cqrs.Request
	req = (*CreateEnumCommand)(nil)
	req = (*UpdateEnumCommand)(nil)
	req = (*DeleteEnumCommand)(nil)
	req = (*GetEnumQuery)(nil)
	req = (*ListEnumsQuery)(nil)
	req = (*EnumExistsCommand)(nil)
	req = (*EnumExistsMultiCommand)(nil)
	util.Unused(req)
}

var createEnumCommandType = cqrs.RequestType{
	Module:    "core",
	Submodule: "enum",
	Action:    "create",
}

type CreateEnumCommand struct {
	Label    model.LangJson `json:"label"`
	EnumType string         `json:"type"`
	Value    string         `json:"value"`
}

func (CreateEnumCommand) Type() cqrs.RequestType {
	return createEnumCommandType
}

type CreateEnumResult model.OpResult[*Enum]

var updateEnumCommandType = cqrs.RequestType{
	Module:    "core",
	Submodule: "enum",
	Action:    "update",
}

type UpdateEnumCommand struct {
	Id    model.Id        `param:"id" json:"id"`
	Etag  model.Etag      `json:"etag,omitempty"`
	Label *model.LangJson `json:"label"`
	Value *string         `json:"value"`
}

func (UpdateEnumCommand) Type() cqrs.RequestType {
	return updateEnumCommandType
}

type UpdateEnumResult model.OpResult[*Enum]

var deleteEnumCommandType = cqrs.RequestType{
	Module:    "core",
	Submodule: "enum",
	Action:    "delete",
}

type DeleteEnumCommand struct {
	Id       *model.Id `json:"id" param:"id"`
	EnumType *string   `json:"type" query:"type"`
}

func (DeleteEnumCommand) Type() cqrs.RequestType {
	return deleteEnumCommandType
}

func (this DeleteEnumCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdPtrValidateRule(&this.Id, this.EnumType == nil),
		EnumTypeValidateRule(&this.EnumType, this.Id == nil),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteEnumResultData struct {
	DeletedAt    time.Time `json:"deletedAt"`
	DeletedCount int       `json:"deletedCount"`
}

type DeleteEnumResult model.OpResult[*DeleteEnumResultData]

var existsCommandType = cqrs.RequestType{
	Module:    "core",
	Submodule: "enum",
	Action:    "exists",
}

type EnumExistsCommand struct {
	Id model.Id `param:"id" json:"id"`
}

func (EnumExistsCommand) Type() cqrs.RequestType {
	return existsCommandType
}

func (this EnumExistsCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type EnumExistsResult model.OpResult[bool]

var existsMultiCommandType = cqrs.RequestType{
	Module:    "core",
	Submodule: "enum",
	Action:    "existsMulti",
}

type EnumExistsMultiCommand struct {
	Ids []model.Id `json:"ids"`
}

func (EnumExistsMultiCommand) Type() cqrs.RequestType {
	return existsMultiCommandType
}

func (this EnumExistsMultiCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRuleMulti(&this.Ids, true, 1, model.MODEL_RULE_ID_ARR_MAX),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type ExistsMultiResultData struct {
	Existing    []model.Id `json:"existing"`
	NotExisting []model.Id `json:"notExisting"`
}

type EnumExistsMultiResult model.OpResult[*ExistsMultiResultData]

var getEnumByIdQueryType = cqrs.RequestType{
	Module:    "core",
	Submodule: "enum",
	Action:    "getEnumById",
}

type GetEnumQuery struct {
	Id       *model.Id `param:"id" json:"id"`
	EnumType *string   `json:"type" query:"type"`
	Value    *string   `json:"value" query:"value"`
}

func (GetEnumQuery) Type() cqrs.RequestType {
	return getEnumByIdQueryType
}

func (this GetEnumQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdPtrValidateRule(&this.Id, this.EnumType == nil && this.Value == nil),
		EnumTypeValidateRule(&this.EnumType, this.Id == nil),
		EnumValueValidateRule(&this.Value, this.Id == nil),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetEnumResult model.OpResult[*Enum]

var listEnumsCommandType = cqrs.RequestType{
	Module:    "core",
	Submodule: "enum",
	Action:    "list",
}

type ListEnumsQuery struct {
	EnumType     *string             `json:"type" query:"type"`
	Page         *int                `json:"page" query:"page"`
	Size         *int                `json:"size" query:"size"`
	SortedByLang *model.LanguageCode `json:"sortedByLang" query:"sortedByLang"`
}

func (ListEnumsQuery) Type() cqrs.RequestType {
	return listEnumsCommandType
}

func (this *ListEnumsQuery) SetDefaults() {
	safe.SetDefaultValue(&this.Page, model.MODEL_RULE_PAGE_INDEX_START)
	safe.SetDefaultValue(&this.Size, model.MODEL_RULE_PAGE_DEFAULT_SIZE)
}

func (this ListEnumsQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		EnumTypeValidateRule(&this.EnumType, false),
		model.LanguageCodeValidateRule(&this.SortedByLang, false),
		model.PageIndexValidateRule(&this.Page),
		model.PageSizeValidateRule(&this.Size),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type ListEnumsResultData = crud.PagedResult[Enum]
type ListEnumsResult model.OpResult[*ListEnumsResultData]
