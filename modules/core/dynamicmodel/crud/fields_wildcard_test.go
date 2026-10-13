package crud

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// wildcardSchema carries two columns, one computed field and one edge, so the test can tell that
// only the edge is left out of the expansion.
func wildcardSchema() *dmodel.ModelSchema {
	return dmodel.DefineModel("test_crud_wildcard").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
		Field(dmodel.DefineField().Name("sku").DataType(dmodel.FieldDataTypeString(1, 100))).
		Field(dmodel.DefineField().Name("template_name").DataType(dmodel.FieldDataTypeString(0, 200)).
			Computed(false, computed.Related("template.name"))).
		EdgeTo(dmodel.Edge("template").ManyToOne("test_crud_wildcard_peer", dmodel.DynamicFields{"id": "id"})).
		Build()
}

func TestExpandFieldsWildcard_LeavesSelectionWithoutWildcardUntouched(t *testing.T) {
	fields := []string{"sku", "template.name"}

	assert.Equal(t, fields, ExpandFieldsWildcard(wildcardSchema(), fields))
	assert.Nil(t, ExpandFieldsWildcard(wildcardSchema(), nil))
}

func TestExpandFieldsWildcard_ExpandsToColumnsAndComputedFields(t *testing.T) {
	got := ExpandFieldsWildcard(wildcardSchema(), []string{"*"})

	assert.Equal(t, []string{"id", "sku", "template_name"}, got)
}

func TestExpandFieldsWildcard_KeepsEdgePathsInPlace(t *testing.T) {
	got := ExpandFieldsWildcard(wildcardSchema(), []string{"*", "template.name", "template"})

	assert.Equal(t, []string{"id", "sku", "template_name", "template.name", "template"}, got)
}

func TestExpandFieldsWildcard_DropsDuplicates(t *testing.T) {
	got := ExpandFieldsWildcard(wildcardSchema(), []string{"sku", "*", "*", "template.name", "sku"})

	assert.Equal(t, []string{"sku", "id", "template_name", "template.name"}, got)
}

// The token survives for the query builder to render as `t0.*`; computed fields are named so the
// computed-field layer evaluates them, and the columns `*` already covers are dropped.
func TestNormalizeFieldsWildcard_KeepsTokenAndNamesComputedFields(t *testing.T) {
	got := NormalizeFieldsWildcard(wildcardSchema(), []string{"sku", "*", "template.name"})

	assert.Equal(t, []string{"*", "template.name", "template_name"}, got)
}

func TestNormalizeFieldsWildcard_IsIdempotent(t *testing.T) {
	once := NormalizeFieldsWildcard(wildcardSchema(), []string{"*", "template"})

	assert.Equal(t, once, NormalizeFieldsWildcard(wildcardSchema(), once))
}

func TestNormalizeFieldsWildcard_LeavesSelectionWithoutWildcardUntouched(t *testing.T) {
	fields := []string{"sku", "template_name"}

	assert.Equal(t, fields, NormalizeFieldsWildcard(wildcardSchema(), fields))
}
