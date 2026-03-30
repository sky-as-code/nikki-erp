-- Create "ident_organizations" table
CREATE TABLE "ident_organizations" (
  "id" character varying NOT NULL,
  "address" character varying NULL,
  "display_name" character varying NOT NULL,
  "legal_name" character varying NULL,
  "phone_number" character varying NULL,
  "slug" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "ident_organizations_display_name_ukey" UNIQUE ("display_name"),
  CONSTRAINT "ident_organizations_slug_ukey" UNIQUE ("slug")
);
-- Create "ident_hierarchy_levels" table
CREATE TABLE "ident_hierarchy_levels" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "parent_id" character varying NULL,
  "org_id" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "ident_hierarchy_levels_org_id_fkey" FOREIGN KEY ("org_id") REFERENCES "ident_organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "ident_hierarchy_levels_parent_id_fkey" FOREIGN KEY ("parent_id") REFERENCES "ident_hierarchy_levels" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "ident_hierarchy_levels_name_org_id_ukey" to table: "ident_hierarchy_levels"
CREATE UNIQUE INDEX "ident_hierarchy_levels_name_org_id_ukey" ON "ident_hierarchy_levels" ("name") WHERE (org_id IS NULL);
-- Create "ident_groups" table
CREATE TABLE "ident_groups" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "is_archived" boolean NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "ident_groups_name_ukey" UNIQUE ("name")
);
-- Create "ident_users" table
CREATE TABLE "ident_users" (
  "id" character varying NOT NULL,
  "avatar_url" character varying NULL,
  "display_name" character varying NOT NULL,
  "email" character varying NOT NULL,
  "status" character varying NOT NULL,
  "is_owner" boolean NULL,
  "is_archived" boolean NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "hierarchy_id" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "ident_users_email_ukey" UNIQUE ("email"),
  CONSTRAINT "ident_users_is_owner_ukey" UNIQUE ("is_owner"),
  CONSTRAINT "ident_users_hierarchy_id_fkey" FOREIGN KEY ("hierarchy_id") REFERENCES "ident_hierarchy_levels" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create "ident_user_group_rel" table
CREATE TABLE "ident_user_group_rel" (
  "user_id" character varying NOT NULL,
  "group_id" character varying NOT NULL,
  PRIMARY KEY ("user_id", "group_id"),
  CONSTRAINT "ident_user_group_rel_group_id_fkey" FOREIGN KEY ("group_id") REFERENCES "ident_groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "ident_user_group_rel_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "ident_users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "ident_user_org_rel" table
CREATE TABLE "ident_user_org_rel" (
  "user_id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  PRIMARY KEY ("user_id", "org_id"),
  CONSTRAINT "ident_user_org_rel_org_id_fkey" FOREIGN KEY ("org_id") REFERENCES "ident_organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "ident_user_org_rel_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "ident_users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
