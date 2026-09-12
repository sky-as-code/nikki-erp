-- Create "notification_notifications" table
CREATE TABLE "notification_notifications" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "source_module" character varying NOT NULL,
  "source_resource_name" character varying NULL,
  "source_resource_key" jsonb NULL,
  "idempotency_key" character varying NULL,
  "title" character varying NOT NULL,
  "message" character varying NOT NULL,
  "severity" character varying NOT NULL,
  "distribution_mode" character varying NOT NULL,
  "requested_channels" jsonb NULL,
  "metadata" jsonb NULL,
  "expires_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "notif_notifs_org_created_idx" to table: "notification_notifications"
CREATE INDEX "notif_notifs_org_created_idx" ON "notification_notifications" ("org_id", "created_at");
-- Create index "notif_notifs_org_src_idem_ukey" to table: "notification_notifications"
CREATE UNIQUE INDEX "notif_notifs_org_src_idem_ukey" ON "notification_notifications" ("org_id", "source_module", "idempotency_key") WHERE (idempotency_key IS NOT NULL);
-- Create "notification_recipients" table
CREATE TABLE "notification_recipients" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "notification_id" character varying NOT NULL,
  "recipient_user_id" character varying NOT NULL,
  "stream_seq" bigint NOT NULL,
  "read_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "notif_recips_notif_user_ukey" UNIQUE ("notification_id", "recipient_user_id"),
  CONSTRAINT "notification_recipients_notification_id_fkey" FOREIGN KEY ("notification_id") REFERENCES "notification_notifications" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "notif_recips_inbox_idx" to table: "notification_recipients"
CREATE INDEX "notif_recips_inbox_idx" ON "notification_recipients" ("org_id", "recipient_user_id", "read_at", "stream_seq");
-- Create "notification_deliveries" table
CREATE TABLE "notification_deliveries" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "notification_recipient_id" character varying NOT NULL,
  "channel_name" character varying NOT NULL,
  "delivery_status" character varying NOT NULL,
  "skip_reason" character varying NULL,
  "attempt_count" integer NOT NULL,
  "last_attempt_at" timestamptz NULL,
  "sent_at" timestamptz NULL,
  "last_error_code" character varying NULL,
  "last_error_message" character varying NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "notification_deliveries_notification_recipient_id_fkey" FOREIGN KEY ("notification_recipient_id") REFERENCES "notification_recipients" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "notif_delivs_recip_channel_idx" to table: "notification_deliveries"
CREATE INDEX "notif_delivs_recip_channel_idx" ON "notification_deliveries" ("notification_recipient_id", "channel_name");
