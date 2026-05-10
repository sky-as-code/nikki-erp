-- Create "essential_contacts" table
CREATE TABLE "essential_contacts" (
  "id" character varying NOT NULL,
  "avatar_url" character varying NULL,
  "display_name" character varying NOT NULL,
  "legal_name" character varying NULL,
  "legal_address" character varying NULL,
  "tax_id" character varying NULL,
  "job_position" character varying NULL,
  "title" character varying NULL,
  "type" character varying NOT NULL,
  "note" character varying NULL,
  "nationality_id" character varying NULL,
  "org_id" character varying NOT NULL,
  "language_id" character varying NULL,
  "website" character varying NULL,
  "is_archived" boolean NOT NULL,
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
-- Create "essential_contact_channels" table
CREATE TABLE "essential_contact_channels" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "contact_id" character varying NOT NULL,
  "type" character varying NOT NULL,
  "value" character varying NULL,
  "value_json" jsonb NULL,
  "note" character varying NULL,
  "is_archived" boolean NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_contact_channels_contact_id_fkey" FOREIGN KEY ("contact_id") REFERENCES "essential_contacts" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "essential_contact_relationships" table
CREATE TABLE "essential_contact_relationships" (
  "id" character varying NOT NULL,
  "contact_id" character varying NOT NULL,
  "target_contact_id" character varying NOT NULL,
  "type" character varying NOT NULL,
  "note" character varying NULL,
  "is_archived" boolean NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_contact_relationships_contact_id_fkey" FOREIGN KEY ("contact_id") REFERENCES "essential_contacts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "essential_contact_relationships_target_contact_id_fkey" FOREIGN KEY ("target_contact_id") REFERENCES "essential_contacts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
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
-- Create "essential_unit_categories" table
CREATE TABLE "essential_unit_categories" (
  "id" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "org_id" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_unit_categories_name_org_id_ukey" UNIQUE ("name", "org_id")
);
-- Create "essential_units" table
CREATE TABLE "essential_units" (
  "id" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "symbol" character varying NOT NULL,
  "status" character varying NOT NULL,
  "base_unit" character varying NULL,
  "multiplier" bigint NULL,
  "category_id" character varying NULL,
  "org_id" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_units_symbol_org_id_ukey" UNIQUE ("symbol", "org_id"),
  CONSTRAINT "essential_units_category_id_fkey" FOREIGN KEY ("category_id") REFERENCES "essential_unit_categories" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
