package advanced

import (
	"fmt"
	"strings"

	"github.com/huandu/go-sqlbuilder"
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
)

// Depth caps. Filters keep PgQueryBuilder's five hops. Selection and ordering may walk further
// than before because to-one hops are plain joins here; the cap only bounds the join count.
const (
	MaxSelectDots = 3
	MaxOrderDots  = 3
	MaxFilterDots = orm.MaxSearchGraphConditionDots
)

// use records where a path (and so a lateral it needs) is referenced. The count and exists
// queries drop laterals that only the projection needs; a fan-out-safe DISTINCT count keeps
// the projected ones as well.
type use int

const (
	useFilter use = 1 << iota
	useOrder
	useSelect
)

const lateralUseAll = useFilter | useOrder | useSelect

type planMode int

const (
	planSelect planMode = iota
	planCount
	planExists
)

// lateralSpec is one LEFT JOIN LATERAL (...) AS alias ON TRUE. Computed laterals expose a
// single column named v; edge laterals (to-many projection) expose v as a jsonb array.
type lateralSpec struct {
	alias string
	body  string
	uses  use
}

type queryPlan struct {
	qb        *AdvancedPgQueryBuilder
	registry  *dmodel.SchemaRegistry
	root      *dmodel.ModelSchema
	join      *orm.JoinPlan
	ctxValues map[string]any
	mode      planMode

	laterals     []*lateralSpec
	lateralByKey map[string]*lateralSpec
	nextComputed int
	nextEdge     int
	// sqlComputedCount enforces computed.ActiveLimits().MaxSqlComputedFieldsPerRequest over
	// distinct laterals, not occurrences.
	sqlComputedCount int
	// exprDepth guards recursive expression compilation (an expression over an expression).
	exprDepth int
	// nested is the folding plan for projected edge fields, built alongside the projection.
	nested *orm.NestedProjection
}

func (this *AdvancedPgQueryBuilder) newPlan(
	schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry, graph *dmodel.SearchGraph,
	opts orm.SqlSelectGraphOpts, mode planMode,
) (*queryPlan, error) {
	if schema == nil {
		return nil, errors.New("AdvancedPgQueryBuilder: schema is required")
	}
	plan := &queryPlan{
		qb:           this,
		registry:     registry,
		root:         schema,
		join:         this.NewJoinPlan(registry, schema),
		ctxValues:    opts.ComputedContext,
		mode:         mode,
		lateralByKey: map[string]*lateralSpec{},
	}
	return plan, plan.prepass(graph, opts)
}

// prepass resolves every path the query mentions before any SQL is written, so joins and
// laterals exist (and the root-alias decision is final) by the time refs are rendered.
func (this *queryPlan) prepass(graph *dmodel.SearchGraph, opts orm.SqlSelectGraphOpts) error {
	if this.mode == planSelect {
		for _, col := range opts.Columns {
			if _, inner, ok := orm.ParseAllowedAggregate(col); ok {
				if _, err := this.resolve(orm.SelectColumnPath(orm.SelectColumn(inner)), useSelect); err != nil {
					return err
				}
				continue
			}
			path := col.JoinPlanningPath()
			if path == "" {
				continue
			}
			if err := this.projectPath(newProjection(), path); err != nil {
				return err
			}
		}
	}
	if graph == nil {
		return nil
	}
	if err := this.prepassNode(graph.GetCondition(), graph.GetAnd(), graph.GetOr()); err != nil {
		return err
	}
	if this.mode == planSelect {
		for _, item := range graph.GetOrder() {
			if item.Field() == "" {
				continue
			}
			if _, err := this.resolve(item.Field(), useOrder); err != nil {
				return err
			}
		}
	}
	return nil
}

func (this *queryPlan) prepassNode(cond dmodel.Condition, and, or []dmodel.SearchNode) error {
	if field := strings.TrimSpace(cond.Field()); field != "" {
		op := cond.Operator()
		if op == dmodel.Linked || op == dmodel.NotLinked {
			// The linked predicate correlates on the root primary key, so the root must be aliased.
			this.join.EnsureRootAliased()
		} else if _, err := this.resolve(field, useFilter); err != nil {
			return err
		}
	}
	for i := range and {
		if err := this.prepassNode(and[i].GetCondition(), and[i].GetAnd(), and[i].GetOr()); err != nil {
			return err
		}
	}
	for i := range or {
		if err := this.prepassNode(or[i].GetCondition(), or[i].GetAnd(), or[i].GetOr()); err != nil {
			return err
		}
	}
	return nil
}

func (this *queryPlan) registryOrErr() (*dmodel.SchemaRegistry, error) {
	if this.registry == nil {
		return nil, orm.WrapClientErrors(orm.ClientErrorsRegistryRequiredForGraph())
	}
	return this.registry, nil
}

func maxDotsFor(u use) int {
	switch u {
	case useSelect:
		return MaxSelectDots
	case useOrder:
		return MaxOrderDots
	default:
		return MaxFilterDots
	}
}

func splitPath(path string, maxDots int) ([]string, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("field path is required")
	}
	if strings.Count(path, ".") > maxDots {
		return nil, orm.WrapClientErrors(orm.ClientErrorsGraphFieldTooDeep(path, maxDots))
	}
	parts := strings.Split(path, ".")
	for _, seg := range parts {
		if seg == "" {
			return nil, orm.WrapClientErrors(orm.ClientErrorsInvalidGraphFieldPath(path))
		}
	}
	return parts, nil
}

