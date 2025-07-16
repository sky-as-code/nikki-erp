package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

type PasswordStoreMixin struct {
	mixin.Schema
}

func (PasswordStoreMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.String("password").
			Optional().
			Nillable(),

		field.Time("password_expired_at").
			Optional().
			Nillable(),

		field.Time("password_updated_at").
			Optional().
			Nillable(),

		field.String("passwordtmp").
			Optional().
			Nillable(),

		field.Time("passwordtmp_expired_at").
			Optional().
			Nillable(),

		field.String("passwordotp").
			Optional().
			Nillable(),

		field.Time("passwordotp_expired_at").
			Optional().
			Nillable(),

		field.String("subject_type").
			Immutable(),

		field.String("subject_ref").
			Immutable(),

		field.String("subject_source_ref").
			Optional().
			Nillable().
			Immutable(),
	}
}

type PasswordStore struct {
	ent.Schema
}

func (PasswordStore) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "authn_password_stores"},
	}
}

func (PasswordStore) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("subject_type", "subject_ref"),
	}
}

func (PasswordStore) Mixin() []ent.Mixin {
	return []ent.Mixin{
		PasswordStoreMixin{},
	}
}
