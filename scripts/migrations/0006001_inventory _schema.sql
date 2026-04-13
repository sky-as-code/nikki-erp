-- Create "invent_products" table
CREATE TABLE "invent_products" (
  "id" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "description" jsonb NULL,
  "org_id" character varying NOT NULL,
  "status" character varying NULL,
  "thumbnail_url" character varying NULL,
  "unit_id" character varying NULL,
  "default_variant_id" character varying NULL,
  "tag_ids" character varying NULL,
  "is_archived" boolean NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invent_products_name_ukey" UNIQUE ("name")
);
-- Create "invent_attribute_groups" table
CREATE TABLE "invent_attribute_groups" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "index" bigint NULL,
  "product_id" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invent_attribute_groups_name_product_id_ukey" UNIQUE ("name", "product_id"),
  CONSTRAINT "invent_attribute_groups_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "invent_products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "invent_attributes" table
CREATE TABLE "invent_attributes" (
  "id" character varying NOT NULL,
  "code_name" character varying NOT NULL,
  "display_name" jsonb NULL,
  "sort_index" bigint NULL,
  "data_type" character varying NOT NULL,
  "is_required" boolean NULL,
  "is_enum" boolean NULL,
  "enum_value_sort" boolean NULL,
  "enum_value" jsonb[] NULL,
  "attribute_group_id" character varying NULL,
  "product_id" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invent_attributes_code_name_product_id_ukey" UNIQUE ("code_name", "product_id"),
  CONSTRAINT "invent_attributes_attribute_group_id_fkey" FOREIGN KEY ("attribute_group_id") REFERENCES "invent_attribute_groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "invent_attributes_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "invent_products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "invent_attribute_values" table
CREATE TABLE "invent_attribute_values" (
  "id" character varying NOT NULL,
  "attribute_id" character varying NOT NULL,
  "value_text" jsonb NULL,
  "value_number" numeric NULL,
  "value_bool" boolean NULL,
  "value_ref" character varying NULL,
  "variant_count" bigint NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invent_attribute_values_attribute_id_fkey" FOREIGN KEY ("attribute_id") REFERENCES "invent_attributes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "invent_product_categories" table
CREATE TABLE "invent_product_categories" (
  "id" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "org_id" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invent_product_categories_name_org_id_ukey" UNIQUE ("name", "org_id")
);
-- Create "invent_product_category_rel" table
CREATE TABLE "invent_product_category_rel" (
  "product_id" character varying NOT NULL,
  "product_category_id" character varying NOT NULL,
  PRIMARY KEY ("product_id", "product_category_id"),
  CONSTRAINT "invent_product_category_rel_product_category_id_fkey" FOREIGN KEY ("product_category_id") REFERENCES "invent_product_categories" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "invent_product_category_rel_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "invent_products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "invent_unit_categories" table
CREATE TABLE "invent_unit_categories" (
  "id" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "org_id" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invent_unit_categories_name_org_id_ukey" UNIQUE ("name", "org_id")
);
-- Create "invent_units" table
CREATE TABLE "invent_units" (
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
  CONSTRAINT "invent_units_symbol_org_id_ukey" UNIQUE ("symbol", "org_id"),
  CONSTRAINT "invent_units_category_id_fkey" FOREIGN KEY ("category_id") REFERENCES "invent_unit_categories" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create "invent_variants" table
CREATE TABLE "invent_variants" (
  "id" character varying NOT NULL,
  "product_id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "sku" character varying NULL,
  "barcode" character varying NULL,
  "proposed_price" numeric NULL,
  "status" character varying NULL,
  "image_url" character varying NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invent_variants_sku_ukey" UNIQUE ("sku"),
  CONSTRAINT "invent_variants_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "invent_products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "invent_variant_attr_val_rel" table
CREATE TABLE "invent_variant_attr_val_rel" (
  "variant_id" character varying NOT NULL,
  "attribute_value_id" character varying NOT NULL,
  PRIMARY KEY ("variant_id", "attribute_value_id"),
  CONSTRAINT "invent_variant_attr_val_rel_attribute_value_id_fkey" FOREIGN KEY ("attribute_value_id") REFERENCES "invent_attribute_values" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "invent_variant_attr_val_rel_variant_id_fkey" FOREIGN KEY ("variant_id") REFERENCES "invent_variants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
