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
		Field(DefineFieldIncludeArchived()).
		Field(DefineFieldSearchContext()).
		Field(DefineFieldSearchLanguage())
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

// DefineFieldSearchContext defines the "context" search query field: the whitelisted request
// values SQL-computed fields may bind inside their filters. It is a free-form JSON object here;
// the per-definition whitelist and type conversion are enforced where the subquery compiles.
func DefineFieldSearchContext() *dmodel.FieldBuilder {
	return dmodel.DefineField().
		Name(basemodel.FieldContext).
		DataType(dmodel.FieldDataTypeJsonMap())
}

func DefineFieldSearchName() *dmodel.FieldBuilder {
	return dmodel.DefineField().
		Name(basemodel.FieldSearchName).
		DataType(dmodel.FieldDataTypeString(1, model.MODEL_RULE_TINY_NAME_LENGTH, dmodel.FieldDataTypeStringOpts{
			Regex: regexp.MustCompile(`^[a-zA-Z0-9_\.]+$`),
		}))
}

// DefineFieldSearchLanguage defines the "language" search query field: which translation of a
// LangJson column to filter and sort on.
//
// Declared here rather than left undeclared because an undeclared field survives validation only by
// accident -- ValidateStruct mutates the caller's own struct, and the mapper happens not to zero
// what the schema did not mention. A refactor to build a fresh target would silently drop it and
// un-localize every search, with no compile error and no failing test.
//
// The rule is BCP47 shape rather than an enum of the supported locales: that set lives in the
// essential module, which this package must not import. A well-formed locale the application ships
// no translations for simply matches nothing, which is the same answer it would give anyway.
func DefineFieldSearchLanguage() *dmodel.FieldBuilder {
	return dmodel.DefineField().
		Name(basemodel.FieldLanguage).
		DataType(dmodel.FieldDataTypeString(2, 35, dmodel.FieldDataTypeStringOpts{
			Regex: regexp.MustCompile(`^[a-zA-Z]{2,3}-[a-zA-Z0-9]{2,8}$`),
		}))
}
