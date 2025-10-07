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

type EntitlementMixin struct {
	mixin.Schema
}

func (EntitlementMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		// If NULL, grant all actions
		field.String("action_id").
			Optional().
			Nillable().
			Immutable(),

		field.String("action_expr").
			Immutable().
			Comment("Format: '{resourceName}:{actionName}' E.g: 'user:create', '*:*'"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.String("created_by").
			Immutable(),

		field.String("name").
			Immutable().
			Unique(),

		field.String("description").
			Optional().
			Nillable(),

		field.String("etag"),

		// If NULL, grant specified action on all resources
		field.String("resource_id").
			Optional().
			Nillable().
			Immutable(),

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

type Entitlement struct {
	ent.Schema
}

func (Entitlement) Fields() []ent.Field {
	return nil
}

func (Entitlement) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("permission_histories", PermissionHistory.Type).
			Ref("entitlement"),

		edge.From("entitlement_assignments", EntitlementAssignment.Type).
			Ref("entitlement"),

		edge.To("action", Action.Type).
			Field("action_id").
			Immutable().
			Unique(),

		edge.To("resource", Resource.Type).
			Field("resource_id").
			Immutable().
			Unique(),
	}
}

func (Entitlement) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "authz_entitlements"},
	}
}

func (Entitlement) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("action_expr", "org_id").Unique(),
	}
}

func (Entitlement) Mixin() []ent.Mixin {
	return []ent.Mixin{
		EntitlementMixin{},
	}
}
