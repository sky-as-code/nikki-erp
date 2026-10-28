-- Create "settings_quotas" table
CREATE TABLE "settings_quotas" (
  "id" character varying NOT NULL,
  "name" character varying NULL,
  "owner_type" character varying NULL,
  "owner_id" character varying NULL,
  "itstorage_capacity_mibytes" bigint NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "settings_quotas_name_owner_ukey" UNIQUE ("name", "owner_type", "owner_id")
);
-- Create "settings_records" table
CREATE TABLE "settings_records" (
  "id" character varying NOT NULL,
  "schema_id" character varying NULL,
  "module_key" character varying NULL,
  "level" character varying NULL,
  "owner_type" character varying NULL,
  "owner_id" character varying NULL,
  "name" character varying NULL,
  "value" jsonb NULL,
  "allow_override" boolean NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "settings_records_mod_name_owner_ukey" UNIQUE ("module_key", "name", "owner_id")
);
-- Create "settings_schemas" table
CREATE TABLE "settings_schemas" (
  "id" character varying NOT NULL,
  "module_key" character varying NULL,
  "level" character varying NULL,
  "schema" jsonb NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "created_by" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "settings_schemas_mod_lvl_ukey" UNIQUE ("module_key", "level")
);
