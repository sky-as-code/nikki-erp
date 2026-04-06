-- Create "authz_resources" table
CREATE TABLE "authz_resources" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "owner_type" character varying NOT NULL,
  "max_scope" character varying NOT NULL,
  "min_scope" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "authz_resources_name_ukey" UNIQUE ("name")
);
-- Create "authz_actions" table
CREATE TABLE "authz_actions" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "resource_id" character varying NOT NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "authz_actions_resource_id_fkey" FOREIGN KEY ("resource_id") REFERENCES "authz_resources" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "authz_entitlements" table
CREATE TABLE "authz_entitlements" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "action_id" character varying NOT NULL,
  "scope" character varying NULL,
  "org_unit_id" character varying NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "authz_entitlements_name_ukey" UNIQUE ("name"),
  CONSTRAINT "authz_entitlements_action_id_fkey" FOREIGN KEY ("action_id") REFERENCES "authz_actions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "authz_entitlements_org_unit_id_fkey" FOREIGN KEY ("org_unit_id") REFERENCES "authz_resources" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "authz_entitlements_action_id_org_unit_id_ukey_notnull" to table: "authz_entitlements"
CREATE UNIQUE INDEX "authz_entitlements_action_id_org_unit_id_ukey_notnull" ON "authz_entitlements" ("action_id", "org_unit_id") WHERE (org_unit_id IS NOT NULL);
-- Create index "authz_entitlements_action_id_org_unit_id_ukey_null" to table: "authz_entitlements"
CREATE UNIQUE INDEX "authz_entitlements_action_id_org_unit_id_ukey_null" ON "authz_entitlements" ("action_id") WHERE (org_unit_id IS NULL);
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
-- Create "ident_org_units" table
CREATE TABLE "ident_org_units" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "path" character varying[] NOT NULL,
  "parent_id" character varying NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "org_id" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "ident_org_units_name_org_id_ukey" UNIQUE ("name", "org_id"),
  CONSTRAINT "ident_org_units_name_ukey" UNIQUE ("name"),
  CONSTRAINT "ident_org_units_org_id_fkey" FOREIGN KEY ("org_id") REFERENCES "ident_organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "ident_org_units_parent_id_fkey" FOREIGN KEY ("parent_id") REFERENCES "ident_org_units" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "ident_users" table
