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

type UnitMixin struct {
	mixin.Schema
}

func (UnitMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.String("base_unit").
			Optional().
			Nillable(),

		field.String("category_id").
			Optional().
			Nillable(),

		field.Time("created_at").
			Default(time.Now),

		field.String("etag"),

		field.Int("multiplier").
			Optional().
			Nillable(),

		field.String("org_id").
			Optional().
			Nillable(),

		field.JSON("name", model.LangJson{}),

		field.String("status").
			Optional().
			Nillable(),

		field.String("symbol"),

		field.Time("updated_at").
			Optional().
			Nillable(),
	}
}

type Unit struct {
	ent.Schema
}

func (Unit) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "inventory_unit"},
	}
}

func (Unit) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product", Product.Type).
			Ref("unit"),

		edge.To("unit_category", UnitCategory.Type).
			Field("category_id").
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (Unit) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UnitMixin{},
	}
}
