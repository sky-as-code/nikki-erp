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

type RoleMixin struct {
	mixin.Schema
}

func (RoleMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.String("created_by").
			Immutable(),

		field.String("display_name"),

		field.String("description"),

		field.String("etag"),

		field.Enum("owner_type").
			Values("user", "group"),

		field.String("owner_id"),

		field.Bool("is_requestable"),

		field.Bool("is_required_attachment"),

		field.Bool("is_required_comment"),
	}
}

type Role struct {
	ent.Schema
}

func (Role) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "authz_roles"},
	}
}

func (Role) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("role_suites", RoleSuite.Type).
			Ref("roles").
			Through("role_rolesuite", RoleRoleSuite.Type),
		edge.From("role_users", RoleUser.Type).
			Ref("role"),
		edge.From("grant_requests", GrantRequest.Type).
			Ref("role"),
		edge.From("revoke_requests", RevokeRequest.Type).
			Ref("role"),
		edge.From("permission_histories", PermissionHistory.Type).
			Ref("role"),
	}
}

func (Role) Mixin() []ent.Mixin {
	return []ent.Mixin{
		RoleMixin{},
	}
}
