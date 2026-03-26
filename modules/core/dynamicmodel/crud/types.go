package crud

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

type GetOneQuery interface {
	dmodel.SchemaGetter
	GetIncludeArchived() bool
	GetColumns() []string
	DeleteFieldData(fields *dmodel.DynamicFields)
}

type GetOneQueryBase struct {
	IncludeArchived bool     `json:"include_archived" query:"include_archived"`
	Columns         []string `json:"columns" query:"columns"`
}

// Implements GetOneQuery interface
func (this GetOneQueryBase) GetIncludeArchived() bool {
	return this.IncludeArchived
}

// Implements GetOneQuery interface
func (this GetOneQueryBase) GetColumns() []string {
	return this.Columns
}

func (this GetOneQueryBase) GetFieldData() dmodel.DynamicFields {
	return dmodel.DynamicFields{
		basemodel.FieldIncludeArchived: this.IncludeArchived,
		basemodel.FieldColumns:         this.Columns,
	}
}

func (this GetOneQueryBase) DeleteFieldData(fields *dmodel.DynamicFields) {
	delete(*fields, basemodel.FieldIncludeArchived)
	delete(*fields, basemodel.FieldColumns)
}

type SearchQuery struct {
	Columns         []string            `json:"columns" query:"columns"`
	Graph           *dmodel.SearchGraph `json:"graph" query:"graph"`
	IncludeArchived *bool               `json:"include_archived" query:"include_archived"`
	Page            *int                `json:"page" query:"page"`
	Size            *int                `json:"size" query:"size"`
	Relations       []string            `json:"relations" query:"relations"`
}

func (this SearchQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		dmodel.DefineModel("core.crud.search_query").
			Extend(SearchQuerySchemaBuilder()).
			Build(),
	)
}

func GetOneQuerySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel("core.crud.get_one_query").
		Field(dmodel.DefineField().
			Name(basemodel.FieldIncludeArchived).
			DataType(dmodel.FieldDataTypeBoolean()).
			Default(false)).
		Field(dmodel.DefineField().
			Name(basemodel.FieldColumns).
			DataType(dmodel.FieldDataTypeString(model.MODEL_RULE_COL_LENGTH_MIN, model.MODEL_RULE_COL_LENGTH_MAX).ArrayType()))
}

func SearchQuerySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel("core.crud.search_query").
		Field(dmodel.DefineField().
			Name(basemodel.FieldIncludeArchived).
			DataType(dmodel.FieldDataTypeBoolean()).
			Default(false)).
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
