package computed

import (
	"sync"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// FieldPlan is the finalized description of one computed field: what it reads, how it is typed,
// and — for related kinds — which edge and leaf supply the value. Built once at schema finalize
// time; the read-time eval engine only consumes it.
type FieldPlan struct {
	FieldName string
	Def       *Definition
	Type      Type

	// PhysicalOperands are the same-schema physical columns evaluation reads, transitively
	// through computed dependencies. The eval planner appends them to the SQL projection.
	PhysicalOperands []string
	// ComputedDeps are same-schema computed fields this one reads; they evaluate first.
	ComputedDeps []string

	// Related description, set only when Def.Kind is ComputeRelated.
	RelatedEdge       string
	RelatedLeaf       string
	RelatedSchemaName string
	// RelatedFkColumn / RelatedRefColumn: the local FK column and the referenced column on the
	// related schema. Single-column only in this phase.
	RelatedFkColumn  string
	RelatedRefColumn string

	// SqlSource describes the collection edge an SQL-compiled kind (aggregate/exists/lookup)
	// correlates through. Set only for those kinds; the subquery emitter consumes it.
	SqlSource *SqlSourcePlan

	// FunctionName is the engine-registered function supplying the value, set only when Def.Kind
	// is ComputeFunction. DependsOn is its optional same-schema trigger field.
	FunctionName string
	DependsOn    string

	// Dependencies is the full metadata for impact analysis: every schema element this field's
	// definition points at (operand fields, the related edge, FK and leaf columns).
	Dependencies []FieldRef
}

// SqlSourcePlan is the resolved correlation between a schema and the collection edge an
// SQL-compiled computed field aggregates over. Single-column keys only in this phase.
type SqlSourcePlan struct {
	Edge             string
	SourceSchemaName string
	// Many is true for a many:many edge, correlated through the junction schema below;
	// false for one:many, correlated by SourceFkColumn = root RootRefColumn.
	Many bool

	// one:many correlation.
	SourceFkColumn string
	RootRefColumn  string

	// many:many correlation: through.ThroughSrcColumn = root.RootPkColumn and
	// through.ThroughDestColumn = source.SourcePkColumn.
	ThroughSchemaName string
	ThroughSrcColumn  string
	ThroughDestColumn string
	SourcePkColumn    string
	RootPkColumn      string
}

// SchemaPlan holds every computed field of one schema plus a dependency-safe evaluation order.
type SchemaPlan struct {
	SchemaName string
	Fields     map[string]*FieldPlan
	// EvalOrder lists the computed fields so that every field appears after the computed fields
	// it depends on.
	EvalOrder []string
}

var (
	plansMu sync.RWMutex
	plans   = map[string]*SchemaPlan{}
)

// PlanFor returns the finalized plan of a schema, or nil when it has no computed fields (or the
// registry has not been finalized yet).
func PlanFor(schemaName string) *SchemaPlan {
	plansMu.RLock()
	defer plansMu.RUnlock()
	return plans[schemaName]
}

func init() {
	// SchemaRegistry.FinalizeRelations runs this after its own relation passes, outside its
	// lock, so the resolver may use the registry's normal accessors. See model/computed_hooks.go.
	dmodel.RegisterComputedFinalizer(func(reg *dmodel.SchemaRegistry) error {
		return FinalizeSchemas(reg)
	})
}

// FinalizeSchemas validates every computed field in the registry and rebuilds the plan cache.
// It is idempotent: FinalizeRelations may run more than once as modules boot.
func FinalizeSchemas(reg *dmodel.SchemaRegistry) error {
	resolver := &resolver{
		reg:    reg,
		limits: ActiveLimits(),
		built:  map[string]*SchemaPlan{},
		done:   map[FieldRef]*FieldPlan{},
	}
	if err := resolver.resolveAll(); err != nil {
		return err
	}

	plansMu.Lock()
	defer plansMu.Unlock()
	plans = resolver.built
	return nil
}

type resolver struct {
	reg    *dmodel.SchemaRegistry
	limits Limits
	built  map[string]*SchemaPlan
	done   map[FieldRef]*FieldPlan
	// stack is the in-flight resolution chain, used for cycle detection and depth limiting.
	stack []FieldRef
}

func (this *resolver) resolveAll() error {
	return this.reg.ForEach(func(schemaName string, schema *dmodel.ModelSchema) error {
		return this.resolveSchema(schema)
	})
}

func (this *resolver) resolveSchema(schema *dmodel.ModelSchema) error {
	for _, field := range schema.ReadableFields() {
		if !field.IsComputed() {
			continue
		}
		if _, err := this.resolveField(schema, field); err != nil {
			return err
		}
	}
	return nil
}

// resolveField validates one computed field and memoizes its plan. Reentry through the stack
// means the definitions form a cycle.
func (this *resolver) resolveField(schema *dmodel.ModelSchema, field *dmodel.ModelField) (*FieldPlan, error) {
	ref := FieldRef{Schema: schema.Name(), Field: field.Name()}
	if plan, ok := this.done[ref]; ok {
		return plan, nil
	}
	for _, inFlight := range this.stack {
		if inFlight == ref {
			return nil, newCycleError(this.stack, ref)
		}
	}
	if len(this.stack) >= this.limits.MaxComputedDependencyDepth {
		return nil, errors.Errorf(
			"computed field %s exceeds the maximum dependency depth of %d",
			ref, this.limits.MaxComputedDependencyDepth)
	}

	this.stack = append(this.stack, ref)
	plan, err := this.buildFieldPlan(schema, field)
	this.stack = this.stack[:len(this.stack)-1]
	if err != nil {
		return nil, err
	}

	this.done[ref] = plan
	this.appendToSchemaPlan(schema.Name(), plan)
	return plan, nil
}

// appendToSchemaPlan records a finished plan. Dependencies resolve before their dependents
// finish, so append order IS a safe evaluation order.
func (this *resolver) appendToSchemaPlan(schemaName string, plan *FieldPlan) {
	schemaPlan, ok := this.built[schemaName]
	if !ok {
		schemaPlan = &SchemaPlan{SchemaName: schemaName, Fields: map[string]*FieldPlan{}}
		this.built[schemaName] = schemaPlan
	}
	schemaPlan.Fields[plan.FieldName] = plan
	schemaPlan.EvalOrder = append(schemaPlan.EvalOrder, plan.FieldName)
}

func (this *resolver) buildFieldPlan(schema *dmodel.ModelSchema, field *dmodel.ModelField) (*FieldPlan, error) {
	def, err := DefOf(field)
	if err != nil {
		return nil, err
	}
	plan := &FieldPlan{FieldName: field.Name(), Def: def}
	switch def.Kind {
	case ComputeExpression:
		err = this.buildExpressionPlan(schema, field, plan)
	case ComputeRelated:
		err = this.buildRelatedPlan(schema, field, plan)
	case ComputeAggregate, ComputeExists, ComputeLookup:
		err = this.buildSqlKindPlan(schema, field, plan)
	case ComputeFunction:
		err = this.buildFunctionPlan(schema, field, plan)
	default:
		err = errors.Errorf("computed field %s.%s has unsupported kind %q", schema.Name(), field.Name(), def.Kind)
	}
	if err != nil {
		return nil, err
	}
	if def.Kind == ComputeFunction {
		// Nothing to reconcile: the value comes from Go, whose return type the resolver never
		// sees, and which may be an array — a shape the scalar inference types cannot represent.
		return plan, nil
	}
	return plan, this.checkDeclaredType(schema, field, plan)
}

// checkDeclaredType reconciles the field's declared data_type with the inferred type (spec §22).
func (this *resolver) checkDeclaredType(
	schema *dmodel.ModelSchema, field *dmodel.ModelField, plan *FieldPlan,
) error {
	declared := Type(field.DataType().String())
	if !declaredCompatible(declared, plan.Type) {
		return errors.Errorf(
			"computed field %s.%s declares type %s but its definition produces %s",
			schema.Name(), field.Name(), declared, orNullName(plan.Type))
	}
	return nil
}

// declaredCompatible is the short, explicit compatibility table between a declared data_type and
// an inferred expression type. Exact match is the rule; each extra pair is deliberate.
func declaredCompatible(declared Type, inferred Type) bool {
	switch {
	case declared == inferred:
		return true
	case declared == TypeEnumString && inferred == TypeString:
		// CASE / concat produce plain strings; the declared enum narrows the value set.
		return true
	case declared == TypeString && inferred.IsTexty():
		return true
	case declared == TypeInt64 && inferred == TypeInt32:
		return true
	case declared == TypeDecimal && inferred.IsNumeric():
		return true
	case declared == TypeDateTime && inferred == TypeDate:
		return true
	}
	return false
}
