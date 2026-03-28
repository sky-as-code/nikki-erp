package crud

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coredyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

type SetIsArchivedCommand struct {
	Id         *model.Id   `json:"id" param:"id"`
	Etag       *model.Etag `json:"etag" param:"etag"`
	IsArchived *bool       `json:"is_archived" param:"is_archived"`
}

type ToDomainModelFunc[TDomain any] func(data dmodel.DynamicFields) TDomain
type BeforeValidationFunc[TDomain any] func(ctx corectx.Context, model TDomain) (TDomain, error)
type AfterValidationFunc[TDomain any] func(ctx corectx.Context, model TDomain) (TDomain, error)
type ValidateExtraFunc[TDomain any] func(ctx corectx.Context, model TDomain, vErrs *ft.ClientErrors) error

type CreateParam[
	TDomain any,
	TDomainPtr coredyn.DynamicModelPtr[TDomain],
] struct {
	// Action name for logging and error messages
	Action         string
	BaseRepoGetter coredyn.BaseRepoGetter

	// Data to create
	Data dmodel.DynamicModelGetter

	// Optional function to do some processing on the domain model before validation.
	BeforeValidation BeforeValidationFunc[TDomainPtr]

	// Optional function to do some processing on the domain model after validation.
	AfterValidation AfterValidationFunc[TDomainPtr]

	// Optional function for advanced validation (business rules) in addition to built-in schema validation.
	ValidateExtra ValidateExtraFunc[TDomainPtr]
}

type DeleteOneQuery struct {
	Id *model.Id `json:"id" param:"id"`
}

type GetOneQuery struct {
	Id      *model.Id `json:"id" param:"id"`
	Columns []string  `json:"columns" query:"columns"`
}

type SearchQuery struct {
	Columns []string            `json:"columns" query:"columns"`
	Graph   *dmodel.SearchGraph `json:"graph" query:"graph"`
	Page    *int                `json:"page" query:"page"`
	Size    *int                `json:"size" query:"size"`
}

func DeleteOneQuerySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel("_").
		Field(dmodel.DefineField().
			Name(basemodel.FieldId).
			DataType(dmodel.FieldDataTypeUlid()))
}

func GetOneQuerySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel("_").
		Field(dmodel.DefineField().
			Name(basemodel.FieldId).
			DataType(dmodel.FieldDataTypeUlid())).
		Field(dmodel.DefineField().
			Name(basemodel.FieldColumns).
			DataType(dmodel.FieldDataTypeString(model.MODEL_RULE_COL_LENGTH_MIN, model.MODEL_RULE_COL_LENGTH_MAX).ArrayType()))
}

func SearchQuerySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel("_").
		Field(dmodel.DefineField().
			Name(basemodel.FieldColumns).
			DataType(dmodel.FieldDataTypeString(model.MODEL_RULE_COL_LENGTH_MIN, model.MODEL_RULE_COL_LENGTH_MAX).ArrayType()).
			Rule(dmodel.FieldRuleArrayLength(0, 20))).
		Field(dmodel.DefineField().
			Name(basemodel.FieldPage).
			DataType(dmodel.FieldDataTypeInteger()).
			Default(model.MODEL_RULE_PAGE_INDEX_START)).
		Field(dmodel.DefineField().
			Name(basemodel.FieldSize).
			DataType(dmodel.FieldDataTypeInteger()).
			Default(model.MODEL_RULE_PAGE_DEFAULT_SIZE))
}

func SetArchivedCommandSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel("_").
		Field(dmodel.DefineField().
			Name(basemodel.FieldId).
			DataType(dmodel.FieldDataTypeUlid()).
			Required()).
		Field(dmodel.DefineField().
			Name(basemodel.FieldEtag).
			DataType(dmodel.FieldDataTypeEtag()).
			VersioningKey()).
		Field(dmodel.DefineField().
			Name(basemodel.FieldIsArchived).
			DataType(dmodel.FieldDataTypeBoolean()).
			Required())
}
