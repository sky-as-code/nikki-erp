package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/mixin"
)

type TagMixin struct {
	mixin.Schema
}

func (TagMixin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "core_tags"},
	}
}

func (TagMixin) Mixin() []ent.Mixin {
	return []ent.Mixin{
		EnumMixin{},
	}
}

type Tag struct {
	ent.Schema
}

func (Tag) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TagMixin{},
	}
}
