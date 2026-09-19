package advanced

import (
	"strings"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// Computed fields in the advanced builder.
//
// An SQL-computed field (aggregate / exists / lookup) becomes one LEFT JOIN LATERAL per
// (owner alias, field) pair: "LEFT JOIN LATERAL (SELECT <correlated subquery> AS v) cN ON TRUE".
// The projection, the WHERE clause and ORDER BY all read cN.v, so a field named in both the
// filter and the selection is evaluated once per root row instead of once per mention. The
// subquery is exactly what PgQueryBuilder emits as a scalar subquery (orm.ComputedSubqueryExpr),
// so tenant correlation, archive scoping, filter escaping and context binding are unchanged.
//
// A lateral whose subquery is an aggregate without GROUP BY, an EXISTS, or a LIMIT 1 lookup
// yields exactly one row per outer row, so it never fans the root out and never needs DISTINCT.

// resolveComputed answers a path whose leaf is a virtual (computed) field on owner.
func (this *queryPlan) resolveComputed(
	path string, owner *dmodel.ModelSchema, ownerAlias string, field *dmodel.ModelField, u use,
) (*resolvedRef, error) {
	fieldPlan := computedPlanFor(owner, field.Name())
	if fieldPlan == nil {
		return this.resolveGoOnly(path, owner, field, u)
	}
	switch fieldPlan.Def.Kind {
	case computed.ComputeAggregate, computed.ComputeExists, computed.ComputeLookup:
		return this.resolveSqlKind(path, owner, ownerAlias, field, fieldPlan, u)
	case computed.ComputeRelated:
		return this.resolveRelated(path, owner, field, fieldPlan, u)
	case computed.ComputeExpression:
		return this.resolveExpression(path, owner, ownerAlias, field, fieldPlan, u)
	default:
		return this.resolveGoOnly(path, owner, field, u)
	}
}

func computedPlanFor(owner *dmodel.ModelSchema, fieldName string) *computed.FieldPlan {
	schemaPlan := computed.PlanFor(owner.Name())
	if schemaPlan == nil {
		return nil
	}
	return schemaPlan.Fields[fieldName]
}

// resolveGoOnly handles a virtual field with no SQL rendering (a Go function, or a computed
// field the plan cache does not know). It stays selectable on the root, where a service fills
// it after the read; anywhere else it is a client error.
func (this *queryPlan) resolveGoOnly(
	path string, owner *dmodel.ModelSchema, field *dmodel.ModelField, u use,
) (*resolvedRef, error) {
	if u == useSelect {
		return &resolvedRef{field: field, kind: refComputedGo, owner: owner, path: path}, nil
	}
	if u == useOrder {
		return nil, orm.WrapClientErrors(orm.ClientErrorsFieldNotSortable(path))
	}
	return nil, orm.WrapClientErrors(clientErrorsComputedNotFilterable(path,
		"computed field is evaluated in Go and cannot be used in a filter"))
}

func (this *queryPlan) resolveSqlKind(
	path string, owner *dmodel.ModelSchema, ownerAlias string, field *dmodel.ModelField,
	fieldPlan *computed.FieldPlan, u use,
) (*resolvedRef, error) {
	registry, err := this.registryOrErr()
	if err != nil {
		return nil, err
	}
	if err := this.rejectFanOutOwner(path); err != nil {
		return nil, err
	}
	if owner == this.root {
		this.join.EnsureRootAliased()
		ownerAlias = this.join.RootAlias()
	}
	key := ownerAlias + "\x00" + field.Name()
	spec, ok := this.lateralByKey[key]
	if !ok {
		this.sqlComputedCount++
		if limit := sqlComputedLimit(); this.sqlComputedCount > limit {
			return nil, orm.WrapClientErrors(orm.ClientErrorsTooManySqlComputedFields(limit))
		}
		subquery, cErrs, err := this.qb.ComputedSubqueryExpr(registry, owner, ownerAlias, fieldPlan, this.ctxValues)
		if err != nil {
			return nil, err
		}
		if len(cErrs) > 0 {
			return nil, orm.WrapClientErrors(cErrs)
		}
		spec = this.addLateral(key, "c", "SELECT "+subquery+" AS v", u)
	} else {
		spec.uses |= u
	}
	return &resolvedRef{
		field: syntheticField(field), kind: refComputedSql, owner: owner, path: path,
		sql: spec.alias + ".v",
	}, nil
}

// syntheticField is the non-virtual twin of a computed field, carrying the field's declared
// data type. The shared predicate compiler converts and validates filter values against a
// field, and refuses virtual ones; this twin lets a computed reference pass through that path
// unchanged, so a decimal aggregate is compared as a decimal and a boolean exists as a boolean.
func syntheticField(field *dmodel.ModelField) *dmodel.ModelField {
	return dmodel.DefineField().Name(field.Name()).DataType(field.DataType()).Build()
}

// rejectFanOutOwner refuses a computed field reached through a to-many edge: the value would
// be one per child row, which has no single meaning on the root row.
func (this *queryPlan) rejectFanOutOwner(path string) error {
	segments := strings.Split(path, ".")
	schema := this.root
	for _, edge := range segments[:len(segments)-1] {
		rel, err := orm.RelationByEdge(schema, edge)
		if err != nil {
			return err
		}
		if orm.RelationFansOut(rel) {
			return orm.WrapClientErrors(clientErrorsComputedNotFilterable(path,
				"computed field cannot be reached through a to-many edge"))
		}
		schema = this.registry.Get(rel.DestSchemaName)
		if schema == nil {
			return orm.ErrUnknownField(path)
		}
	}
	return nil
}

func clientErrorsComputedNotFilterable(field, message string) ft.ClientErrors {
	return ft.ClientErrors{
		*ft.NewValidationError(field, "err_computed_field_not_filterable", message),
	}
}
