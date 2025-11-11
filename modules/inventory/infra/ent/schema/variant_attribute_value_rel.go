package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type VariantAttributeRelMixin struct {
	mixin.Schema
}

func (VariantAttributeRelMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("variant_id").Immutable(),

		field.String("attribute_value_id").Immutable(),
	}
}

type VariantAttributeRel struct {
	ent.Schema
}

func (VariantAttributeRel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		field.ID("variant_id", "attribute_value_id"),
		entsql.Annotation{Table: "variant_attribute_value_rel"},
	}
}

func (VariantAttributeRel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("variant", Variant.Type).
			Field("variant_id").
			Immutable().
			Required().
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
		edge.To("attribute_value", AttributeValue.Type).
			Field("attribute_value_id").
			Immutable().
			Required().
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (VariantAttributeRel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VariantAttributeRelMixin{},
	}
}
