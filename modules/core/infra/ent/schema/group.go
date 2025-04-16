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
			MaxLen(36).
			NotEmpty().
			Immutable().
			Unique().
			Comment("Primary key using UUID format").
			StorageKey("id"),

		field.String("name").
			Unique().
			NotEmpty().
			MaxLen(50).
			Comment("Group name"),

		field.String("description").
			Optional().
			MaxLen(255).
			Comment("Group description"),

		field.Enum("status").
			Values("active", "inactive").
			Default("active"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),

		field.Time("deleted_at").
			Optional().
			Nillable(),

		field.String("created_by").
			Optional(),

		field.String("updated_by").
			Optional(),

		field.String("deleted_by").
			Optional(),
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
