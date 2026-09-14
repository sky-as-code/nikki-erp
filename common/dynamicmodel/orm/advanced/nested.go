package advanced

import (
	"fmt"
	"strings"

	"github.com/huandu/go-sqlbuilder"
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/computed"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/dynamicmodel/orm"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
)

// Nested edge projection (F3).
//
// A to-one edge in the select list is a LEFT JOIN; its leaves are projected as flat columns
// aliased to the request path ("template.code", "template.uom.name") and the destination
// primary keys always ride along ("template.id"), so a NULL key tells the repository the
// relation is absent. A to-many edge (one:many, many:many) is one LEFT JOIN LATERAL that
// aggregates the matching child rows into a jsonb array — jsonb rather than json because a
// DISTINCT query needs an equality operator — projected under the edge path ("quants",
// "template.quants"). Neither shape multiplies root rows, so no DISTINCT is induced, and a page
// of any size is still exactly one statement.
//
// The same walk that writes the SQL records an orm.NestedProjection, which
// PlanNestedProjection hands to the repository so the flat row can be folded back into the
// nested shape the API promises.

// MaxNestedEdgeLaterals caps the to-many edges one request may project: each one is an
// aggregation per root row.
var MaxNestedEdgeLaterals = 5

// MaxNestedEdgeLateralsForTest swaps the cap and returns the previous value.
func MaxNestedEdgeLateralsForTest(limit int) int {
	prev := MaxNestedEdgeLaterals
	MaxNestedEdgeLaterals = limit
	return prev
}

// pathShape classifies a requested select path.
type pathShape struct {
	// toOneEdges is the to-one chain leading to the owner of the leaf (or of the to-many edge).
	toOneEdges []string
	// toManyEdge is set when the path crosses a fan-out edge; it is then the last edge.
	toManyEdge string
	toManyRel  dmodel.ModelRelation
	// leaf is the field name on the owner; "" for a bare edge request.
	leaf string
	// bareEdge is the relation named by a path whose last segment is an edge field.
	bareEdge *dmodel.ModelRelation
}

func (this *queryPlan) classifySelectPath(path string) (*pathShape, error) {
	segments, err := splitPath(path, MaxSelectDots)
	if err != nil {
		return nil, err
	}
	shape := &pathShape{}
	schema := this.root
	for i, seg := range segments {
		field, ok := schema.Field(seg)
		if !ok {
			return nil, orm.ErrUnknownField(path)
		}
		isLast := i == len(segments)-1
		if !field.IsEdgeModel() {
			if !isLast {
				return nil, orm.ErrUnknownField(path)
			}
			shape.leaf = seg
			return shape, nil
		}
		if _, err := this.registryOrErr(); err != nil {
			return nil, err
		}
		rel, err := orm.RelationByEdge(schema, seg)
		if err != nil {
			return nil, err
		}
		if shape.toManyEdge != "" {
			return nil, orm.WrapClientErrors(clientErrorsNestedBeyondToMany(path))
		}
		if orm.RelationFansOut(rel) {
			shape.toManyEdge = seg
			shape.toManyRel = rel
		} else {
			shape.toOneEdges = append(shape.toOneEdges, seg)
		}
		if isLast {
			relCopy := rel
			shape.bareEdge = &relCopy
			return shape, nil
		}
		schema = this.registry.Get(rel.DestSchemaName)
		if schema == nil {
			return nil, errors.Errorf("classifySelectPath: schema %q not in registry", rel.DestSchemaName)
		}
	}
	return shape, nil
}

func clientErrorsNestedBeyondToMany(path string) ft.ClientErrors {
	return ft.ClientErrors{
		*ft.NewValidationError(path, "err_graph_nested_beyond_to_many",
			"a to-many edge can only be the last edge of a selected path"),
	}
}

func clientErrorsTooManyNestedEdges(limit int) ft.ClientErrors {
	return ft.ClientErrors{
		*ft.NewValidationError("fields", "err_too_many_nested_edges",
			fmt.Sprintf("request projects more than %d to-many edges", limit)),
	}
}

