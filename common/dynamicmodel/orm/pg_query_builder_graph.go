package orm

import (
	"fmt"
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
)

const MaxSelectGraphColumnDots = 1
const MaxOrderGraphFieldDots = 1
const MaxSearchGraphConditionDots = 5

type graphJoinSpec struct {
	tableWithAlias string
	onExpr         string
}

type joinPlanner struct {
	qb              *PgQueryBuilder
	registry        *dmodel.SchemaRegistry
	root            *dmodel.ModelSchema
	pathKeyAli      map[string]string
	pathKeySch      map[string]*dmodel.ModelSchema
	joins           []graphJoinSpec
	m2mTenantWheres []string
	nextTableIdx    int
	rootAlias       string
	// hasFanOutJoin is set when a planned join can multiply root rows. Such a join makes the
	// row set and COUNT(*) both report duplicates, so the query needs DISTINCT to keep the
	// caller's grain.
	hasFanOutJoin bool
}

// relationFansOut reports whether traversing rel can yield several destination rows per source
// row. The cases mirror joinOnExpr's switch, so the two cannot disagree about direction.
func relationFansOut(rel dmodel.ModelRelation) bool {
	switch rel.RelationType {
	case dmodel.RelationTypeOneToMany, dmodel.RelationTypeManyToMany:
		return true
	case dmodel.RelationTypeOneToOne:
		// An inverse one:one joins dest.fk = parent.pk. Uniqueness there is a modelling promise
		// rather than something the planner can see, and treating it as fan-out would force
		// DISTINCT on every such query. Matches joinOnExpr's handling.
		return false
	default:
		return false
	}
}

// needsDistinct reports whether this plan produces duplicate root rows.
func (p *joinPlanner) needsDistinct() bool {
	return p != nil && p.hasFanOutJoin
}

func newJoinPlanner(qb *PgQueryBuilder, registry *dmodel.SchemaRegistry, root *dmodel.ModelSchema) *joinPlanner {
	return &joinPlanner{
		qb:              qb,
		registry:        registry,
		root:            root,
		pathKeyAli:      make(map[string]string),
		pathKeySch:      make(map[string]*dmodel.ModelSchema),
		m2mTenantWheres: nil,
		nextTableIdx:    1,
	}
}

func (p *joinPlanner) usesJoins() bool {
	return len(p.joins) > 0
}

func (p *joinPlanner) ensureRootAliased() {
	if p.rootAlias == "" {
		p.rootAlias = "t0"
	}
}

func (p *joinPlanner) allocJoinedAlias() string {
	alias := fmt.Sprintf("t%d", p.nextTableIdx)
	p.nextTableIdx++
	return alias
}

func parseDottedPath(field string, maxDots int) ([]string, error) {
	if strings.TrimSpace(field) == "" {
		return nil, errors.New("parseDottedPath: field name is required")
	}
	if strings.Count(field, ".") > maxDots {
		return nil, wrapClientSqlErrors(clientErrorsGraphFieldTooDeep(field, maxDots))
	}
	parts := strings.Split(field, ".")
	for _, seg := range parts {
		if seg == "" {
			return nil, wrapClientSqlErrors(clientErrorsInvalidGraphFieldPath(field))
		}
	}
	return parts, nil
}

func clientErrorsGraphFieldTooDeep(field string, maxDots int) ft.ClientErrors {
	return ft.ClientErrors{
		*ft.NewValidationError(field, ft.ErrorKey("err_graph_field_path_too_deep"),
			fmt.Sprintf("field path exceeds maximum of %d dot separators", maxDots)),
	}
}

func clientErrorsInvalidGraphFieldPath(field string) ft.ClientErrors {
	return ft.ClientErrors{
		*ft.NewValidationError(field, ft.ErrorKey("err_invalid_graph_field_path"),
			"field path must not contain empty segments"),
	}
}

