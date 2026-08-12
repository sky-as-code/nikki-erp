package services

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

func quantSchema(t *testing.T) *dmodel.ModelSchema {
	t.Helper()
	// Normally done by CoreModule.RegisterModels during app start-up; the quant schema extends the
	// base models and cannot be parsed until they are registered.
	_ = basemodel.RegisterJsonBaseSchemas()
	return models.StockQuantSchemaBuilder().Build()
}

func lockKey() QuantLockKey {
	return QuantLockKey{
		OrgId:            model.Id("org-1"),
		ProductVariantId: model.Id("var-1"),
		LocationId:       model.Id("loc-1"),
	}
}

func TestBuildQuantLockQueryLocksRowsInADeterministicOrder(t *testing.T) {
	ctx := corectx.NewRequestContext(context.Background())

	query, args := buildQuantLockQuery(quantSchema(t), ctx, lockKey())

	assert.Contains(t, query, "FOR UPDATE", "without the lock clause this is just a read")
	assert.Contains(t, query, "ORDER BY incoming_date ASC NULLS LAST, id ASC",
		"a total order is what stops two overlapping reservations deadlocking")
	assert.Contains(t, query, "FROM inventory_stock_quants")
	assert.Len(t, args, 3, "org, variant and location")
}

func TestBuildQuantLockQueryBindsEveryValue(t *testing.T) {
	// Values must travel as bound arguments, never interpolated into the statement.
	ctx := corectx.NewRequestContext(context.Background())

	query, args := buildQuantLockQuery(quantSchema(t), ctx, lockKey())

	assert.NotContains(t, query, "org-1", "the org id must be bound, not inlined")
	assert.NotContains(t, query, "var-1")
	for i := range args {
		assert.Containsf(t, query, "$"+string(rune('1'+i)), "placeholder $%d must appear", i+1)
	}
}

func TestBuildQuantLockQuerySelectsTheColumnsAllocationNeeds(t *testing.T) {
	ctx := corectx.NewRequestContext(context.Background())

	query, _ := buildQuantLockQuery(quantSchema(t), ctx, lockKey())

	for _, column := range lockedQuantColumns() {
		assert.Containsf(t, query, column, "%s is read inside the lock", column)
	}
	assert.NotContains(t, query, models.StockQuantFieldAvailableQuantity,
		"available_quantity is virtual and has no column to select")
}

func TestBuildQuantLockQueryAddsTheTenantPredicateWhenTheSchemaHasOne(t *testing.T) {
	// In nikkierp the quant schema has no tenant key, so the query carries three predicates. In
	// coremart the shadow module adds one, and this asserts the predicate follows the schema rather
	// than being hardcoded — the difference between the two binaries locking the right rows and one
	// of them locking another tenant's.
	schema := quantSchema(t)
	tenantKey := schema.TenantKey()

	ctx := corectx.NewRequestContextF(context.Background(), "inventory", dmodel.DynamicFields{
		"tenant_id": "tenant-1",
	})
	query, args := buildQuantLockQuery(schema, ctx, lockKey())

	if tenantKey == "" {
		assert.Len(t, args, 3, "no tenant key in this build, so no tenant predicate")
		assert.NotContains(t, query, "tenant_id")
		return
	}
	require.Len(t, args, 4, "the tenant value must be bound alongside the others")
	assert.Contains(t, query, tenantKey+" = $4")
	assert.Equal(t, "tenant-1", args[3])
}

func TestBuildQuantLockQueryOmitsTheTenantPredicateWithoutAConstraint(t *testing.T) {
	// A missing constraint must not become a literal NULL comparison, which matches nothing and
	// would silently reserve against an empty result.
	schema := quantSchema(t)
	ctx := corectx.NewRequestContext(context.Background())

	query, args := buildQuantLockQuery(schema, ctx, lockKey())

	assert.Len(t, args, 3)
	assert.Equal(t, 3, strings.Count(query, "$"), "exactly one placeholder per bound value")
}

func TestLockQuantsForUpdateRefusesToRunOutsideATransaction(t *testing.T) {
	// Without a transaction ExtractClient hands back a pooled connection, the lock is released as
	// soon as the statement returns, and the caller reserves against figures it does not hold. That
	// failure is invisible until two requests collide, so it is refused loudly instead.
	ctx := corectx.NewRequestContext(context.Background())

	locked, err := LockQuantsForUpdate(ctx, nil, lockKey())

	require.Error(t, err, "locking outside a transaction must fail rather than silently not lock")
	assert.Nil(t, locked)
	assert.Contains(t, err.Error(), "ambient transaction")
}

func TestAvailableOnALockedQuant(t *testing.T) {
	quant := lockedQuant("q1", "100", "30")
	assert.Equal(t, "70", quant.Available().String())
}
