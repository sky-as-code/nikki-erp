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

type GroupMixin struct {
	mixin.Schema
}

func (GroupMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.String("description").
			Optional().
			Nillable().
			Comment("Group description"),

		field.String("email").
			Optional().
			Nillable(),

		field.String("etag"),

		field.String("name").
			Unique().
			NotEmpty(),

		field.String("org_id").
			Optional().
			Nillable(),

		field.Time("updated_at").
			Optional().
			Nillable().
			UpdateDefault(time.Now),
	}
}

func (GroupMixin) Edges() []ent.Edge {
	return nil
}

type Group struct {
	ent.Schema
}

func (Group) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ident_groups"},
	}
}

func (Group) Fields() []ent.Field {
	return nil
}

func (Group) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("users", User.Type).
			Ref("groups").
			Through("user_groups", UserGroup.Type),

		// A group may belong to an organization (optional)
		edge.To("org", Organization.Type).
			Field("org_id").
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (Group) Mixin() []ent.Mixin {
	return []ent.Mixin{
		GroupMixin{},
	}
}
