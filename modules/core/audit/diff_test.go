package audit

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
)

func auditTestSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema("core.audit_diff_test", func() *dmodel.ModelSchemaBuilder {
		return dmodel.DefineModel("core.audit_diff_test").
			Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeUlid()).PrimaryKey()).
			Field(dmodel.DefineField().Name("name").DataType(dmodel.FieldDataTypeString(0, 100))).
			Field(dmodel.DefineField().Name("qty").DataType(dmodel.FieldDataTypeInt64(0, 1000))).
			Field(dmodel.DefineField().Name("hash").DataType(dmodel.FieldDataTypeSecret(0, 100))).
			Field(dmodel.DefineField().Name("token").DataType(dmodel.FieldDataTypeString(0, 100)).
				Metadata(map[string]any{MetadataAuditExcluded: true})).
			Field(dmodel.DefineField().Name("status").DataType(dmodel.FieldDataTypeString(0, 50)).
				Metadata(map[string]any{MetadataTrackHistory: true})).
			Field(dmodel.DefineField().Name("etag").DataType(dmodel.FieldDataTypeEtag()).VersioningKey()).
			Field(dmodel.DefineField().Name("updated_at").DataType(dmodel.FieldDataTypeDateTime()))
	})
}

func TestDiffUpdateKeepsOnlyChangedFields(t *testing.T) {
	before := dmodel.DynamicFields{"id": "01A", "name": "old", "qty": int64(5), "etag": "1"}
	after := dmodel.DynamicFields{"id": "01A", "name": "new", "qty": 5, "etag": "2", "updated_at": "now"}

	oldValues, newValues := Diff(auditTestSchema(), before, after)

	assert.Equal(t, dmodel.DynamicFields{"name": "old"}, oldValues)
	assert.Equal(t, dmodel.DynamicFields{"name": "new"}, newValues)
}

func TestDiffUpdateWithNoChangeIsEmpty(t *testing.T) {
	before := dmodel.DynamicFields{"name": "same", "qty": "10.50"}
	after := dmodel.DynamicFields{"name": "same", "qty": 10.5}

	oldValues, newValues := Diff(auditTestSchema(), before, after)

	assert.True(t, Entry{Before: oldValues, After: newValues}.IsEmpty())
}

func TestDiffExcludesSecretAndFlaggedFields(t *testing.T) {
	after := dmodel.DynamicFields{"name": "n", "hash": "h", "token": "t", "unknown": "u"}

	_, newValues := Diff(auditTestSchema(), nil, after)

	assert.Equal(t, dmodel.DynamicFields{"name": "n"}, newValues)
}

func TestDiffCreateAndDeleteSkipNulls(t *testing.T) {
	row := dmodel.DynamicFields{"id": "01A", "name": "n", "qty": nil}

	oldOnCreate, newOnCreate := Diff(auditTestSchema(), nil, row)
	oldOnDelete, newOnDelete := Diff(auditTestSchema(), row, nil)

	assert.Empty(t, oldOnCreate)
	assert.Equal(t, dmodel.DynamicFields{"name": "n"}, newOnCreate)
	assert.Equal(t, dmodel.DynamicFields{"name": "n"}, oldOnDelete)
	assert.Empty(t, newOnDelete)
}

func TestSameValueAcrossGoTypes(t *testing.T) {
	var nilPtr *string
	assert.True(t, SameValue(model.Id("01A"), "01A"))
	assert.True(t, SameValue(int64(3), 3))
	assert.True(t, SameValue(nil, nilPtr))
	assert.False(t, SameValue(nil, ""))
	assert.False(t, SameValue("a", "b"))
}

func TestIsExcludedFieldWithoutSchemaUsesNameList(t *testing.T) {
	assert.True(t, IsExcludedField(nil, "tenant_id"))
	assert.True(t, IsExcludedField(nil, "created_by"))
	assert.False(t, IsExcludedField(nil, "name"))
}

func TestDiffTrackedKeepsOnlyTrackedFields(t *testing.T) {
	schema := auditTestSchema()
	before := dmodel.DynamicFields{"id": "01A", "name": "old", "status": "draft"}
	after := dmodel.DynamicFields{"id": "01A", "name": "new", "status": "confirmed"}

	oldValues, newValues := DiffTracked(schema, before, after)

	assert.Equal(t, dmodel.DynamicFields{"status": "draft"}, oldValues)
	assert.Equal(t, dmodel.DynamicFields{"status": "confirmed"}, newValues)
}

func TestDiffTrackedIsEmptyWhenOnlyUntrackedFieldsChanged(t *testing.T) {
	before := dmodel.DynamicFields{"id": "01A", "name": "old", "status": "draft"}
	after := dmodel.DynamicFields{"id": "01A", "name": "new", "status": "draft"}

	oldValues, newValues := DiffTracked(auditTestSchema(), before, after)

	assert.Empty(t, oldValues)
	assert.Empty(t, newValues)
}

func TestTrackedFieldLookups(t *testing.T) {
	schema := auditTestSchema()

	assert.True(t, HasTrackedField(schema))
	assert.True(t, TouchesTrackedField(schema, dmodel.DynamicFields{"id": "01A", "status": "x"}))
	assert.False(t, TouchesTrackedField(schema, dmodel.DynamicFields{"id": "01A", "name": "x"}))
	assert.False(t, IsTrackedField(schema, "unknown"))
	assert.False(t, HasTrackedField(nil))
}
