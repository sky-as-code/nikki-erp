package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

type LoginAttemptMixin struct {
	mixin.Schema
}

func (LoginAttemptMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Strings("methods").
			Immutable(),

		field.String("current_method").
			Optional().
			Nillable(),

		field.String("device_ip").
			Optional().
			Nillable(),

		field.String("device_name").
			Optional().
			Nillable(),

		field.String("device_location").
			Optional().
			Nillable(),

		field.Time("expired_at"),

		field.Bool("is_genuine").
			Comment("Whether user has confirmed it is them"),

		field.String("subject_type").
			Immutable(),

		field.String("subject_ref").
			Immutable(),

		field.String("subject_source_ref").
			Optional().
			Nillable().
			Immutable(),

		field.String("status"),

		field.Time("updated_at").
			Optional().
			Nillable(),
	}
}

type LoginAttempt struct {
	ent.Schema
}

func (LoginAttempt) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "authn_attempts"},
	}
}

func (LoginAttempt) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("subject_type", "subject_ref"), // List all login attempts for a subject
	}
}

func (LoginAttempt) Mixin() []ent.Mixin {
	return []ent.Mixin{
		LoginAttemptMixin{},
	}
}
