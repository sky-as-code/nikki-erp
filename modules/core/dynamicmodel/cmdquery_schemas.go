package dynamicmodel

import (
	"regexp"

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
		Field(DefineFieldSearchFields())
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
		Field(DefineFieldSearchFields()).
		Field(DefineFieldSearchGraph()).
		Field(DefineFieldSearchPage()).
		Field(DefineFieldSearchSize()).
		Field(DefineFieldSearchName()).
		Field(DefineFieldIncludeArchived())
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

func DefineFieldSearchFields() *dmodel.FieldBuilder {
	return dmodel.DefineField().
		Name(basemodel.FieldFields).
		DataType(dmodel.FieldDataTypeString(model.MODEL_RULE_FIELDS_LENGTH_MIN, model.MODEL_RULE_FIELDS_LENGTH_MAX).ArrayType()).
		Rule(dmodel.FieldRuleArrayLength(0, 50))
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

// DefineFieldIncludeArchived defines the "include_archived" search query field.
//
// It deliberately carries no Default(false): ModelField.Validate materializes a non-empty
// default into the *bool, which would erase the nil/false distinction before crud.Search
// ever sees it. The public-API default is applied in crud.Search instead, so that callers
// going straight to the repository keep the unfiltered legacy behaviour.
func DefineFieldIncludeArchived() *dmodel.FieldBuilder {
	return dmodel.DefineField().
		Name(basemodel.FieldIncludeArchived).
		DataType(dmodel.FieldDataTypeBoolean())
}

func DefineFieldSearchName() *dmodel.FieldBuilder {
	return dmodel.DefineField().
		Name(basemodel.FieldSearchName).
		DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH, dmodel.FieldDataTypeStringOpts{
			Regex: regexp.MustCompile(`^[a-zA-Z0-9_\.]+$`),
		}))
}
