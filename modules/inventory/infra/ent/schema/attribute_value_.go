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

type AttributeValueMixin struct {
	mixin.Schema
}

func (AttributeValueMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.String("attribute_id").
			Immutable().
			StorageKey("attribute_id"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Optional().
			Nillable(),

		field.JSON("value_text", model.LangJson{}).
			Optional(),

		field.Float("value_number").
			Optional().
			Nillable(),

		field.Bool("value_bool").
			Optional().
			Nillable(),

		field.String("value_ref").
			Optional().
			Nillable(),

		field.Int("variant_count"),

		field.String("etag"),
	}
}

type AttributeValue struct {
	ent.Schema
}

func (AttributeValue) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "inventory_attribute_value"},
	}
}

func (AttributeValue) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("attribute", Attribute.Type).
			Field("attribute_id").
			Immutable().
			Required().
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),

		edge.From("variant", Variant.Type).
			Ref("attribute_value").
			Through("variant_attribute_rel", VariantAttributeRel.Type),
	}
}

func (AttributeValue) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AttributeValueMixin{},
	}
}
