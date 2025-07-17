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

type ResourceMixin struct {
	mixin.Schema
}

func (ResourceMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.String("name").
			Immutable(),

		field.String("description").
			Optional(),

		field.String("etag"),

		field.Enum("resource_type").
			Values("nikki_application", "custom").
			Immutable().
			Comment("A resource can be a Nikki Application (lifecycle and access managed by Nikki) or a custom string"),

		field.String("resource_ref").
			Optional().
			Immutable().
			Comment("Some resource type requires identifier (ID)"),

		field.Enum("scope_type").
			Values("domain", "org", "hierarchy", "private").
			Immutable().
			Comment("This field cannot be changed to avoid breaking existing entitlements"),
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

func (Resource) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Unique(),
	}
}

func (Resource) Mixin() []ent.Mixin {
	return []ent.Mixin{
		ResourceMixin{},
	}
}
