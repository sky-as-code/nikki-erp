package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type VariantMixin struct {
	mixin.Schema
}

func (VariantMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.String("barcode").
			Optional().
			Nillable(),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.String("etag"),

		field.Float("proposed_price"),

		field.String("product_id").
			Immutable().
			StorageKey("product_id"),

		field.String("sku").
			Immutable().
			StorageKey("sku"),

		field.String("status").
			Default("active"),

		field.Time("updated_at").
			Optional().
			Nillable(),
	}
}

type Variant struct {
	ent.Schema
}

func (Variant) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "inventory_variant"},
	}
}

func (Variant) Fields() []ent.Field {
	return nil
}

func (Variant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("product", Product.Type).
			Field("product_id").
			Immutable().
			Required().
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),

		edge.To("attribute_value", AttributeValue.Type).
			Through("variant_attribute_rel", VariantAttributeRel.Type),
	}
}

func (Variant) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VariantMixin{},
	}
}
