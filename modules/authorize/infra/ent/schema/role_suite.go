package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

type RoleSuiteMixin struct {
	mixin.Schema
}

func (RoleSuiteMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.String("created_by").
			Immutable(),

		field.String("name"),

		field.String("description").
			Optional(),

		field.String("etag"),

		field.Enum("owner_type").
			Values("user", "group"),

		field.String("owner_ref"),

		field.Bool("is_requestable"),

		field.Bool("is_required_attachment"),

		field.Bool("is_required_comment"),

		// NULL means regardless of level
		field.String("org_id").
			Optional().
			Nillable().
			Immutable(),
	}
}

type RoleSuite struct {
	ent.Schema
}

func (RoleSuite) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "authz_role_suites"},
	}
}

func (RoleSuite) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("rolesuite_users", RoleSuiteUser.Type).
			Ref("role_suite"),
		edge.From("permission_histories", PermissionHistory.Type).
			Ref("role_suite"),
		edge.From("grant_requests", GrantRequest.Type).
			Ref("role_suite"),
		edge.From("revoke_requests", RevokeRequest.Type).
			Ref("role_suite"),

		edge.To("roles", Role.Type).
			Through("role_rolesuite", RoleRoleSuite.Type),
	}
}

func (RoleSuite) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", "org_id").Unique(),
	}
}

func (RoleSuite) Mixin() []ent.Mixin {
	return []ent.Mixin{
		RoleSuiteMixin{},
	}
}
