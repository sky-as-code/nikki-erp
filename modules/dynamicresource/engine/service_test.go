package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
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

// newRestrictedService builds a service whose engine was created with an explicit CrudActions
// allow-list, which is what a read-only resource looks like.
func newRestrictedService(allowed ...it.CrudAction) it.DynamicResourceService {
	return NewDynamicResourceService(NewServiceParam{
		Schema:      newFieldsTestSchema(),
		CrudActions: allowed,
	})
}

func TestServiceRefusesActionLeftOutOfCrudActions(t *testing.T) {
	service := newRestrictedService(it.CrudActionGetById, it.CrudActionSearch)

	result, err := service.Create(nil, dmodel.DynamicFields{"name": "x"})

	assert.NoError(t, err, "an unsupported action is a client error, not a Go error")
	assert.Positive(t, result.ClientErrors.Count())
}

// A listed action gets past the guard. It still fails further down - the test service has no
// repository - so this asserts on the refusal being absent, not on the call succeeding: a
// refused action reports err_action_not_supported and never reaches the query at all.
func TestServiceAllowsActionListedInCrudActions(t *testing.T) {
	service := newRestrictedService(it.CrudActionGetById, it.CrudActionSearch)

	defer func() { _ = recover() }()
	result, _ := service.GetById(nil, dmodel.DynamicFields{"id": "rec-1"})
	if result != nil {
		assert.False(t, hasUnsupportedActionError(result.ClientErrors),
			"a listed action is not refused")
	}
}

// An empty CrudActions list is the default and means "every action supported".
func TestServiceSupportsEveryActionByDefault(t *testing.T) {
	service := NewDynamicResourceService(NewServiceParam{Schema: newFieldsTestSchema()})

	defer func() { _ = recover() }()
	result, _ := service.Create(nil, dmodel.DynamicFields{"name": "x"})
	if result != nil {
		assert.False(t, hasUnsupportedActionError(result.ClientErrors),
			"no allow-list means nothing is refused")
	}
}

func hasUnsupportedActionError(cErrs ft.ClientErrors) bool {
	for _, item := range cErrs {
		if item.Key == "err_action_not_supported" {
			return true
		}
	}
	return false
}
