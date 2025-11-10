package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

type ModuleOrgRelMixin struct {
	mixin.Schema
}

func (ModuleOrgRelMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id").
			Default("").
			Comment("Not used. Just because Ent requires an ID field."),

		field.String("module_id").Immutable(),

		field.String("org_id").Immutable(),
	}
}

type ModuleOrgRel struct {
	ent.Schema
}

func (ModuleOrgRel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("module", Module.Type).
			Field("module_id").
			Immutable().
			Required().
			Unique().
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

func (ModuleOrgRel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("module_id", "org_id").Unique(),
	}
}

func (ModuleOrgRel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "essential_module_org_rel"},
	}
}

func (ModuleOrgRel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		ModuleOrgRelMixin{},
	}
}
