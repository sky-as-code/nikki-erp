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
	"github.com/sky-as-code/nikki-erp/modules/drive/enum"
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

		field.String("visibility").
			Default(enum.DriveFileVisibilityName[enum.DriveFileVisibilityOwner]).
			NotEmpty(),

		field.String("status").
			Default(enum.DriveFileStatusName[enum.DriveFileStatusDefault]).
			NotEmpty(),
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
		index.Fields("owner_ref", "parent_file_ref", "is_folder", "name").Unique().
			Annotations(entsql.IndexWhere("parent_file_ref is NOT NULL")),

		index.Fields("owner_ref", "is_folder", "name").Unique(),

		index.Fields("parent_file_ref"),

		index.Fields("status"),
	}
}

func (DriveFile) Mixin() []ent.Mixin {
	return []ent.Mixin{
		DriveFileMixin{},
	}
}
