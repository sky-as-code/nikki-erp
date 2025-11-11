package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/sky-as-code/nikki-erp/common/model"
)

type UnitCategoryMixin struct {
	mixin.Schema
}

func (UnitCategoryMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.JSON("name", model.LangJson{}),

		field.Time("updated_at").
			Optional().
			Nillable(),
	}
}

type UnitCategory struct {
	ent.Schema
}

func (UnitCategory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "inventory_unit_category"},
	}
}

func (UnitCategory) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("unit", Unit.Type).
			Ref("unit_category"),
	}
}

func (UnitCategory) Mixin() []ent.Mixin {
	return []ent.Mixin{
		ProductMixin{},
	}
}
