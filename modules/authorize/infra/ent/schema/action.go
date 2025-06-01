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

type ActionMixin struct {
	mixin.Schema
}

func (ActionMixin) Fields() []ent.Field {
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

		field.String("resource_id").
			Immutable(),
	}
}

type Action struct {
	ent.Schema
}

func (Action) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "authz_actions"},
	}
}

func (Action) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entitlements", Entitlement.Type).
			Ref("action"),

		edge.To("resource", Resource.Type).
			Field("resource_id").
			Immutable().
			Required().
			Unique(),
	}
}

func (Action) Mixin() []ent.Mixin {
	return []ent.Mixin{
		ActionMixin{},
	}
}