func clientErrorsRegistryRequiredForGraph() ft.ClientErrors {
	return ft.ClientErrors{
		*ft.NewAnonymousValidationError(ft.ErrorKey("err_schema_registry_required"),
			"schema registry is required for relation field paths", nil),
	}
}

func clientErrorsUnsupportedRelationInGraph(edge string, relType dmodel.RelationType) ft.ClientErrors {
	return ft.ClientErrors{
		*ft.NewValidationError(edge, ft.ErrorKey("err_graph_relation_not_supported"),
			fmt.Sprintf("relation type %s is not supported in graph queries", relType)),
	}
}

func relationByEdge(schema *dmodel.ModelSchema, edge string) (dmodel.ModelRelation, error) {
	for _, rel := range schema.Relations() {
		if rel.Edge == edge {
			return rel, nil
		}
	}
	return dmodel.ModelRelation{}, errors.Wrap(&errClientUnknownField{Field: edge}, "relationByEdge")
}

func validateImplicitEdgeModelField(schema *dmodel.ModelSchema, edge string) error {
	field, ok := schema.Field(edge)
	if !ok {
		return errors.Wrap(&errClientUnknownField{Field: edge}, "validateImplicitEdgeModelField")
	}
	if !dmodel.IsFieldDataTypeModel(field.DataType()) {
		return wrapClientSqlErrors(ft.ClientErrors{
			*ft.NewValidationError(edge, ft.ErrorKey("err_graph_edge_not_model_field"),
				"field must be a model (relation) type for graph traversal"),
		})
	}
	return nil
}

func (p *joinPlanner) joinOnExpr(
	parentAlias, destAlias string, _, _ *dmodel.ModelSchema, rel dmodel.ModelRelation,
) (string, error) {
	switch rel.RelationType {
	case dmodel.RelationTypeManyToOne:
		return joinExprManyToOneOrOneToOne(parentAlias, destAlias, rel), nil
	case dmodel.RelationTypeOneToOne:
		if rel.IsInverse {
			return joinExprOneToMany(parentAlias, destAlias, rel), nil
		}
		return joinExprManyToOneOrOneToOne(parentAlias, destAlias, rel), nil
	case dmodel.RelationTypeOneToMany:
		return joinExprOneToMany(parentAlias, destAlias, rel), nil
	case dmodel.RelationTypeManyToMany:
		return "", wrapClientSqlErrors(clientErrorsUnsupportedRelationInGraph(rel.Edge, rel.RelationType))
	default:
		return "", wrapClientSqlErrors(clientErrorsUnsupportedRelationInGraph(rel.Edge, rel.RelationType))
	}
}

func joinExprManyToOneOrOneToOne(parentAlias, destAlias string, rel dmodel.ModelRelation) string {
	pairs := rel.EffectiveForeignKeys()
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = fmt.Sprintf("%s.%s = %s.%s",
			parentAlias, pgQuote(p.FkColumn), destAlias, pgQuote(p.ReferencedColumn))
	}
	return strings.Join(parts, " AND ")
}

func joinExprOneToMany(parentAlias, destAlias string, rel dmodel.ModelRelation) string {
	pairs := rel.EffectiveForeignKeys()
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = fmt.Sprintf("%s.%s = %s.%s",
			destAlias, pgQuote(p.FkColumn), parentAlias, pgQuote(p.ReferencedColumn))
	}
	return strings.Join(parts, " AND ")
}

func (p *joinPlanner) ensureJoinedEdges(edgeChain []string) (alias string, sch *dmodel.ModelSchema, err error) {
	if len(edgeChain) == 0 {
		if p.rootAlias != "" {
			return p.rootAlias, p.root, nil
		}
		return "", p.root, nil
	}
	p.ensureRootAliased()
	key := strings.Join(edgeChain, ".")
	if existing, ok := p.pathKeyAli[key]; ok {
		return existing, p.pathKeySch[key], nil
	}
	return p.appendJoinForEdgePrefix(edgeChain, key)
}

