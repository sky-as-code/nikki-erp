-- Create "essential_modules" table
CREATE TABLE "essential_modules" (
  "id" character varying NOT NULL,
  "label" jsonb NOT NULL,
  "name" character varying NOT NULL,
  "version" character varying NOT NULL,
  "is_orphaned" boolean NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "essential_modules_name_key" to table: "essential_modules"
CREATE UNIQUE INDEX "essential_modules_name_key" ON "essential_modules" ("name");
-- Create "essential_module_org_rel" table
CREATE TABLE "essential_module_org_rel" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "module_id" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_module_org_rel_essential_modules_module" FOREIGN KEY ("module_id") REFERENCES "essential_modules" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "moduleorgrel_module_id_org_id" to table: "essential_module_org_rel"
CREATE UNIQUE INDEX "moduleorgrel_module_id_org_id" ON "essential_module_org_rel" ("module_id", "org_id");
