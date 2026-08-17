package computed

import (
	"strings"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// Related-kind resolution. The current phase evaluates one forward to-one hop — "edge.leaf",
// with the FK column on this schema — because the batched read-time fill (one extra query per
// page per edge) only needs that shape. Deeper chains and inverse edges are rejected with a
// clear message rather than silently mis-evaluated.

func (this *resolver) buildRelatedPlan(
	schema *dmodel.ModelSchema, field *dmodel.ModelField, plan *FieldPlan,
) error {
	edgeName, leafName, err := this.splitRelatedPath(schema, field, plan.Def.Related)
	if err != nil {
		return err
	}
	relation, err := findToOneRelation(schema, edgeName)
	if err != nil {
		return errors.Wrapf(err, "computed field %s.%s", schema.Name(), field.Name())
	}
	fkColumn, refColumn, err := singleForeignKeyPair(relation)
	if err != nil {
		return errors.Wrapf(err, "computed field %s.%s", schema.Name(), field.Name())
	}
	leaf, destSchema, err := this.resolveRelatedLeaf(relation.DestSchemaName, leafName)
	if err != nil {
		return errors.Wrapf(err, "computed field %s.%s", schema.Name(), field.Name())
	}

	plan.Type = Type(leaf.DataType().String())
	plan.RelatedEdge = edgeName
	plan.RelatedLeaf = leafName
	plan.RelatedSchemaName = destSchema.Name()
	plan.RelatedFkColumn = fkColumn
	plan.RelatedRefColumn = refColumn
	plan.PhysicalOperands = []string{fkColumn}
	plan.Dependencies = []FieldRef{
		{Schema: schema.Name(), Field: edgeName},
		{Schema: schema.Name(), Field: fkColumn},
		{Schema: destSchema.Name(), Field: refColumn},
		{Schema: destSchema.Name(), Field: leafName},
	}
	return nil
}

func (this *resolver) splitRelatedPath(
	schema *dmodel.ModelSchema, field *dmodel.ModelField, path string,
) (string, string, error) {
	segments := strings.Split(path, ".")
	for _, segment := range segments {
		if segment == "" {
			return "", "", errors.Errorf(
				"computed field %s.%s: related path %q has an empty segment",
				schema.Name(), field.Name(), path)
		}
	}
	edgeCount := len(segments) - 1
	if edgeCount < 1 {
		return "", "", errors.Errorf(
			"computed field %s.%s: related path %q must be \"edge.field\"",
			schema.Name(), field.Name(), path)
	}
	if edgeCount > 1 || edgeCount > this.limits.MaxRelatedPathDepth {
		return "", "", errors.Errorf(
			"computed field %s.%s: related path %q traverses %d edges; only one forward to-one "+
				"edge is supported in this phase", schema.Name(), field.Name(), path, edgeCount)
	}
	return segments[0], segments[1], nil
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

func (this *resolver) resolveRelatedLeaf(
	destSchemaName string, leafName string,
) (*dmodel.ModelField, *dmodel.ModelSchema, error) {
	destSchema := this.reg.Get(destSchemaName)
	if destSchema == nil {
		return nil, nil, errors.Errorf("Unknown schema %q", destSchemaName)
	}
	leaf, ok := destSchema.Field(leafName)
	if !ok {
		return nil, nil, errors.Errorf("Unknown field %q", destSchemaName+"."+leafName)
	}
	if leaf.IsEdgeModel() {
		return nil, nil, errors.Errorf("field %q is an edge, not a scalar leaf", leafName)
	}
	if leaf.IsComputed() {
		return nil, nil, errors.Errorf(
			"field %q on schema %q is itself derived; a related computed field must end at a "+
				"physical column in this phase", leafName, destSchemaName)
	}
	return leaf, destSchema, nil
}
