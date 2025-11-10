package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/sky-as-code/nikki-erp/common/model"
)

type ModuleMixin struct {
	mixin.Schema
}

func (ModuleMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.JSON("label", model.LangJson{}),

		field.String("name").
			Unique(),

		field.String("version"),

		field.Bool("is_orphaned"),
	}
}

type Module struct {
	ent.Schema
}

func (Module) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "essential_modules"},
	}
}

func (Module) Mixin() []ent.Mixin {
	return []ent.Mixin{
		ModuleMixin{},
	}
}
