package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"

	"github.com/sky-as-code/nikki-erp/common/model"
)

type EnumMixin struct {
	mixin.Schema
}

func (EnumMixin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "core_enums"},
	}
}

func (EnumMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.String("etag"),

		field.JSON("label", model.LangJson{}),

		field.String("value"),

		field.String("type"),
	}
}

func (EnumMixin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("value", "type").
			Unique().
			StorageKey("enum_value_type"),
	}
}

type Enum struct {
	ent.Schema
}

func (Enum) Mixin() []ent.Mixin {
	return []ent.Mixin{
		EnumMixin{},
	}
}
