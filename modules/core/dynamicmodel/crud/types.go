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

// type SearchQuery interface {
// 	GetOneQuery
// 	GetGraph() dmodel.SearchGraph
// 	GetPage() int
// 	GetSize() int
// }

type SearchQuery struct {
	// fields dmodel.DynamicFields
	Columns         []string            `json:"columns" query:"columns"`
	Graph           *dmodel.SearchGraph `json:"graph" query:"graph"`
	IncludeArchived *bool               `json:"include_archived" query:"include_archived"`
	Page            *int                `json:"page" query:"page"`
	Size            *int                `json:"size" query:"size"`
	Relations       []string            `json:"relations" query:"relations"`
}

// func (this SearchQuery) Columns() []string {
// 	return this.fields.GetStrings(basemodel.FieldColumns)
// }

// func (this SearchQuery) IncludeArchived() *bool {
// 	return this.fields.GetBool(basemodel.FieldIncludeArchived)
// }

// func (this SearchQuery) Graph() *dmodel.SearchGraph {
// 	val := this.fields.GetAny(basemodel.FieldGraph)
// 	graph, ok := val.(dmodel.SearchGraph)
// 	if ok {
// 		return &graph
// 	}
// 	return nil
// }

// func (this SearchQuery) Page() *int {
// 	return this.fields.GetInt(basemodel.FieldPage)
// }

// func (this SearchQuery) Size() *int {
// 	return this.fields.GetInt(basemodel.FieldSize)
// }

func (this SearchQuery) GetFieldData() dmodel.DynamicFields {
	fields := dmodel.DynamicFields{
		basemodel.FieldColumns:         this.Columns,
		basemodel.FieldGraph:           this.Graph,
		basemodel.FieldIncludeArchived: this.IncludeArchived,
		basemodel.FieldPage:            this.Page,
		basemodel.FieldSize:            this.Size,
	}
	return fields
}

func (this SearchQuery) SetFieldData(fields dmodel.DynamicFields) {
	this.Columns = fields.GetStrings(basemodel.FieldColumns)
	this.IncludeArchived = fields.GetBool(basemodel.FieldIncludeArchived)
	this.Page = fields.GetInt(basemodel.FieldPage)
	this.Size = fields.GetInt(basemodel.FieldSize)
	// this.Graph = fields.GetAny(basemodel.FieldGraph).(dmodel.SearchGraph)
}

func (this SearchQuery) GetSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		dmodel.DefineModel("core.crud.search_query").
			Extend(SearchQuerySchemaBuilder()).
			Build(),
	)
}

// func (this SearchQueryBase) DeleteFieldData(fields *dmodel.DynamicFields) {
// 	delete(*fields, basemodel.FieldColumns)
// 	delete(*fields, basemodel.FieldIncludeArchived)
// 	delete(*fields, basemodel.FieldGraph)
// 	delete(*fields, basemodel.FieldPage)
// 	delete(*fields, basemodel.FieldSize)
// }

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
