// This file contains the only raw SQL in the Inventory module, and it is a deliberate exception
// to the engine-first rule in docs/wiki/07 §6.7 rather than drift.
//
// The reason is narrow and checkable: reserving stock safely requires SELECT ... FOR UPDATE, and
// the query builder cannot emit it. orm.QueryBuilder exposes no lock clause and SqlSelectGraphOpts
// has no lock field, so there is no combination of engine calls that produces a row lock. Without
// one, two concurrent reservations both read the same available quantity and both reserve it,
// which is precisely the over-reservation BR §8.6 and AC-STOCK-007 exist to prevent.
//
// Everything else about stock still goes through the engine. Only the lock is hand-written, and
// only because the alternative is a race that no amount of application-level care can close.
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
//
// Its quantities are read inside the lock and are therefore current as of the lock being taken —
// which is the only moment at which they can be trusted. A value read before the lock is stale by
// definition, however recently it was fetched.
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
// returns them, ordered oldest first.
//
// Three properties make this correct, and none of them is optional:
//
//   - **It must run inside a transaction.** ExtractClient returns the transaction's client when one
//     is set on the context, and the connection pool's otherwise. On a pooled connection the locks
//     would be released the instant the statement returned, leaving a caller that believes it holds
//     them and does not. The function refuses to run without one rather than silently degrading to
//     no locking at all, because that failure is invisible until two requests collide in production.
//
//   - **The order is deterministic.** Two reservations touching overlapping rows acquire the locks
//     in the same sequence and therefore queue, rather than each holding what the other wants.
//     Ordering by incoming_date alone is not enough — rows can share a timestamp — so id breaks the
//     tie and gives a total order.
//
//   - **The tenant predicate comes from the schema.** Coremart's shadow module adds tenant_id to
//     every table; nikkierp's does not have it. Reading the key from the schema means the same code
//     is correct in both binaries, where a hardcoded column list would either fail to compile
//     against one or, worse, lock another tenant's rows in the other.
//
// The caller must recompute availability from what this returns and never reuse a figure read
// before the call (BR §8.6).
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

// buildQuantLockQuery assembles the locking SELECT with positional placeholders.
//
// The column and table names come from the schema and from this module's own constants, never from
// caller input, so there is nothing here an untrusted value can reach: every value travels as a
// bound argument.
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

// nullDecimalOrZero reads a null quantity as zero, matching AvailableQuantity: a balance that has
// never been written has moved nothing, which is not the same as being unknown.
func nullDecimalOrZero(value decimal.NullDecimal) decimal.Decimal {
	if !value.Valid {
		return decimal.Zero
	}
	return value.Decimal
}
