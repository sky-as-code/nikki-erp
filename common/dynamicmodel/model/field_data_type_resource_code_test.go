package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceCode_AcceptsStrictForm(t *testing.T) {
	dataType := FieldDataTypeResourceCode(1, 20)
	for _, input := range []string{"WH-CENTRAL", "VN-VAT-10", "A1", "  MOMO  "} {
		validated, clientErr := dataType.Validate(Value(input))
		require.Nilf(t, clientErr, "%q must be accepted", input)
		assert.NotContains(t, (*validated.Get()).(string), " ", "TryConvert trims")
	}
}

func TestResourceCode_RejectsWithoutCaseConversion(t *testing.T) {
	dataType := FieldDataTypeResourceCode(1, 20)
	for _, input := range []string{"wh-a", "WH_A", "WH A", "WH.A", "vdmc"} {
		_, clientErr := dataType.Validate(Value(input))
		require.NotNilf(t, clientErr, "%q must be rejected", input)
		assert.Equal(t, "common:err_invalid_resource_code", clientErr.Key)
	}
}

func TestResourceCode_EnforcesLength(t *testing.T) {
	_, clientErr := FieldDataTypeResourceCode(1, 5).Validate(Value("ABCDEF"))
	assert.NotNil(t, clientErr)
}

func TestResourceCode_FieldErrorNamesTheField(t *testing.T) {
	schema := ParseModelJson(`{
		"name": "test_resource_code",
		"fields": [{"name": "code", "data_type": {"type":"resource_code","min":1,"max":20}}]
	}`).Build()

	_, clientErrs := schema.Validate(map[string]any{"code": "bad_code"})
	require.Len(t, clientErrs, 1)
	assert.Equal(t, "code", clientErrs[0].Field)
	assert.Equal(t, "common:err_invalid_resource_code", clientErrs[0].Key)
}

func TestResourceCode_ArrayTrimsEachItem(t *testing.T) {
	converted, err := FieldDataTypeResourceCode(1, 20).ArrayType().TryConvert([]any{" A ", "B-1"}, nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"A", "B-1"}, *converted.Get())
}

func TestResourceCode_EmptyAllowedOnlyWhenOptional(t *testing.T) {
	_, clientErr := FieldDataTypeResourceCode(0, 20).Validate(Value(""))
	assert.Nil(t, clientErr, "min 0 declares the code optional")

	_, clientErr = FieldDataTypeResourceCode(1, 20).Validate(Value(""))
	assert.NotNil(t, clientErr)
}

func TestNormalizeResourceCode(t *testing.T) {
	assert.Equal(t, "VN-VAT-10", NormalizeResourceCode(" vn_vat_10 "))
	assert.Equal(t, "VDMC", NormalizeResourceCode("vdmc"))
}
