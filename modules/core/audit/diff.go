package audit

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// systemFields are written by the platform on every save, so their change says nothing about
// what the actor did. org_id never changes and is the entry's own column.
var systemFields = map[string]bool{
	basemodel.FieldOrgId:     true,
	basemodel.FieldId:        true,
	"tenant_id":              true,
	basemodel.FieldEtag:      true,
	basemodel.FieldCreatedAt: true,
	basemodel.FieldUpdatedAt: true,
	basemodel.FieldCreatedBy: true,
}

// IsExcludedField reports a field that never appears in a diff: a system field, a key, a column
// the schema does not store, a `secret` value, or one flagged MetadataAuditExcluded. A nil schema
// applies the name list only.
func IsExcludedField(schema *dmodel.ModelSchema, name string) bool {
	if systemFields[name] {
		return true
	}
	if schema == nil {
		return false
	}
	field, ok := schema.Field(name)
	if !ok || field.IsVirtual() {
		return true
	}
	if schema.IsPrimaryKey(name) || schema.IsTenantKey(name) || schema.IsVersioningKey(name) {
		return true
	}
	if field.DataType() != nil && field.DataType().String() == dmodel.FieldDataTypeNameSecret {
		return true
	}
	flag, ok := field.MetadataValue(MetadataAuditExcluded)
	return ok && flag == true
}

// Diff returns the old and new values of the fields that changed, excluding IsExcludedField.
//
// A nil before is a create: every non-null field of after is new. A nil after is a delete: every
// non-null field of before is gone. Otherwise only the keys after carries are compared, since an
// update sends just the fields it sets.
func Diff(schema *dmodel.ModelSchema, before, after dmodel.DynamicFields) (dmodel.DynamicFields, dmodel.DynamicFields) {
	oldValues, newValues := dmodel.DynamicFields{}, dmodel.DynamicFields{}
	switch {
	case before == nil:
		copyNonNull(schema, after, newValues)
	case after == nil:
		copyNonNull(schema, before, oldValues)
	default:
		for name, newValue := range after {
			if IsExcludedField(schema, name) || SameValue(before[name], newValue) {
				continue
			}
			oldValues[name] = before[name]
			newValues[name] = newValue
		}
	}
	return oldValues, newValues
}

func copyNonNull(schema *dmodel.ModelSchema, src, dest dmodel.DynamicFields) {
	for name, value := range src {
		if IsExcludedField(schema, name) || isNull(value) {
			continue
		}
		dest[name] = value
	}
}

// SameValue compares two field values as stored rather than as Go values: the row read back from
// the database and the validated input rarely share a Go type (int vs int64, string vs model.Id,
// a datetime struct vs its string). Numbers compare by value, so "10.50" equals 10.5.
func SameValue(left, right any) bool {
	if isNull(left) || isNull(right) {
		return isNull(left) && isNull(right)
	}
	if reflect.DeepEqual(left, right) {
		return true
	}
	leftJson, leftErr := json.Marshal(left)
	rightJson, rightErr := json.Marshal(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	if string(leftJson) == string(rightJson) {
		return true
	}
	return sameNumber(string(leftJson), string(rightJson))
}

func sameNumber(left, right string) bool {
	leftNum, leftErr := decimal.NewFromString(strings.Trim(left, `"`))
	rightNum, rightErr := decimal.NewFromString(strings.Trim(right, `"`))
	return leftErr == nil && rightErr == nil && leftNum.Equal(rightNum)
}

func isNull(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice:
		return reflected.IsNil()
	}
	return false
}
