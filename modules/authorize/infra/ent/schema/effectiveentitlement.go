package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type EffectiveEntitlement struct {
	ent.View
}

func (EffectiveEntitlement) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id"),
		field.String("action_expr"),
		field.String("resource_id"),
		field.String("scope_ref").Nillable(),
		field.String("source"),
	}
}

func (EffectiveEntitlement) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "authz_effective_user_entitlements", // the view name
		},
	}
}
