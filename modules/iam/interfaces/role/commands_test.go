package role

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

func TestLiftEntitlementsDelta_LiftsEntitlementsOutOfRoleFields(t *testing.T) {
	fields := dmodel.DynamicFields{
		"id":   "01JWNZ3XVWDP73JGHRRBFAHQY1",
		"etag": "e1",
		UpdateRoleFieldEntitlements: map[string]any{
			"add": []any{
				map[string]any{"resource_id": "01JWNZ3XVWDP73JGHRRBFAHQY2", "action_id": nil, "description": "all"},
			},
			"remove": []any{"01JWNZ3XVWDP73JGHRRBFAHQY3"},
		},
	}
	require.True(t, HasEntitlementsDelta(fields))

	delta, malformed := LiftEntitlementsDelta(fields)

	_, stillThere := fields[UpdateRoleFieldEntitlements]
	assert.False(t, stillThere)
	assert.False(t, malformed)
	require.NotNil(t, delta)
	require.Len(t, delta.Add, 1)
	assert.Nil(t, delta.Add[0].ActionId)
	assert.Equal(t, "all", *delta.Add[0].Description)
	assert.Len(t, delta.Remove, 1)
}

func TestLiftEntitlementsDelta_NoDeltaKeepsPlainUpdate(t *testing.T) {
	fields := dmodel.DynamicFields{"id": "x", "entitlements": nil}

	assert.False(t, HasEntitlementsDelta(fields))
	delta, malformed := LiftEntitlementsDelta(fields)

	assert.Nil(t, delta)
	assert.False(t, malformed)
}

func TestLiftEntitlementsDelta_FlagsMalformedDelta(t *testing.T) {
	fields := dmodel.DynamicFields{"id": "x", "entitlements": "everything"}

	assert.True(t, HasEntitlementsDelta(fields), "a malformed delta must still be authorized as one")
	_, malformed := LiftEntitlementsDelta(fields)

	assert.True(t, malformed)
}
