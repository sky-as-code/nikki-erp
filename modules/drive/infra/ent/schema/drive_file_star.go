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

type DriveFileStarMixin struct {
	mixin.Schema
}

func (DriveFileStarMixin) Fields() []ent.Field {
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

		field.String("file_ref").
			NotEmpty(),

		// Soft reference to identity user; no FK edge.
		field.String("user_ref").
			NotEmpty(),
	}
}

type DriveFileStar struct {
	ent.Schema
}

func (DriveFileStar) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "dri_file_stars"},
	}
}

func (DriveFileStar) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("drive_files", DriveFile.Type).
			Ref("drive_file_stars").
			Field("file_ref").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (DriveFileStar) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("file_ref", "user_ref").Unique(),
	}
}

func (DriveFileStar) Mixin() []ent.Mixin {
	return []ent.Mixin{
		DriveFileStarMixin{},
	}
}
