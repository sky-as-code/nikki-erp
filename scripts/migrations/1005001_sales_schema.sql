-- Create "sales_channels" table
CREATE TABLE "sales_channels" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "managed_by_module" character varying NULL,
  "status" character varying NOT NULL,
  "is_system" boolean NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_channels_code_ukey" UNIQUE ("code")
);
-- Create "sales_points" table
CREATE TABLE "sales_points" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_channel_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "code" character varying NULL,
  "external_reference_id" character varying NULL,
  "external_reference_type" character varying NULL,
  "status" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_points_sales_channel_id_fkey" FOREIGN KEY ("sales_channel_id") REFERENCES "sales_channels" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create unique index "sales_points_tid_channel_ext_ref_ukey" to table: "sales_points"
CREATE UNIQUE INDEX "sales_points_tid_channel_ext_ref_ukey" ON "sales_points" ("sales_channel_id", "external_reference_id") WHERE "external_reference_id" IS NOT NULL;
-- Create unique index "sales_points_tid_channel_code_ukey" to table: "sales_points"
CREATE UNIQUE INDEX "sales_points_tid_channel_code_ukey" ON "sales_points" ("sales_channel_id", "code") WHERE "code" IS NOT NULL;
-- Create index "sales_points_tid_channel_status_idx" to table: "sales_points"
CREATE INDEX "sales_points_tid_channel_status_idx" ON "sales_points" ("sales_channel_id", "status");
