package dynamicresource

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

func TestOrgBoundRequiresAnOrgOnASchemaWithoutOne(t *testing.T) {
	domSvc, appSvc := newAppService(newPlainSchema(), NewAppServiceParam{IsOrgBound: true})

	result, err := appSvc.Create(memberContext(), dmodel.DynamicFields{"name": "x"})

	require.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count())
	assert.Zero(t, domSvc.calls)
}

// The org is the acting org, never a column: it leaves the payload and filters nothing.
func TestOrgBoundRecordsTheActingOrgAndDropsItFromParams(t *testing.T) {
	domSvc, appSvc := newAppService(newPlainSchema(), NewAppServiceParam{IsOrgBound: true})
	ctx := memberContext()

	result, err := appSvc.Create(ctx, dmodel.DynamicFields{"name": "x", basemodel.FieldOrgId: "org_mine"})

	require.NoError(t, err)
	assert.Zero(t, result.ClientErrors.Count())
	assert.NotContains(t, domSvc.seen, basemodel.FieldOrgId)
	acting, ok := dyn.ActingOrgOf(ctx)
	assert.True(t, ok)
	assert.Equal(t, model.Id("org_mine"), acting)
}

func TestOrgBoundRefusesAnOrgTheCallerDoesNotBelongTo(t *testing.T) {
	domSvc, appSvc := newAppService(newPlainSchema(), NewAppServiceParam{IsOrgBound: true})

	result, err := appSvc.Create(memberContext(), dmodel.DynamicFields{basemodel.FieldOrgId: "org_someone_else"})

	require.NoError(t, err)
	assert.Positive(t, result.ClientErrors.Count())
	assert.Zero(t, domSvc.calls)
}

// The served schema tells every client to send org_id, whichever service it reaches the resource by.
func TestOrgBoundFlagsTheServedSchema(t *testing.T) {
	_, plain := newAppService(newPlainSchema(), NewAppServiceParam{})
	_, bound := newAppService(newPlainSchema(), NewAppServiceParam{IsOrgBound: true})

	plainSchema, err := plain.GetSchema(memberContext(), GetSchemaQuery{})
	require.NoError(t, err)
	boundSchema, err := bound.GetSchema(memberContext(), GetSchemaQuery{})
	require.NoError(t, err)

	assert.NotContains(t, toJsonMap(t, plainSchema.Data), fieldIsOrgBound)
	served := toJsonMap(t, boundSchema.Data)
	assert.Equal(t, true, served[fieldIsOrgBound])
	assert.Equal(t, toJsonMap(t, plainSchema.Data)["name"], served["name"])
}

func toJsonMap(t *testing.T, value any) map[string]any {
	raw, err := json.Marshal(value)
	require.NoError(t, err)
	result := map[string]any{}
	require.NoError(t, json.Unmarshal(raw, &result))
	return result
}

// A client-sent ${ctx.org_id} never wins over the org the server authorized.
func TestMergeComputedContextOverridesTheClientOrg(t *testing.T) {
	ctx := memberContext()
	assert.Equal(t, map[string]any{"org_id": ""}, dyn.MergeComputedContext(ctx, nil), "no org reads as none")

	dyn.WithActingOrg(ctx, "org_mine")
	merged := dyn.MergeComputedContext(ctx, map[string]any{"org_id": "x", "k": 1})
	assert.Equal(t, map[string]any{"org_id": "org_mine", "k": 1}, merged)
}
