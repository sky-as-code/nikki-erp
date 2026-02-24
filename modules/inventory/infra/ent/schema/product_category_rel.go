package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type ProductCategoryRelMixin struct {
	mixin.Schema
}

func (ProductCategoryRelMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("product_id").Immutable(),

		field.String("product_category_id").Immutable(),
	}
}

type ProductCategoryRel struct {
	ent.Schema
}

func (ProductCategoryRel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		field.ID("product_category_id", "product_id"),
		entsql.Annotation{Table: "product_category_rel"},
	}
}

func (ProductCategoryRel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("product", Product.Type).
			Field("product_id").
			Immutable().
			Required().
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),

		edge.To("product_category", ProductCategory.Type).
			Field("product_category_id").
			Immutable().
			Required().
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (ProductCategoryRel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		ProductCategoryRelMixin{},
	}
}
