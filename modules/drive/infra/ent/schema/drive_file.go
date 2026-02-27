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

type DriveFileMixin struct {
	mixin.Schema
}

func (DriveFileMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable().
			StorageKey("id"),

		field.String("etag"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),

		field.Time("updated_at").
			Default(time.Now),

		field.Time("deleted_at").
			Default(time.Now()),

		field.String("scope_type"),

		field.String("scope_ref"),

		field.String("owner_ref").
			NotEmpty(),

		field.String("parent_file_ref").
			Nillable().
			Optional(),

		field.String("name").
			NotEmpty(),

		field.String("mime"),

		field.Bool("is_folder").
			Default(false),

		field.Int64("size").
			NonNegative(),

		field.String("path"),

		field.String("storage"),

		field.String("visiblity"),
	}
}

type DriveFile struct {
	ent.Schema
}

func (DriveFile) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "dri_files"},
	}
}

func (DriveFile) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("children_file", DriveFile.Type),

		edge.From("parent_file", DriveFile.Type).
			Ref("children_file").
			Field("parent_file_ref").
			Unique(),

		edge.To("drive_file_shares", DriveFileShare.Type),
	}
}

func (DriveFile) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("scope_ref", "parent_file_ref", "is_folder", "name").Unique(),
		index.Fields("scope_ref", "owner_ref", "parent_file_ref").Unique(),
		index.Fields("deleted_at"),
	}
}

func (DriveFile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		DriveFileMixin{},
	}
}