CREATE TABLE "ident_users" (
  "id" character varying NOT NULL,
  "avatar_url" character varying NULL,
  "display_name" character varying NOT NULL,
  "email" character varying NOT NULL,
  "status" character varying NOT NULL,
  "is_owner" boolean NULL,
  "org_unit_id" character varying NULL,
  "is_archived" boolean NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "ident_users_email_ukey" UNIQUE ("email"),
  CONSTRAINT "ident_users_is_owner_ukey" UNIQUE ("is_owner"),
  CONSTRAINT "ident_users_org_unit_id_fkey" FOREIGN KEY ("org_unit_id") REFERENCES "ident_org_units" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create "ident_groups" table
CREATE TABLE "ident_groups" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "owner_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "ident_groups_name_ukey" UNIQUE ("name"),
  CONSTRAINT "ident_groups_owner_id_fkey" FOREIGN KEY ("owner_id") REFERENCES "ident_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "authz_roles" table
CREATE TABLE "authz_roles" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "dedicated_group_id" character varying NULL,
  "dedicated_user_id" character varying NULL,
  "owner_group_id" character varying NULL,
  "owner_user_id" character varying NULL,
  "is_requestable" boolean NOT NULL,
  "is_required_attachment" boolean NOT NULL,
  "is_required_comment" boolean NOT NULL,
  "org_id" character varying NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "authz_roles_dedicated_group_id_ukey" UNIQUE ("dedicated_group_id"),
  CONSTRAINT "authz_roles_dedicated_user_id_ukey" UNIQUE ("dedicated_user_id"),
  CONSTRAINT "authz_roles_dedicated_group_id_fkey" FOREIGN KEY ("dedicated_group_id") REFERENCES "ident_groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "authz_roles_dedicated_user_id_fkey" FOREIGN KEY ("dedicated_user_id") REFERENCES "ident_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "authz_roles_owner_group_id_fkey" FOREIGN KEY ("owner_group_id") REFERENCES "ident_groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "authz_roles_owner_user_id_fkey" FOREIGN KEY ("owner_user_id") REFERENCES "ident_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "authz_roles_name_org_id_ukey_notnull" to table: "authz_roles"
CREATE UNIQUE INDEX "authz_roles_name_org_id_ukey_notnull" ON "authz_roles" ("name", "org_id") WHERE (org_id IS NOT NULL);
-- Create index "authz_roles_name_org_id_ukey_null" to table: "authz_roles"
CREATE UNIQUE INDEX "authz_roles_name_org_id_ukey_null" ON "authz_roles" ("name") WHERE (org_id IS NULL);
-- Create "authz_entitlement_role_rel" table
CREATE TABLE "authz_entitlement_role_rel" (
  "entitlement_id" character varying NOT NULL,
  "role_id" character varying NOT NULL,
  PRIMARY KEY ("entitlement_id", "role_id"),
  CONSTRAINT "authz_entitlement_role_rel_entitlement_id_fkey" FOREIGN KEY ("entitlement_id") REFERENCES "authz_entitlements" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "authz_entitlement_role_rel_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "authz_roles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "authz_grant_requests" table
CREATE TABLE "authz_grant_requests" (
  "id" character varying NOT NULL,
  "role_id" character varying NOT NULL,
  "receiver_group_id" character varying NULL,
  "receiver_user_id" character varying NULL,
  "status" character varying NOT NULL,
  "type" character varying NOT NULL,
  "attachment_url" character varying NULL,
  "grant_expires_at" timestamptz NULL,
  "request_comment" character varying NULL,
  "requestor_id" character varying NOT NULL,
  "rejection_reason" character varying NULL,
  "responded_at" timestamptz NULL,
  "responder_id" character varying NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "authz_grant_requests_receiver_group_id_fkey" FOREIGN KEY ("receiver_group_id") REFERENCES "ident_groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "authz_grant_requests_receiver_user_id_fkey" FOREIGN KEY ("receiver_user_id") REFERENCES "ident_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "authz_grant_requests_requestor_id_fkey" FOREIGN KEY ("requestor_id") REFERENCES "ident_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "authz_grant_requests_responder_id_fkey" FOREIGN KEY ("responder_id") REFERENCES "ident_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "authz_grant_requests_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "authz_roles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "authz_role_assignments" table
CREATE TABLE "authz_role_assignments" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "role_id" character varying NOT NULL,
  "receiver_group_id" character varying NOT NULL,
  "receiver_user_id" character varying NOT NULL,
  "approver_id" character varying NULL,
  "role_request_id" character varying NULL,
  "expires_at" timestamptz NULL,
  PRIMARY KEY ("id", "role_id"),
  CONSTRAINT "authz_role_assignments_role_id_receiver_group_id_ukey" UNIQUE ("role_id", "receiver_group_id"),
  CONSTRAINT "authz_role_assignments_role_id_receiver_user_id_ukey" UNIQUE ("role_id", "receiver_user_id"),
  CONSTRAINT "authz_role_assignments_approver_id_fkey" FOREIGN KEY ("approver_id") REFERENCES "ident_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "authz_role_assignments_receiver_group_id_fkey" FOREIGN KEY ("receiver_group_id") REFERENCES "ident_groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "authz_role_assignments_receiver_user_id_fkey" FOREIGN KEY ("receiver_user_id") REFERENCES "ident_users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "authz_role_assignments_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "authz_roles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "authz_role_assignments_role_request_id_fkey" FOREIGN KEY ("role_request_id") REFERENCES "authz_grant_requests" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "ident_group_user_rel" table
CREATE TABLE "ident_group_user_rel" (
  "group_id" character varying NOT NULL,
  "user_id" character varying NOT NULL,
  PRIMARY KEY ("group_id", "user_id"),
  CONSTRAINT "ident_group_user_rel_group_id_fkey" FOREIGN KEY ("group_id") REFERENCES "ident_groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "ident_group_user_rel_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "ident_users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "ident_org_user_rel" table
CREATE TABLE "ident_org_user_rel" (
  "org_id" character varying NOT NULL,
  "user_id" character varying NOT NULL,
  PRIMARY KEY ("org_id", "user_id"),
  CONSTRAINT "ident_org_user_rel_org_id_fkey" FOREIGN KEY ("org_id") REFERENCES "ident_organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "ident_org_user_rel_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "ident_users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
