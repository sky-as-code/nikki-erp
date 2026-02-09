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

type AttributeMixin struct {
	mixin.Schema
}

func (AttributeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.String("code_name"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.String("data_type"),

		field.JSON("display_name", model.LangJson{}).
			Optional(),

		field.Bool("enum_value_sort").
			Default(false),

		field.JSON("enum_text_value", []model.LangJson{}).
			Optional(),

		field.JSON("enum_number_value", []float64{}).
			Optional(),

		field.String("etag"),

		field.String("group_id").
			Optional().
			Nillable(),

		field.Bool("is_enum").
			Default(false),

		field.Bool("is_required").
			Default(false),

		field.String("product_id"),

		field.Int("sort_index").
			Default(0),

		field.Time("updated_at").
			Optional().
			Nillable(),
	}
}

type Attribute struct {
	ent.Schema
}

func (Attribute) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "inventory_attribute"},
	}
}

func (Attribute) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("attribute_group", AttributeGroup.Type).
			Field("group_id").
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),

		edge.To("product", Product.Type).
			Field("product_id").
			Unique().
			Required().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),

		edge.From("attribute_values", AttributeValue.Type).
			Ref("attribute"),
	}
}

func (Attribute) Mixin() []ent.Mixin {
	return []ent.Mixin{
		AttributeMixin{},
	}
}
