// ent/schema/hierarchylevel.go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type HierarchyLevelMixin struct {
	mixin.Schema
}

func (HierarchyLevelMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("Set value for this column when the process is running to delete all resources under this hierarchy level"),

		field.String("deleted_by").
			Optional().
			Nillable().
			Comment("Set value for this column when the process is running to delete all resources under this hierarchy level"),

		field.String("etag"),

		field.String("name").
			Unique(),

		field.String("org_id").
			Immutable(),

		field.String("parent_id").
			Optional().
			Nillable(),
	}
}

func (HierarchyLevelMixin) Edges() []ent.Edge {
	return nil
}

type HierarchyLevel struct {
	ent.Schema
}

func (HierarchyLevel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ident_hierarchy_levels"},
	}
}

func (HierarchyLevel) Fields() []ent.Field {
	return nil
}

func (HierarchyLevel) Edges() []ent.Edge {
	return []ent.Edge{
		// Self-referential parent level
		edge.From("children", HierarchyLevel.Type).
			Ref("parent"),

		edge.From("users", User.Type).
			Ref("hierarchy"),

		edge.To("deleter", User.Type).
			Field("deleted_by").
			Unique(),

		edge.To("parent", HierarchyLevel.Type).
			Field("parent_id").
			Unique(), // O2M relationship

		edge.To("org", Organization.Type).
			Field("org_id").
			Immutable().
			Required().
			Unique(), // O2M relationship
	}
}

func (HierarchyLevel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		HierarchyLevelMixin{},
	}
}
