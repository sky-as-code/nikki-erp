-- Create "inventory_products" table
CREATE TABLE "inventory_products" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "description" jsonb NULL,
  "thumbnail_url" character varying NULL,
  "unit_id" character varying NULL,
  "default_variant_id" character varying NULL,
  "tag_ids" character varying NULL,
  "is_archived" boolean NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_products_name_ukey" UNIQUE ("name")
);
-- Create "inventory_attribute_groups" table
CREATE TABLE "inventory_attribute_groups" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "index" bigint NULL,
  "product_id" character varying NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_attribute_groups_name_product_id_ukey" UNIQUE ("name", "product_id"),
  CONSTRAINT "inventory_attribute_groups_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "inventory_products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "inventory_attributes" table
CREATE TABLE "inventory_attributes" (
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
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_attributes_code_name_product_id_ukey" UNIQUE ("code_name", "product_id"),
  CONSTRAINT "inventory_attributes_attribute_group_id_fkey" FOREIGN KEY ("attribute_group_id") REFERENCES "inventory_attribute_groups" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "inventory_attributes_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "inventory_products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "inventory_attribute_values" table
CREATE TABLE "inventory_attribute_values" (
  "id" character varying NOT NULL,
  "attribute_id" character varying NOT NULL,
  "value_text" jsonb NULL,
  "value_decimal" numeric NULL,
  "value_integer" bigint NULL,
  "value_bool" boolean NULL,
  "value_ref" character varying NULL,
  "variant_count" bigint NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_attribute_values_attribute_id_fkey" FOREIGN KEY ("attribute_id") REFERENCES "inventory_attributes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "inventory_product_categories" table
CREATE TABLE "inventory_product_categories" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_product_categories_name_org_id_ukey" UNIQUE ("name", "org_id")
);
-- Create "inventory_product_category_rel" table
CREATE TABLE "inventory_product_category_rel" (
  "product_id" character varying NOT NULL,
  "product_category_id" character varying NOT NULL,
  PRIMARY KEY ("product_id", "product_category_id"),
  CONSTRAINT "inventory_product_category_rel_product_category_id_fkey" FOREIGN KEY ("product_category_id") REFERENCES "inventory_product_categories" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "inventory_product_category_rel_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "inventory_products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "inventory_variants" table
CREATE TABLE "inventory_variants" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "product_id" character varying NOT NULL,
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
  CONSTRAINT "inventory_variants_sku_ukey" UNIQUE ("sku"),
  CONSTRAINT "inventory_variants_product_id_fkey" FOREIGN KEY ("product_id") REFERENCES "inventory_products" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "inventory_variant_attr_val_rel" table
CREATE TABLE "inventory_variant_attr_val_rel" (
  "variant_id" character varying NOT NULL,
  "attribute_value_id" character varying NOT NULL,
  PRIMARY KEY ("variant_id", "attribute_value_id"),
  CONSTRAINT "inventory_variant_attr_val_rel_attribute_value_id_fkey" FOREIGN KEY ("attribute_value_id") REFERENCES "inventory_attribute_values" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "inventory_variant_attr_val_rel_variant_id_fkey" FOREIGN KEY ("variant_id") REFERENCES "inventory_variants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
