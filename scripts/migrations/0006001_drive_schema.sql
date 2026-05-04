-- Create "dri_files" table
CREATE TABLE "dri_files" (
  "id" character varying NOT NULL,
  "owner_ref" character varying NOT NULL,
  "parent_file_ref" character varying NULL,
  "name" character varying NOT NULL,
  "mime" character varying NULL,
  "is_folder" boolean NOT NULL,
  "size" bigint NOT NULL,
  "storage_path" character varying NULL,
  "storage_key" character varying NULL,
  "storage" character varying NULL,
  "visibility" character varying NOT NULL,
  "status" character varying NOT NULL,
  "deleted_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "dri_files_parent_file_ref_fkey" FOREIGN KEY ("parent_file_ref") REFERENCES "dri_files" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "dri_file_ancestors" table
CREATE TABLE "dri_file_ancestors" (
  "id" character varying NOT NULL,
  "file_ref" character varying NOT NULL,
  "ancestor_ref" character varying NOT NULL,
  "depth" bigint NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "dri_file_ancestors_file_ref_ancestor_ref_ukey" UNIQUE ("file_ref", "ancestor_ref"),
  CONSTRAINT "dri_file_ancestors_file_ref_fkey" FOREIGN KEY ("file_ref") REFERENCES "dri_files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "drivefileancestor_ancestor_ref_idx" to table: "dri_file_ancestors"
CREATE INDEX "drivefileancestor_ancestor_ref_idx" ON "dri_file_ancestors" ("ancestor_ref");
-- Create "dri_file_shares" table
CREATE TABLE "dri_file_shares" (
  "id" character varying NOT NULL,
  "file_ref" character varying NOT NULL,
  "user_ref" character varying NOT NULL,
  "permission" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "dri_file_shares_file_ref_fkey" FOREIGN KEY ("file_ref") REFERENCES "dri_files" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

