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

		// Most of the relationships are nullable to allow the referenced record to be deleted,
		// while the history entry is kept for the audit trail.

		// If grant_request_id is NULL, that means the approver assigned
		// either entitlement, role or role suite manually to receiving user.
		// Similarly to revoke_request_id
		field.String("approver_id").
			Optional().
			Nillable().
			Comment("Must be set NULL before the approver account is deleted"),

		field.String("approver_email").
			Comment("Approver email must be copied here before the approver account is deleted"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Enum("effect").
			Values("grant", "revoke").
			Immutable(),

		field.Enum("reason").
			Values(
				"ent_added", "ent_removed", "ent_deleted",
				"ent_added_group", "ent_removed_group", "ent_deleted_group",
				"ent_added_role", "ent_removed_role", "ent_deleted_role",
				"ent_added_role_group", "ent_removed_role_group", "ent_deleted_role_group", "ent_deleted_suite_group",
				"role_added", "role_removed", "role_deleted",
				"role_added_group", "role_removed_group", "role_deleted_group",
				"suite_added", "suite_removed", "suite_deleted",
				"suite_more_role", "suite_less_role",
				"suite_more_role_group", "suite_less_role_group",
				"suite_added_group", "suite_removed_group", "suite_deleted_group",
			).
			Immutable().
			Comment("Permission is granted, revoked because entitlement, role or role suite is added, removed or deleted"),

		field.String("entitlement_id").
			Optional().
			Nillable().
			Comment("Must be set NULL before the entitlement is deleted"),

		field.String("entitlement_expr").
			Comment("Entitlement expression must be copied here before the entitlement is deleted"),

		field.String("entitlement_assignment_id").
			Optional().
			Nillable().
			Comment("Must be set NULL before the entitlement assignment is deleted"),

		field.String("resolved_expr").
			Comment("Resolved expression must be copied here before the entitlement assignment is deleted"),

		field.String("receiver_id").
			Optional().
			Nillable().
			Comment("Must be set NULL before the receiver account is deleted"),

		field.String("receiver_email").
			Comment("Receiver email must be copied here before the receiver account is deleted"),

		field.String("grant_request_id").
			Optional().
			Nillable(),

		field.String("revoke_request_id").
			Optional().
			Nillable(),

		field.String("role_id").
			Optional().
			Nillable().
			Comment("Must be set NULL before the role is deleted"),

		field.String("role_name").
			Comment("Role name must be copied here before the role is deleted"),

		field.String("role_suite_id").
			Optional().
			Nillable().
			Comment("Must be set NULL before the role suite is deleted"),

		field.String("role_suite_name").
			Comment("Role suite name must be copied here before the role suite is deleted"),
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
			Unique(),
		edge.To("entitlement_assignment", EntitlementAssignment.Type).
			Field("entitlement_assignment_id").
			Unique(),
		edge.To("role", Role.Type).
			Field("role_id").
			Unique(),
		edge.To("role_suite", RoleSuite.Type).
			Field("role_suite_id").
			Unique(),
		edge.To("grant_request", GrantRequest.Type).
			Field("grant_request_id").
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.SetNull,
			}),
		edge.To("revoke_request", RevokeRequest.Type).
			Field("revoke_request_id").
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.SetNull,
			}),

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
