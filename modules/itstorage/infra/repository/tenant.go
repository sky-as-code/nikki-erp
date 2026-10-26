package repository

import (
	"fmt"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// tenantScope returns the tenant column and the acting tenant for a set-based statement, which
// the multi-tenant decorator does not see. Both are empty in a binary without tenants. A table
// with a tenant column and a request without a tenant is refused rather than widened.
func tenantScope(ctx corectx.Context, schema *dmodel.ModelSchema, action string) (string, string, error) {
	column := schema.TenantKey()
	if column == "" {
		return "", "", nil
	}
	constraints := ctx.GetDomainConstraints()
	raw, ok := constraints[column]
	if !ok || raw == nil {
		return "", "", errors.Errorf("%s: the request carries no tenant id", action)
	}
	tenantId := fmt.Sprint(raw)
	if ptr, isPtr := raw.(*string); isPtr {
		tenantId = *ptr
	}
	if tenantId == "" {
		return "", "", errors.Errorf("%s: the request carries no tenant id", action)
	}
	return column, tenantId, nil
}

// execCount runs a statement on the request's client, which is the open transaction when the
// context carries one, and returns the rows it touched.
func execCount(ctx corectx.Context, repo composable.CrudRepository, query string, args []any, action string) (int, error) {
	result, err := repo.GetBaseRepo().ExtractClient(ctx).Exec(ctx, query, args...)
	if err != nil {
		return 0, errors.Wrap(err, action)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, nil
	}
	return int(affected), nil
}

// queryStrings runs a single-column SELECT and returns its values.
func queryStrings(
	ctx corectx.Context, repo composable.CrudRepository, query string, args []any, action string,
) ([]string, error) {
	rows, err := repo.GetBaseRepo().ExtractClient(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, action)
	}
	defer rows.Close()

	values := []string{}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, errors.Wrap(err, action)
		}
		values = append(values, value)
	}
	return values, errors.Wrap(rows.Err(), action)
}
