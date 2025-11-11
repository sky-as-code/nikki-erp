package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/sky-as-code/nikki-erp/common/model"
)

type ProductCategoryMixin struct {
	mixin.Schema
}

func (ProductCategoryMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.JSON("name", model.LangJson{}),

		field.String("parent_id").
			Optional().
			Nillable(),

		field.Time("updated_at").
			Optional().
			Nillable(),
	}
}

type ProductCategory struct {
	ent.Schema
}

func (ProductCategory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "inventory_product_category"},
	}
}

// func (ProductCategory) Edges() []ent.Edge {
// 	return []ent.Edge{
// 		edge.To("children", ProductCategory.Type).
// 			Field("parent_id").
// 			Unique().
// 			Annotations(entsql.Annotation{
// 				OnDelete: entsql.Cascade,
// 			}),

// 		edge.From("parent", ProductCategory.Type).
// 			Ref("children"),
// 	}
// }

func (ProductCategory) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AttributeMixin{},
	}
}
