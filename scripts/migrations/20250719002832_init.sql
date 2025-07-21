-- Modify "ident_hierarchy_levels" table
ALTER TABLE "ident_hierarchy_levels" ADD COLUMN "deleted_at" timestamptz NULL, ADD COLUMN "deleted_by" character varying NULL, ADD COLUMN "updated_at" timestamptz NULL;
