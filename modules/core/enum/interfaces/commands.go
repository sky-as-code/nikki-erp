package interfaces

import (
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
)

type CreateEnumCommand struct {
	// The name replacing "enum" to appear in all error messages
	EntityName string         `json:"entityName"`
	Label      model.LangJson `json:"label"`
	Type       string         `json:"type"`
	Value      *string        `json:"value"`
}

type CreateEnumResult = GetEnumResult

type UpdateEnumCommand struct {
	Id model.Id `param:"id" json:"id"`
	// The name replacing "enum" to appear in all error messages
	EntityName string          `json:"entityName"`
	Etag       model.Etag      `json:"etag"`
	Label      *model.LangJson `json:"label"`
	Value      *string         `json:"value"`
}

type UpdateEnumResult = GetEnumResult

type DeleteEnumCommand struct {
	Id *model.Id `json:"id" param:"id"`
	// The name replacing "enum" to appear in all error messages
	EntityName string  `json:"entityName"`
	Type       *string `json:"type" query:"type"`
}

func (this DeleteEnumCommand) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdPtrValidateRule(&this.Id, this.Type == nil),
		EnumTypeValidateRule(&this.Type, this.Id == nil),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type DeleteEnumResultData struct {
	Id           *model.Id `json:"id"`
	DeletedAt    time.Time `json:"deletedAt"`
	DeletedCount int       `json:"deletedCount"`
}

type DeleteEnumResult = crud.OpResult[*DeleteEnumResultData]

func ToCrudDeletionResult(src *DeleteEnumResult) *crud.DeletionResult {
	var data *crud.DeletionResultData
	if src.HasData {
		data = &crud.DeletionResultData{
			Id:        safe.GetVal(src.Data.Id, ""),
			DeletedAt: src.Data.DeletedAt,
		}
	}
	return &crud.DeletionResult{
		ClientError: src.ClientError,
		Data:        data,
		HasData:     src.HasData,
	}
}

type EnumExistsQuery struct {
	Id model.Id `param:"id" json:"id"`
	// The name replacing "enum" to appear in all error messages
	EntityName string `json:"entityName"`
}

func (this EnumExistsQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRule(&this.Id, true),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type EnumExistsResult = crud.OpResult[bool]

type EnumExistsMultiQuery struct {
	Ids []model.Id `json:"ids"`
	// The name replacing "enum" to appear in all error messages
	EntityName string `json:"entityName"`
}

func (this EnumExistsMultiQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdValidateRuleMulti(&this.Ids, true, 1, model.MODEL_RULE_ID_ARR_MAX),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type ExistsMultiResultData struct {
	Existing    []model.Id `json:"existing"`
	NotExisting []model.Id `json:"notExisting"`
}

type EnumExistsMultiResult = crud.OpResult[*ExistsMultiResultData]

type GetEnumQuery struct {
	Id *model.Id `param:"id" json:"id"`
	// The name replacing "enum" to appear in all error messages
	EntityName string  `json:"entityName"`
	Type       *string `json:"type" query:"type"`
	Value      *string `json:"value" query:"value"`
}

func (this GetEnumQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		model.IdPtrValidateRule(&this.Id, this.Type == nil && this.Value == nil),
		EnumTypeValidateRule(&this.Type, this.Id == nil),
		EnumValueValidateRule(&this.Value, this.Id == nil),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type GetEnumResult = crud.OpResult[*Enum]

type ListEnumsQuery struct {
	// The name replacing "enum" to appear in all error messages
	EntityName string `json:"entityName"`
	// Part of the Enum Label used for filtering
	PartialLabel *string             `json:"partialLabel" query:"partialLabel"`
	Page         *int                `json:"page" query:"page"`
	Size         *int                `json:"size" query:"size"`
	SortByLang   *model.LanguageCode `json:"sortByLang" query:"sortByLang"`
	Type         *string             `json:"type" query:"type"`
}

func (this *ListEnumsQuery) SetDefaults() {
	safe.SetDefaultValue(&this.Page, model.MODEL_RULE_PAGE_INDEX_START)
	safe.SetDefaultValue(&this.Size, model.MODEL_RULE_PAGE_DEFAULT_SIZE)
}

func (this ListEnumsQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		EnumTypeValidateRule(&this.Type, false),
		model.LanguageCodePtrValidateRule(&this.SortByLang, false),
		crud.PageIndexValidateRule(&this.Page),
		crud.PageSizeValidateRule(&this.Size),
		val.Field(&this.PartialLabel,
			val.Length(1, model.MODEL_RULE_TINY_NAME_LENGTH),
		),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type ListEnumsResultData = crud.PagedResult[Enum]
type ListEnumsResult = crud.OpResult[*ListEnumsResultData]

type SearchEnumsQuery struct {
	// The name replacing "enum" to appear in all error messages
	EntityName string  `json:"entityName"`
	Graph      *string `json:"graph" query:"graph"`
	Page       *int    `json:"page" query:"page"`
	Size       *int    `json:"size" query:"size"`
	TypePrefix *string `json:"typePrefix" query:"typePrefix"`
}

func (this *SearchEnumsQuery) SetDefaults() {
	safe.SetDefaultValue(&this.Page, model.MODEL_RULE_PAGE_INDEX_START)
	safe.SetDefaultValue(&this.Size, model.MODEL_RULE_PAGE_DEFAULT_SIZE)
}

func (this SearchEnumsQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		crud.PageIndexValidateRule(&this.Page),
		crud.PageSizeValidateRule(&this.Size),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}

type SearchEnumsResult = ListEnumsResult

type ListDerivedEnumsQuery struct {
	Page       *int                `json:"page" query:"page"`
	Size       *int                `json:"size" query:"size"`
	SortByLang *model.LanguageCode `json:"sortByLang" query:"sortByLang"`
}

func (this *ListDerivedEnumsQuery) SetDefaults() {
	safe.SetDefaultValue(&this.Page, model.MODEL_RULE_PAGE_INDEX_START)
	safe.SetDefaultValue(&this.Size, model.MODEL_RULE_PAGE_DEFAULT_SIZE)
}

func (this ListDerivedEnumsQuery) Validate() ft.ValidationErrors {
	rules := []*val.FieldRules{
		crud.PageIndexValidateRule(&this.Page),
		crud.PageSizeValidateRule(&this.Size),
		model.LanguageCodePtrValidateRule(&this.SortByLang, false),
	}

	return val.ApiBased.ValidateStruct(&this, rules...)
}
