package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

func guardSchema(t *testing.T) *dmodel.ModelSchema {
	t.Helper()
	_ = basemodel.RegisterJsonBaseSchemas()
	return models.WarehouseProductGuardSchemaBuilder().Build()
}

func guardKey(warehouse, variant string) GuardKey {
	return GuardKey{OrgId: model.Id("org-1"), WarehouseId: model.Id(warehouse), ProductVariantId: model.Id(variant)}
}

func TestSortGuardKeysDedupesAndOrdersByWarehouseThenVariant(t *testing.T) {
	keys := []GuardKey{
		guardKey("wh-2", "var-1"),
		guardKey("wh-1", "var-9"),
		guardKey("wh-1", "var-2"),
		guardKey("wh-2", "var-1"),
	}

	sorted := sortGuardKeys(keys)

	require.Len(t, sorted, 3, "the duplicate scope is locked once")
	assert.Equal(t, []GuardKey{
		guardKey("wh-1", "var-2"),
		guardKey("wh-1", "var-9"),
		guardKey("wh-2", "var-1"),
	}, sorted, "a total order is what stops two overlapping requests deadlocking")
}

func TestBuildGuardLockQueryLocksEveryScopeInOneOrderedStatement(t *testing.T) {
	ctx := corectx.NewRequestContext(context.Background())
	keys := sortGuardKeys([]GuardKey{guardKey("wh-2", "var-1"), guardKey("wh-1", "var-1")})

	query, args := buildGuardLockQuery(guardSchema(t), ctx, keys)

	assert.Contains(t, query, "FOR UPDATE")
	assert.Contains(t, query, "FROM inventory_warehouse_product_guards")
	assert.Contains(t, query, "ORDER BY warehouse_id ASC, product_variant_id ASC")
	assert.Contains(t, query, "(warehouse_id, product_variant_id) IN (($2, $3), ($4, $5))")
	assert.Equal(t, []any{"org-1", "wh-1", "var-1", "wh-2", "var-1"}, args)
	assert.NotContains(t, query, "org-1", "values are bound, never inlined")
}

func TestBuildGuardUpsertQueryCreatesTheRowWithoutTouchingAnExistingOne(t *testing.T) {
	ctx := corectx.NewRequestContext(context.Background())

	query, args, err := buildGuardUpsertQuery(guardSchema(t), ctx, guardKey("wh-1", "var-1"))

	require.NoError(t, err)
	assert.Contains(t, query, "INSERT INTO inventory_warehouse_product_guards")
	assert.Contains(t, query, "ON CONFLICT DO NOTHING")
	assert.Len(t, args, 5, "id, warehouse, variant, org and created_at")
	assert.Equal(t, "wh-1", args[1])
	assert.Equal(t, "var-1", args[2])
	assert.Equal(t, "org-1", args[3])
}

func TestLockGuardsForUpdateRefusesToRunOutsideATransaction(t *testing.T) {
	ctx := corectx.NewRequestContext(context.Background())

	_, err := LockGuardsForUpdate(ctx, nil, []GuardKey{guardKey("wh-1", "var-1")})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambient transaction")
}

func TestLockGuardsForUpdateRefusesScopesOfSeveralOrgs(t *testing.T) {
	keys := []GuardKey{
		guardKey("wh-1", "var-1"),
		{OrgId: model.Id("org-2"), WarehouseId: model.Id("wh-1"), ProductVariantId: model.Id("var-2")},
	}

	err := assertSingleOrg(sortGuardKeys(keys))

	require.Error(t, err)
}
