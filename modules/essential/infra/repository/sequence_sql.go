// The Sequence repositories hold the only raw SQL in Essential, a deliberate exception to the
// engine-first rule (docs/wiki/07 §6.7): issuing a number needs INSERT ... ON CONFLICT DO NOTHING /
// DO UPDATE ... WHERE ... RETURNING, SELECT ... FOR SHARE and a transaction-scoped advisory lock,
// none of which orm.QueryBuilder can emit. Table and column names come from the schema and this
// module's constants; every value is a bound argument.
//
// The multi-tenant decorator does not see these statements, so each one carries the tenant key
// itself when the schema has one (the hosted binary), read from the request's domain constraints.
package repository

import (
	"strconv"
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// errNoTransaction refuses a statement whose locks or atomicity mean nothing outside a transaction.
func errNoTransaction(operation string) error {
	return errors.New(operation + " requires an ambient transaction: without one its row locks are " +
		"released as soon as the statement returns")
}

// sqlArgs numbers bound arguments as they are added, so a statement is assembled in reading order.
type sqlArgs struct {
	values []any
}

func (this *sqlArgs) bind(value any) string {
	this.values = append(this.values, value)
	return "$" + strconv.Itoa(len(this.values))
}

// tenantScope is the tenant column and value a statement must carry, both empty in a binary
// without a tenant key.
type tenantScope struct {
	column string
	value  any
}

func resolveTenant(ctx corectx.Context, schema *dmodel.ModelSchema, operation string) (tenantScope, error) {
	column := schema.TenantKey()
	if column == "" {
		return tenantScope{}, nil
	}
	constraints := ctx.GetDomainConstraints()
	value, ok := constraints[column]
	if !ok || value == nil || value == "" {
		return tenantScope{}, errors.New(operation + ": the request carries no tenant id")
	}
	return tenantScope{column: column, value: value}, nil
}

// predicate is "AND tenant = $n", or nothing without a tenant key.
func (this tenantScope) predicate(args *sqlArgs) string {
	if this.column == "" {
		return ""
	}
	return " AND " + this.column + " = " + args.bind(this.value)
}

// conflictColumns prefixes the tenant key, which the DDL generator puts first in every unique key.
func (this tenantScope) conflictColumns(columns ...string) string {
	if this.column == "" {
		return strings.Join(columns, ", ")
	}
	return strings.Join(append([]string{this.column}, columns...), ", ")
}

// insertColumns appends the tenant column and its placeholder to an INSERT's lists.
func (this tenantScope) insertColumns(columns []string, values []string, args *sqlArgs) ([]string, []string) {
	if this.column == "" {
		return columns, values
	}
	return append(columns, this.column), append(values, args.bind(this.value))
}

// actingPrincipal is the id stamped into created_by, nil when nobody is authenticated.
func actingPrincipal(ctx corectx.Context) any {
	perms := ctx.GetPermissions()
	if perms.UserId != "" {
		return string(perms.UserId)
	}
	if perms.Principal.Id != "" {
		return string(perms.Principal.Id)
	}
	return nil
}

func nullableId(id *string) any {
	if id == nil || *id == "" {
		return nil
	}
	return *id
}
