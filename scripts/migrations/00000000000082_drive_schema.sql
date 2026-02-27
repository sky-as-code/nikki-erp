-- Modify "dri_files" table
ALTER TABLE "dri_files" ADD COLUMN "materialized_path" character varying NULL;
-- Create "dri_file_stars" table
CREATE TABLE "dri_file_stars" (
  "id" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "user_ref" character varying NOT NULL,
  "file_ref" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "dri_file_stars_dri_files_drive_file_stars" FOREIGN KEY ("file_ref") REFERENCES "dri_files" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "drivefilestar_file_ref_user_ref" to table: "dri_file_stars"
CREATE UNIQUE INDEX "drivefilestar_file_ref_user_ref" ON "dri_file_stars" ("file_ref", "user_ref");
