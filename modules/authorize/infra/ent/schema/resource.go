package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type ResourceMixin struct {
	mixin.Schema
}

func (ResourceMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.String("name"),

		field.Enum("scope_type").
			Values("domain", "org", "hierarchy", "private"),
	}
}

type Resource struct {
	ent.Schema
}

func (Resource) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "authz_resources"},
	}
}

func (Resource) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("actions", Action.Type).
			Ref("resource").
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
		edge.From("entitlements", Entitlement.Type).
			Ref("resource"),
	}
}

func (Resource) Mixin() []ent.Mixin {
	return []ent.Mixin{
		ResourceMixin{},
	}
}
