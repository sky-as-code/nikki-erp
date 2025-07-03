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

		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("Set value for this column when the process is running to delete all resources under this hierarchy level"),

		field.String("address").
			Optional().
			Nillable(),

		field.String("display_name").
			Comment("Human-friendly-readable organization name"),

		field.String("legal_name").
			Optional().
			Nillable(),

		field.String("phone_number").
			Optional().
			Nillable(),

		field.String("etag"),

		field.Enum("status").
			Values("active", "inactive"),

		field.String("slug").
			Unique().
			Comment("URL-safe organization name"),

		field.Time("updated_at").
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

		edge.From("groups", Group.Type).
			Ref("org"),
	}
}

func (Organization) Mixin() []ent.Mixin {
	return []ent.Mixin{
		OrganizationMixin{},
	}
}
