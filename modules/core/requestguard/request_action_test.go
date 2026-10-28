package requestguard

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

func ownerCtx() corectx.Context {
	ctx := corectx.NewRequestContext(context.Background())
	ctx.SetPermissions(corectx.ContextPermissions{
		IsOwner:   true,
		Principal: corectx.Principal{Kind: corectx.PrincipalKindUser, Id: model.Id("u1")},
	})
	return ctx
}

func TestPassedCustomActionLabelsTheContext(t *testing.T) {
	ctx := ownerCtx()

	assert.Nil(t, AssertPermission(ctx, PermFor("update", "inventory_warehouse", ResourceScopeOrg)))
	assert.Equal(t, "", corectx.RequestActionOf(ctx), "a CRUD permission names nothing")
	assert.Nil(t, AssertPermission(ctx, PermFor("read_inbox", "notification_recipient", ResourceScopeOrg)))
	assert.Equal(t, "", corectx.RequestActionOf(ctx), "a read-like action names nothing")
	assert.Nil(t, AssertPermission(ctx, PermFor("suspend", "inventory_warehouse", ResourceScopeOrg)))
	assert.Equal(t, "suspend", corectx.RequestActionOf(ctx))
}

func TestRefusedCustomActionLabelsNothing(t *testing.T) {
	ctx := corectx.NewRequestContext(context.Background())

	assert.NotNil(t, AssertPermission(ctx, PermFor("suspend", "inventory_warehouse", ResourceScopeOrg)))
	assert.Equal(t, "", corectx.RequestActionOf(ctx))
}
