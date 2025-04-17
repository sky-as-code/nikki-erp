package schema

import (
	"time"

	"fmt"

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
			MaxLen(36).
			Immutable().
			StorageKey("id"),

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
			Immutable(),

		field.String("display_name").
			NotEmpty().
			MaxLen(50),

		field.String("email").
			Unique().
			NotEmpty().
			MaxLen(100),

		field.String("etag").
			MaxLen(100).
			DefaultFunc(func() string {
				return fmt.Sprint(time.Now().UnixNano())
			}),

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
			NotEmpty().
			MaxLen(1000),

		field.Time("password_changed_at").
			Default(time.Now).
			Comment("Last password change timestamp"),

		field.Enum("status").
			Values("active", "inactive", "suspended", "pending").
			Default("pending"),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),

		field.String("username").
			Unique().
			NotEmpty().
			MaxLen(50).
			Comment("Login username"),
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
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UserMixin{},
	}
}
