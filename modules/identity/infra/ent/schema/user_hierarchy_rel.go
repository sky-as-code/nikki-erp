package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type UserHierarchyMixin struct {
	mixin.Schema
}

func (UserHierarchyMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").Immutable(),
		field.String("hierarchy_id").Immutable(),
	}
}

func (UserHierarchyMixin) Edges() []ent.Edge {
	return nil
}

type UserHierarchy struct {
	ent.Schema
}

func (UserHierarchy) Fields() []ent.Field {
	return nil
}

func (UserHierarchy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Immutable().
			Required().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
		edge.To("hierarchy", HierarchyLevel.Type).
			Field("hierarchy_id").
			Unique().
			Immutable().
			Required().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (UserHierarchy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		field.ID("user_id", "hierarchy_id"),
		entsql.Annotation{Table: "ident_user_hierarchy_rel"},
	}
}

func (UserHierarchy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UserHierarchyMixin{},
	}
}
