package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"

	coreSchema "github.com/sky-as-code/nikki-erp/modules/core/infra/ent/schema"
)

type UserStatusEnum struct {
	ent.Schema
}

func (UserStatusEnum) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("users", User.Type).
			Ref("user_status"),
	}
}

func (UserStatusEnum) Mixin() []ent.Mixin {
	return []ent.Mixin{
		coreSchema.EnumMixin{},
	}
}
