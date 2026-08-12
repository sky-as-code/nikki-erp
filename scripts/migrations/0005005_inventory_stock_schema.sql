-- Stock Management balance core (stock BR §4.2.1, §4.2.2, §7).
--
-- Generated with `make gen-sql module=inventory` and split out of the full inventory DDL. It is a
-- new file rather than an edit to 0005001: that migration is already hashed in atlas.sum and
-- applied, so changing it would invalidate every database that has run it.
--
-- Quantities are numeric, never floating point: a stock balance is the running total of many
-- movements, and binary floating point cannot represent the decimal quantities a business counts
-- in without drift accumulating over those additions (BR §7.3).

-- Create "inventory_stock_locations" table
CREATE TABLE "inventory_stock_locations" (
  "id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "location_type" character varying NOT NULL,
  "parent_location_id" character varying NULL,
  "description" jsonb NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_stock_locs_tid_code_org_id_ukey" UNIQUE ("code", "org_id"),
  CONSTRAINT "inventory_stock_locations_parent_location_id_fkey" FOREIGN KEY ("parent_location_id") REFERENCES "inventory_stock_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);

-- Create "inventory_stock_operation_types" table
CREATE TABLE "inventory_stock_operation_types" (
  "id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "operation_code" character varying NOT NULL,
  "reservation_method" character varying NOT NULL,
  "reserve_before_days" integer NULL,
  "backorder_policy" character varying NOT NULL,
  "shipping_policy" character varying NULL,
  "default_source_location_id" character varying NULL,
  "default_destination_location_id" character varying NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_stock_op_types_tid_code_org_id_ukey" UNIQUE ("code", "org_id"),
  CONSTRAINT "inventory_stock_operation_types_default_source_location_id_fkey" FOREIGN KEY ("default_source_location_id") REFERENCES "inventory_stock_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_operation_types_default_destination_location_id_fkey" FOREIGN KEY ("default_destination_location_id") REFERENCES "inventory_stock_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);

-- Create "inventory_stock_quants" table.
--
-- lot_ref, package_ref and owner_ref are NOT NULL and default to the empty string rather than
-- being nullable: they are part of the unique key, and SQL does not constrain NULL rows against
-- each other, so a nullable column here would let the same balance be created twice.
CREATE TABLE "inventory_stock_quants" (
  "id" character varying NOT NULL,
  "product_variant_id" character varying NOT NULL,
  "location_id" character varying NOT NULL,
  "lot_ref" character varying NOT NULL,
  "package_ref" character varying NOT NULL,
  "owner_ref" character varying NOT NULL,
  "base_uom_id" character varying NULL,
  "on_hand_quantity" numeric NULL,
  "reserved_quantity" numeric NULL,
  "incoming_date" timestamptz NULL,
  "counted_quantity" numeric NULL,
  "count_quantity_set" boolean NULL,
  "count_snapshot_quantity" numeric NULL,
  "count_reason_code" character varying NULL,
  "count_reason_text" character varying NULL,
  "next_count_date" date NULL,
  "last_count_date" date NULL,
  "count_assigned_user_id" character varying NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_stock_quants_tid_pvar_loc_lot_pkg_own_ukey" UNIQUE ("product_variant_id", "location_id", "lot_ref", "package_ref", "owner_ref", "org_id"),
  CONSTRAINT "inventory_stock_quants_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_quants_location_id_fkey" FOREIGN KEY ("location_id") REFERENCES "inventory_stock_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);

-- Balance lookups go both ways: "what is in this location" and "where is this product".
CREATE INDEX "invty_stock_quants_pvar_id_loc_id_idx" ON "inventory_stock_quants" ("product_variant_id", "location_id");
CREATE INDEX "invty_stock_quants_loc_id_pvar_id_idx" ON "inventory_stock_quants" ("location_id", "product_variant_id");
-- Filtering on this is what produces a cycle-count worklist.
CREATE INDEX "invty_stock_quants_next_count_date_idx" ON "inventory_stock_quants" ("next_count_date");
