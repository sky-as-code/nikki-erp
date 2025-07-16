-- Create "ident_organizations" table
CREATE TABLE "ident_organizations" ("id" character varying NOT NULL, "created_at" timestamptz NOT NULL, "deleted_at" timestamptz NULL, "address" character varying NULL, "display_name" character varying NOT NULL, "legal_name" character varying NULL, "phone_number" character varying NULL, "etag" character varying NOT NULL, "status" character varying NOT NULL, "slug" character varying NOT NULL, "updated_at" timestamptz NULL, PRIMARY KEY ("id"));
-- Create index "ident_organizations_slug_key" to table: "ident_organizations"
CREATE UNIQUE INDEX "ident_organizations_slug_key" ON "ident_organizations" ("slug");
-- Create "ident_groups" table
CREATE TABLE "ident_groups" ("id" character varying NOT NULL, "created_at" timestamptz NOT NULL, "description" character varying NULL, "email" character varying NULL, "etag" character varying NOT NULL, "name" character varying NOT NULL, "updated_at" timestamptz NULL, "org_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "ident_groups_ident_organizations_org" FOREIGN KEY ("org_id") REFERENCES "ident_organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "ident_groups_name_key" to table: "ident_groups"
CREATE UNIQUE INDEX "ident_groups_name_key" ON "ident_groups" ("name");
-- Create "ident_hierarchy_levels" table
CREATE TABLE "ident_hierarchy_levels" ("id" character varying NOT NULL, "created_at" timestamptz NOT NULL, "etag" character varying NOT NULL, "name" character varying NOT NULL, "parent_id" character varying NULL, "org_id" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "ident_hierarchy_levels_ident_hierarchy_levels_parent" FOREIGN KEY ("parent_id") REFERENCES "ident_hierarchy_levels" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, CONSTRAINT "ident_hierarchy_levels_ident_organizations_org" FOREIGN KEY ("org_id") REFERENCES "ident_organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
-- Create index "hierarchylevel_name_org_id" to table: "ident_hierarchy_levels"
CREATE UNIQUE INDEX "hierarchylevel_name_org_id" ON "ident_hierarchy_levels" ("name", "org_id");
-- Create "ident_users" table
CREATE TABLE "ident_users" ("id" character varying NOT NULL, "avatar_url" character varying NULL, "created_at" timestamptz NOT NULL, "display_name" character varying NOT NULL, "email" character varying NOT NULL, "etag" character varying NOT NULL, "failed_login_attempts" bigint NOT NULL DEFAULT 0, "is_owner" boolean NULL, "last_login_at" timestamptz NULL, "locked_until" timestamptz NULL, "must_change_password" boolean NOT NULL DEFAULT true, "password_hash" character varying NOT NULL, "password_changed_at" timestamptz NOT NULL, "status" character varying NOT NULL, "updated_at" timestamptz NULL, "hierarchy_id" character varying NULL, PRIMARY KEY ("id"), CONSTRAINT "ident_users_ident_hierarchy_levels_hierarchy" FOREIGN KEY ("hierarchy_id") REFERENCES "ident_hierarchy_levels" ("id") ON UPDATE NO ACTION ON DELETE SET NULL);
-- Create index "ident_users_email_key" to table: "ident_users"
CREATE UNIQUE INDEX "ident_users_email_key" ON "ident_users" ("email");
-- Create index "user_is_owner" to table: "ident_users"
CREATE UNIQUE INDEX "user_is_owner" ON "ident_users" ("is_owner");
-- Create "ident_user_group_rel" table
CREATE TABLE "ident_user_group_rel" ("user_id" character varying NOT NULL, "group_id" character varying NOT NULL, PRIMARY KEY ("user_id", "group_id"), CONSTRAINT "ident_user_group_rel_ident_groups_group" FOREIGN KEY ("group_id") REFERENCES "ident_groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "ident_user_group_rel_ident_users_user" FOREIGN KEY ("user_id") REFERENCES "ident_users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create "ident_user_org_rel" table
CREATE TABLE "ident_user_org_rel" ("user_id" character varying NOT NULL, "org_id" character varying NOT NULL, PRIMARY KEY ("user_id", "org_id"), CONSTRAINT "ident_user_org_rel_ident_organizations_org" FOREIGN KEY ("org_id") REFERENCES "ident_organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "ident_user_org_rel_ident_users_user" FOREIGN KEY ("user_id") REFERENCES "ident_users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
