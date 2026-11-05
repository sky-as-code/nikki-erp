// The only raw SQL of the promotion family, a deliberate exception to the engine-first rule: taking
// the last use of a promotion or voucher safely needs UPDATE ... WHERE used_count < usage_limit
// RETURNING, which the query builder cannot emit (D5). The multi-tenant decorator does not see these
// statements, so each carries the tenant key itself when the schema has one.
package repository

import (
	"strconv"

	"go.bryk.io/pkg/errors"

	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"

	itPromotion "github.com/sky-as-code/nikki-erp/modules/commerce/sales/interfaces/promotion"
)

const (
	usageColumnId        = "id"
	usageColumnUsedCount = "used_count"
	usageColumnLimit     = "usage_limit"
)

var (
	_ itPromotion.UsageCounterRepository = (*SalesPromotionRepositoryImpl)(nil)
	_ itPromotion.UsageCounterRepository = (*SalesPromotionChildRepositoryImpl)(nil)
)

func (this *SalesPromotionRepositoryImpl) TryConsumeUsage(ctx corectx.Context, id string) (bool, error) {
	return changeUsage(ctx, this.GetBaseRepo(), id, true)
}

func (this *SalesPromotionRepositoryImpl) ReleaseUsage(ctx corectx.Context, id string) (bool, error) {
	return changeUsage(ctx, this.GetBaseRepo(), id, false)
}

func (this *SalesPromotionChildRepositoryImpl) TryConsumeUsage(ctx corectx.Context, id string) (bool, error) {
	return changeUsage(ctx, this.GetBaseRepo(), id, true)
}

func (this *SalesPromotionChildRepositoryImpl) ReleaseUsage(ctx corectx.Context, id string) (bool, error) {
	return changeUsage(ctx, this.GetBaseRepo(), id, false)
}

func changeUsage(ctx corectx.Context, repo dyn.BaseDynamicRepository, id string, consume bool) (bool, error) {
	if ctx == nil || ctx.GetDbTranx() == nil {
		return false, errors.New("changing a usage count requires an ambient transaction: " +
			"outside one a failed confirm could not give the use back")
	}
	query, args := buildUsageSql(ctx, repo, id, consume)
	rows, err := repo.ExtractClient(ctx).Query(ctx, query, args...)
	if err != nil {
		return false, errors.Wrap(err, "changing a usage count")
	}
	defer rows.Close()
	changed := rows.Next()
	return changed, errors.Wrap(rows.Err(), "changing a usage count")
}

// buildUsageSql assembles the conditional UPDATE. Names come from constants and the schema, every
// value is a bound argument.
func buildUsageSql(ctx corectx.Context, repo dyn.BaseDynamicRepository, id string, consume bool) (string, []any) {
	schema := repo.Schema()
	args := []any{id}
	used := "COALESCE(" + usageColumnUsedCount + ", 0)"
	set := usageColumnUsedCount + " = " + used + " + 1"
	guard := " AND (" + usageColumnLimit + " IS NULL OR " + used + " < " + usageColumnLimit + ")"
	if !consume {
		set = usageColumnUsedCount + " = " + used + " - 1"
		guard = " AND " + used + " > 0"
	}
	where := usageColumnId + " = $1" + guard
	if tenantKey := schema.TenantKey(); tenantKey != "" {
		if value, ok := ctx.GetDomainConstraints()[tenantKey]; ok && value != nil {
			args = append(args, value)
			where += " AND " + tenantKey + " = $" + strconv.Itoa(len(args))
		}
	}
	query := "UPDATE " + schema.TableName() + " SET " + set + " WHERE " + where + " RETURNING " + usageColumnId
	return query, args
}
