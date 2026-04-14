-- Create "essential_modules" table
CREATE TABLE "essential_modules" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "label" jsonb NOT NULL,
  "is_orphaned" boolean NOT NULL,
  "version" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_modules_name_ukey" UNIQUE ("name")
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
