package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

type MethodSettingMixin struct {
	mixin.Schema
}

func (MethodSettingMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.String("method"),

		field.Int("order"),

		field.Int("max_failures"),

		field.Int("lock_duration_secs").
			Optional().
			Nillable(),

		field.String("subject_type").
			Immutable(),

		field.String("subject_ref").
			Optional().
			Nillable().
			Immutable(),

		field.String("subject_source_ref").
			Optional().
			Nillable().
			Immutable(),
	}
}

type MethodSetting struct {
	ent.Schema
}

func (MethodSetting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "authn_method_settings"},
	}
}

func (MethodSetting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("subject_type", "subject_ref"), // List all settings for a subject
	}
}

func (MethodSetting) Mixin() []ent.Mixin {
	return []ent.Mixin{
		MethodSettingMixin{},
	}
}
