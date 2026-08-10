package context

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ds "github.com/sky-as-code/nikki-erp/common/datastructure"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
)

// CloneRequestContext derives a scoped context from a live request, which is how a service opens
// a transaction without mutating the context it was handed. The copy must be total: the write
// path reads the caller's identity off the context to stamp created_by/updated_by, and a missing
// permission set produces a null audit column rather than an error.

type fakeDbTransaction struct {
	committed bool
}

func (this *fakeDbTransaction) Commit() error {
	this.committed = true
	return nil
}

func (this *fakeDbTransaction) Rollback() error {
	return nil
}

// newPopulatedContext builds a context with every field a request carries set to a
// distinguishable value, so that a field the clone forgets shows up as a zero value.
func newPopulatedContext() (Context, *fakeDbTransaction) {
	tranx := &fakeDbTransaction{}
	ctx := NewRequestContextM(context.Background(), "inventory")
	ctx.SetDbTranx(tranx)
	ctx.SetPermissions(ContextPermissions{
		IsOwner:      true,
		Entitlements: ds.NewSetFrom("inventory_product_template:update"),
		UserId:       model.Id("01JQZ0X0000000000000000001"),
		UserOrgIds:   ds.NewSetFrom(model.Id("01JQZ0X0000000000000000002")),
	})
	ctx.SetUser(dmodel.DynamicFields{"id": "01JQZ0X0000000000000000001", "email": "tester@example.com"})
	ctx.SetDomainConstraints(dmodel.DynamicFields{"org_id": "01JQZ0X0000000000000000002"})
	return ctx, tranx
}

// The whole point of the helper: everything a request carries survives the copy. UserId is
// called out because it is the one whose loss is silent — basemodel writes a null updated_by
// rather than failing, so no other test would catch it.
func TestCloneCarriesEveryRequestField(t *testing.T) {
	original, tranx := newPopulatedContext()

	clone := CloneRequestContext(original)

	require.NotNil(t, clone)
	assert.Equal(t, model.Id("01JQZ0X0000000000000000001"), clone.GetPermissions().UserId,
		"identity must survive: basemodel stamps created_by/updated_by from it")
	assert.True(t, clone.GetPermissions().IsOwner)
	assert.True(t, clone.GetPermissions().Entitlements.Contains("inventory_product_template:update"))
	assert.Equal(t, original.GetUser(), clone.GetUser())
	assert.Equal(t, "inventory", clone.GetModuleName())
	assert.Equal(t, db.DbTransaction(tranx), clone.GetDbTranx())
	assert.Equal(t, original.GetLogger(), clone.GetLogger())
	// Domain constraints live in the inner context's values, so they ride along with Context.
	assert.Equal(t, original.GetDomainConstraints(), clone.GetDomainConstraints())
}

// Giving the clone its own transaction must not reach back into the caller's context. This is
// what makes a scoped transaction safe, and what SetDbTranx on the original would break.
func TestCloneTransactionDoesNotLeakToTheOriginal(t *testing.T) {
	original, originalTranx := newPopulatedContext()
	scopedTranx := &fakeDbTransaction{}

	clone := CloneRequestContext(original)
	clone.SetDbTranx(scopedTranx)

	assert.Equal(t, db.DbTransaction(scopedTranx), clone.GetDbTranx())
	assert.Equal(t, db.DbTransaction(originalTranx), original.GetDbTranx(),
		"the caller's context must keep the transaction it had")
}

// A context with nothing set clones to a context with nothing set, rather than panicking on a
// nil logger or an empty permission set.
func TestCloneOfAnEmptyContext(t *testing.T) {
	original := NewRequestContext(context.Background())

	clone := CloneRequestContext(original)

	require.NotNil(t, clone)
	assert.Nil(t, clone.GetDbTranx())
	assert.Empty(t, clone.GetModuleName())
	assert.Equal(t, model.Id(""), clone.GetPermissions().UserId)
}
