-- Create "iam_actions" table
CREATE TABLE "iam_actions" (
  "id" character varying NOT NULL,
  "name" jsonb NULL,
  "code" character varying NULL,
  "description" character varying NULL,
  "resource_id" character varying NULL,
  "etag" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_actions_resource_id_code_ukey" UNIQUE ("resource_id", "code")
);
-- Create "iam_attempts" table
CREATE TABLE "iam_attempts" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "updated_by" character varying NULL,
  "methods" character varying[] NULL,
  "current_method" character varying NULL,
  "device_ip" character varying NULL,
  "device_name" character varying NULL,
  "device_location" character varying NULL,
  "expires_at" timestamptz NULL,
  "revoked_at" timestamptz NULL,
  "principal_type" character varying NULL,
  "status" character varying NULL,
  "username" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create "iam_entitlements" table
CREATE TABLE "iam_entitlements" (
  "id" character varying NOT NULL,
  "name" character varying NULL,
  "description" character varying NULL,
  "expression" character varying NULL,
  "field_name" character varying NULL,
  "action_id" character varying NULL,
  "resource_id" character varying NULL,
  "role_id" character varying NULL,
  "scope" character varying NULL,
  "org_id" character varying NULL,
  "org_unit_id" character varying NULL,
  "is_archived" boolean NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "updated_by" character varying NULL,
  "etag" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_entitlements_role_id_expression_ukey" UNIQUE ("role_id", "expression"),
  CONSTRAINT "iam_entitlements_role_id_name_ukey" UNIQUE ("role_id", "name")
);
-- Create "iam_grant_requests" table
CREATE TABLE "iam_grant_requests" (
  "id" character varying NOT NULL,
  "role_id" character varying NULL,
  "receiver_group_id" character varying NULL,
  "receiver_user_id" character varying NULL,
  "status" character varying NULL,
  "type" character varying NULL,
  "attachment_url" character varying NULL,
  "grant_expires_at" timestamptz NULL,
  "request_comment" character varying NULL,
  "requestor_id" character varying NULL,
  "rejection_reason" character varying NULL,
  "responded_at" timestamptz NULL,
  "responder_id" character varying NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "updated_by" character varying NULL,
  "etag" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create "iam_group_user_rel" table
CREATE TABLE "iam_group_user_rel" (
  "id" character varying NOT NULL,
  "group_id" character varying NULL,
  "user_id" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_group_user_rel_group_id_user_id_ukey" UNIQUE ("group_id", "user_id")
);
-- Create "iam_groups" table
CREATE TABLE "iam_groups" (
  "id" character varying NOT NULL,
  "name" jsonb NULL,
  "description" jsonb NULL,
  "owner_id" character varying NULL,
  "is_archived" boolean NULL,
  "etag" character varying NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "updated_by" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_groups_name_ukey" UNIQUE ("name")
);
-- Create "iam_method_settings" table
CREATE TABLE "iam_method_settings" (
  "id" character varying NOT NULL,
  "method" character varying NULL,
  "order" integer NULL,
  "max_failures" integer NULL,
  "lock_duration_secs" bigint NULL,
  "subject_type" character varying NULL,
  "subject_ref" character varying NULL,
  "subject_source_ref" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create "iam_org_units" table
CREATE TABLE "iam_org_units" (
  "id" character varying NOT NULL,
  "name" character varying NULL,
  "description" character varying NULL,
  "path" character varying[] NULL,
  "parent_id" character varying NULL,
  "org_id" character varying NULL,
  "etag" character varying NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "updated_by" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_org_units_name_org_id_ukey" UNIQUE ("name", "org_id"),
  CONSTRAINT "iam_org_units_name_ukey" UNIQUE ("name"),
  CONSTRAINT "iam_org_units_parent_id_fkey" FOREIGN KEY ("parent_id") REFERENCES "iam_org_units" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "iam_org_user_rel" table
CREATE TABLE "iam_org_user_rel" (
  "id" character varying NOT NULL,
  "org_id" character varying NULL,
  "user_id" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_org_user_rel_org_id_user_id_ukey" UNIQUE ("org_id", "user_id")
);
-- Create "iam_organizations" table
CREATE TABLE "iam_organizations" (
  "id" character varying NOT NULL,
  "address" jsonb NULL,
  "display_name" jsonb NULL,
  "legal_name" jsonb NULL,
  "phone_number" character varying NULL,
  "slug" character varying NULL,
  "status" character varying NULL,
  "status_message" character varying NULL,
  "status_message_vars" jsonb NULL,
  "etag" character varying NULL,
  "is_archived" boolean NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "updated_by" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_organizations_slug_ukey" UNIQUE ("slug")
);
-- Create "iam_password_stores" table
CREATE TABLE "iam_password_stores" (
  "id" character varying NOT NULL,
  "principal_type" character varying NULL,
  "principal_id" character varying NULL,
  "type" character varying NULL,
  "hash" character varying NULL,
  "last_used_at" timestamptz NULL,
  "expires_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_pwd_stores_tid_princ_type_princ_id_type_ukey" UNIQUE ("principal_type", "principal_id", "type")
);
-- Create "iam_permission_histories" table
CREATE TABLE "iam_permission_histories" (
  "id" character varying NOT NULL,
  "approver_id" character varying NULL,
  "approver_email" character varying NULL,
  "effect" character varying NULL,
  "reason" character varying NULL,
  "entitlement_id" character varying NULL,
  "entitlement_expr" character varying NULL,
  "entitlement_assignment_id" character varying NULL,
  "resolved_expr" character varying NULL,
  "receiver_id" character varying NULL,
  "receiver_email" character varying NULL,
  "grant_request_id" character varying NULL,
  "revoke_request_id" character varying NULL,
  "role_id" character varying NULL,
  "role_name" character varying NULL,
  "created_at" timestamptz NULL,
  "created_by" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create "iam_resources" table
CREATE TABLE "iam_resources" (
  "id" character varying NOT NULL,
  "name" jsonb NULL,
  "code" character varying NULL,
  "description" character varying NULL,
  "owner_type" character varying NULL,
  "max_scope" character varying NULL,
  "min_scope" character varying NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "updated_by" character varying NULL,
  "etag" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_resources_code_ukey" UNIQUE ("code")
);
-- Create "iam_role_group_assignments" table
CREATE TABLE "iam_role_group_assignments" (
  "id" character varying NOT NULL,
  "role_id" character varying NULL,
  "receiver_group_id" character varying NULL,
  "approver_id" character varying NULL,
  "role_request_id" character varying NULL,
  "expires_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "created_by" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_role_grp_assigns_tid_role_id_rcv_grp_id_ukey" UNIQUE ("role_id", "receiver_group_id")
);
-- Create "iam_role_user_assignments" table
CREATE TABLE "iam_role_user_assignments" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NULL,
  "created_by" character varying NULL,
  "role_id" character varying NULL,
  "receiver_user_id" character varying NULL,
  "approver_id" character varying NULL,
  "role_request_id" character varying NULL,
  "expires_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_role_usr_assigns_tid_role_id_rcv_usr_id_ukey" UNIQUE ("role_id", "receiver_user_id")
);
-- Create "iam_roles" table
CREATE TABLE "iam_roles" (
  "id" character varying NOT NULL,
  "name" character varying NULL,
  "description" character varying NULL,
  "owner_group_id" character varying NULL,
  "owner_user_id" character varying NULL,
  "is_private" boolean NULL,
  "is_requestable" boolean NULL,
  "is_system" boolean NULL,
  "is_required_attachment" boolean NULL,
  "is_required_comment" boolean NULL,
  "org_id" character varying NULL,
  "is_archived" boolean NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "updated_by" character varying NULL,
  "managed_by_type" character varying NULL,
  "managed_by_id" character varying NULL,
  "etag" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create "iam_user_permissions" table
CREATE TABLE "iam_user_permissions" (
  "user_id" character varying NOT NULL,
  "ent_id" character varying NOT NULL,
  "source_kind" character varying NOT NULL,
  "source_id" character varying NOT NULL,
  "expires_at" timestamptz NULL,
  "ent_expression" character varying NULL,
  "action_id" character varying NULL,
  "resource_id" character varying NULL,
  "resource_code" character varying NULL,
  "role_group_assignment_id" character varying NULL,
  "role_user_assignment_id" character varying NULL,
  "scope" character varying NULL,
  "org_id" character varying NULL,
  "group_membership_id" character varying NULL,
  "org_unit_id" character varying NULL,
  PRIMARY KEY ("user_id", "ent_id", "source_kind", "source_id")
);
-- Create "iam_users" table
CREATE TABLE "iam_users" (
  "id" character varying NOT NULL,
  "avatar_file_id" character varying NULL,
  "display_name" character varying NULL,
  "email" character varying NULL,
  "type" character varying NULL,
  "status" character varying NULL,
  "is_owner" boolean NULL,
  "org_unit_id" character varying NULL,
  "is_archived" boolean NULL,
  "etag" character varying NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "updated_by" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "iam_users_email_ukey" UNIQUE ("email"),
  CONSTRAINT "iam_users_is_owner_ukey" UNIQUE ("is_owner"),
  CONSTRAINT "iam_users_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "iam_users_updated_by_fkey" FOREIGN KEY ("updated_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Modify "iam_actions" table
ALTER TABLE "iam_actions" ADD CONSTRAINT "iam_actions_resource_id_fkey" FOREIGN KEY ("resource_id") REFERENCES "iam_resources" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "iam_attempts" table
ALTER TABLE "iam_attempts" ADD CONSTRAINT "iam_attempts_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "iam_attempts_updated_by_fkey" FOREIGN KEY ("updated_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "iam_entitlements" table
ALTER TABLE "iam_entitlements" ADD CONSTRAINT "iam_entitlements_action_id_fkey" FOREIGN KEY ("action_id") REFERENCES "iam_actions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "iam_entitlements_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "iam_entitlements_org_id_fkey" FOREIGN KEY ("org_id") REFERENCES "iam_organizations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "iam_entitlements_org_unit_id_fkey" FOREIGN KEY ("org_unit_id") REFERENCES "iam_org_units" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "iam_entitlements_resource_id_fkey" FOREIGN KEY ("resource_id") REFERENCES "iam_resources" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "iam_entitlements_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "iam_roles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "iam_entitlements_updated_by_fkey" FOREIGN KEY ("updated_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "iam_grant_requests" table
ALTER TABLE "iam_grant_requests" ADD CONSTRAINT "iam_grant_requests_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "iam_grant_requests_receiver_group_id_fkey" FOREIGN KEY ("receiver_group_id") REFERENCES "iam_groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "iam_grant_requests_receiver_user_id_fkey" FOREIGN KEY ("receiver_user_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "iam_grant_requests_requestor_id_fkey" FOREIGN KEY ("requestor_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "iam_grant_requests_responder_id_fkey" FOREIGN KEY ("responder_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "iam_grant_requests_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "iam_roles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "iam_grant_requests_updated_by_fkey" FOREIGN KEY ("updated_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "iam_group_user_rel" table
ALTER TABLE "iam_group_user_rel" ADD CONSTRAINT "iam_group_user_rel_group_id_fkey" FOREIGN KEY ("group_id") REFERENCES "iam_groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "iam_group_user_rel_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "iam_groups" table
ALTER TABLE "iam_groups" ADD CONSTRAINT "iam_groups_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "iam_groups_owner_id_fkey" FOREIGN KEY ("owner_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "iam_groups_updated_by_fkey" FOREIGN KEY ("updated_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "iam_org_units" table
ALTER TABLE "iam_org_units" ADD CONSTRAINT "iam_org_units_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "iam_org_units_org_id_fkey" FOREIGN KEY ("org_id") REFERENCES "iam_organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "iam_org_units_updated_by_fkey" FOREIGN KEY ("updated_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "iam_org_user_rel" table
ALTER TABLE "iam_org_user_rel" ADD CONSTRAINT "iam_org_user_rel_org_id_fkey" FOREIGN KEY ("org_id") REFERENCES "iam_organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "iam_org_user_rel_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "iam_organizations" table
ALTER TABLE "iam_organizations" ADD CONSTRAINT "iam_organizations_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "iam_organizations_updated_by_fkey" FOREIGN KEY ("updated_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "iam_permission_histories" table
ALTER TABLE "iam_permission_histories" ADD CONSTRAINT "iam_permission_histories_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "iam_permission_histories_grant_request_id_fkey" FOREIGN KEY ("grant_request_id") REFERENCES "iam_grant_requests" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "iam_permission_histories_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "iam_roles" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "iam_resources" table
ALTER TABLE "iam_resources" ADD CONSTRAINT "iam_resources_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "iam_resources_updated_by_fkey" FOREIGN KEY ("updated_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "iam_role_group_assignments" table
ALTER TABLE "iam_role_group_assignments" ADD CONSTRAINT "iam_role_group_assignments_approver_id_fkey" FOREIGN KEY ("approver_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "iam_role_group_assignments_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "iam_role_group_assignments_receiver_group_id_fkey" FOREIGN KEY ("receiver_group_id") REFERENCES "iam_groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "iam_role_group_assignments_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "iam_roles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "iam_role_group_assignments_role_request_id_fkey" FOREIGN KEY ("role_request_id") REFERENCES "iam_grant_requests" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "iam_role_user_assignments" table
ALTER TABLE "iam_role_user_assignments" ADD CONSTRAINT "iam_role_user_assignments_approver_id_fkey" FOREIGN KEY ("approver_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "iam_role_user_assignments_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "iam_role_user_assignments_receiver_user_id_fkey" FOREIGN KEY ("receiver_user_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "iam_role_user_assignments_role_id_fkey" FOREIGN KEY ("role_id") REFERENCES "iam_roles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "iam_role_user_assignments_role_request_id_fkey" FOREIGN KEY ("role_request_id") REFERENCES "iam_grant_requests" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "iam_roles" table
ALTER TABLE "iam_roles" ADD CONSTRAINT "iam_roles_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD CONSTRAINT "iam_roles_owner_group_id_fkey" FOREIGN KEY ("owner_group_id") REFERENCES "iam_groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "iam_roles_owner_user_id_fkey" FOREIGN KEY ("owner_user_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "iam_roles_updated_by_fkey" FOREIGN KEY ("updated_by") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- Modify "iam_user_permissions" table
ALTER TABLE "iam_user_permissions" ADD CONSTRAINT "iam_user_permissions_action_id_fkey" FOREIGN KEY ("action_id") REFERENCES "iam_actions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "iam_user_permissions_ent_id_fkey" FOREIGN KEY ("ent_id") REFERENCES "iam_entitlements" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "iam_user_permissions_group_membership_id_fkey" FOREIGN KEY ("group_membership_id") REFERENCES "iam_group_user_rel" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "iam_user_permissions_org_id_fkey" FOREIGN KEY ("org_id") REFERENCES "iam_organizations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "iam_user_permissions_org_unit_id_fkey" FOREIGN KEY ("org_unit_id") REFERENCES "iam_org_units" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "iam_user_permissions_resource_id_fkey" FOREIGN KEY ("resource_id") REFERENCES "iam_resources" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "iam_user_permissions_role_group_assignment_id_fkey" FOREIGN KEY ("role_group_assignment_id") REFERENCES "iam_role_group_assignments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "iam_user_permissions_role_user_assignment_id_fkey" FOREIGN KEY ("role_user_assignment_id") REFERENCES "iam_role_user_assignments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD CONSTRAINT "iam_user_permissions_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "iam_users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- Modify "iam_users" table
ALTER TABLE "iam_users" ADD CONSTRAINT "iam_users_org_unit_id_fkey" FOREIGN KEY ("org_unit_id") REFERENCES "iam_org_units" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
