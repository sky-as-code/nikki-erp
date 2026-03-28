-- Create "dri_files" table
CREATE TABLE "public"."dri_files" (
  "id" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "deleted_at" timestamptz NOT NULL,
  "owner_ref" character varying NOT NULL,
  "name" character varying NOT NULL,
  "mime" character varying NOT NULL,
  "is_folder" boolean NOT NULL DEFAULT false,
  "size" bigint NOT NULL,
  "storage_path" character varying NOT NULL,
  "storage_key" character varying NOT NULL,
  "storage" character varying NOT NULL,
  "visibility" character varying NOT NULL DEFAULT 'owner',
  "status" character varying NOT NULL DEFAULT 'active',
  "parent_file_ref" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "dri_files_dri_files_children_file" FOREIGN KEY ("parent_file_ref") REFERENCES "public"."dri_files" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "drivefile_owner_ref_is_folder_name" to table: "dri_files"
CREATE UNIQUE INDEX "drivefile_owner_ref_is_folder_name" ON "public"."dri_files" ("owner_ref", "is_folder", "name");
-- Create index "drivefile_owner_ref_parent_file_ref_is_folder_name" to table: "dri_files"
CREATE UNIQUE INDEX "drivefile_owner_ref_parent_file_ref_is_folder_name" ON "public"."dri_files" ("owner_ref", "parent_file_ref", "is_folder", "name") WHERE (parent_file_ref IS NOT NULL);
-- Create index "drivefile_parent_file_ref" to table: "dri_files"
CREATE INDEX "drivefile_parent_file_ref" ON "public"."dri_files" ("parent_file_ref");
-- Create index "drivefile_status" to table: "dri_files"
CREATE INDEX "drivefile_status" ON "public"."dri_files" ("status");
-- Create "dri_file_shares" table
CREATE TABLE "public"."dri_file_shares" (
  "id" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "user_ref" character varying NOT NULL,
  "permission" character varying NOT NULL DEFAULT 'view',
  "file_ref" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "dri_file_shares_dri_files_drive_file_shares" FOREIGN KEY ("file_ref") REFERENCES "public"."dri_files" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "drivefileshare_file_ref_user_ref" to table: "dri_file_shares"
CREATE UNIQUE INDEX "drivefileshare_file_ref_user_ref" ON "public"."dri_file_shares" ("file_ref", "user_ref");
