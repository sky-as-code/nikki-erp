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
			Immutable().
			StorageKey("id"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.String("created_by").
			Immutable(),

		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("Set value for this column when the process is running to delete all resources under this hierarchy level"),

		field.String("deleted_by").
			Optional().
			Nillable().
			Comment("Set value for this column when the process is running to delete all resources under this hierarchy level"),

		field.String("display_name").
			Comment("Human-friendly-readable organization name"),

		field.String("etag"),

		field.Enum("status").
			Values("active", "inactive"),

		field.String("slug").
			Unique().
			Comment("URL-safe organization name"),

		field.Time("updated_at").
			Optional().
			Nillable().
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
		entsql.Annotation{Table: "ident_organizations"},
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

		edge.From("hierarchies", HierarchyLevel.Type).
			Ref("org"),

		edge.To("deleter", User.Type).
			Field("deleted_by").
			Unique(),

		edge.To("groups", Group.Type).
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (Organization) Mixin() []ent.Mixin {
	return []ent.Mixin{
		OrganizationMixin{},
	}
}
