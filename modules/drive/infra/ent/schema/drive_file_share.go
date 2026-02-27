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

type DriveFileShareMixin struct {
	mixin.Schema
}

func (DriveFileShareMixin) Fields() []ent.Field {
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

		field.String("scope_type"),

		field.String("scope_ref"),

		field.String("file_ref").
			NotEmpty(),

		field.String("user_ref"),

		field.String("permission").
			Default(enum.DriveFileSharePermName[enum.DriveFileSharePermDefault]),
	}
}

type DriveFileShare struct {
	ent.Schema
}

func (DriveFileShare) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "dri_file_shares"},
	}
}

func (DriveFileShare) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("drive_files", DriveFile.Type).
			Ref("drive_file_shares").
			Field("file_ref").
			Unique().
			Required(),
	}
}

func (DriveFileShare) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("scope_ref", "user_ref").Unique(),
		index.Fields("scope_ref", "file_ref").Unique(),
	}
}

func (DriveFileShare) Mixin() []ent.Mixin {
	return []ent.Mixin{
		DriveFileShareMixin{},
	}
}