func (p *joinPlanner) appendJoinForEdgePrefix(edgeChain []string, cacheKey string) (string, *dmodel.ModelSchema, error) {
	parentAlias, parentSch, err := p.ensureJoinedEdges(edgeChain[:len(edgeChain)-1])
	if err != nil {
		return "", nil, err
	}
	edge := edgeChain[len(edgeChain)-1]
	rel, err := relationByEdge(parentSch, edge)
	if err != nil {
		return "", nil, err
	}
	if err := validateImplicitEdgeModelField(parentSch, edge); err != nil {
		return "", nil, err
	}
	if relationFansOut(rel) {
		p.hasFanOutJoin = true
	}
	if rel.RelationType == dmodel.RelationTypeManyToMany {
		return p.appendManyToManyJoin(parentAlias, parentSch, rel, cacheKey)
	}
	destSch := p.registry.Get(rel.DestSchemaName)
	if destSch == nil {
		return "", nil, errors.Errorf("joinPlanner: schema '%s' not found in registry", rel.DestSchemaName)
	}
	destAlias := p.allocJoinedAlias()
	onExpr, err := p.joinOnExpr(parentAlias, destAlias, parentSch, destSch, rel)
	if err != nil {
		return "", nil, err
	}
	tableRef := fmt.Sprintf("%s AS %s", p.qb.tableExpression(destSch), destAlias)
	p.joins = append(p.joins, graphJoinSpec{tableWithAlias: tableRef, onExpr: onExpr})
	p.pathKeyAli[cacheKey] = destAlias
	p.pathKeySch[cacheKey] = destSch
	return destAlias, destSch, nil
}

func (p *joinPlanner) appendManyToManyJoin(
	parentAlias string, parentSch *dmodel.ModelSchema, rel dmodel.ModelRelation, cacheKey string,
) (string, *dmodel.ModelSchema, error) {
	if rel.M2mDestFieldPrefix == "" || rel.M2mThroughSchemaName == "" || rel.M2mSrcFieldPrefix == "" {
		return "", nil, errors.Errorf(
			"joinPlanner: many-to-many relation '%s' is not finalized", rel.Edge)
	}
	throughSch := p.registry.Get(rel.M2mThroughSchemaName)
	if throughSch == nil {
		return "", nil, errors.Errorf("joinPlanner: through schema '%s' not found", rel.M2mThroughSchemaName)
	}
	destSch := p.registry.Get(rel.DestSchemaName)
	if destSch == nil {
		return "", nil, errors.Errorf("joinPlanner: schema '%s' not found", rel.DestSchemaName)
	}
	throughAlias := p.allocJoinedAlias()
	srcPks := parentSch.PrimaryKeys()
	if len(srcPks) == 0 {
		return "", nil, errors.Errorf("joinPlanner: schema '%s' has no primary key", parentSch.Name())
	}
	throughOn := make([]string, 0, len(srcPks))
	for _, pk := range srcPks {
		tc := dmodel.PrefixedThroughColumn(rel.M2mSrcFieldPrefix, pk)
		throughOn = append(throughOn, fmt.Sprintf("%s.%s = %s.%s",
			parentAlias, pgQuote(pk), throughAlias, pgQuote(tc)))
	}
	srcTk := parentSch.TenantKey()
	if srcTk != "" {
		throughOn = append(throughOn, fmt.Sprintf("%s.%s = %s.%s",
			parentAlias, pgQuote(srcTk), throughAlias, pgQuote(srcTk)))
	}
	onThrough := strings.Join(throughOn, " AND ")
	throughRef := fmt.Sprintf("%s AS %s", p.qb.tableExpression(throughSch), throughAlias)
	p.joins = append(p.joins, graphJoinSpec{tableWithAlias: throughRef, onExpr: onThrough})
	destAlias := p.allocJoinedAlias()
	peerPks := destSch.PrimaryKeys()
	if len(peerPks) == 0 {
		return "", nil, errors.Errorf("joinPlanner: peer schema '%s' has no primary key", destSch.Name())
	}
	destOn := make([]string, 0, len(peerPks))
	for _, pk := range peerPks {
		tc := dmodel.PrefixedThroughColumn(rel.M2mDestFieldPrefix, pk)
		destOn = append(destOn, fmt.Sprintf("%s.%s = %s.%s",
			throughAlias, pgQuote(tc), destAlias, pgQuote(pk)))
	}
	destTk := destSch.TenantKey()
	if srcTk != "" && destTk != "" {
		destOn = append(destOn, fmt.Sprintf("%s.%s = %s.%s",
			throughAlias, pgQuote(srcTk), destAlias, pgQuote(destTk)))
	}
	onDest := strings.Join(destOn, " AND ")
	destRef := fmt.Sprintf("%s AS %s", p.qb.tableExpression(destSch), destAlias)
	p.joins = append(p.joins, graphJoinSpec{tableWithAlias: destRef, onExpr: onDest})
	p.pathKeyAli[cacheKey] = destAlias
	p.pathKeySch[cacheKey] = destSch
	return destAlias, destSch, nil
}

