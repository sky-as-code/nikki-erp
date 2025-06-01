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

type PermissionHistoryMixin struct {
	mixin.Schema
}

func (PermissionHistoryMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		// If grant_request_id is NULL, that means the approver assigned
		// either entitlement, role or role suite manually to receiving user.
		// Similarly to revoke_request_id
		field.String("approver_id").
			Immutable(),

		// Denormalized field. For historical data when approver account is deleted.
		field.String("approver_email").
			Immutable(),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Enum("effect").
			Values("GRANT", "REVOKE", "ENT_DELETED", "ROLE_DELETED", "SUITE_DELETED").
			Immutable(),

		field.String("entitlement_id").
			Immutable().
			Optional().
			Nillable(),

		field.String("entitlement_expr").
			Immutable().
			Optional().
			Nillable(),

		field.String("receiver_id").Immutable(),

		field.String("grant_request_id").
			Immutable().
			Optional().
			Nillable(),

		field.String("revoke_request_id").
			Immutable().
			Optional().
			Nillable(),

		field.String("role_id").
			Immutable().
			Optional().
			Nillable(),

		field.String("role_name").
			Immutable().
			Optional().
			Nillable(),

		field.String("role_suite_id").
			Immutable().
			Optional().
			Nillable(),

		field.String("role_suite_name").
			Immutable().
			Optional().
			Nillable(),
	}
}

type PermissionHistory struct {
	ent.Schema
}

func (PermissionHistory) Fields() []ent.Field {
	return nil
}

func (PermissionHistory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entitlement", Entitlement.Type).
			Field("entitlement_id").
			Immutable().
			Unique(),
		edge.To("role", Role.Type).
			Field("role_id").
			Immutable().
			Unique(),
		edge.To("role_suite", RoleSuite.Type).
			Field("role_suite_id").
			Immutable().
			Unique(),
		edge.To("grant_request", GrantRequest.Type).
			Field("grant_request_id").
			Immutable().
			Unique(),
		edge.To("revoke_request", RevokeRequest.Type).
			Field("revoke_request_id").
			Immutable().
			Unique(),

		// NOT TODO: No need to add edge to User because each module should be loose-coupled
		// in terms of data.
	}
}

func (PermissionHistory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "authz_permission_histories"},
	}
}

func (PermissionHistory) Mixin() []ent.Mixin {
	return []ent.Mixin{
		PermissionHistoryMixin{},
	}
}
