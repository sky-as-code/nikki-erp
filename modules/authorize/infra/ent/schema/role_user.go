package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

type RoleUserMixin struct {
	mixin.Schema
}

func (RoleUserMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("approver_id").Immutable(),

		field.String("receiver_ref").Immutable(),

		field.Enum("receiver_type").
			Values("user", "group").
			Immutable(),

		field.String("role_id").Immutable(),
	}
}

type RoleUser struct {
	ent.Schema
}

func (RoleUser) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("role", Role.Type).
			Field("role_id").
			Immutable().
			Required().
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),

		// NOT TODO: No need to add edge to User because each module should be loose-coupled
		// in terms of data.
	}
}

func (RoleUser) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "authz_role_user"},
	}
}

func (RoleUser) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("role_id", "receiver_ref").Unique(),
	}
}

func (RoleUser) Mixin() []ent.Mixin {
	return []ent.Mixin{
		RoleUserMixin{},
	}
}
