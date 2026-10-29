package composable

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

func newManagedSchema(withManagedField bool) *dmodel.ModelSchema {
	builder := dmodel.DefineModel("cmp_managed_resource").
		Field(dmodel.DefineField().Name("id").DataType(dmodel.FieldDataTypeString(0, 50)))
	if withManagedField {
		builder = builder.Field(dmodel.DefineField().Name(basemodel.FieldManagedByType).
			DataType(dmodel.FieldDataTypeString(1, 100)))
	}
	return builder.Build()
}

func runManagedRule(
	t *testing.T, withField bool, mode managedDeleteMode, hook ValidateExtraFn, stored dmodel.DynamicFields,
) ft.ClientErrors {
	service := &DefaultDomainServiceImpl{schema: newManagedSchema(withField)}
	rule := service.withManagedDeleteRule(mode, hook)
	require.NotNil(t, rule)
	vErrs := ft.ClientErrors{}
	require.NoError(t, rule(newTestRequestCtx(), nil, NewDynamicEntityFrom(stored), &vErrs))
	return vErrs
}

func managedRow(managedBy any) dmodel.DynamicFields {
	return dmodel.DynamicFields{"id": "01", basemodel.FieldManagedByType: managedBy}
}

func TestDeleteRefusesAManagedRecord(t *testing.T) {
	for _, managedBy := range []string{basemodel.ManagedByTypeSystem, "commerce/inventory_warehouse"} {
		vErrs := runManagedRule(t, true, managedDeleteRefuse, nil, managedRow(managedBy))
		require.Equal(t, 1, vErrs.Count(), managedBy)
		assert.Equal(t, "common:err_record_managed_delete", vErrs[0].Key)
	}
}

func TestDeleteAcceptsAnUnmanagedRecord(t *testing.T) {
	vErrs := runManagedRule(t, true, managedDeleteRefuse, nil, managedRow(nil))
	assert.Zero(t, vErrs.Count())
}

func TestDeleteRuleIsUnchangedForASchemaWithoutTheField(t *testing.T) {
	service := &DefaultDomainServiceImpl{schema: newManagedSchema(false)}
	assert.Nil(t, service.withManagedDeleteRule(managedDeleteRefuse, nil))
}

func TestDeleteManagedAcceptsOnlyAManagedRecord(t *testing.T) {
	vErrs := runManagedRule(t, true, managedDeleteRequire, nil, managedRow("vendingmachine/vdmc_kiosks"))
	assert.Zero(t, vErrs.Count())

	vErrs = runManagedRule(t, true, managedDeleteRequire, nil, managedRow(nil))
	require.Equal(t, 1, vErrs.Count())
	assert.Equal(t, "common:err_record_not_managed", vErrs[0].Key)

	vErrs = runManagedRule(t, false, managedDeleteRequire, nil, dmodel.DynamicFields{"id": "01"})
	require.Equal(t, 1, vErrs.Count(), "a schema without managed_by_type has no managed record")
	assert.Equal(t, "common:err_record_not_managed", vErrs[0].Key)
}

func TestManagedRuleKeepsTheCallersOwnRefusal(t *testing.T) {
	calls := []string{}
	vErrs := runManagedRule(t, true, managedDeleteRefuse, recording(&calls, "hook"), managedRow("system"))
	assert.Equal(t, []string{"hook"}, calls)
	assert.Equal(t, 1, vErrs.Count())

	vErrs = runManagedRule(t, true, managedDeleteRefuse, violation("err_own_rule"), managedRow("system"))
	require.Equal(t, 1, vErrs.Count())
	assert.Equal(t, "common:err_own_rule", vErrs[0].Key)
}
