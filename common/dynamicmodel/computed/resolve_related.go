package computed

import (
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// Related-kind resolution. A related path is a chain of forward to-one edges ending at a
// physical column: "template.name", "template.uom.name". The FIRST hop is what the read-time
// fill batches (one query per page per edge, keyed by this schema's FK column); every further
// hop rides on that read as a nested field path ("uom.name"), which the source repository
// projects in the same statement. The leaf may itself be a related field of the last schema —
// it is flattened into the chain at finalize, so evaluation and SQL only ever see physical
// columns. Inverse edges and to-many edges are rejected with a clear message.

func (this *resolver) buildRelatedPlan(
	schema *dmodel.ModelSchema, field *dmodel.ModelField, plan *FieldPlan,
) error {
	segments, err := this.splitRelatedPath(schema, field, plan.Def.Related)
	if err != nil {
		return err
	}
	chain, err := this.flattenRelatedPath(schema, field, segments)
	if err != nil {
		return errors.Wrapf(err, "computed field %s.%s", schema.Name(), field.Name())
	}
	if hops := len(chain) - 1; hops > this.limits.MaxRelatedPathDepth {
		return errors.Errorf(
			"computed field %s.%s: related path %q resolves to %d edges (%s); the limit is %d",
			schema.Name(), field.Name(), plan.Def.Related, hops, strings.Join(chain, "."),
			this.limits.MaxRelatedPathDepth)
	}

	current := schema
	deps := []FieldRef{}
	for i, edge := range chain[:len(chain)-1] {
		relation, err := findToOneRelation(current, edge)
		if err != nil {
			return errors.Wrapf(err, "computed field %s.%s", schema.Name(), field.Name())
		}
		fkColumn, refColumn, err := singleForeignKeyPair(relation)
		if err != nil {
			return errors.Wrapf(err, "computed field %s.%s", schema.Name(), field.Name())
		}
		dest := this.reg.Get(relation.DestSchemaName)
		if dest == nil {
			return errors.Errorf("computed field %s.%s: Unknown schema %q", schema.Name(), field.Name(), relation.DestSchemaName)
		}
		deps = append(deps,
			FieldRef{Schema: current.Name(), Field: edge},
			FieldRef{Schema: current.Name(), Field: fkColumn},
			FieldRef{Schema: dest.Name(), Field: refColumn},
		)
		if i == 0 {
			plan.RelatedEdge = edge
			plan.RelatedSchemaName = dest.Name()
			plan.RelatedFkColumn = fkColumn
			plan.RelatedRefColumn = refColumn
			plan.PhysicalOperands = []string{fkColumn}
		}
		current = dest
	}
	leafName := chain[len(chain)-1]
	leaf, err := physicalLeaf(current, leafName)
	if err != nil {
		return errors.Wrapf(err, "computed field %s.%s", schema.Name(), field.Name())
	}
	plan.Type = Type(leaf.DataType().String())
	plan.RelatedLeaf = strings.Join(chain[1:], ".")
	plan.Dependencies = append(deps, FieldRef{Schema: current.Name(), Field: leafName})
	return nil
}

// splitRelatedPath validates the path's shape: non-empty segments, at least one edge.
func (this *resolver) splitRelatedPath(
	schema *dmodel.ModelSchema, field *dmodel.ModelField, path string,
) ([]string, error) {
	segments := strings.Split(path, ".")
	for _, segment := range segments {
		if segment == "" {
			return nil, errors.Errorf(
				"computed field %s.%s: related path %q has an empty segment",
				schema.Name(), field.Name(), path)
		}
	}
	if len(segments) < 2 {
		return nil, errors.Errorf(
			"computed field %s.%s: related path %q must be \"edge.field\"",
			schema.Name(), field.Name(), path)
	}
	return segments, nil
}

// flattenRelatedPath walks the edges to the schema owning the leaf and, when that leaf is itself
// a related field, splices in its (already flattened) chain. Resolving the leaf through the
// resolver keeps cycle detection and the dependency-depth limit in force.
func (this *resolver) flattenRelatedPath(
	schema *dmodel.ModelSchema, field *dmodel.ModelField, segments []string,
) ([]string, error) {
	current := schema
	for _, edge := range segments[:len(segments)-1] {
		relation, err := findToOneRelation(current, edge)
		if err != nil {
			return nil, err
		}
		dest := this.reg.Get(relation.DestSchemaName)
		if dest == nil {
			return nil, errors.Errorf("Unknown schema %q", relation.DestSchemaName)
		}
		current = dest
	}
	leafName := segments[len(segments)-1]
	leaf, ok := current.Field(leafName)
	if !ok {
		return nil, errors.Errorf("Unknown field %q", current.Name()+"."+leafName)
	}
	if leaf.IsEdgeModel() {
		return nil, errors.Errorf("field %q is an edge, not a scalar leaf", leafName)
	}
	if !leaf.IsComputed() {
		return segments, nil
	}
	def, err := DefOf(leaf)
	if err != nil {
		return nil, err
	}
	if def.Kind != ComputeRelated {
		return nil, errors.Errorf(
			"field %q on schema %q is a %s computed field; a related path may only end at a "+
				"physical column or at another related field", leafName, current.Name(), def.Kind)
	}
	leafPlan, err := this.resolveField(current, leaf)
	if err != nil {
		return nil, err
	}
	chain := append([]string{}, segments[:len(segments)-1]...)
	chain = append(chain, leafPlan.RelatedEdge)
	return append(chain, strings.Split(leafPlan.RelatedLeaf, ".")...), nil
}

func physicalLeaf(schema *dmodel.ModelSchema, leafName string) (*dmodel.ModelField, error) {
	leaf, ok := schema.Field(leafName)
	if !ok {
		return nil, errors.Errorf("Unknown field %q", schema.Name()+"."+leafName)
	}
	if leaf.IsEdgeModel() || leaf.IsComputed() {
		return nil, errors.Errorf("field %q on schema %q is not a physical column", leafName, schema.Name())
	}
	return leaf, nil
}

func findToOneRelation(schema *dmodel.ModelSchema, edgeName string) (*dmodel.ModelRelation, error) {
	for _, relation := range schema.Relations() {
		if relation.Edge != edgeName {
			continue
		}
		if relation.IsInverse {
			return nil, errors.Errorf(
				"edge %q is an inverse edge; the foreign key lives on the other schema, which this phase does not batch", edgeName)
		}
		toOne := relation.RelationType == dmodel.RelationTypeManyToOne ||
			relation.RelationType == dmodel.RelationTypeOneToOne
		if !toOne {
			return nil, errors.Errorf(
				"edge %q is a %s relation; a related computed field copies from a to-one edge",
				edgeName, relation.RelationType)
		}
		result := relation
		return &result, nil
	}
	return nil, errors.Errorf("Unknown relation %q", schema.Name()+"."+edgeName)
}

func singleForeignKeyPair(relation *dmodel.ModelRelation) (string, string, error) {
	pairs := relation.EffectiveForeignKeys()
	if len(pairs) != 1 {
		return "", "", errors.Errorf(
			"edge %q uses a composite foreign key; related computed fields support single-column keys only",
			relation.Edge)
	}
	return pairs[0].FkColumn, pairs[0].ReferencedColumn, nil
}
