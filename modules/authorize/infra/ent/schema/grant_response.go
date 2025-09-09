package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

type GrantResponseMixin struct {
	mixin.Schema
}

func (GrantResponseMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.String("request_id").
			Immutable(),

		field.Bool("is_approved").
			Immutable(),

		field.String("reason").
			Immutable().
			Optional().
			Nillable(),

		field.String("responder_id").
			Immutable(),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.String("etag"),
	}
}

type GrantResponse struct {
	ent.Schema
}

func (GrantResponse) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "authz_grant_responses"},
	}
}

func (GrantResponse) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("request_id", "responder_id").Unique(),
	}
}

func (GrantResponse) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("grant_request", GrantRequest.Type).
			Field("request_id").
			Immutable().
			Required().
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (GrantResponse) Mixin() []ent.Mixin {
	return []ent.Mixin{
		GrantResponseMixin{},
	}
}
