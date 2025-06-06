-- Modify "groups" table
ALTER TABLE "groups" DROP COLUMN "parent_id", ADD COLUMN "etag" character varying NOT NULL, ADD COLUMN "org_id" character varying NULL, ADD CONSTRAINT "groups_organizations_groups" FOREIGN KEY ("org_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Create index "groups_org_id_key" to table: "groups"
CREATE UNIQUE INDEX "groups_org_id_key" ON "groups" ("org_id");
