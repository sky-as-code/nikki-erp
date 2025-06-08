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

type UserMixin struct {
	mixin.Schema
}

func (UserMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.String("avatar_url").
			Optional().
			Nillable().
			Comment("URL to user's profile picture"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.String("display_name"),

		field.String("email").
			Unique(),

		field.String("etag"),

		field.Int("failed_login_attempts").
			Default(0).
			Comment("Count of consecutive failed login attempts"),

		field.String("hierarchy_id").
			Optional().
			Nillable(),

		field.Bool("is_owner").
			Optional().
			Immutable().
			Comment("Whether the user is an owner with root privileges in this deployment"),

		field.Time("last_login_at").
			Optional().
			Nillable(),

		field.Time("locked_until").
			Optional().
			Nillable().
			Comment("Account locked until this timestamp"),

		field.Bool("must_change_password").
			Default(true).
			Comment("Force password change on next login"),

		field.String("password_hash").
			Sensitive(),

		field.Time("password_changed_at").
			Comment("Last password change timestamp"),

		field.Enum("status").
			Values("active", "inactive", "locked"),

		field.Time("updated_at").
			Optional().
			Nillable(),
	}
}

type User struct {
	ent.Schema
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ident_users"},
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_owner").Unique(),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("groups", Group.Type).
			Through("user_groups", UserGroup.Type),

		edge.To("hierarchy", HierarchyLevel.Type).
			Field("hierarchy_id").
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.SetNull,
			}),

		edge.To("orgs", Organization.Type).
			Through("user_orgs", UserOrg.Type),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UserMixin{},
	}
}
