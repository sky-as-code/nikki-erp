-- Create "itstorage_consumptions" table
CREATE TABLE "itstorage_consumptions" (
  "id" character varying NOT NULL,
  "owner_type" character varying NOT NULL,
  "owner_id" character varying NOT NULL,
  "used_bytes" bigint NOT NULL,
  "reserved_bytes" bigint NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "itstorage_consumptions_owner_ukey" UNIQUE ("owner_type", "owner_id")
);
-- Create "itstorage_spaces" table
CREATE TABLE "itstorage_spaces" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "parent_id" character varying NULL,
  "path" character varying NOT NULL,
  "size_bytes" bigint NOT NULL,
  "owner_type" character varying NOT NULL,
  "owner_id" character varying NOT NULL,
  "managed_by_type" character varying NOT NULL,
  "managed_by_id" character varying NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "itstorage_spaces_parent_id_fkey" FOREIGN KEY ("parent_id") REFERENCES "itstorage_spaces" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- Create index "itstorage_spaces_name_parent_id_ukey_notnull" to table: "itstorage_spaces"
CREATE UNIQUE INDEX "itstorage_spaces_name_parent_id_ukey_notnull" ON "itstorage_spaces" ("name", "parent_id") WHERE (parent_id IS NOT NULL);
-- Create index "itstorage_spaces_name_parent_id_ukey_null" to table: "itstorage_spaces"
CREATE UNIQUE INDEX "itstorage_spaces_name_parent_id_ukey_null" ON "itstorage_spaces" ("name") WHERE (parent_id IS NULL);
-- Create index "itstorage_spaces_owner_idx" to table: "itstorage_spaces"
CREATE INDEX "itstorage_spaces_owner_idx" ON "itstorage_spaces" ("owner_type", "owner_id");
-- Create "itstorage_files" table
CREATE TABLE "itstorage_files" (
  "id" character varying NOT NULL,
  "org_id" character varying NULL,
  "name" character varying NOT NULL,
  "mime_type" character varying NOT NULL,
  "size_bytes" bigint NOT NULL,
  "space_id" character varying NOT NULL,
  "object_key" character varying NOT NULL,
  "status" character varying NOT NULL,
  "in_use_at" timestamptz NULL,
  "uploaded_by" character varying NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "itstorage_files_space_id_fkey" FOREIGN KEY ("space_id") REFERENCES "itstorage_spaces" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- Create index "itstorage_files_org_id_idx" to table: "itstorage_files"
CREATE INDEX "itstorage_files_org_id_idx" ON "itstorage_files" ("org_id");
-- Create index "itstorage_files_space_id_idx" to table: "itstorage_files"
CREATE INDEX "itstorage_files_space_id_idx" ON "itstorage_files" ("space_id");
-- Create index "itstorage_files_status_idx" to table: "itstorage_files"
CREATE INDEX "itstorage_files_status_idx" ON "itstorage_files" ("status");
