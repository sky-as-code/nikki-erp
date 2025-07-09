package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type UserOrgMixin struct {
	mixin.Schema
}

func (UserOrgMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").Immutable(),
		field.String("org_id").Immutable(),
	}
}

func (UserOrgMixin) Edges() []ent.Edge {
	return nil
}

type UserOrg struct {
	ent.Schema
}

func (UserOrg) Fields() []ent.Field {
	return nil
}

func (UserOrg) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("user", User.Type).
			Field("user_id").
			Unique().
			Immutable().
			Required().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
		edge.To("org", Organization.Type).
			Field("org_id").
			Unique().
			Immutable().
			Required().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (UserOrg) Annotations() []schema.Annotation {
	return []schema.Annotation{
		field.ID("user_id", "org_id"),
		entsql.Annotation{Table: "ident_user_org_rel"},
	}
}

func (UserOrg) Mixin() []ent.Mixin {
	return []ent.Mixin{
		UserOrgMixin{},
	}
}
