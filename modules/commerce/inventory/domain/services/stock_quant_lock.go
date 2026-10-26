// The only raw SQL in the Inventory module, a deliberate exception to the engine-first rule:
// reserving stock safely requires SELECT ... FOR UPDATE, and the query builder cannot emit it —
// orm.QueryBuilder exposes no lock clause and SqlSelectGraphOpts has no lock field. Without the
// lock, two concurrent reservations read the same available quantity and both reserve it.
package services

import (
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

// LockedQuant is one stock balance row held under a row lock until the enclosing transaction ends.
// Its quantities are read inside the lock, the only moment they can be trusted; a value read before
// the lock is stale by definition.
type LockedQuant struct {
	Id         model.Id
	OnHand     decimal.Decimal
	Reserved   decimal.Decimal
	LotRef     string
	PackageRef string
	OwnerRef   string
}

// Available is the quantity this locked row can still promise to a new demand.
func (this LockedQuant) Available() decimal.Decimal {
	return this.OnHand.Sub(this.Reserved)
}

// QuantLockKey identifies the set of balances a move needs to draw from.
type QuantLockKey struct {
	OrgId            model.Id
	ProductVariantId model.Id
	LocationId       model.Id
}

// LockQuantsForUpdate takes a row lock on every stock balance for a variant in a location and
// returns them, ordered oldest first. Three properties are load-bearing:
//
//   - It must run inside a transaction. ExtractClient falls back to the connection pool otherwise,
//     where the locks are released the instant the statement returns, leaving a caller that
//     believes it holds them and does not. It refuses to run rather than degrade silently.
//   - The order is deterministic, so two reservations over overlapping rows queue instead of
//     deadlocking. incoming_date alone is not enough — rows can share a timestamp — so id breaks
//     the tie.
//   - The tenant predicate comes from the schema, because coremart's shadow module adds tenant_id
//     and nikkierp's does not; a hardcoded column list would lock another tenant's rows in one of
//     the two binaries.
//
// Callers must recompute availability from what this returns and never reuse a figure read before
// the call.
func LockQuantsForUpdate(
	ctx corectx.Context, repo dyn.BaseDynamicRepository, key QuantLockKey,
) ([]LockedQuant, error) {
	if ctx == nil || ctx.GetDbTranx() == nil {
		return nil, errors.New(
			"LockQuantsForUpdate requires an ambient transaction: without one the row locks would be " +
				"released as soon as the statement returns, and the caller would reserve against stale figures")
	}

	query, args := buildQuantLockQuery(repo.Schema(), ctx, key)
	rows, err := repo.ExtractClient(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to lock stock quants for update")
	}
	defer rows.Close()

	locked := make([]LockedQuant, 0, 4)
	for rows.Next() {
		quant, err := scanLockedQuant(rows)
		if err != nil {
			return nil, err
		}
		locked = append(locked, quant)
	}
	return locked, errors.Wrap(rows.Err(), "failed to read locked stock quants")
}

// buildQuantLockQuery assembles the locking SELECT with positional placeholders. Column and table
// names come from the schema and this module's constants, never caller input, and every value
// travels as a bound argument.
func buildQuantLockQuery(
	schema *dmodel.ModelSchema, ctx corectx.Context, key QuantLockKey,
) (string, []any) {
	conditions := []string{
		models.StockQuantFieldOrgId,
		models.StockQuantFieldProductVariantId,
		models.StockQuantFieldLocationId,
	}
	args := []any{
		string(key.OrgId),
		string(key.ProductVariantId),
		string(key.LocationId),
	}

	if tenantKey := schema.TenantKey(); tenantKey != "" {
		if value := tenantValue(ctx, tenantKey); value != nil {
			conditions = append(conditions, tenantKey)
			args = append(args, value)
		}
	}

	predicates := make([]string, 0, len(conditions))
	for i, column := range conditions {
		predicates = append(predicates, column+" = $"+strconv.Itoa(i+1))
	}

	query := "SELECT " + strings.Join(lockedQuantColumns(), ", ") +
		" FROM " + schema.TableName() +
		" WHERE " + strings.Join(predicates, " AND ") +
		" ORDER BY " + models.StockQuantFieldIncomingDate + " ASC NULLS LAST, " + models.StockQuantFieldId + " ASC" +
		" FOR UPDATE"
	return query, args
}

func lockedQuantColumns() []string {
	return []string{
		models.StockQuantFieldId,
		models.StockQuantFieldOnHandQuantity,
		models.StockQuantFieldReservedQuantity,
		models.StockQuantFieldLotRef,
		models.StockQuantFieldPackageRef,
		models.StockQuantFieldOwnerRef,
	}
}

func tenantValue(ctx corectx.Context, tenantKey string) any {
	constraints := ctx.GetDomainConstraints()
	if constraints == nil {
		return nil
	}
	value, ok := constraints[tenantKey]
	if !ok {
		return nil
	}
	return value
}

// rowScanner is the part of *sql.Rows this file uses, named so the scan can be tested without a
// database.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanLockedQuant(row rowScanner) (LockedQuant, error) {
	var (
		id               string
		onHand, reserved decimal.NullDecimal
		lot, pkg, owner  string
	)
	if err := row.Scan(&id, &onHand, &reserved, &lot, &pkg, &owner); err != nil {
		return LockedQuant{}, errors.Wrap(err, "failed to scan a locked stock quant")
	}
	return LockedQuant{
		Id:         model.Id(id),
		OnHand:     nullDecimalOrZero(onHand),
		Reserved:   nullDecimalOrZero(reserved),
		LotRef:     lot,
		PackageRef: pkg,
		OwnerRef:   owner,
	}, nil
}

// nullDecimalOrZero reads a null quantity as zero, matching AvailableQuantity: a balance never
// written has moved nothing.
func nullDecimalOrZero(value decimal.NullDecimal) decimal.Decimal {
	if !value.Valid {
		return decimal.Zero
	}
	return value.Decimal
}
