package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type RelationshipMixin struct {
	mixin.Schema
}

func (RelationshipMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id").
			Comment("ULID"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("deleted_at").
			Optional().
			Nillable(),

		field.String("deleted_by").
			Optional().
			Nillable(),

		field.String("etag"),

		field.String("note").
			Optional().
			Nillable(),

		field.String("target_party_id"),

		field.Enum("type").
			Values("employee", "spouse", "parent", "sibling", "emergency", "subsidiary"),

		field.Time("updated_at").
			Optional().
			Nillable(),
	}
}

// Relationship holds the schema definition for relationships between parties.
type Relationship struct {
	ent.Schema
}

func (Relationship) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "contacts_relationships"},
	}
}

func (Relationship) Fields() []ent.Field {
	return nil
}

func (Relationship) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("party", Party.Type).
			Field("target_party_id").
			Unique().
			Required().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (Relationship) Mixin() []ent.Mixin {
	return []ent.Mixin{
		RelationshipMixin{},
	}
}
