package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type RoleRoleSuiteMixin struct {
	mixin.Schema
}

func (RoleRoleSuiteMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("role_id").Immutable(),

		field.String("role_suite_id").Immutable(),
	}
}

type RoleRoleSuite struct {
	ent.Schema
}

func (RoleRoleSuite) Fields() []ent.Field {
	return nil
}

func (RoleRoleSuite) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("role", Role.Type).
			Field("role_id").
			Immutable().
			Required().
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
		edge.To("role_suite", RoleSuite.Type).
			Field("role_suite_id").
			Immutable().
			Required().
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (RoleRoleSuite) Annotations() []schema.Annotation {
	return []schema.Annotation{
		field.ID("role_suite_id", "role_id"),
		entsql.Annotation{Table: "authz_role_rolesuite"},
	}
}

func (RoleRoleSuite) Mixin() []ent.Mixin {
	return []ent.Mixin{
		RoleRoleSuiteMixin{},
	}
}
