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
			NotEmpty().
			Immutable().
			StorageKey("id"),

		// field.String("org_id").
		// 	NotEmpty().
		// 	Immutable(),

		field.String("name").
			Unique().
			NotEmpty().
			MaxLen(50).
			Comment("Group name"),

		field.String("description").
			Optional().
			Nillable().
			MaxLen(255).
			Comment("Group description"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.String("created_by"),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),

		field.String("updated_by").
			Optional().
			Nillable(),

		field.String("parent_id").
			Optional().
			Nillable(),
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
		entsql.Annotation{Table: "groups"},
	}
}

func (Group) Fields() []ent.Field {
	return nil
}

func (Group) Edges() []ent.Edge {
	return []ent.Edge{
		// A group must belong to an organization
		// edge.From("organization", Organization.Type).
		// 	Ref("groups").
		// 	Required().
		// 	Unique(). // O2M relationship
		// 	Immutable().
		// 	Field("org_id"),

		// A group may belong to a parent group (NULL for top-level)
		edge.From("parent", Group.Type).
			Ref("subgroups").
			Unique(). // O2M relationship
			Field("parent_id"),

		// A group can have multiple subgroups
		edge.To("subgroups", Group.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),

		edge.From("users", User.Type).
			Ref("groups").
			Through("user_groups", UserGroup.Type),
	}
}

func (Group) Mixin() []ent.Mixin {
	return []ent.Mixin{
		GroupMixin{},
	}
}
