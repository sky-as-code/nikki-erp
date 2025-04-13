package schema

import "entgo.io/ent"

// Org holds the schema definition for the Org entity.
type Org struct {
	ent.Schema
}

// Fields of the Org.
func (Org) Fields() []ent.Field {
	return nil
}

// Edges of the Org.
func (Org) Edges() []ent.Edge {
	return nil
}
