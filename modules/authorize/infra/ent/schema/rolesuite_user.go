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

type RoleSuiteUserMixin struct {
	mixin.Schema
}

func (RoleSuiteUserMixin) Fields() []ent.Field {
	return []ent.Field{
		// If request_id is NULL, that means the approver assigned this role manually to receiving user.
		field.String("approver_id").Immutable(),

		field.String("receiver_id").Immutable(),

		field.Enum("receiver_type").
			Values("user", "group").
			Immutable(),

		field.String("role_suite_id").Immutable(),
	}
}

type RoleSuiteUser struct {
	ent.Schema
}

func (RoleSuiteUser) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("role_suite", RoleSuite.Type).
			Field("role_suite_id").
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

func (RoleSuiteUser) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "authz_role_suite_user"},
	}
}

func (RoleSuiteUser) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("role_suite_id", "receiver_id").Unique(),
	}
}

func (RoleSuiteUser) Mixin() []ent.Mixin {
	return []ent.Mixin{
		RoleSuiteUserMixin{},
	}
}