// projection is the ordered select list under construction: plain expressions interleaved with
// to-many groups, which render last so leaves mentioned anywhere in the list merge into one
// lateral per edge.
type projection struct {
	items  []projItem
	groups map[string]*toManyGroup
	seen   map[string]struct{}
}

type projItem struct {
	expr     string
	groupKey string
}

type toManyGroup struct {
	key        string
	ownerEdges []string
	rel        dmodel.ModelRelation
	dest       *dmodel.ModelSchema
	leaves     []string
	leafSet    map[string]struct{}
}

func newProjection() *projection {
	return &projection{groups: map[string]*toManyGroup{}, seen: map[string]struct{}{}}
}

func (this *projection) addExpr(expr string) {
	if _, ok := this.seen[expr]; ok {
		return
	}
	this.seen[expr] = struct{}{}
	this.items = append(this.items, projItem{expr: expr})
}

// projectPath adds one requested path to the projection, planning joins and laterals as needed.
func (this *queryPlan) projectPath(proj *projection, path string) error {
	shape, err := this.classifySelectPath(path)
	if err != nil {
		return err
	}
	if shape.toManyEdge != "" {
		return this.projectToMany(proj, path, shape)
	}
	if shape.bareEdge != nil {
		return this.projectBareToOne(proj, path, shape)
	}
	if len(shape.toOneEdges) == 0 {
		exprs, err := this.rootSelectExprs(path)
		if err != nil {
			return err
		}
		for _, e := range exprs {
			proj.addExpr(e)
		}
		return nil
	}
	return this.projectToOneLeaf(proj, path, shape.toOneEdges, shape.leaf)
}

func (this *queryPlan) rootSelectExprs(path string) ([]string, error) {
	ref, err := this.resolve(path, useSelect)
	if err != nil {
		return nil, err
	}
	switch ref.kind {
	case refComputedGo:
		return nil, nil
	case refComputedSql, refComputedExpr:
		return []string{ref.sql + " AS " + orm.PgQuote(path)}, nil
	default:
		return []string{ref.sql}, nil
	}
}

// projectToOneLeaf projects one leaf reached through to-one edges, registering every hop's
// primary keys on the way.
func (this *queryPlan) projectToOneLeaf(proj *projection, path string, edges []string, leaf string) error {
	if err := this.ensureToOneHops(proj, edges); err != nil {
		return err
	}
	ref, err := this.resolve(path, useSelect)
	if err != nil {
		return err
	}
	if ref.kind == refComputedGo {
		return nil
	}
	proj.addExpr(ref.sql + " AS " + orm.PgQuote(path))
	// The leaf is recorded under the hop schema's own field, so a related or expression field is
	// typed by its declared data type, not by the column it was compiled from.
	this.noteToOneLeaf(strings.Join(edges, "."), path, leaf)
	return nil
}

// ensureToOneHops joins each prefix of the chain and projects its destination primary keys.
func (this *queryPlan) ensureToOneHops(proj *projection, edges []string) error {
	for i := 1; i <= len(edges); i++ {
		hop := edges[:i]
		alias, dest, err := this.join.EnsureJoinedEdges(hop)
		if err != nil {
			return err
		}
		hopKey := strings.Join(hop, ".")
		entry := this.toOneEntry(hopKey, dest)
		for _, pk := range dest.PrimaryKeys() {
			pkAlias := hopKey + "." + pk
			proj.addExpr(alias + "." + orm.PgQuote(pk) + " AS " + orm.PgQuote(pkAlias))
			if !containsString(entry.PkAliases, pkAlias) {
				entry.PkAliases = append(entry.PkAliases, pkAlias)
			}
			entry.Leaves[pkAlias] = pk
		}
		this.nested.ToOne[hopKey] = entry
	}
	return nil
}

func (this *queryPlan) ensureNested() *orm.NestedProjection {
	if this.nested == nil {
		this.nested = &orm.NestedProjection{ToOne: map[string]orm.ToOneEdgeProjection{}, ToMany: map[string]orm.ToManyEdgeProjection{}}
	}
	return this.nested
}

