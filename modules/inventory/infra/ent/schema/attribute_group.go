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

type AttributeGroupMixin struct {
	mixin.Schema
}

func (AttributeGroupMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.String("id").
			Immutable().
			StorageKey("id"),

		field.Int("index"),

		field.JSON("name", model.LangJson{}),

		field.String("product_id").
			Optional().
			Nillable(),

		field.Time("updated_at").
			Optional().
			Nillable(),
	}
}

type AttributeGroup struct {
	ent.Schema
}

func (AttributeGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "inventory_attribute_group"},
	}
}

func (AttributeGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("attribute", Attribute.Type).
			Ref("attribute_group"),

		edge.To("product", Product.Type).
			Field("product_id").
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (AttributeGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AttributeGroupMixin{},
	}
}
