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

type EntitlementAssignmentMixin struct {
	mixin.Schema
}

func (EntitlementAssignmentMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.String("entitlement_id").
			Immutable(),

		field.Enum("subject_type").
			Values("nikki_user", "nikki_group", "nikki_role", "custom").
			Immutable(),

		field.String("subject_ref").
			Immutable(),

		field.String("resolved_expr").
			Immutable().
			Comment("Format: '{subjectRef}:{scopeRef}:{resourceName}:{actionName}' E.g: '01JWNXT3EY7FG47VDJTEPTDC98:01JWNZ5KW6WC643VXGKV1D0J64.user:create'"),

		field.String("action_name").
			Optional().
			Nillable().
			Comment("Denormalized action name for easier search and display"),

		field.String("resource_name").
			Optional().
			Nillable().
			Comment("Denormalized resource name for easier search and display"),

		// NULL means regardless of scope
		// field.String("scope_ref").
		// 	Optional().
		// 	Nillable().
		// 	Immutable(),

		// NULL means regardless of level
		field.String("org_id").
			Optional().
			Nillable().
			Immutable(),
	}
}

type EntitlementAssignment struct {
	ent.Schema
}

func (EntitlementAssignment) Fields() []ent.Field {
	return nil
}

func (EntitlementAssignment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entitlement", Entitlement.Type).
			Field("entitlement_id").
			Immutable().
			Required().
			Unique(),
		edge.From("permission_histories", PermissionHistory.Type).
			Ref("entitlement_assignment"),
	}
}

func (EntitlementAssignment) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "authz_entitlement_assignments"},
	}
}

func (EntitlementAssignment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("entitlement_id", "subject_type", "subject_ref", "org_id").Unique(),
		index.Fields("resolved_expr"),
		index.Fields("action_name"),
		index.Fields("resource_name"),
	}
}

func (EntitlementAssignment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		EntitlementAssignmentMixin{},
	}
}