func (p *joinPlanner) ensureFullPath(field string, maxDots int) error {
	segments, err := parseDottedPath(field, maxDots)
	if err != nil {
		return err
	}
	if len(segments) == 1 {
		return p.validateRootColumn(segments[0])
	}
	edges := segments[:len(segments)-1]
	leaf := segments[len(segments)-1]
	_, destSch, err := p.ensureJoinedEdges(edges)
	if err != nil {
		return err
	}
	return p.validateLeafColumn(destSch, leaf, field)
}

// validateRootColumn answers "is this a legal path on the root schema?", which is a different
// question from resolveFieldSqlRef's "give me SQL for it". A virtual scalar is legal to select —
// it is dropped from the projection later and filled by a service — so it is accepted here while
// resolveFieldSqlRef still rejects it.
func (p *joinPlanner) validateRootColumn(col string) error {
	field, ok := p.root.Field(col)
	if !ok || field.IsEdgeModel() {
		return errors.Wrap(&errClientUnknownField{Field: col}, "validateRootColumn")
	}
	return nil
}

func (p *joinPlanner) validateLeafColumn(schema *dmodel.ModelSchema, leaf, fullPath string) error {
	field, ok := schema.Column(leaf)
	if !ok || field.IsEdgeModel() {
		return errors.Wrap(&errClientUnknownField{Field: fullPath}, "validateLeafColumn")
	}
	return nil
}

func (p *joinPlanner) resolveFieldSqlRef(field string, maxDots int) (*dmodel.ModelField, string, error) {
	segments, err := parseDottedPath(field, maxDots)
	if err != nil {
		return nil, "", err
	}
	if len(segments) == 1 {
		col := segments[0]
		fieldObj, ok := p.root.Column(col)
		if !ok || fieldObj.IsEdgeModel() {
			return nil, "", errors.Wrap(&errClientUnknownField{Field: field}, "resolveFieldSqlRef")
		}
		if p.usesJoins() {
			p.ensureRootAliased()
			return fieldObj, fmt.Sprintf("%s.%s", p.rootAlias, pgQuote(col)), nil
		}
		return fieldObj, pgQuote(col), nil
	}
	edges := segments[:len(segments)-1]
	leaf := segments[len(segments)-1]
	alias, destSch, err := p.ensureJoinedEdges(edges)
	if err != nil {
		return nil, "", err
	}
	fieldObj, ok := destSch.Column(leaf)
	if !ok || fieldObj.IsEdgeModel() {
		return nil, "", errors.Wrap(&errClientUnknownField{Field: field}, "resolveFieldSqlRef")
	}
	return fieldObj, fmt.Sprintf("%s.%s", alias, pgQuote(leaf)), nil
}

