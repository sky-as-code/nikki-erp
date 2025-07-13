package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"

	coreSchema "github.com/sky-as-code/nikki-erp/modules/core/infra/ent/schema"
)

type IdentStatusEnum struct {
	ent.Schema
}

func (IdentStatusEnum) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("orgs", Organization.Type).
			Ref("org_status"),
		edge.From("users", User.Type).
			Ref("user_status"),
	}
}

func (IdentStatusEnum) Mixin() []ent.Mixin {
	return []ent.Mixin{
		coreSchema.EnumMixin{},
	}
}
