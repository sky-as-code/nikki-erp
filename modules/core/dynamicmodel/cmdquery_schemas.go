package dynamicmodel

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

func DeleteOneQuerySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel("_").
		Field(dmodel.DefineField().
			Name(basemodel.FieldId).
			DataType(dmodel.FieldDataTypeUlid()))
}

func ExistsQuerySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel("_").
		Field(dmodel.DefineField().
			Name("ids").
			DataType(dmodel.FieldDataTypeUlid().ArrayType()).
			Rule(dmodel.FieldRuleArrayLength(1, 50)).
			Required())
}

func GetOneQuerySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel("_").
		Field(dmodel.DefineField().
			Name(basemodel.FieldId).
			DataType(dmodel.FieldDataTypeUlid()).
			Required()).
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
