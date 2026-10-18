package role

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func TestUpdateRoleCommand_SetFieldData_LiftsEntitlementsOutOfRoleFields(t *testing.T) {
	cmd := &UpdateRoleCommand{}
	cmd.SetFieldData(dmodel.DynamicFields{
		"id":   "01JWNZ3XVWDP73JGHRRBFAHQY1",
		"etag": "e1",
		UpdateRoleFieldEntitlements: map[string]any{
			"add": []any{
				map[string]any{"resource_id": "01JWNZ3XVWDP73JGHRRBFAHQY2", "action_id": nil, "description": "all"},
			},
			"remove": []any{"01JWNZ3XVWDP73JGHRRBFAHQY3"},
		},
	})

	_, stillThere := cmd.GetFieldData()[UpdateRoleFieldEntitlements]
	assert.False(t, stillThere)
	assert.False(t, cmd.HasMalformedEntitlements())
	require.True(t, cmd.HasEntitlementsDelta())
	require.Len(t, cmd.Entitlements.Add, 1)
	assert.Nil(t, cmd.Entitlements.Add[0].ActionId)
	assert.Equal(t, "all", *cmd.Entitlements.Add[0].Description)
	assert.Len(t, cmd.Entitlements.Remove, 1)
}

func TestUpdateRoleCommand_SetFieldData_NoDeltaKeepsPlainUpdate(t *testing.T) {
	cmd := &UpdateRoleCommand{}
	cmd.SetFieldData(dmodel.DynamicFields{"id": "x", "entitlements": nil})

	assert.False(t, cmd.HasEntitlementsDelta())
	assert.False(t, cmd.HasMalformedEntitlements())
}

func TestUpdateRoleCommand_SetFieldData_FlagsMalformedDelta(t *testing.T) {
	cmd := &UpdateRoleCommand{}
	cmd.SetFieldData(dmodel.DynamicFields{"id": "x", "entitlements": "everything"})

	assert.True(t, cmd.HasMalformedEntitlements())
	assert.False(t, cmd.HasEntitlementsDelta())
}
