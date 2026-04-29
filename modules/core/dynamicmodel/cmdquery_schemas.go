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
			RequiredAlways())
}

func GetOneQuerySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel("_").
		Field(dmodel.DefineField().
			Name(basemodel.FieldId).
			DataType(dmodel.FieldDataTypeUlid()).
			RequiredAlways()).
		Field(DefineFieldSearchColumns())
}

func ManageAssocsSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel("_").
		Field(dmodel.DefineField().
			Name(basemodel.FieldId).
			DataType(dmodel.FieldDataTypeUlid()).
			RequiredAlways()).
		Field(dmodel.DefineField().
			Name(basemodel.FieldAssociations).
			DataType(dmodel.FieldDataTypeUlid().ArrayType()).
			Rule(dmodel.FieldRuleArrayLength(0, 50))).
		Field(dmodel.DefineField().
			Name(basemodel.FieldDesociations).
			DataType(dmodel.FieldDataTypeUlid().ArrayType()).
			Rule(dmodel.FieldRuleArrayLength(0, 50)))
}

func SearchQuerySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel("_").
		Field(DefineFieldSearchColumns()).
		Field(DefineFieldSearchGraph()).
		Field(DefineFieldSearchPage()).
		Field(DefineFieldSearchSize())
}

func SetArchivedCommandSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.DefineModel("_").
		Field(dmodel.DefineField().
			Name(basemodel.FieldId).
			DataType(dmodel.FieldDataTypeUlid()).
			RequiredAlways()).
		Field(dmodel.DefineField().
			Name(basemodel.FieldEtag).
			DataType(dmodel.FieldDataTypeEtag()).
			VersioningKey()).
		Field(dmodel.DefineField().
			Name(basemodel.FieldIsArchived).
			DataType(dmodel.FieldDataTypeBoolean()).
			RequiredAlways())
}

func DefineFieldSearchColumns() *dmodel.FieldBuilder {
	return dmodel.DefineField().
		Name(basemodel.FieldColumns).
		DataType(dmodel.FieldDataTypeString(model.MODEL_RULE_COL_LENGTH_MIN, model.MODEL_RULE_COL_LENGTH_MAX).ArrayType()).
		Rule(dmodel.FieldRuleArrayLength(0, 20))
}

func DefineFieldSearchPage() *dmodel.FieldBuilder {
	return dmodel.DefineField().
		Name(basemodel.FieldPage).
		DataType(dmodel.FieldDataTypeInt32(model.MODEL_RULE_PAGE_INDEX_START, model.MODEL_RULE_PAGE_INDEX_END)).
		Default(model.MODEL_RULE_PAGE_INDEX_START)
}

func DefineFieldSearchSize() *dmodel.FieldBuilder {
	return dmodel.DefineField().
		Name(basemodel.FieldSize).
		DataType(dmodel.FieldDataTypeInt32(model.MODEL_RULE_PAGE_MIN_SIZE, model.MODEL_RULE_PAGE_MAX_SIZE)).
		Default(model.MODEL_RULE_PAGE_DEFAULT_SIZE)
}

func DefineFieldSearchGraph() *dmodel.FieldBuilder {
	return dmodel.DefineField().
		Name(basemodel.FieldGraph).
		DataType(dmodel.FieldDataTypeModel())
}
