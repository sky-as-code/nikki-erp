-- Create "auditlog_actions" table
CREATE TABLE "auditlog_actions" (
  "id" character varying NOT NULL,
  "org_id" character varying NULL,
  "action" character varying NULL,
  "resource_code" character varying NULL,
  "resource_id" character varying NULL,
  "actor_type" character varying NULL,
  "actor_id" character varying NULL,
  "actor_display_name" character varying NULL,
  "before" jsonb NULL,
  "after" jsonb NULL,
  "request_id" character varying NULL,
  "source" character varying NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "auditlog_actions_actor_idx" to table: "auditlog_actions"
CREATE INDEX "auditlog_actions_actor_idx" ON "auditlog_actions" ("actor_id", "created_at");
-- Create index "auditlog_actions_org_idx" to table: "auditlog_actions"
CREATE INDEX "auditlog_actions_org_idx" ON "auditlog_actions" ("org_id", "created_at");
-- Create index "auditlog_actions_res_idx" to table: "auditlog_actions"
CREATE INDEX "auditlog_actions_res_idx" ON "auditlog_actions" ("resource_code", "resource_id", "created_at");
