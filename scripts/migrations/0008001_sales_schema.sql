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
  "external_reference" character varying NULL,
  "status" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_points_sales_channel_id_fkey" FOREIGN KEY ("sales_channel_id") REFERENCES "sales_channels" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
CREATE INDEX "sales_points_tid_channel_status_idx" ON "sales_points" ("sales_channel_id", "status");
CREATE INDEX "sales_points_tid_channel_ext_ref_idx" ON "sales_points" ("sales_channel_id", "external_reference");
-- The two sales point uniqueness rules are written by hand because the dynamic model cannot
-- express them. Its partial_uniques builder scopes a REQUIRED field by a NULLABLE one and always
-- emits a pair of indexes: one over (not_null + nullable) WHERE nullable IS NOT NULL, and another
-- over the not_null fields ALONE where it IS NULL. Declaring these there would therefore also
-- create a unique index on sales_channel_id by itself for the NULL rows, capping a channel at one
-- sales point with no external reference and one with no code. Here only the IS NOT NULL half is
-- created, which is what the requirement asks for.
--
-- The external_reference index is load-bearing rather than merely defensive: the (channel,
-- reference) pair IS the idempotency mechanism of CreateSalesPoint, so a vending module retrying
-- after a timeout resolves to the point it already created instead of making a second one.
CREATE UNIQUE INDEX "sales_points_channel_ext_ref_ukey" ON "sales_points" ("sales_channel_id", "external_reference") WHERE "external_reference" IS NOT NULL;
CREATE UNIQUE INDEX "sales_points_channel_code_ukey" ON "sales_points" ("sales_channel_id", "code") WHERE "code" IS NOT NULL;
