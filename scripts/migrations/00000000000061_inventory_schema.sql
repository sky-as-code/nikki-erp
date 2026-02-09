-- Create "inventory_product_category" table
CREATE TABLE "inventory_product_category" (
  "id" character varying NOT NULL,
  "code_name" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "data_type" character varying NOT NULL,
  "display_name" jsonb NULL,
  "enum_value_sort" boolean NOT NULL DEFAULT false,
  "enum_text_value" jsonb NULL,
  "enum_number_value" jsonb NULL,
  "etag" character varying NOT NULL,
  "group_id" character varying NULL,
  "is_enum" boolean NOT NULL DEFAULT false,
  "is_required" boolean NOT NULL DEFAULT false,
  "product_id" character varying NOT NULL,
  "sort_index" bigint NOT NULL DEFAULT 0,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "inventory_unit_category" table
CREATE TABLE "inventory_unit_category" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "default_variant_id" character varying NULL,
  "description" jsonb NULL,
  "etag" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "org_id" character varying NOT NULL,
  "status" character varying NOT NULL DEFAULT 'archived',
  "tag_ids" character varying NULL,
  "thumbnail_url" character varying NULL,
  "unit_id" character varying NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "inventory_unit" table
CREATE TABLE "inventory_unit" (
  "id" character varying NOT NULL,
  "base_unit" character varying NULL,
  "created_at" timestamptz NOT NULL,
  "etag" character varying NOT NULL,
  "multiplier" bigint NULL,
  "org_id" character varying NULL,
  "name" jsonb NOT NULL,
  "status" character varying NULL,
  "symbol" character varying NOT NULL,
  "updated_at" timestamptz NULL,
  "category_id" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_unit_inventory_unit_category_unit_category" FOREIGN KEY ("category_id") REFERENCES "inventory_unit_category" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "inventory_product" table
CREATE TABLE "inventory_product" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "default_variant_id" character varying NULL,
  "description" jsonb NULL,
  "etag" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "org_id" character varying NOT NULL,
  "status" character varying NOT NULL DEFAULT 'archived',
  "tag_ids" character varying NULL,
  "thumbnail_url" character varying NULL,
  "updated_at" timestamptz NULL,
  "unit_id" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_product_inventory_unit_unit" FOREIGN KEY ("unit_id") REFERENCES "inventory_unit" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "inventory_attribute_group" table
CREATE TABLE "inventory_attribute_group" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "index" bigint NOT NULL,
  "name" jsonb NOT NULL,
  "updated_at" timestamptz NULL,
  "product_id" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_attribute_group_inventory_product_product" FOREIGN KEY ("product_id") REFERENCES "inventory_product" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "inventory_attribute" table
CREATE TABLE "inventory_attribute" (
  "id" character varying NOT NULL,
  "code_name" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "data_type" character varying NOT NULL,
  "display_name" jsonb NULL,
  "enum_value_sort" boolean NOT NULL DEFAULT false,
  "enum_text_value" jsonb NULL,
  "enum_number_value" jsonb NULL,
  "etag" character varying NOT NULL,
  "is_enum" boolean NOT NULL DEFAULT false,
  "is_required" boolean NOT NULL DEFAULT false,
  "sort_index" bigint NOT NULL DEFAULT 0,
  "updated_at" timestamptz NULL,
  "group_id" character varying NULL,
  "product_id" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_attribute_inventory_attribute_group_attribute_group" FOREIGN KEY ("group_id") REFERENCES "inventory_attribute_group" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "inventory_attribute_inventory_product_product" FOREIGN KEY ("product_id") REFERENCES "inventory_product" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "inventory_attribute_value" table
CREATE TABLE "inventory_attribute_value" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "value_text" jsonb NULL,
  "value_number" double precision NULL,
  "value_bool" boolean NULL,
  "value_ref" character varying NULL,
  "variant_count" bigint NOT NULL,
  "etag" character varying NOT NULL,
  "attribute_id" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_attribute_value_inventory_attribute_attribute" FOREIGN KEY ("attribute_id") REFERENCES "inventory_attribute" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "inventory_variant" table
CREATE TABLE "inventory_variant" (
  "id" character varying NOT NULL,
  "barcode" character varying NULL,
  "created_at" timestamptz NOT NULL,
  "etag" character varying NOT NULL,
  "proposed_price" double precision NOT NULL,
  "sku" character varying NOT NULL,
  "status" character varying NOT NULL DEFAULT 'active',
  "updated_at" timestamptz NULL,
  "product_id" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_variant_inventory_product_product" FOREIGN KEY ("product_id") REFERENCES "inventory_product" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "variant_attribute_value_rel" table
CREATE TABLE "variant_attribute_value_rel" (
  "variant_id" character varying NOT NULL,
  "attribute_value_id" character varying NOT NULL,
  PRIMARY KEY ("variant_id", "attribute_value_id"),
  CONSTRAINT "variant_attribute_value_rel_in_744514f5b5e112d76cfa9f07290de30a" FOREIGN KEY ("attribute_value_id") REFERENCES "inventory_attribute_value" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "variant_attribute_value_rel_inventory_variant_variant" FOREIGN KEY ("variant_id") REFERENCES "inventory_variant" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
