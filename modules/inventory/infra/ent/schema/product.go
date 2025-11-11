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

type ProductMixin struct {
	mixin.Schema
}

func (ProductMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.String("default_variant_id").
			Optional().
			Nillable(),

		field.JSON("description", model.LangJson{}).
			Optional(),

		field.String("etag"),

		field.JSON("name", model.LangJson{}),

		field.String("org_id").
			Immutable().
			StorageKey("org_id"),

		field.String("status").
			Default("archived"),

		field.String("tag_ids").
			Optional().
			Nillable(),

		field.String("thumbnail_url").
			Optional().
			Nillable(),

		field.String("unit_id"),

		field.Time("updated_at").
			Optional().
			Nillable(),
	}
}

type Product struct {
	ent.Schema
}

func (Product) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "inventory_product"},
	}
}

func (Product) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("variant", Variant.Type).
			Ref("product"),

		edge.From("attribute", Attribute.Type).
			Ref("product"),

		edge.From("attribute_group", AttributeGroup.Type).
			Ref("product"),

		edge.To("unit", Unit.Type).
			Field("unit_id").
			Unique().
			Required().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (Product) Mixin() []ent.Mixin {
	return []ent.Mixin{
		ProductMixin{},
	}
}
