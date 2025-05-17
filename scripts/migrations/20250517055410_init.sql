-- Create "groups" table
CREATE TABLE "groups" ("id" character varying NOT NULL, "name" character varying NOT NULL, "description" character varying NULL, "created_at" timestamptz NOT NULL, "created_by" character varying NOT NULL, "updated_at" timestamptz NOT NULL, "updated_by" character varying NULL, "parent_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "groups_groups_subgroups" FOREIGN KEY ("parent_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "groups_name_key" to table: "groups"
CREATE UNIQUE INDEX "groups_name_key" ON "groups" ("name");
-- Create "hierarchy_levels" table
CREATE TABLE "hierarchy_levels" ("id" character varying NOT NULL, "org_id" character varying NOT NULL, "name" character varying NOT NULL, "parent_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "hierarchy_levels_hierarchy_levels_child" FOREIGN KEY ("parent_id") REFERENCES "hierarchy_levels" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "hierarchy_levels_name_key" to table: "hierarchy_levels"
CREATE UNIQUE INDEX "hierarchy_levels_name_key" ON "hierarchy_levels" ("name");
-- Create "users" table
CREATE TABLE "users" ("id" character varying NOT NULL, "avatar_url" character varying NULL, "created_at" timestamptz NOT NULL, "created_by" character varying NOT NULL, "display_name" character varying NOT NULL, "email" character varying NOT NULL, "etag" character varying NOT NULL, "failed_login_attempts" bigint NOT NULL DEFAULT 0, "last_login_at" timestamptz NULL, "locked_until" timestamptz NULL, "must_change_password" boolean NOT NULL DEFAULT true, "password_hash" character varying NOT NULL, "password_changed_at" timestamptz NOT NULL, "status" character varying NOT NULL DEFAULT 'inactive', "updated_at" timestamptz NOT NULL, "updated_by" character varying NULL, PRIMARY KEY ("id"));
-- Create index "users_email_key" to table: "users"
CREATE UNIQUE INDEX "users_email_key" ON "users" ("email");
-- Create "user_groups" table
CREATE TABLE "user_groups" ("user_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("user_id", "group_id"), CONSTRAINT "user_groups_groups_group" FOREIGN KEY ("group_id") REFERENCES "groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "user_groups_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "organizations" table
CREATE TABLE "organizations" ("id" character varying NOT NULL, "created_at" timestamptz NOT NULL, "created_by" character varying NOT NULL, "display_name" character varying NOT NULL, "etag" character varying NOT NULL, "status" character varying NOT NULL DEFAULT 'inactive', "slug" character varying NOT NULL, "updated_at" timestamptz NOT NULL, "updated_by" character varying NULL, PRIMARY KEY ("id"));
-- Create index "organizations_slug_key" to table: "organizations"
CREATE UNIQUE INDEX "organizations_slug_key" ON "organizations" ("slug");
-- Create "user_orgs" table
CREATE TABLE "user_orgs" ("user_id" character varying NOT NULL, "org_id" character varying NOT NULL, PRIMARY KEY ("user_id", "org_id"), CONSTRAINT "user_orgs_organizations_org" FOREIGN KEY ("org_id") REFERENCES "organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "user_orgs_users_user" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
