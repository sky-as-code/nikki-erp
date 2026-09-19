// The second raw-SQL file in the Inventory module, for the same reason as stock_quant_lock.go: a
// warehouse-level commitment can only be made safe with SELECT ... FOR UPDATE, and the query
// builder cannot emit a lock clause. Locking quants alone is not enough — two requests can pick two
// different quants and still both spend the same warehouse-level availability — so every
// operation that reads or changes a (warehouse, variant) scope's availability locks its guard
// row first, and only then its quants. That order is fixed everywhere; reversing it anywhere is a
// deadlock waiting for load.
package services

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// GuardKey names one (org, warehouse, variant) scope whose availability is about to be read or
// changed.
type GuardKey struct {
	OrgId            model.Id
	WarehouseId      model.Id
	ProductVariantId model.Id
}

// LockedGuard is a guard row held under a row lock until the enclosing transaction ends.
type LockedGuard struct {
	Id  model.Id
	Key GuardKey
}

// LockGuardsForUpdate takes a row lock on the guard of every scope in keys and returns them in
// lock order. Three properties are load-bearing:
//
//   - It must run inside a transaction, for the same reason as LockQuantsForUpdate: outside one
//     the lock is gone the instant the statement returns.
//   - Keys are deduplicated and locked in one statement ordered by (warehouse_id,
//     product_variant_id), so two requests over overlapping scopes queue instead of deadlocking.
//   - A scope that has no guard row yet gets one first. A brand-new variant at a warehouse must
//     still serialise with everyone else on that scope, so the row is created on first use rather
//     than when stock first arrives.
//
// Every key must belong to one org; scopes of several orgs are locked by separate calls.
func LockGuardsForUpdate(
	ctx corectx.Context, repo dyn.BaseDynamicRepository, keys []GuardKey,
) ([]LockedGuard, error) {
	if ctx == nil || ctx.GetDbTranx() == nil {
		return nil, errors.New(
			"LockGuardsForUpdate requires an ambient transaction: without one the guard lock would be " +
				"released as soon as the statement returns, and two requests could commit the same stock")
	}

	keys = sortGuardKeys(keys)
	if len(keys) == 0 {
		return nil, nil
	}
	if err := assertSingleOrg(keys); err != nil {
		return nil, err
	}

	schema := repo.Schema()
	client := repo.ExtractClient(ctx)
	for _, key := range keys {
		query, args, err := buildGuardUpsertQuery(schema, ctx, key)
		if err != nil {
			return nil, err
		}
		if _, err := client.Exec(ctx, query, args...); err != nil {
			return nil, errors.Wrap(err, "failed to ensure a warehouse product guard row")
		}
	}

	query, args := buildGuardLockQuery(schema, ctx, keys)
	rows, err := client.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to lock warehouse product guards for update")
	}
	defer rows.Close()

	locked := make([]LockedGuard, 0, len(keys))
	for rows.Next() {
		var id, warehouseId, variantId string
		if err := rows.Scan(&id, &warehouseId, &variantId); err != nil {
			return nil, errors.Wrap(err, "failed to scan a locked warehouse product guard")
		}
		locked = append(locked, LockedGuard{
			Id: model.Id(id),
			Key: GuardKey{
				OrgId:            keys[0].OrgId,
				WarehouseId:      model.Id(warehouseId),
				ProductVariantId: model.Id(variantId),
			},
		})
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to read locked warehouse product guards")
	}
	if len(locked) != len(keys) {
		return nil, errors.Errorf(
			"locked %d warehouse product guards but %d scopes were requested", len(locked), len(keys))
	}
	return locked, nil
}

// DbNowUnderLock reads the database clock. It is meant to be called after the guard lock has been
// acquired: a time taken before queuing for the lock is stale by however long the wait was, and
// an expiry decided on it could revive a reservation that had already lapsed.
func DbNowUnderLock(ctx corectx.Context, repo dyn.BaseDynamicRepository) (time.Time, error) {
	if ctx == nil || ctx.GetDbTranx() == nil {
		return time.Time{}, errors.New("DbNowUnderLock requires an ambient transaction")
	}
	var now time.Time
	err := repo.ExtractClient(ctx).QueryRow(ctx, "SELECT clock_timestamp()").Scan(&now)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "failed to read the database clock")
	}
	return now.UTC(), nil
}