func (this *queryPlan) toOneEntry(hopKey string, dest *dmodel.ModelSchema) orm.ToOneEdgeProjection {
	entry, ok := this.ensureNested().ToOne[hopKey]
	if !ok {
		entry = orm.ToOneEdgeProjection{DestSchemaName: dest.Name(), Leaves: map[string]string{}}
	}
	return entry
}

func (this *queryPlan) noteToOneLeaf(hopKey, alias, fieldName string) {
	entry, ok := this.ensureNested().ToOne[hopKey]
	if !ok {
		return
	}
	entry.Leaves[alias] = fieldName
	this.nested.ToOne[hopKey] = entry
}

// projectBareToOne expands "edge" to every readable leaf of the destination: physical columns
// and SQL-renderable computed fields; Go-only computed fields are skipped as on the root.
func (this *queryPlan) projectBareToOne(proj *projection, path string, shape *pathShape) error {
	if err := this.ensureToOneHops(proj, shape.toOneEdges); err != nil {
		return err
	}
	dest := this.registry.Get(shape.bareEdge.DestSchemaName)
	if dest == nil {
		return errors.Errorf("projectBareToOne: schema %q not in registry", shape.bareEdge.DestSchemaName)
	}
	for _, name := range projectableLeafNames(dest) {
		if this.skipInBareExpansion(dest, name) {
			continue
		}
		err := this.projectToOneLeaf(proj, path+"."+name, shape.toOneEdges, name)
		if err != nil && isNotRenderable(err) {
			// An expression with no SQL rendering (a date builtin) is left to the Go evaluator,
			// exactly as a Go function field is; only an explicit request reports it.
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func isNotRenderable(err error) bool {
	_, cErrs, ierr := orm.SqlGraphOutcome("", nil, err)
	if ierr != nil || cErrs == nil {
		return false
	}
	for _, item := range *cErrs {
		if strings.HasSuffix(item.Key, "err_computed_field_not_filterable") {
			return true
		}
	}
	return false
}

// skipInBareExpansion drops, from a bare-edge expansion only, an SQL-computed field whose
// filter needs a context value the request did not bind. Naming the field explicitly still
// reports the missing key; a bare edge merely asks for "whatever is available".
func (this *queryPlan) skipInBareExpansion(schema *dmodel.ModelSchema, name string) bool {
	fieldPlan := computedPlanFor(schema, name)
	if fieldPlan == nil || fieldPlan.SqlSource == nil {
		return false
	}
	for _, key := range contextKeysOf(fieldPlan) {
		if _, ok := this.ctxValues[key]; !ok {
			return true
		}
	}
	return false
}

func contextKeysOf(fieldPlan *computed.FieldPlan) []string {
	switch {
	case fieldPlan.Def.Aggregate != nil:
		return fieldPlan.Def.Aggregate.Context
	case fieldPlan.Def.Exists != nil:
		return fieldPlan.Def.Exists.Context
	case fieldPlan.Def.Lookup != nil:
		return fieldPlan.Def.Lookup.Context
	}
	return nil
}

// projectableLeafNames lists a schema's readable scalar fields minus the tenant key, which the
// repository never returns to a client.
func projectableLeafNames(schema *dmodel.ModelSchema) []string {
	tenantKey := schema.TenantKey()
	out := []string{}
	for _, field := range schema.ReadableFields() {
		if field.IsEdgeModel() || field.Name() == tenantKey {
			continue
		}
		out = append(out, field.Name())
	}
	return out
}

func (this *queryPlan) projectToMany(proj *projection, path string, shape *pathShape) error {
	ownerKey := strings.Join(shape.toOneEdges, ".")
	groupKey := shape.toManyEdge
	if ownerKey != "" {
		groupKey = ownerKey + "." + shape.toManyEdge
	}
	group, ok := proj.groups[groupKey]
	if !ok {
		if len(proj.groups) >= MaxNestedEdgeLaterals {
			return orm.WrapClientErrors(clientErrorsTooManyNestedEdges(MaxNestedEdgeLaterals))
		}
		dest := this.registry.Get(shape.toManyRel.DestSchemaName)
		if dest == nil {
			return errors.Errorf("projectToMany: schema %q not in registry", shape.toManyRel.DestSchemaName)
		}
		if err := this.ensureToOneHops(proj, shape.toOneEdges); err != nil {
			return err
		}
		// The lateral correlates on the owner row; when that is the root, the alias must be
		// decided now so every root column rendered from here on is qualified.
		if len(shape.toOneEdges) == 0 {
			this.join.EnsureRootAliased()
		}
		group = &toManyGroup{
			key: groupKey, ownerEdges: shape.toOneEdges, rel: shape.toManyRel, dest: dest,
			leafSet: map[string]struct{}{},
		}
		for _, pk := range dest.PrimaryKeys() {
			group.addLeaf(pk)
		}
		proj.groups[groupKey] = group
		proj.items = append(proj.items, projItem{groupKey: groupKey})
	}
	if shape.leaf == "" {
		for _, name := range projectableLeafNames(group.dest) {
			if this.skipInBareExpansion(group.dest, name) {
				continue
			}
			group.addLeaf(name)
		}
		return nil
	}
	field, ok := group.dest.Field(shape.leaf)
	if !ok || field.IsEdgeModel() {
		return orm.ErrUnknownField(path)
	}
	group.addLeaf(shape.leaf)
	return nil
}

func (this *toManyGroup) addLeaf(name string) {
	if _, ok := this.leafSet[name]; ok {
		return
	}
	this.leafSet[name] = struct{}{}
	this.leaves = append(this.leaves, name)
}

// renderToMany writes the group's lateral and returns the projected expression.
func (this *queryPlan) renderToMany(group *toManyGroup) (string, error) {
	ownerAlias, owner, err := this.ownerOf(group.ownerEdges)
	if err != nil {
		return "", err
	}
	if len(group.ownerEdges) == 0 {
		this.join.EnsureRootAliased()
		ownerAlias = this.join.RootAlias()
	}
	const destAlias = "d"
	const junctionAlias = "j"

	pairs := make([]string, 0, len(group.leaves)*2)
	leaves := make([]string, 0, len(group.leaves))
	for _, name := range group.leaves {
		field, _ := group.dest.Field(name)
		var expr string
		if field.IsVirtual() {
			fieldPlan := computedPlanFor(group.dest, name)
			if fieldPlan == nil || fieldPlan.SqlSource == nil {
				continue
			}
			this.sqlComputedCount++
			if limit := sqlComputedLimit(); this.sqlComputedCount > limit {
				return "", orm.WrapClientErrors(orm.ClientErrorsTooManySqlComputedFields(limit))
			}
			sub, cErrs, err := this.qb.ComputedSubqueryExpr(this.registry, group.dest, destAlias, fieldPlan, this.ctxValues)
			if err != nil {
				return "", err
			}
			if len(cErrs) > 0 {
				return "", orm.WrapClientErrors(cErrs)
			}
			expr = sub
		} else {
			expr = destAlias + "." + orm.PgQuote(name)
		}
		pairs = append(pairs, orm.PgStringLiteral(name)+", "+expr)
		leaves = append(leaves, name)
	}
	orderBy := make([]string, 0, len(group.dest.PrimaryKeys()))
	for _, pk := range group.dest.PrimaryKeys() {
		orderBy = append(orderBy, destAlias+"."+orm.PgQuote(pk))
	}

	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select(fmt.Sprintf("COALESCE(jsonb_agg(jsonb_build_object(%s) ORDER BY %s), '[]'::jsonb) AS v",
		strings.Join(pairs, ", "), strings.Join(orderBy, ", ")))
	sb.From(this.qb.TableExpression(group.dest) + " AS " + destAlias)
	wheres, err := this.toManyCorrelation(sb, group, owner, ownerAlias, destAlias, junctionAlias)
	if err != nil {
		return "", err
	}
	sb.Where(wheres...)
	body, _, err := build(sb, "renderToMany")
	if err != nil {
		return "", err
	}
	spec := this.addLateral("e\x00"+group.key, "e", body, useSelect)
	this.ensureNested().ToMany[group.key] = orm.ToManyEdgeProjection{
		DestSchemaName: group.dest.Name(), ColumnAlias: group.key, Leaves: leaves,
	}
	return spec.alias + ".v AS " + orm.PgQuote(group.key), nil
}

// toManyCorrelation writes the junction join (many:many) and returns the WHERE predicates that
// tie the child rows to the owner row, tenant included when both sides carry one.
func (this *queryPlan) toManyCorrelation(
	sb *sqlbuilder.SelectBuilder, group *toManyGroup, owner *dmodel.ModelSchema,
	ownerAlias, destAlias, junctionAlias string,
) ([]string, error) {
	rel := group.rel
	ownerTk := owner.TenantKey()
	destTk := group.dest.TenantKey()
	var wheres []string
	if rel.RelationType == dmodel.RelationTypeManyToMany {
		through := this.registry.Get(rel.M2mThroughSchemaName)
		if through == nil || rel.M2mSrcFieldPrefix == "" || rel.M2mDestFieldPrefix == "" {
			return nil, errors.Errorf("toManyCorrelation: many-to-many relation %q is not finalized", rel.Edge)
		}
		var on []string
		for _, pk := range group.dest.PrimaryKeys() {
			on = append(on, fmt.Sprintf("%s.%s = %s.%s", junctionAlias,
				orm.PgQuote(dmodel.PrefixedThroughColumn(rel.M2mDestFieldPrefix, pk)), destAlias, orm.PgQuote(pk)))
		}
		if ownerTk != "" && destTk != "" {
			on = append(on, fmt.Sprintf("%s.%s = %s.%s", junctionAlias, orm.PgQuote(ownerTk), destAlias, orm.PgQuote(destTk)))
		}
		sb.Join(this.qb.TableExpression(through)+" AS "+junctionAlias, strings.Join(on, " AND "))
		for _, pk := range owner.PrimaryKeys() {
			wheres = append(wheres, fmt.Sprintf("%s.%s = %s.%s", junctionAlias,
				orm.PgQuote(dmodel.PrefixedThroughColumn(rel.M2mSrcFieldPrefix, pk)), ownerAlias, orm.PgQuote(pk)))
		}
		if ownerTk != "" {
			wheres = append(wheres, fmt.Sprintf("%s.%s = %s.%s", junctionAlias, orm.PgQuote(ownerTk), ownerAlias, orm.PgQuote(ownerTk)))
		}
		return wheres, nil
	}
	for _, pair := range rel.EffectiveForeignKeys() {
		wheres = append(wheres, fmt.Sprintf("%s.%s = %s.%s", destAlias, orm.PgQuote(pair.FkColumn),
			ownerAlias, orm.PgQuote(pair.ReferencedColumn)))
	}
	if ownerTk != "" && destTk != "" {
		wheres = append(wheres, fmt.Sprintf("%s.%s = %s.%s", destAlias, orm.PgQuote(destTk), ownerAlias, orm.PgQuote(ownerTk)))
	}
	return wheres, nil
}

// PlanNestedProjection implements orm.NestedProjectionCapable: it walks the column list exactly
// as SqlSelectGraph does and returns the folding plan, or nil when no edge is projected.
func (this *AdvancedPgQueryBuilder) PlanNestedProjection(
	schema *dmodel.ModelSchema, registry *dmodel.SchemaRegistry, columns []orm.SelectColumn,
) (*orm.NestedProjection, *ft.ClientErrors, error) {
	plan, err := this.newPlan(schema, registry, nil, orm.SqlSelectGraphOpts{Columns: columns}, planSelect)
	if err == nil {
		err = plan.applySelectColumns(sqlbuilder.PostgreSQL.NewSelectBuilder(), columns)
	}
	if err != nil {
		_, cErrs, ierr := orm.SqlGraphOutcome("", nil, err)
		return nil, cErrs, ierr
	}
	if plan.nested == nil || (len(plan.nested.ToOne) == 0 && len(plan.nested.ToMany) == 0) {
		return nil, nil, nil
	}
	return plan.nested, nil, nil
}

var _ orm.NestedProjectionCapable = (*AdvancedPgQueryBuilder)(nil)
