package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/sky-as-code/nikki-erp/modules/contacts/domain"
)

type CommChannelMixin struct {
	mixin.Schema
}

func (CommChannelMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id").
			Comment("ULID"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.String("deleted_by").
			Optional().
			Nillable(),

		field.Time("deleted_at").
			Optional().
			Nillable(),

		field.String("etag"),

		field.String("note").
			Optional().
			Nillable(),

		field.String("org_id").
			Comment("Organization ID"),

		field.String("party_id").
			Comment("Party ID"),

		field.String("type").
			Comment("Channel type including email, phone, facebook, twitter, post , etc"),

		field.Time("updated_at").
			Optional().
			Nillable(),

		field.String("value").
			Optional().
			Nillable(),

		field.JSON("value_json", domain.ValueJsonData{}).
			Optional(),
	}
}

// CommChannel holds the schema definition for a communication channel.
type CommChannel struct {
	ent.Schema
}

func (CommChannel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "contacts_comm_channels"},
	}
}

func (CommChannel) Fields() []ent.Field {
	return nil
}

func (CommChannel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("party", Party.Type).
			Field("party_id").
			Unique().
			Required().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (CommChannel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		CommChannelMixin{},
	}
}
