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

type RevokeRequestMixin struct {
	mixin.Schema
}

func (RevokeRequestMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.String("attachment_url").
			Immutable().
			Optional().
			Nillable(),

		field.String("comment").
			Immutable().
			Optional().
			Nillable(),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.String("created_by").
			Immutable(),

		field.String("etag"),

		field.String("receiver_id").
			Immutable(),

		field.Enum("target_type").
			Values("role", "suite").
			Immutable(),

		field.String("target_role_id").
			Nillable().
			Optional().
			Comment("Must be set NULL before the role is deleted"),

		field.String("target_role_name").
			Comment("Role name must be copied here before the role is deleted"),

		field.String("target_suite_id").
			Nillable().
			Optional().
			Comment("Must be set NULL before the role suite is deleted"),

		field.String("target_suite_name").
			Comment("Role suite name must be copied here before the role suite is deleted"),

		field.Enum("status").
			Values("pending", "approved", "rejected"),
	}
}

type RevokeRequest struct {
	ent.Schema
}

func (RevokeRequest) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "authz_revoke_requests"},
	}
}

func (RevokeRequest) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("permission_histories", PermissionHistory.Type).
			Ref("revoke_request"),

		edge.To("role", Role.Type).
			Field("target_role_id").
			Unique(),
		edge.To("role_suite", RoleSuite.Type).
			Field("target_suite_id").
			Unique(),
	}
}

func (RevokeRequest) Mixin() []ent.Mixin {
	return []ent.Mixin{
		RevokeRequestMixin{},
	}
}