// isVirtualScalarPath reports whether a single-segment path names a virtual scalar on the root
// schema. Such a field has no column to project: it is filled by a service after the read, so it
// is dropped from the SELECT while still being a legal thing to ask for.
//
// Scalars only, deliberately. This resolves through Field, so an edge-model field is reachable
// here, and an edge is virtual too — but dropping one from the projection would silently discard
// a caller's request instead of letting resolveFieldSqlRef report it.
func (p *joinPlanner) isVirtualScalarPath(path string) bool {
	if strings.Contains(path, ".") {
		return false
	}
	field, ok := p.root.Field(path)
	return ok && field.IsVirtual() && !field.IsEdgeModel()
}

// primaryKeySelectRefs returns SQL refs for the root schema's primary keys, used when every
// requested column turned out to be virtual and the projection would otherwise be empty.
func (p *joinPlanner) primaryKeySelectRefs() []string {
	keys := p.root.PrimaryKeys()
	refs := make([]string, 0, len(keys))
	for _, key := range keys {
		if _, ref, err := p.resolveFieldSqlRef(key, 0); err == nil {
			refs = append(refs, ref)
		}
	}
	return refs
}

func (p *joinPlanner) selectExprForColumn(requested string) (string, error) {
	segments, err := parseDottedPath(requested, MaxSelectGraphColumnDots)
	if err != nil {
		return "", err
	}
	_, ref, err := p.resolveFieldSqlRef(requested, MaxSelectGraphColumnDots)
	if err != nil {
		return "", err
	}
	if len(segments) == 1 {
		return ref, nil
	}
	aliasLabel := strings.Join(segments, ".")
	return fmt.Sprintf("%s AS %s", ref, pgQuote(aliasLabel)), nil
}

func isLinkedEdgeOnlyCondition(cond dmodel.Condition) bool {
	if len(cond) < 2 {
		return false
	}
	field := strings.TrimSpace(cond.Field())
	if field == "" || strings.Contains(field, ".") {
		return false
	}
	op := cond.Operator()
	return op == dmodel.Linked || op == dmodel.NotLinked
}

func graphRequiresRootAliasForLinkedNotLinked(graph *dmodel.SearchGraph) bool {
	if graph == nil {
		return false
	}
	if isLinkedEdgeOnlyCondition(graph.GetCondition()) {
		return true
	}
	for i := range graph.GetAnd() {
		if searchNodeRequiresLinkedNotLinkedRootAlias(&graph.GetAnd()[i]) {
			return true
		}
	}
	for i := range graph.GetOr() {
		if searchNodeRequiresLinkedNotLinkedRootAlias(&graph.GetOr()[i]) {
			return true
		}
	}
	return false
}

func searchNodeRequiresLinkedNotLinkedRootAlias(node *dmodel.SearchNode) bool {
	if node == nil {
		return false
	}
	if isLinkedEdgeOnlyCondition(node.GetCondition()) {
		return true
	}
	for i := range node.GetAnd() {
		if searchNodeRequiresLinkedNotLinkedRootAlias(&node.GetAnd()[i]) {
			return true
		}
	}
	for i := range node.GetOr() {
		if searchNodeRequiresLinkedNotLinkedRootAlias(&node.GetOr()[i]) {
			return true
		}
	}
	return false
}

func collectGraphFieldPaths(graph *dmodel.SearchGraph, into map[string]struct{}) error {
	if graph == nil {
		return nil
	}
	if graph.GetCondition().Field() != "" && !isLinkedEdgeOnlyCondition(graph.GetCondition()) {
		if err := noteGraphFieldName(graph.GetCondition().Field(), MaxSearchGraphConditionDots, into); err != nil {
			return err
		}
	}
	if err := walkSearchNodesForFields(graph.GetAnd(), into); err != nil {
		return err
	}
	if err := walkSearchNodesForFields(graph.GetOr(), into); err != nil {
		return err
	}
	for _, item := range graph.GetOrder() {
		if len(item) == 0 || item[0] == "" {
			continue
		}
		if err := noteGraphFieldName(item[0], MaxOrderGraphFieldDots, into); err != nil {
			return err
		}
	}
	return nil
}

