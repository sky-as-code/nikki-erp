-- Create "ident_groups" table
CREATE TABLE "ident_groups" ("id" character varying NOT NULL, "name" character varying NOT NULL, "description" character varying NULL, "email" character varying NULL, "etag" character varying NOT NULL, "created_at" timestamptz NOT NULL, "created_by" character varying NOT NULL, "updated_at" timestamptz NULL, "updated_by" character varying NULL, "parent_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "ident_groups_ident_groups_subgroups" FOREIGN KEY ("parent_id") REFERENCES "ident_groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "ident_groups_name_key" to table: "ident_groups"
CREATE UNIQUE INDEX "ident_groups_name_key" ON "ident_groups" ("name");
-- Create "ident_hierarchy_levels" table
CREATE TABLE "ident_hierarchy_levels" ("id" character varying NOT NULL, "deleted_at" timestamptz NULL, "etag" character varying NOT NULL, "name" character varying NOT NULL, "deleted_by" character varying NULL, "parent_id" character varying NULL, "org_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "ident_hierarchy_levels_ident_hierarchy_levels_parent" FOREIGN KEY ("parent_id") REFERENCES "ident_hierarchy_levels" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "ident_hierarchy_levels_name_key" to table: "ident_hierarchy_levels"
CREATE UNIQUE INDEX "ident_hierarchy_levels_name_key" ON "ident_hierarchy_levels" ("name");
-- Create "ident_organizations" table
CREATE TABLE "ident_organizations" ("id" character varying NOT NULL, "created_at" timestamptz NOT NULL, "created_by" character varying NOT NULL, "deleted_at" timestamptz NULL, "display_name" character varying NOT NULL, "etag" character varying NOT NULL, "status" character varying NOT NULL, "slug" character varying NOT NULL, "updated_at" timestamptz NULL, "updated_by" character varying NULL, "deleted_by" character varying NULL, PRIMARY KEY ("id"));
-- Create index "ident_organizations_slug_key" to table: "ident_organizations"
CREATE UNIQUE INDEX "ident_organizations_slug_key" ON "ident_organizations" ("slug");
-- Create "ident_user_group" table
CREATE TABLE "ident_user_group" ("user_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("user_id", "group_id"));
-- Create "ident_user_org" table
CREATE TABLE "ident_user_org" ("user_id" character varying NOT NULL, "org_id" character varying NOT NULL, PRIMARY KEY ("user_id", "org_id"));
-- Create "ident_users" table
CREATE TABLE "ident_users" ("id" character varying NOT NULL, "avatar_url" character varying NULL, "created_at" timestamptz NOT NULL, "created_by" character varying NOT NULL, "display_name" character varying NOT NULL, "email" character varying NOT NULL, "etag" character varying NOT NULL, "failed_login_attempts" bigint NOT NULL DEFAULT 0, "is_owner" boolean NULL, "last_login_at" timestamptz NULL, "locked_until" timestamptz NULL, "must_change_password" boolean NOT NULL DEFAULT true, "password_hash" character varying NOT NULL, "password_changed_at" timestamptz NOT NULL, "status" character varying NOT NULL, "updated_at" timestamptz NULL, "updated_by" character varying NULL, "hierarchy_id" character varying NULL, PRIMARY KEY ("id"));
-- Create index "ident_users_email_key" to table: "ident_users"
CREATE UNIQUE INDEX "ident_users_email_key" ON "ident_users" ("email");
-- Create index "user_is_owner" to table: "ident_users"
CREATE UNIQUE INDEX "user_is_owner" ON "ident_users" ("is_owner");
-- Modify "ident_hierarchy_levels" table
ALTER TABLE "ident_hierarchy_levels" ADD CONSTRAINT "ident_hierarchy_levels_ident_organizations_org" FOREIGN KEY ("org_id") REFERENCES "ident_organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "ident_hierarchy_levels_ident_users_deleter" FOREIGN KEY ("deleted_by") REFERENCES "ident_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "ident_organizations" table
ALTER TABLE "ident_organizations" ADD CONSTRAINT "ident_organizations_ident_users_deleter" FOREIGN KEY ("deleted_by") REFERENCES "ident_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "ident_user_group" table
ALTER TABLE "ident_user_group" ADD CONSTRAINT "ident_user_group_ident_groups_group" FOREIGN KEY ("group_id") REFERENCES "ident_groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "ident_user_group_ident_users_user" FOREIGN KEY ("user_id") REFERENCES "ident_users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "ident_user_org" table
ALTER TABLE "ident_user_org" ADD CONSTRAINT "ident_user_org_ident_organizations_org" FOREIGN KEY ("org_id") REFERENCES "ident_organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "ident_user_org_ident_users_user" FOREIGN KEY ("user_id") REFERENCES "ident_users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "ident_users" table
ALTER TABLE "ident_users" ADD CONSTRAINT "ident_users_ident_hierarchy_levels_hierarchy" FOREIGN KEY ("hierarchy_id") REFERENCES "ident_hierarchy_levels" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
