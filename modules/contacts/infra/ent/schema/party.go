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

// Party holds the schema definition for the Party entity.
type PartyMixin struct {
	mixin.Schema
}

func (PartyMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id").
			Comment("ULID"),

		field.String("avatarUrl").
			Optional().
			Nillable().
			Comment("URL to avatar image"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("deleted_at").
			Optional().
			Nillable(),

		field.String("deleted_by").
			Optional().
			Nillable(),

		field.String("display_name").
			MaxLen(50).
			Comment("Display name, max 50 characters"),

		field.String("etag"),

		field.String("job_position").
			Optional().
			Nillable().
			Comment("Job title (Individual)"),

		field.String("language_id").
			Optional().
			Nillable(),

		field.String("legal_address").
			Optional().
			Nillable().
			Comment("Registered business address (Company)"),

		field.String("legal_name").
			Optional().
			Nillable().
			MaxLen(100).
			Comment("Legal name (Company)"),

		field.String("nationality_id").
			Optional().
			Nillable(),

		field.String("note").
			Optional().
			Nillable().
			Comment("Notes"),

		field.String("org_id").
			Optional().
			Nillable().
			Comment("Organization ID"),

		field.String("tax_id").
			Optional().
			Nillable().
			Comment("Tax Identification Number"),

		field.String("title").
			Optional().
			Nillable(),

		field.String("type").
			Comment(`"individual" for person, "company" for organization`),

		field.Time("updated_at").
			Optional().
			Nillable(),

		field.String("website").
			Optional().
			Nillable().
			Comment("Website URL"),
	}
}

type Party struct {
	ent.Schema
}

func (Party) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "contacts_parties"},
	}
}

func (Party) Fields() []ent.Field {
	return nil
}

func (Party) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("comm_channels", CommChannel.Type).
			Ref("party"),

		edge.From("relationships_as_source", Relationship.Type).
			Ref("source_party"),

		edge.From("relationships_as_target", Relationship.Type).
			Ref("target_party"),
	}
}

func (Party) Mixin() []ent.Mixin {
	return []ent.Mixin{
		PartyMixin{},
	}
}
