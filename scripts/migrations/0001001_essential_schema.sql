-- Create "essential_currencies" table
CREATE TABLE "essential_currencies" (
  "id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "numeric_code" character varying NULL,
  "name" jsonb NOT NULL,
  "symbol" character varying NULL,
  "decimal_places" integer NOT NULL,
  "is_active" boolean NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essent_currencies_code_ukey" UNIQUE ("code")
);
-- Create index "essent_currencies_is_active_idx" to table: "essential_currencies"
CREATE INDEX "essent_currencies_is_active_idx" ON "essential_currencies" ("is_active");
-- Create "essential_enums" table
CREATE TABLE "essential_enums" (
  "id" character varying NOT NULL,
  "label" jsonb NOT NULL,
  "value" character varying NULL,
  "type" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "essential_languages" table
CREATE TABLE "essential_languages" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "iso_code" character varying NOT NULL,
  "direction" character varying NOT NULL,
  "decimal_separator" character varying NOT NULL,
  "thousands_separator" character varying NOT NULL,
  "date_format" character varying NOT NULL,
  "time_format" character varying NOT NULL,
  "short_time_format" character varying NOT NULL,
  "first_day_of_week" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_languages_iso_code_ukey" UNIQUE ("iso_code")
);
-- Create "essential_model_metadata" table
CREATE TABLE "essential_model_metadata" (
  "id" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "code" character varying NOT NULL,
  "code_prefix" character varying NULL,
  "code_last_seq" integer NOT NULL,
  "padding" integer NOT NULL,
  "is_archived" boolean NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "essential_modules" table
CREATE TABLE "essential_modules" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "label" jsonb NOT NULL,
  "is_orphaned" boolean NOT NULL,
  "is_internal" boolean NOT NULL,
  "version" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_modules_name_ukey" UNIQUE ("name")
);
-- Create "essential_tags" table
CREATE TABLE "essential_tags" (
  "id" character varying NOT NULL,
  "label" jsonb NOT NULL,
  "type" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "essential_field_metadata" table
CREATE TABLE "essential_field_metadata" (
  "id" character varying NOT NULL,
  "model_metadata_id" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "code" character varying NOT NULL,
  "data_type" character varying NOT NULL,
  "is_required" boolean NOT NULL,
  "display_order" integer NOT NULL,
  "is_archived" boolean NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_field_metadata_model_metadata_id_code_ukey" UNIQUE ("model_metadata_id", "code"),
  CONSTRAINT "essential_field_metadata_model_metadata_id_fkey" FOREIGN KEY ("model_metadata_id") REFERENCES "essential_model_metadata" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "essential_uomcats" table
CREATE TABLE "essential_uomcats" (
  "id" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "reference_uom_id" character varying NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_uomcats_name_org_id_ukey" UNIQUE ("name", "org_id")
);
-- Create "essential_uoms" table
CREATE TABLE "essential_uoms" (
  "id" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "symbol" character varying NOT NULL,
  "category_id" character varying NOT NULL,
  "uom_type" character varying NOT NULL,
  "factor" numeric NOT NULL,
  "rounding" numeric NOT NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_uoms_symbol_org_id_ukey" UNIQUE ("symbol", "org_id"),
  CONSTRAINT "essential_uoms_category_id_fkey" FOREIGN KEY ("category_id") REFERENCES "essential_uomcats" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
