package schema

import (
	"entgo.io/ent"

	coreSchema "github.com/sky-as-code/nikki-erp/modules/core/infra/ent/schema"
)

type AuthNEnum struct {
	ent.Schema
}

func (AuthNEnum) Edges() []ent.Edge {
	return nil
}

func (AuthNEnum) Mixin() []ent.Mixin {
	return []ent.Mixin{
		coreSchema.EnumMixin{},
	}
}
