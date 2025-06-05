-- Modify "groups" table
ALTER TABLE "groups" DROP COLUMN "parent_id", ADD COLUMN "etag" character varying NOT NULL, ADD COLUMN "org_id" character varying NULL, ADD CONSTRAINT "groups_organizations_groups" FOREIGN KEY ("org_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Create index "groups_org_id_key" to table: "groups"
CREATE UNIQUE INDEX "groups_org_id_key" ON "groups" ("org_id");
-- Create "group_subgroups" table
CREATE TABLE "group_subgroups" ("group_id" character varying NOT NULL, "subgroup_id" character varying NOT NULL, PRIMARY KEY ("group_id", "subgroup_id"), CONSTRAINT "group_subgroups_group_id" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "group_subgroups_subgroup_id" FOREIGN KEY ("subgroup_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
