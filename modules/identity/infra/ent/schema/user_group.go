package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type UserGroupMixin struct {
	mixin.Schema
}

func (UserGroupMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").
			NotEmpty(),

		field.String("group_id").
			NotEmpty(),
	}
}

func (UserGroupMixin) Edges() []ent.Edge {
	return nil
}

type UserGroup struct {
	ent.Schema
}

func (UserGroup) Fields() []ent.Field {
	return nil
}

func (UserGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
		edge.To("group", Group.Type).
			Field("group_id").
			Unique().
			Required().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (UserGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		field.ID("user_id", "group_id"),
		entsql.Annotation{Table: "user_groups"},
	}
}

func (UserGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UserGroupMixin{},
	}
}
