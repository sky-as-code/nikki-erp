package model

// pruneUnresolvableLooseRelationsUnlocked drops every loose edge (NoDbConstraint) whose destination
// schema is not registered.
//
// A loose edge exists for a reference against the migration order, e.g. a module created before
// IAM pointing at iam_user. Generating that module's migration registers only its dependency
// closure, which lacks the destination; the edge carries no column or constraint, so dropping it
// there changes nothing in the DDL. A full application registers every schema and keeps it.
func pruneUnresolvableLooseRelationsUnlocked(reg *SchemaRegistry) {
	for _, sch := range reg.schemas {
		sch.toRelations = keepResolvableRelations(reg, sch.toRelations)
	}
}

func keepResolvableRelations(reg *SchemaRegistry, rels []ModelRelation) []ModelRelation {
	kept := make([]ModelRelation, 0, len(rels))
	for _, rel := range rels {
		if rel.NoDbConstraint && reg.schemas[rel.DestSchemaName] == nil {
			continue
		}
		kept = append(kept, rel)
	}
	return kept
}
