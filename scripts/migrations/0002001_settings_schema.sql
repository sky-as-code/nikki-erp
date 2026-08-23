-- Create "settings_records" table
CREATE TABLE "settings_records" (
  "id" character varying NOT NULL,
  "schema_id" character varying NOT NULL,
  "module_key" character varying NOT NULL,
  "level" character varying NOT NULL,
  "owner_type" character varying NOT NULL,
  "owner_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "value" jsonb NOT NULL,
  "allow_override" boolean NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "settings_records_mod_name_owner_ukey" UNIQUE ("module_key", "name", "owner_id")
);
-- Create "settings_schemas" table
CREATE TABLE "settings_schemas" (
  "id" character varying NOT NULL,
  "module_key" character varying NOT NULL,
  "level" character varying NOT NULL,
  "schema" jsonb NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "settings_schemas_mod_lvl_ukey" UNIQUE ("module_key", "level")
);
