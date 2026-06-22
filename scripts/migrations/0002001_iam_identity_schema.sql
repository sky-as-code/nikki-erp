-- Create "iam_attempts" table
CREATE TABLE "iam_attempts" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "methods" character varying[] NOT NULL,
  "current_method" character varying NOT NULL,
  "device_ip" character varying NULL,
  "device_name" character varying NULL,
  "device_location" character varying NULL,
  "expires_at" timestamptz NOT NULL,
  "principal_type" character varying NOT NULL,
  "status" character varying NOT NULL,
  "username" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create "iam_method_settings" table
CREATE TABLE "iam_method_settings" (
  "id" character varying NOT NULL,
  "method" character varying NOT NULL,
  "order" integer NOT NULL,
  "max_failures" integer NOT NULL,
  "lock_duration_secs" bigint NULL,
  "subject_type" character varying NOT NULL,
  "subject_ref" character varying NULL,
  "subject_source_ref" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create "iam_resources" table
CREATE TABLE "iam_resources" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "code" character varying NOT NULL,
  "description" character varying NULL,
  "owner_type" character varying NOT NULL,
  "max_scope" character varying NOT NULL,
  "min_scope" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_resources_code_ukey" UNIQUE ("code"),
  CONSTRAINT "iam_resources_name_ukey" UNIQUE ("name")
);
-- Create "iam_actions" table
CREATE TABLE "iam_actions" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "code" character varying NOT NULL,
  "description" character varying NULL,
  "resource_id" character varying NOT NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_actions_resource_id_code_ukey" UNIQUE ("resource_id", "code"),
  CONSTRAINT "iam_actions_resource_id_name_ukey" UNIQUE ("resource_id", "name"),
  CONSTRAINT "iam_actions_resource_id_fkey" FOREIGN KEY ("resource_id") REFERENCES "iam_resources" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "iam_organizations" table
CREATE TABLE "iam_organizations" (
  "id" character varying NOT NULL,
  "address" character varying NULL,
  "display_name" character varying NOT NULL,
  "legal_name" character varying NULL,
  "phone_number" character varying NULL,
  "slug" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_organizations_display_name_ukey" UNIQUE ("display_name"),
  CONSTRAINT "iam_organizations_slug_ukey" UNIQUE ("slug")
);
-- Create "iam_org_units" table
CREATE TABLE "iam_org_units" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "path" character varying[] NOT NULL,
  "parent_id" character varying NULL,
  "org_id" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_org_units_name_org_id_ukey" UNIQUE ("name", "org_id"),
  CONSTRAINT "iam_org_units_name_ukey" UNIQUE ("name"),
  CONSTRAINT "iam_org_units_org_id_fkey" FOREIGN KEY ("org_id") REFERENCES "iam_organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "iam_org_units_parent_id_fkey" FOREIGN KEY ("parent_id") REFERENCES "iam_org_units" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "iam_users" table