func walkSearchNodesForFields(nodes []dmodel.SearchNode, into map[string]struct{}) error {
	for i := range nodes {
		node := &nodes[i]
		if node.GetCondition().Field() != "" && !isLinkedEdgeOnlyCondition(node.GetCondition()) {
			if err := noteGraphFieldName(node.GetCondition().Field(), MaxSearchGraphConditionDots, into); err != nil {
				return err
			}
		}
		if err := walkSearchNodesForFields(node.GetAnd(), into); err != nil {
			return err
		}
		if err := walkSearchNodesForFields(node.GetOr(), into); err != nil {
			return err
		}
	}
	return nil
}

func noteGraphFieldName(name string, maxDots int, into map[string]struct{}) error {
	if name == "" {
		return nil
	}
	if strings.Count(name, ".") > maxDots {
		return wrapClientSqlErrors(clientErrorsGraphFieldTooDeep(name, maxDots))
	}
	into[name] = struct{}{}
	return nil
}

func collectSelectAndGraphPaths(graph *dmodel.SearchGraph, opts SqlSelectGraphOpts) (map[string]struct{}, error) {
	paths := make(map[string]struct{})
	for _, col := range opts.Columns {
		path := col.joinPlanningPath()
		if path == "" {
			continue
		}
		if err := noteGraphFieldName(path, MaxSelectGraphColumnDots, paths); err != nil {
			return nil, err
		}
	}
	if err := collectGraphFieldPaths(graph, paths); err != nil {
		return nil, err
	}
	return paths, nil
}

func graphPathsNeedRegistry(paths map[string]struct{}) bool {
	for path := range paths {
		if strings.Contains(path, ".") {
			return true
		}
	}
	return false
}

func (this *PgQueryBuilder) planGraphJoins(
	root *dmodel.ModelSchema, registry *dmodel.SchemaRegistry, graph *dmodel.SearchGraph, opts SqlSelectGraphOpts,
) (*joinPlanner, error) {
	paths, err := collectSelectAndGraphPaths(graph, opts)
	if err != nil {
		return nil, err
	}
	if graphPathsNeedRegistry(paths) && registry == nil {
		return nil, wrapClientSqlErrors(clientErrorsRegistryRequiredForGraph())
	}
	planner := newJoinPlanner(this, registry, root)
	maxDotsFor := func(path string) int {
		if pathInSelectColumns(path, opts.Columns) {
			return MaxSelectGraphColumnDots
		}
		if pathInOrder(path, graph) {
			return MaxOrderGraphFieldDots
		}
		return MaxSearchGraphConditionDots
	}
	for path := range paths {
		if err := planner.ensureFullPath(path, maxDotsFor(path)); err != nil {
			return nil, err
		}
	}
	if graph != nil && graphRequiresRootAliasForLinkedNotLinked(graph) {
		planner.ensureRootAliased()
	}
	return planner, nil
}

func pathInOrder(path string, graph *dmodel.SearchGraph) bool {
	if graph == nil {
		return false
	}
	for _, item := range graph.GetOrder() {
		if len(item) > 0 && item[0] == path {
			return true
		}
	}
	return false
}

func pathInSelectColumns(path string, columns []SelectColumn) bool {
	for _, c := range columns {
		if c.joinPlanningPath() == path {
			return true
		}
		if strings.TrimSpace(c.rawString()) == path {
			return true
		}
	}
	return false
}

type graphSelectCtx struct {
	planner  *joinPlanner
	language *model.LanguageCode
}

func (this *PgQueryBuilder) prepareColNameForGraph(
	ctx *graphSelectCtx, schema *dmodel.ModelSchema, fieldName string,
) (*dmodel.ModelField, string, error) {
	if ctx == nil || ctx.planner == nil {
		return this.prepareColName(schema, fieldName)
	}
	return ctx.planner.resolveFieldSqlRef(fieldName, MaxSearchGraphConditionDots)
}