// refKind says what a resolved path stands for.
type refKind int

const (
	// refPhysical is a stored column, on the root or on a joined edge.
	refPhysical refKind = iota
	// refComputedSql is an SQL-computed field materialised as a lateral column.
	refComputedSql
	// refComputedGo is a virtual field with no SQL rendering: legal to select (a service fills
	// it), illegal to filter or sort on.
	refComputedGo
	// refComputedExpr is an expression field compiled inline to SQL. It is filterable and
	// sortable; on the root projection it is still left to the Go evaluator.
	refComputedExpr
)

type resolvedRef struct {
	field *dmodel.ModelField
	sql   string
	kind  refKind
	// owner is the schema the field belongs to; path is the request's dotted path.
	owner *dmodel.ModelSchema
	path  string
}

// resolve turns a dotted path into a SQL reference for the given use, planning whatever joins
// and laterals it needs. Every consumer — projection, predicate, order — goes through here so a
// path always maps to one alias.
func (this *queryPlan) resolve(path string, u use) (*resolvedRef, error) {
	return this.resolveWith(path, u, maxDotsFor(u))
}

func (this *queryPlan) resolveWith(path string, u use, maxDots int) (*resolvedRef, error) {
	segments, err := splitPath(path, maxDots)
	if err != nil {
		return nil, err
	}
	ownerAlias, owner, err := this.ownerOf(segments[:len(segments)-1])
	if err != nil {
		return nil, err
	}
	leaf := segments[len(segments)-1]
	field, ok := owner.Field(leaf)
	if !ok || field.IsEdgeModel() {
		return nil, orm.ErrUnknownField(path)
	}
	if field.IsVirtual() {
		return this.resolveComputed(path, owner, ownerAlias, field, u)
	}
	return &resolvedRef{
		field: field, kind: refPhysical, owner: owner, path: path,
		sql: this.columnRef(ownerAlias, leaf),
	}, nil
}

// ownerOf joins the edge chain and returns the alias and schema its last hop lands on. An empty
// chain is the root, whose alias may still be empty when nothing has forced one.
func (this *queryPlan) ownerOf(edges []string) (string, *dmodel.ModelSchema, error) {
	if len(edges) == 0 {
		return this.join.RootAlias(), this.root, nil
	}
	if _, err := this.registryOrErr(); err != nil {
		return "", nil, err
	}
	return this.join.EnsureJoinedEdges(edges)
}

// columnRef qualifies a column with its alias. Root columns stay bare until something aliases
// the root, which keeps a plain single-table query byte-identical to PgQueryBuilder's.
func (this *queryPlan) columnRef(alias, column string) string {
	if alias == "" {
		alias = this.join.RootAlias()
	}
	if alias == "" {
		return orm.PgQuote(column)
	}
	return alias + "." + orm.PgQuote(column)
}

// addLateral registers (or reuses) a LEFT JOIN LATERAL keyed by its owner alias and field name,
// and returns the alias its single column v is read through.
func (this *queryPlan) addLateral(key, prefix, body string, u use) *lateralSpec {
	if existing, ok := this.lateralByKey[key]; ok {
		existing.uses |= u
		return existing
	}
	var alias string
	if prefix == "c" {
		alias = fmt.Sprintf("c%d", this.nextComputed)
		this.nextComputed++
	} else {
		alias = fmt.Sprintf("e%d", this.nextEdge)
		this.nextEdge++
	}
	spec := &lateralSpec{alias: alias, body: body, uses: u}
	this.laterals = append(this.laterals, spec)
	this.lateralByKey[key] = spec
	return spec
}

// applyFrom writes FROM, the planned LEFT JOINs, the laterals whose uses intersect wanted, and
// the tenant predicates a junction join needs.
func (this *queryPlan) applyFrom(sb *sqlbuilder.SelectBuilder, wanted use) {
	this.qb.ApplyFromWithJoins(sb, this.join)
	for _, lat := range this.laterals {
		if lat.uses&wanted == 0 {
			continue
		}
		sb.JoinWithOption(sqlbuilder.LeftJoin, fmt.Sprintf("LATERAL (%s) AS %s", lat.body, lat.alias), "TRUE")
	}
	for _, w := range this.join.M2MTenantWheres() {
		sb.Where(w)
	}
}

// useResolver adapts the plan to orm.GraphRefResolver for one use, so the shared predicate and
// order compilers resolve paths through the plan.
type useResolver struct {
	plan *queryPlan
	u    use
}

func (this *queryPlan) forUse(u use) orm.GraphRefResolver {
	return &useResolver{plan: this, u: u}
}

func (this *useResolver) ResolveFilterRef(name string) (*dmodel.ModelField, string, bool, error) {
	return this.resolveFor(name)
}

func (this *useResolver) ResolveOrderRef(name string) (*dmodel.ModelField, string, bool, error) {
	return this.resolveFor(name)
}

func (this *useResolver) resolveFor(name string) (*dmodel.ModelField, string, bool, error) {
	ref, err := this.plan.resolve(name, this.u)
	if err != nil {
		return nil, "", false, err
	}
	if ref.kind == refComputedGo {
		return nil, "", false, orm.WrapClientErrors(orm.ClientErrorsVirtualFieldUnavailable(name))
	}
	return ref.field, ref.sql, true, nil
}
