package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type EffectiveUserEntitlement struct {
	ent.View
}

func (EffectiveUserEntitlement) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id"),
		field.String("action_expr"),
		field.String("resource_id").Nillable(),
		field.String("resource_name").Nillable(),
		field.String("scope_ref").Nillable(),
		field.String("scope_type").Nillable(),
		field.String("action_id").Nillable(),
		field.String("action_name").Nillable(),
		field.String("source"),
	}
}

func (EffectiveUserEntitlement) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "authz_effective_user_entitlements", // the view name
		},
	}
}
