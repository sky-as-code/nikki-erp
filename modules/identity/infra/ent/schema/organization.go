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

type OrganizationMixin struct {
	mixin.Schema
}

func (OrganizationMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			NotEmpty().
			Immutable().
			StorageKey("id"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.String("created_by").
			NotEmpty().
			Immutable(),

		field.String("display_name").
			NotEmpty().
			MaxLen(50).
			Comment("Human-friendly-readable organization name"),

		field.String("etag").
			MaxLen(100),

		field.Enum("status").
			Values("active", "inactive").
			Default("inactive"),

		field.String("slug").
			NotEmpty().
			MaxLen(50).
			Unique().
			Comment("URL-safe organization name"),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),

		field.String("updated_by").
			Optional().
			Nillable(),
	}
}

func (OrganizationMixin) Edges() []ent.Edge {
	return nil
}

type Organization struct {
	ent.Schema
}

func (Organization) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "organizations"},
	}
}

func (Organization) Fields() []ent.Field {
	return nil
}

func (Organization) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("users", User.Type).
			Ref("orgs").
			Through("user_orgs", UserOrg.Type),
		// edge.To("orgs", Organization.Type). // Self-referential parent org (NULL for top-level)
		// 	Annotations(entsql.Annotation{
		// 		OnDelete: entsql.Cascade,
		// 	}),
		// edge.To("users", User.Type).
		// 	Annotations(entsql.Annotation{
		// 		OnDelete: entsql.Cascade,
		// 	}),
	}
}

func (Organization) Mixin() []ent.Mixin {
	return []ent.Mixin{
		OrganizationMixin{},
	}
}
