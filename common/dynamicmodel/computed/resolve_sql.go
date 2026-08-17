package computed

import (
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// Resolution of the SQL-compiled kinds (aggregate / exists / lookup). Every identifier — the
// source edge, filter fields, order_by fields, the aggregated operand — resolves through the
// schema registry here, at finalize time. This is the primary injection barrier: the subquery
// emitter later renders only names that survived this pass.

func (this *resolver) buildSqlKindPlan(
	schema *dmodel.ModelSchema, field *dmodel.ModelField, plan *FieldPlan,
) error {
	err := this.buildSqlKindPlanInner(schema, field, plan)
	return errors.Wrapf(err, "computed field %s.%s", schema.Name(), field.Name())
}

func (this *resolver) buildSqlKindPlanInner(
	schema *dmodel.ModelSchema, field *dmodel.ModelField, plan *FieldPlan,
) error {
	switch plan.Def.Kind {
	case ComputeAggregate:
		return this.buildAggregatePlan(schema, plan)
	case ComputeExists:
		return this.buildExistsPlan(schema, plan)
	case ComputeLookup:
		return this.buildLookupPlan(schema, plan)
	}
	return errors.Errorf("unsupported SQL compute kind %q", plan.Def.Kind)
}

func (this *resolver) buildAggregatePlan(schema *dmodel.ModelSchema, plan *FieldPlan) error {
	node := plan.Def.Aggregate
	source, sourceSchema, err := this.resolveSqlSource(schema, node.Source, plan)
	if err != nil {
		return err
	}
	operandType, err := this.resolveAggregateOperand(sourceSchema, node, plan)
	if err != nil {
		return err
	}
	resultType, err := aggregateResultType(node.Function, operandType)
	if err != nil {
		return err
	}
	if err := this.validateSqlFilter(sourceSchema, node.Filter, node.Context, plan); err != nil {
		return err
	}
	if err := checkDefaultValue(node.Default, resultType); err != nil {
		return err
	}
	plan.Type = resultType
	plan.SqlSource = source
	return nil
}

func (this *resolver) buildExistsPlan(schema *dmodel.ModelSchema, plan *FieldPlan) error {
	node := plan.Def.Exists
	source, sourceSchema, err := this.resolveSqlSource(schema, node.Source, plan)
	if err != nil {
		return err
	}
	if err := this.validateSqlFilter(sourceSchema, node.Filter, node.Context, plan); err != nil {
		return err
	}
	plan.Type = TypeBoolean
	plan.SqlSource = source
	return nil
}

func (this *resolver) buildLookupPlan(schema *dmodel.ModelSchema, plan *FieldPlan) error {
	node := plan.Def.Lookup
	source, sourceSchema, err := this.resolveSqlSource(schema, node.Source, plan)
	if err != nil {
		return err
	}
	copied, err := this.resolveSourceScalar(sourceSchema, node.Field, plan)
	if err != nil {
		return err
	}
	for _, order := range node.OrderBy {
		if _, err := this.resolveSourceScalar(sourceSchema, order.Field, plan); err != nil {
			return errors.Wrap(err, "order_by")
		}
	}
	if err := this.validateSqlFilter(sourceSchema, node.Filter, node.Context, plan); err != nil {
		return err
	}
	resultType := Type(copied.DataType().String())
	if err := checkDefaultValue(node.Default, resultType); err != nil {
		return err
	}
	plan.Type = resultType
	plan.SqlSource = source
	return nil
}

// resolveSqlSource resolves the collection edge and derives the correlation columns the subquery
// emitter needs. Only one direct edge is supported this phase (MaxAggregateRelationDepth is
// effectively 1): a dotted source is rejected before any lookup happens.
func (this *resolver) resolveSqlSource(
	schema *dmodel.ModelSchema, edgeName string, plan *FieldPlan,
) (*SqlSourcePlan, *dmodel.ModelSchema, error) {
	if strings.Contains(edgeName, ".") {
		return nil, nil, errors.Errorf(
			"source %q traverses more than one edge; only a direct collection edge is supported in this phase",
			edgeName)
	}
	relation, err := findCollectionRelation(schema, edgeName)
	if err != nil {
		return nil, nil, err
	}
	sourceSchema := this.reg.Get(relation.DestSchemaName)
	if sourceSchema == nil {
		return nil, nil, errors.Errorf("Unknown schema %q", relation.DestSchemaName)
	}
	source, err := buildSourceCorrelation(schema, relation, sourceSchema)
	if err != nil {
		return nil, nil, err
	}
	plan.Dependencies = append(plan.Dependencies,
		FieldRef{Schema: schema.Name(), Field: edgeName})
	return source, sourceSchema, nil
}

func findCollectionRelation(schema *dmodel.ModelSchema, edgeName string) (*dmodel.ModelRelation, error) {
	for _, relation := range schema.Relations() {
		if relation.Edge != edgeName {
			continue
		}
		collection := relation.RelationType == dmodel.RelationTypeOneToMany ||
			relation.RelationType == dmodel.RelationTypeManyToMany
		if !collection {
			return nil, errors.Errorf(
				"edge %q is a %s relation; aggregate/exists/lookup need a collection edge (one:many or many:many)",
				edgeName, relation.RelationType)
		}
		result := relation
		return &result, nil
	}
	return nil, errors.Errorf("Unknown relation %q", schema.Name()+"."+edgeName)
}

func buildSourceCorrelation(
	root *dmodel.ModelSchema, relation *dmodel.ModelRelation, sourceSchema *dmodel.ModelSchema,
) (*SqlSourcePlan, error) {
	if relation.RelationType == dmodel.RelationTypeManyToMany {
		return buildManyToManyCorrelation(root, relation, sourceSchema)
	}
	pairs := relation.EffectiveForeignKeys()
	if len(pairs) != 1 {
		return nil, errors.Errorf(
			"edge %q uses a composite foreign key; SQL computed fields support single-column keys only",
			relation.Edge)
	}
	rootRef := pairs[0].ReferencedColumn
	if rootRef == "" {
		rootPk, err := singlePrimaryKey(root, relation.Edge)
		if err != nil {
			return nil, err
		}
		rootRef = rootPk
	}
	return &SqlSourcePlan{
		Edge:             relation.Edge,
		SourceSchemaName: sourceSchema.Name(),
		SourceFkColumn:   pairs[0].FkColumn,
		RootRefColumn:    rootRef,
	}, nil
}

func buildManyToManyCorrelation(
	root *dmodel.ModelSchema, relation *dmodel.ModelRelation, sourceSchema *dmodel.ModelSchema,
) (*SqlSourcePlan, error) {
	if relation.M2mThroughSchemaName == "" || relation.M2mSrcFieldPrefix == "" || relation.M2mDestFieldPrefix == "" {
		return nil, errors.Errorf("edge %q: many-to-many junction metadata is not finalized", relation.Edge)
	}
	rootPk, err := singlePrimaryKey(root, relation.Edge)
	if err != nil {
		return nil, err
	}
	sourcePk, err := singlePrimaryKey(sourceSchema, relation.Edge)
	if err != nil {
		return nil, err
	}
	return &SqlSourcePlan{
		Edge:              relation.Edge,
		SourceSchemaName:  sourceSchema.Name(),
		Many:              true,
		ThroughSchemaName: relation.M2mThroughSchemaName,
		ThroughSrcColumn:  dmodel.PrefixedThroughColumn(relation.M2mSrcFieldPrefix, rootPk),
		ThroughDestColumn: dmodel.PrefixedThroughColumn(relation.M2mDestFieldPrefix, sourcePk),
		SourcePkColumn:    sourcePk,
		RootPkColumn:      rootPk,
	}, nil
}

func singlePrimaryKey(schema *dmodel.ModelSchema, edgeName string) (string, error) {
	pks := schema.PrimaryKeys()
	if len(pks) != 1 {
		return "", errors.Errorf(
			"edge %q: schema %q has a composite primary key; SQL computed fields support single-column keys only",
			edgeName, schema.Name())
	}
	return pks[0], nil
}

// resolveSourceScalar resolves a field of the SOURCE schema for use inside a subquery: it must
// be a physical scalar. A computed field of the source cannot appear — an aggregate of it would
// be aggregate-of-derived, which cannot compile to SQL in this phase (spec §37).
func (this *resolver) resolveSourceScalar(
	sourceSchema *dmodel.ModelSchema, name string, plan *FieldPlan,
) (*dmodel.ModelField, error) {
	if strings.Contains(name, ".") {
		return nil, errors.Errorf(
			"field %q is a dotted path; subquery fields must live on the source schema itself in this phase", name)
	}
	referenced, ok := sourceSchema.Field(name)
	if !ok {
		return nil, errors.Errorf("Unknown field %q", sourceSchema.Name()+"."+name)
	}
	if referenced.IsEdgeModel() {
		return nil, errors.Errorf("field %q is an edge, not a scalar", name)
	}
	if referenced.IsComputed() {
		return nil, errors.Errorf(
			"cannot reference computed field %q inside an aggregate/exists/lookup; it has no column to query", name)
	}
	if referenced.IsVirtual() {
		return nil, errors.Errorf(
			"field %q is filled by service code after the read; a subquery cannot reference it", name)
	}
	plan.Dependencies = append(plan.Dependencies,
		FieldRef{Schema: sourceSchema.Name(), Field: name})
	return referenced, nil
}

// resolveAggregateOperand types the aggregated value: the named source field, or the restricted
// inner expression (validated by validateSqlInnerExpr) inferred over source-schema fields.
func (this *resolver) resolveAggregateOperand(
	sourceSchema *dmodel.ModelSchema, node *AggregateExpr, plan *FieldPlan,
) (Type, error) {
	if node.Function == AggCount {
		return TypeInt64, nil
	}
	if node.Field != "" {
		operand, err := this.resolveSourceScalar(sourceSchema, node.Field, plan)
		if err != nil {
			return TypeUnknown, err
		}
		return Type(operand.DataType().String()), nil
	}
	if err := validateSqlInnerExpr(node.Expr); err != nil {
		return TypeUnknown, err
	}
	return InferType(node.Expr, func(name string) (Type, error) {
		operand, err := this.resolveSourceScalar(sourceSchema, name, plan)
		if err != nil {
			return TypeUnknown, err
		}
		return Type(operand.DataType().String()), nil
	})
}

// validateSqlInnerExpr restricts the aggregate inner expression to the SQL-compilable subset:
// source field refs, scalar literals, arithmetic (including negate), coalesce and nullif. The Go
// registry functions never grow SQL twins — that is what keeps the two execution worlds honest.
func validateSqlInnerExpr(expr Expr) error {
	switch node := expr.(type) {
	case FieldExpr, LiteralExpr:
		return nil
	case BinaryExpr:
		if !node.Op.IsArithmetic() {
			return errors.Errorf(
				"operator %q cannot be used inside an aggregate expression; only arithmetic is SQL-compilable", node.Op)
		}
		if err := validateSqlInnerExpr(node.Left); err != nil {
			return err
		}
		return validateSqlInnerExpr(node.Right)
	case UnaryExpr:
		if node.Op != OpNegate {
			return errors.Errorf("operator %q cannot be used inside an aggregate expression", node.Op)
		}
		return validateSqlInnerExpr(node.Operand)
	case FunctionExpr:
		return validateSqlInnerFunction(node)
	}
	return errors.Errorf("%T cannot be used inside an aggregate expression", expr)
}

func validateSqlInnerFunction(node FunctionExpr) error {
	if node.Name != "coalesce" && node.Name != "nullif" {
		return errors.Errorf(
			"function %q cannot be used inside an aggregate expression; only coalesce and nullif compile to SQL",
			node.Name)
	}
	for _, arg := range node.Args {
		if err := validateSqlInnerExpr(arg); err != nil {
			return err
		}
	}
	return nil
}

func aggregateResultType(function AggregateFunction, operandType Type) (Type, error) {
	switch function {
	case AggCount, AggCountDistinct:
		return TypeInt64, nil
	case AggAvg:
		if !operandType.IsNumeric() {
			return TypeUnknown, errors.Errorf("aggregate \"avg\" needs a numeric operand, not %s", operandType)
		}
		return TypeDecimal, nil
	case AggSum:
		return sumResultType(operandType)
	case AggMin, AggMax:
		if !operandType.IsNumeric() && !operandType.IsTexty() && !operandType.IsTemporal() {
			return TypeUnknown, errors.Errorf(
				"aggregate %q needs an orderable operand (numeric, text or date/time), not %s",
				string(function), operandType)
		}
		return operandType, nil
	}
	return TypeUnknown, errors.Errorf("aggregate function %q is not supported", string(function))
}

// sumResultType mirrors PostgreSQL: SUM(int32) -> int64, SUM(int64)/SUM(decimal) -> decimal.
func sumResultType(operandType Type) (Type, error) {
	switch operandType {
	case TypeInt32:
		return TypeInt64, nil
	case TypeInt64, TypeDecimal:
		return TypeDecimal, nil
	}
	return TypeUnknown, errors.Errorf("aggregate \"sum\" needs a numeric operand, not %s", operandType)
}

// checkDefaultValue verifies the declared default is a scalar whose type unifies with the
// subquery's result type, so COALESCE never mixes types at query time.
func checkDefaultValue(value any, resultType Type) error {
	if value == nil {
		return nil
	}
	defaultType, err := literalType(value)
	if err != nil {
		return errors.Wrap(err, "default")
	}
	if _, ok := unify(resultType, defaultType); !ok {
		return errors.Errorf("default value %v (%s) does not match the computed type %s",
			value, defaultType, resultType)
	}
	return nil
}
