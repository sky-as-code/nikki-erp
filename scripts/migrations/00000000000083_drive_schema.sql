-- Modify "dri_files" table
ALTER TABLE "dri_files" DROP COLUMN "materialized_path";
-- Create "dri_file_ancestors" table
CREATE TABLE "dri_file_ancestors" (
  "id" character varying NOT NULL,
  "ancestor_ref" character varying NOT NULL,
  "depth" bigint NOT NULL DEFAULT 0,
  "file_ref" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "dri_file_ancestors_dri_files_drive_file_ancestors" FOREIGN KEY ("file_ref") REFERENCES "dri_files" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "drivefileancestor_ancestor_ref" to table: "dri_file_ancestors"
CREATE INDEX "drivefileancestor_ancestor_ref" ON "dri_file_ancestors" ("ancestor_ref");
-- Create index "drivefileancestor_file_ref_ancestor_ref" to table: "dri_file_ancestors"
CREATE UNIQUE INDEX "drivefileancestor_file_ref_ancestor_ref" ON "dri_file_ancestors" ("file_ref", "ancestor_ref");