CREATE TABLE "iam_users" (
  "id" character varying NOT NULL,
  "avatar_url" character varying NULL,
  "display_name" character varying NOT NULL,
  "email" character varying NOT NULL,
  "status" character varying NOT NULL,
  "is_owner" boolean NULL,
  "org_unit_id" character varying NULL,
  "password" character varying NULL,
  "password_expires_at" timestamptz NULL,
  "password_updated_at" timestamptz NULL,
  "passwordtmp" character varying NULL,
  "passwordtmp_expires_at" timestamptz NULL,
  "passwordotp" character varying NULL,
  "passwordotp_expires_at" timestamptz NULL,
  "passwordotp_recovery" character varying[] NULL,
  "is_archived" boolean NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_users_email_ukey" UNIQUE ("email"),
  CONSTRAINT "iam_users_is_owner_ukey" UNIQUE ("is_owner"),
  CONSTRAINT "iam_users_org_unit_id_fkey" FOREIGN KEY ("org_unit_id") REFERENCES "iam_org_units" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create "iam_groups" table
CREATE TABLE "iam_groups" (
  "id" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "description" jsonb NULL,
  "owner_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_groups_name_ukey" UNIQUE ("name"),
  CONSTRAINT "iam_groups_owner_id_fkey" FOREIGN KEY ("owner_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "iam_roles" table
CREATE TABLE "iam_roles" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "owner_group_id" character varying NULL,
  "owner_user_id" character varying NULL,
  "is_private" boolean NOT NULL,
  "is_requestable" boolean NOT NULL,
  "is_required_attachment" boolean NULL,
  "is_required_comment" boolean NULL,
  "org_id" character varying NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_roles_owner_group_id_fkey" FOREIGN KEY ("owner_group_id") REFERENCES "iam_groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "iam_roles_owner_user_id_fkey" FOREIGN KEY ("owner_user_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "iam_roles_name_org_id_ukey_notnull" to table: "iam_roles"
CREATE UNIQUE INDEX "iam_roles_name_org_id_ukey_notnull" ON "iam_roles" ("name", "org_id") WHERE (org_id IS NOT NULL);
-- Create index "iam_roles_name_org_id_ukey_null" to table: "iam_roles"
CREATE UNIQUE INDEX "iam_roles_name_org_id_ukey_null" ON "iam_roles" ("name") WHERE (org_id IS NULL);
-- Create "iam_entitlements" table
CREATE TABLE "iam_entitlements" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "expression" character varying NOT NULL,
  "action_id" character varying NULL,
  "resource_id" character varying NULL,
  "role_id" character varying NOT NULL,
  "scope" character varying NOT NULL,
  "org_id" character varying NULL,
  "org_unit_id" character varying NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_entitlements_role_id_expression_ukey" UNIQUE ("role_id", "expression"),
  CONSTRAINT "iam_entitlements_role_id_name_ukey" UNIQUE ("role_id", "name"),
  CONSTRAINT "iam_entitlements_action_id_fkey" FOREIGN KEY ("action_id") REFERENCES "iam_actions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "iam_entitlements_org_id_fkey" FOREIGN KEY ("org_id") REFERENCES "iam_organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "iam_entitlements_org_unit_id_fkey" FOREIGN KEY ("org_unit_id") REFERENCES "iam_org_units" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "iam_entitlements_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "iam_roles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "iam_grant_requests" table
CREATE TABLE "iam_grant_requests" (
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
  CONSTRAINT "iam_grant_requests_receiver_group_id_fkey" FOREIGN KEY ("receiver_group_id") REFERENCES "iam_groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "iam_grant_requests_receiver_user_id_fkey" FOREIGN KEY ("receiver_user_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "iam_grant_requests_requestor_id_fkey" FOREIGN KEY ("requestor_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "iam_grant_requests_responder_id_fkey" FOREIGN KEY ("responder_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "iam_grant_requests_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "iam_roles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "iam_group_user_rel" table
CREATE TABLE "iam_group_user_rel" (
  "id" character varying NOT NULL,
  "group_id" character varying NOT NULL,
  "user_id" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_group_user_rel_group_id_user_id_ukey" UNIQUE ("group_id", "user_id"),
  CONSTRAINT "iam_group_user_rel_group_id_fkey" FOREIGN KEY ("group_id") REFERENCES "iam_groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "iam_group_user_rel_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "iam_org_user_rel" table
CREATE TABLE "iam_org_user_rel" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "user_id" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_org_user_rel_org_id_user_id_ukey" UNIQUE ("org_id", "user_id"),
  CONSTRAINT "iam_org_user_rel_org_id_fkey" FOREIGN KEY ("org_id") REFERENCES "iam_organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "iam_org_user_rel_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "iam_role_group_assignments" table
CREATE TABLE "iam_role_group_assignments" (
  "id" character varying NOT NULL,
  "role_id" character varying NOT NULL,
  "receiver_group_id" character varying NOT NULL,
  "approver_id" character varying NULL,
  "role_request_id" character varying NULL,
  "expires_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_role_group_assignments_role_id_receiver_group_id_ukey" UNIQUE ("role_id", "receiver_group_id"),
  CONSTRAINT "iam_role_group_assignments_approver_id_fkey" FOREIGN KEY ("approver_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "iam_role_group_assignments_receiver_group_id_fkey" FOREIGN KEY ("receiver_group_id") REFERENCES "iam_groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "iam_role_group_assignments_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "iam_roles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "iam_role_group_assignments_role_request_id_fkey" FOREIGN KEY ("role_request_id") REFERENCES "iam_grant_requests" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "iam_role_user_assignments" table
CREATE TABLE "iam_role_user_assignments" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "role_id" character varying NOT NULL,
  "receiver_user_id" character varying NOT NULL,
  "approver_id" character varying NULL,
  "role_request_id" character varying NULL,
  "expires_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_role_user_assignments_role_id_receiver_user_id_ukey" UNIQUE ("role_id", "receiver_user_id"),
  CONSTRAINT "iam_role_user_assignments_approver_id_fkey" FOREIGN KEY ("approver_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "iam_role_user_assignments_receiver_user_id_fkey" FOREIGN KEY ("receiver_user_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "iam_role_user_assignments_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "iam_roles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "iam_role_user_assignments_role_request_id_fkey" FOREIGN KEY ("role_request_id") REFERENCES "iam_grant_requests" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "iam_user_permissions" table
CREATE TABLE "iam_user_permissions" (
  "user_id" character varying NOT NULL,
  "ent_id" character varying NOT NULL,
  "ent_expression" character varying NOT NULL,
  "action_id" character varying NULL,
  "resource_id" character varying NULL,
  "resource_code" character varying NULL,
  "role_group_assignment_id" character varying NULL,
  "role_user_assignment_id" character varying NULL,
  "scope" character varying NOT NULL,
  "org_id" character varying NULL,
  "org_membership_id" character varying NULL,
  "group_membership_id" character varying NULL,
  "org_unit_id" character varying NULL,
  PRIMARY KEY ("user_id", "ent_id"),
  CONSTRAINT "iam_user_permissions_user_id_ent_expression_ukey" UNIQUE ("user_id", "ent_expression"),
  CONSTRAINT "iam_user_permissions_action_id_fkey" FOREIGN KEY ("action_id") REFERENCES "iam_actions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "iam_user_permissions_ent_id_fkey" FOREIGN KEY ("ent_id") REFERENCES "iam_entitlements" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "iam_user_permissions_group_membership_id_fkey" FOREIGN KEY ("group_membership_id") REFERENCES "iam_group_user_rel" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "iam_user_permissions_org_id_fkey" FOREIGN KEY ("org_id") REFERENCES "iam_organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "iam_user_permissions_org_membership_id_fkey" FOREIGN KEY ("org_membership_id") REFERENCES "iam_org_user_rel" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "iam_user_permissions_org_unit_id_fkey" FOREIGN KEY ("org_unit_id") REFERENCES "iam_org_units" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "iam_user_permissions_resource_id_fkey" FOREIGN KEY ("resource_id") REFERENCES "iam_resources" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "iam_user_permissions_role_group_assignment_id_fkey" FOREIGN KEY ("role_group_assignment_id") REFERENCES "iam_role_group_assignments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "iam_user_permissions_role_user_assignment_id_fkey" FOREIGN KEY ("role_user_assignment_id") REFERENCES "iam_role_user_assignments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "iam_user_permissions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
