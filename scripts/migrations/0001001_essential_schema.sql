-- Create "essential_currencies" table
CREATE TABLE "essential_currencies" (
  "id" character varying NOT NULL,
  "code" character varying NULL,
  "numeric_code" character varying NULL,
  "name" jsonb NULL,
  "symbol" character varying NULL,
  "decimal_places" integer NULL,
  "is_active" boolean NULL,
  "is_archived" boolean NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "updated_by" character varying NULL,
  "etag" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essent_currencies_code_ukey" UNIQUE ("code")
);
-- Create index "essent_currencies_is_active_idx" to table: "essential_currencies"
CREATE INDEX "essent_currencies_is_active_idx" ON "essential_currencies" ("is_active");
-- Create "essential_enums" table
CREATE TABLE "essential_enums" (
  "id" character varying NOT NULL,
  "label" jsonb NULL,
  "value" character varying NULL,
  "type" character varying NULL,
  "etag" character varying NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "updated_by" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create "essential_languages" table
CREATE TABLE "essential_languages" (
  "id" character varying NOT NULL,
  "name" character varying NULL,
  "iso_code" character varying NULL,
  "direction" character varying NULL,
  "decimal_separator" character varying NULL,
  "thousands_separator" character varying NULL,
  "date_format" character varying NULL,
  "time_format" character varying NULL,
  "short_time_format" character varying NULL,
  "first_day_of_week" character varying NULL,
  "is_enabled" boolean NULL,
  "etag" character varying NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "updated_by" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_languages_iso_code_ukey" UNIQUE ("iso_code")
);
-- Create "essential_model_metadata" table
CREATE TABLE "essential_model_metadata" (
  "id" character varying NOT NULL,
  "name" jsonb NULL,
  "code" character varying NULL,
  "code_prefix" character varying NULL,
  "code_last_seq" integer NULL,
  "padding" integer NULL,
  "is_archived" boolean NULL,
  "etag" character varying NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "updated_by" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create "essential_modules" table
CREATE TABLE "essential_modules" (
  "id" character varying NOT NULL,
  "name" character varying NULL,
  "label" jsonb NULL,
  "is_orphaned" boolean NULL,
  "is_internal" boolean NULL,
  "version" character varying NULL,
  "has_ui" boolean NULL,
  "model_prefixes" character varying[] NULL,
  "category" character varying NULL,
  "is_enabled" boolean NULL,
  "dependencies" character varying[] NULL,
  "dep_level" integer NULL,
  "etag" character varying NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "updated_by" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_modules_name_ukey" UNIQUE ("name")
);
-- Create "essential_tags" table
CREATE TABLE "essential_tags" (
  "id" character varying NOT NULL,
  "label" jsonb NULL,
  "type" character varying NULL,
  "etag" character varying NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "updated_by" character varying NULL,
  PRIMARY KEY ("id")
);
-- Create "essential_field_metadata" table
CREATE TABLE "essential_field_metadata" (
  "id" character varying NOT NULL,
  "model_metadata_id" character varying NULL,
  "name" jsonb NULL,
  "code" character varying NULL,
  "data_type" character varying NULL,
  "is_required" boolean NULL,
  "display_order" integer NULL,
  "is_archived" boolean NULL,
  "etag" character varying NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "updated_by" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_field_metadata_model_metadata_id_code_ukey" UNIQUE ("model_metadata_id", "code"),
  CONSTRAINT "essential_field_metadata_model_metadata_id_fkey" FOREIGN KEY ("model_metadata_id") REFERENCES "essential_model_metadata" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "essential_uomcats" table
CREATE TABLE "essential_uomcats" (
  "id" character varying NOT NULL,
  "name" jsonb NULL,
  "reference_uom_id" character varying NULL,
  "org_id" character varying NULL,
  "is_archived" boolean NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "updated_by" character varying NULL,
  "etag" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_uomcats_name_org_id_ukey" UNIQUE ("name", "org_id")
);
-- Create "essential_uoms" table
CREATE TABLE "essential_uoms" (
  "id" character varying NOT NULL,
  "name" jsonb NULL,
  "symbol" character varying NULL,
  "category_id" character varying NULL,
  "uom_type" character varying NULL,
  "factor" numeric NULL,
  "rounding" numeric NULL,
  "org_id" character varying NULL,
  "is_archived" boolean NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  "updated_by" character varying NULL,
  "etag" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_uoms_symbol_org_id_ukey" UNIQUE ("symbol", "org_id"),
  CONSTRAINT "essential_uoms_category_id_fkey" FOREIGN KEY ("category_id") REFERENCES "essential_uomcats" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
