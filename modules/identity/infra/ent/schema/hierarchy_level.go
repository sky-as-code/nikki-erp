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

		field.String("org_id").
			Immutable(),

		field.String("etag"),

		field.String("name").
			Unique(),

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
		edge.From("parent", HierarchyLevel.Type).
			Ref("child").
			Unique(). // O2M relationship
			Field("parent_id"),

		edge.To("child", HierarchyLevel.Type),
	}
}

func (HierarchyLevel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		HierarchyLevelMixin{},
	}
}
