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

type DriveFileAncestorMixin struct {
	mixin.Schema
}

func (DriveFileAncestorMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.String("file_ref").
			NotEmpty().
			Immutable(),

		field.String("ancestor_ref").
			NotEmpty().
			Immutable(),

		field.Int("depth").
			NonNegative().
			Default(0),
	}
}

type DriveFileAncestor struct {
	ent.Schema
}

func (DriveFileAncestor) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "dri_file_ancestors"},
	}
}

func (DriveFileAncestor) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("drive_file", DriveFile.Type).
			Ref("drive_file_ancestors").
			Field("file_ref").
			Unique().
			Required().
			Immutable().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (DriveFileAncestor) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("file_ref", "ancestor_ref").Unique(),
		index.Fields("ancestor_ref"),
	}
}

func (DriveFileAncestor) Mixin() []ent.Mixin {
	return []ent.Mixin{
		DriveFileAncestorMixin{},
	}
}
