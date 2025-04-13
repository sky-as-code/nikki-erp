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
			MaxLen(36).
			NotEmpty(),

		field.String("group_id").
			MaxLen(36).
			NotEmpty(),
	}
}

func (UserGroupMixin) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Required(),
		edge.To("group", Group.Type).
			Field("group_id").
			Unique().
			Required(),
	}
}

type UserGroup struct {
	ent.Schema
}

func (UserGroup) Fields() []ent.Field {
	return nil
}

func (UserGroup) Edges() []ent.Edge {
	return nil
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