// sortGuardKeys returns the distinct keys in lock order. Exported to tests through the lock
// functions; it is the single definition of "lock order" for guards.
func sortGuardKeys(keys []GuardKey) []GuardKey {
	seen := make(map[GuardKey]struct{}, len(keys))
	distinct := make([]GuardKey, 0, len(keys))
	for _, key := range keys {
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		distinct = append(distinct, key)
	}
	sort.Slice(distinct, func(i, j int) bool {
		if distinct[i].WarehouseId != distinct[j].WarehouseId {
			return distinct[i].WarehouseId < distinct[j].WarehouseId
		}
		return distinct[i].ProductVariantId < distinct[j].ProductVariantId
	})
	return distinct
}

func assertSingleOrg(keys []GuardKey) error {
	for _, key := range keys[1:] {
		if key.OrgId != keys[0].OrgId {
			return errors.New("LockGuardsForUpdate takes the scopes of one org per call")
		}
	}
	return nil
}

// buildGuardUpsertQuery creates the guard row for a scope if it does not exist. ON CONFLICT DO
// NOTHING without a target relies on the composite unique over (warehouse, variant, org[, tenant]);
// an existing row is left alone and is locked by the SELECT that follows.
func buildGuardUpsertQuery(
	schema *dmodel.ModelSchema, ctx corectx.Context, key GuardKey,
) (string, []any, error) {
	id, err := model.NewId()
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to generate a guard id")
	}

	columns := []string{
		models.WarehouseProductGuardFieldId,
		models.WarehouseProductGuardFieldWarehouseId,
		models.WarehouseProductGuardFieldProductVariantId,
		models.WarehouseProductGuardFieldOrgId,
		"created_at",
	}
	args := []any{
		string(*id),
		string(key.WarehouseId),
		string(key.ProductVariantId),
		string(key.OrgId),
		time.Now().UTC(),
	}
	if tenantKey := schema.TenantKey(); tenantKey != "" {
		if value := tenantValue(ctx, tenantKey); value != nil {
			columns = append(columns, tenantKey)
			args = append(args, value)
		}
	}

	placeholders := make([]string, 0, len(columns))
	for i := range columns {
		placeholders = append(placeholders, "$"+strconv.Itoa(i+1))
	}
	query := "INSERT INTO " + schema.TableName() +
		" (" + strings.Join(columns, ", ") + ")" +
		" VALUES (" + strings.Join(placeholders, ", ") + ")" +
		" ON CONFLICT DO NOTHING"
	return query, args, nil
}

// buildGuardLockQuery assembles the locking SELECT over every requested scope of one org, ordered
// so that overlapping callers take the same rows in the same sequence. Column and table names
// come from the schema and this module's constants, never caller input.
func buildGuardLockQuery(
	schema *dmodel.ModelSchema, ctx corectx.Context, keys []GuardKey,
) (string, []any) {
	args := []any{string(keys[0].OrgId)}
	predicates := []string{models.WarehouseProductGuardFieldOrgId + " = $1"}

	if tenantKey := schema.TenantKey(); tenantKey != "" {
		if value := tenantValue(ctx, tenantKey); value != nil {
			args = append(args, value)
			predicates = append(predicates, tenantKey+" = $"+strconv.Itoa(len(args)))
		}
	}

	tuples := make([]string, 0, len(keys))
	for _, key := range keys {
		args = append(args, string(key.WarehouseId))
		first := len(args)
		args = append(args, string(key.ProductVariantId))
		tuples = append(tuples, "($"+strconv.Itoa(first)+", $"+strconv.Itoa(first+1)+")")
	}
	predicates = append(predicates,
		"("+models.WarehouseProductGuardFieldWarehouseId+", "+models.WarehouseProductGuardFieldProductVariantId+")"+
			" IN ("+strings.Join(tuples, ", ")+")")

	query := "SELECT " + strings.Join([]string{
		models.WarehouseProductGuardFieldId,
		models.WarehouseProductGuardFieldWarehouseId,
		models.WarehouseProductGuardFieldProductVariantId,
	}, ", ") +
		" FROM " + schema.TableName() +
		" WHERE " + strings.Join(predicates, " AND ") +
		" ORDER BY " + models.WarehouseProductGuardFieldWarehouseId + " ASC, " +
		models.WarehouseProductGuardFieldProductVariantId + " ASC" +
		" FOR UPDATE"
	return query, args
}
