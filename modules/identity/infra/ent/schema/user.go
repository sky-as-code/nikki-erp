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

type UserMixin struct {
	mixin.Schema
}

func (UserMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			NotEmpty().
			Immutable().
			StorageKey("id"),

		// field.String("org_id").
		// 	NotEmpty().
		// 	Immutable(),

		field.String("avatar_url").
			Optional().
			Nillable().
			MaxLen(255).
			Comment("URL to user's profile picture"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.String("created_by").
			NotEmpty().
			MaxLen(36).
			Immutable(),

		field.String("display_name").
			NotEmpty().
			MaxLen(50),

		field.String("email").
			Unique().
			NotEmpty().
			MaxLen(100),

		field.String("etag").
			MaxLen(100),

		field.Int("failed_login_attempts").
			Default(0).
			Comment("Count of consecutive failed login attempts"),

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
			Sensitive().
			NotEmpty(),

		field.Time("password_changed_at").
			Comment("Last password change timestamp"),

		field.Enum("status").
			Values("active", "inactive", "locked").
			Default("inactive"),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),

		field.String("updated_by").
			Optional().
			Nillable(),
	}
}

func (UserMixin) Edges() []ent.Edge {
	return nil
}

type User struct {
	ent.Schema
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "users"},
	}
}

func (User) Fields() []ent.Field {
	return nil
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("groups", Group.Type).
			Through("user_groups", UserGroup.Type),
		edge.To("orgs", Organization.Type).
			Through("user_orgs", UserOrg.Type),
		// edge.From("organization", Organization.Type).
		// 	Ref("users").
		// 	Required().
		// 	Unique(). // O2M relationship
		// 	Immutable().
		// 	Field("org_id"),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UserMixin{},
	}
}
