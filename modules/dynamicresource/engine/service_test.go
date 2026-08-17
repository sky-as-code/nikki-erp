package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
)

// newFieldsTestSchema builds a schema with three column-backed fields, so that a test
// can tell a configured field list apart from the all-columns fallback.
func newFieldsTestSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel("test_default_fields").
		Field(
			dmodel.DefineField().
				Name("name").
				DataType(dmodel.FieldDataTypeLangJson(1, model.MODEL_RULE_LONG_NAME_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name("description").
				DataType(dmodel.FieldDataTypeLangJson(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Field(
			dmodel.DefineField().
				Name("secret").
				DataType(dmodel.FieldDataTypeLangJson(0, model.MODEL_RULE_DESC_LENGTH)),
		).
		Build()
}

func TestNewServiceUsesConfiguredDefaultFields(t *testing.T) {
	schema := newFieldsTestSchema()

	service := NewDynamicResourceService(NewServiceParam{
		Schema:        schema,
		DefaultFields: []string{"name", "description"},
	}).(*DynamicResourceServiceImpl)

	assert.Equal(t, []string{"name", "description"}, service.defaultFields)
	assert.NotContains(t, service.defaultFields, "secret")
}

func TestNewServiceFallsBackToAllColumns(t *testing.T) {
	schema := newFieldsTestSchema()

	service := NewDynamicResourceService(NewServiceParam{
		Schema: schema,
	}).(*DynamicResourceServiceImpl)

	assert.Equal(t, columnNames(schema), service.defaultFields)
	assert.Contains(t, service.defaultFields, "secret")
}

func TestParamsToSearchQuery_DecodesComputedContext(t *testing.T) {
	query, err := paramsToSearchQuery(dmodel.DynamicFields{
		"fields":  []string{"id", "available_qty"},
		"context": map[string]any{"warehouse_id": "wh1", "company_id": "co1"},
	})
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"warehouse_id": "wh1", "company_id": "co1"}, query.Context)
}

func TestParamsToSearchQuery_NoContextStaysNil(t *testing.T) {
	query, err := paramsToSearchQuery(dmodel.DynamicFields{"fields": []string{"id"}})
	assert.NoError(t, err)
	assert.Nil(t, query.Context)
}
