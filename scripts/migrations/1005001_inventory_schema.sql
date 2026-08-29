-- Create "inventory_storage_categories" table
CREATE TABLE "inventory_storage_categories" (
  "id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "max_weight" numeric NULL,
  "allow_new_item_policy" character varying NOT NULL,
  "description" jsonb NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_stor_cats_tid_code_org_id_ukey" UNIQUE ("code", "org_id")
);
-- Create "inventory_warehouses" table
CREATE TABLE "inventory_warehouses" (
  "id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "warehouse_role" character varying NOT NULL,
  "parent_warehouse_id" character varying NULL,
  "address" character varying NULL,
  "manager_user_id" character varying NULL,
  "incoming_flow" character varying NOT NULL,
  "outgoing_flow" character varying NOT NULL,
  "status" character varying NOT NULL,
  "notes" character varying NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_whs_tid_code_org_id_ukey" UNIQUE ("code", "org_id"),
  CONSTRAINT "inventory_warehouses_parent_warehouse_id_fkey" FOREIGN KEY ("parent_warehouse_id") REFERENCES "inventory_warehouses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "invty_whs_parent_warehouse_id_idx" to table: "inventory_warehouses"
CREATE INDEX "invty_whs_parent_warehouse_id_idx" ON "inventory_warehouses" ("parent_warehouse_id");
-- Create "inventory_locations" table
CREATE TABLE "inventory_locations" (
  "id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "location_usage" character varying NOT NULL,
  "purpose" character varying NULL,
  "warehouse_id" character varying NULL,
  "parent_location_id" character varying NULL,
  "complete_path" character varying NULL,
  "hierarchy_depth" integer NULL,
  "barcode" character varying NULL,
  "storage_category_id" character varying NULL,
  "removal_strategy" character varying NULL,
  "is_replenishment_destination" boolean NOT NULL,
  "is_system_generated" boolean NOT NULL,
  "status" character varying NOT NULL,
  "description" jsonb NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_locs_tid_code_org_id_ukey" UNIQUE ("code", "org_id"),
  CONSTRAINT "inventory_locations_parent_location_id_fkey" FOREIGN KEY ("parent_location_id") REFERENCES "inventory_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_locations_storage_category_id_fkey" FOREIGN KEY ("storage_category_id") REFERENCES "inventory_storage_categories" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_locations_warehouse_id_fkey" FOREIGN KEY ("warehouse_id") REFERENCES "inventory_warehouses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "invty_locs_barcode_idx" to table: "inventory_locations"
CREATE INDEX "invty_locs_barcode_idx" ON "inventory_locations" ("barcode");
-- Create index "invty_locs_parent_location_id_idx" to table: "inventory_locations"
CREATE INDEX "invty_locs_parent_location_id_idx" ON "inventory_locations" ("parent_location_id");
-- Create index "invty_locs_warehouse_id_idx" to table: "inventory_locations"
CREATE INDEX "invty_locs_warehouse_id_idx" ON "inventory_locations" ("warehouse_id");
-- Create "inventory_product_attributes" table
CREATE TABLE "inventory_product_attributes" (
  "id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "data_type" character varying NOT NULL,
  "variant_creation_mode" character varying NOT NULL,
  "display_type" character varying NULL,
  "sequence" integer NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_prod_attrs_tid_code_org_id_ukey" UNIQUE ("code", "org_id")
);
-- Create "inventory_product_attribute_values" table
CREATE TABLE "inventory_product_attribute_values" (
  "id" character varying NOT NULL,
  "attribute_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "sequence" integer NULL,
  "price_extra" numeric NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_prod_attr_vals_tid_attr_id_code_ukey" UNIQUE ("attribute_id", "code"),
  CONSTRAINT "inventory_product_attribute_values_attribute_id_fkey" FOREIGN KEY ("attribute_id") REFERENCES "inventory_product_attributes" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "inventory_brands" table
CREATE TABLE "inventory_brands" (
  "id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "logo_id" character varying NULL,
  "country_id" character varying NULL,
  "website" character varying NULL,
  "description" jsonb NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_brands_code_org_id_ukey" UNIQUE ("code", "org_id")
);
-- Create "inventory_product_categories" table
CREATE TABLE "inventory_product_categories" (
  "id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "parent_category_id" character varying NULL,
  "sequence" integer NULL,
  "description" jsonb NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_product_categories_code_org_id_ukey" UNIQUE ("code", "org_id"),
  CONSTRAINT "inventory_product_categories_parent_category_id_fkey" FOREIGN KEY ("parent_category_id") REFERENCES "inventory_product_categories" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "inventory_product_types" table
CREATE TABLE "inventory_product_types" (
  "id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "description" jsonb NULL,
  "supports_stock" boolean NULL,
  "supports_sale" boolean NULL,
  "supports_purchase" boolean NULL,
  "supports_manufacturing" boolean NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_product_types_code_ukey" UNIQUE ("code")
);
-- Create "inventory_product_templates" table
CREATE TABLE "inventory_product_templates" (
  "id" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "short_name" jsonb NULL,
  "product_type_id" character varying NOT NULL,
  "category_id" character varying NOT NULL,
  "brand_id" character varying NULL,
  "sale_ok" boolean NULL,
  "purchase_ok" boolean NULL,
  "description" jsonb NULL,
  "sales_description" jsonb NULL,
  "purchase_description" jsonb NULL,
  "default_image_id" character varying NULL,
  "base_sales_price" numeric NULL,
  "default_weight" numeric NULL,
  "default_length" numeric NULL,
  "default_width" numeric NULL,
  "default_height" numeric NULL,
  "status" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_product_templates_brand_id_fkey" FOREIGN KEY ("brand_id") REFERENCES "inventory_brands" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_product_templates_category_id_fkey" FOREIGN KEY ("category_id") REFERENCES "inventory_product_categories" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_product_templates_product_type_id_fkey" FOREIGN KEY ("product_type_id") REFERENCES "inventory_product_types" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "inventory_product_template_attributes" table
CREATE TABLE "inventory_product_template_attributes" (
  "id" character varying NOT NULL,
  "product_template_id" character varying NOT NULL,
  "attribute_id" character varying NOT NULL,
  "sequence" integer NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_prod_tpl_attrs_tid_ptpl_id_attr_id_ukey" UNIQUE ("product_template_id", "attribute_id"),
  CONSTRAINT "inventory_product_template_attributes_attribute_id_fkey" FOREIGN KEY ("attribute_id") REFERENCES "inventory_product_attributes" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_product_template_attributes_product_template_id_fkey" FOREIGN KEY ("product_template_id") REFERENCES "inventory_product_templates" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "inventory_product_template_attribute_values" table
CREATE TABLE "inventory_product_template_attribute_values" (
  "id" character varying NOT NULL,
  "template_attribute_id" character varying NOT NULL,
  "attribute_value_id" character varying NOT NULL,
  "sales_price_extra" numeric NULL,
  "sequence" integer NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_prod_tpl_attr_vals_tid_tattr_id_aval_id_ukey" UNIQUE ("template_attribute_id", "attribute_value_id"),
  CONSTRAINT "inventory_product_template_attribute_values_attribute_value_id_" FOREIGN KEY ("attribute_value_id") REFERENCES "inventory_product_attribute_values" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_product_template_attribute_values_template_attribute_" FOREIGN KEY ("template_attribute_id") REFERENCES "inventory_product_template_attributes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "inventory_product_variants" table
CREATE TABLE "inventory_product_variants" (
  "id" character varying NOT NULL,
  "product_template_id" character varying NOT NULL,
  "combination_key" character varying NOT NULL,
  "sku" character varying NULL,
  "primary_barcode" character varying NULL,
  "is_materialized" boolean NULL,
  "variant_image_id" character varying NULL,
  "cost" numeric NULL,
  "weight" numeric NULL,
  "length" numeric NULL,
  "width" numeric NULL,
  "height" numeric NULL,
  "status" character varying NOT NULL,
  "archive_source" character varying NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_prod_variants_tid_ptpl_id_comb_key_ukey" UNIQUE ("product_template_id", "combination_key"),
  CONSTRAINT "inventory_product_variants_product_template_id_fkey" FOREIGN KEY ("product_template_id") REFERENCES "inventory_product_templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "inventory_product_variant_attribute_values" table
CREATE TABLE "inventory_product_variant_attribute_values" (
  "id" character varying NOT NULL,
  "product_variant_id" character varying NOT NULL,
  "template_attribute_value_id" character varying NOT NULL,
  "sales_price_extra" numeric NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_prod_var_attr_vals_tid_pvar_id_tav_id_ukey" UNIQUE ("product_variant_id", "template_attribute_value_id"),
  CONSTRAINT "inventory_product_variant_attribute_values_product_variant_id_f" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "inventory_product_variant_attribute_values_template_attribute_v" FOREIGN KEY ("template_attribute_value_id") REFERENCES "inventory_product_template_attribute_values" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "inventory_putaway_rules" table
CREATE TABLE "inventory_putaway_rules" (
  "id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "warehouse_id" character varying NOT NULL,
  "source_location_id" character varying NOT NULL,
  "destination_location_id" character varying NOT NULL,
  "storage_category_id" character varying NULL,
  "product_id" character varying NULL,
  "product_category_id" character varying NULL,
  "package_type_id" character varying NULL,
  "priority" integer NOT NULL,
  "sublocation_strategy" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_putaway_tid_code_org_id_ukey" UNIQUE ("code", "org_id"),
  CONSTRAINT "inventory_putaway_rules_destination_location_id_fkey" FOREIGN KEY ("destination_location_id") REFERENCES "inventory_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_putaway_rules_source_location_id_fkey" FOREIGN KEY ("source_location_id") REFERENCES "inventory_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_putaway_rules_storage_category_id_fkey" FOREIGN KEY ("storage_category_id") REFERENCES "inventory_storage_categories" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_putaway_rules_warehouse_id_fkey" FOREIGN KEY ("warehouse_id") REFERENCES "inventory_warehouses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "invty_putaway_wh_src_priority_idx" to table: "inventory_putaway_rules"
CREATE INDEX "invty_putaway_wh_src_priority_idx" ON "inventory_putaway_rules" ("warehouse_id", "source_location_id", "priority");
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
  CONSTRAINT "inventory_stock_operation_types_default_destination_location_id" FOREIGN KEY ("default_destination_location_id") REFERENCES "inventory_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_operation_types_default_source_location_id_fkey" FOREIGN KEY ("default_source_location_id") REFERENCES "inventory_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "inventory_stock_transfers" table
CREATE TABLE "inventory_stock_transfers" (
  "id" character varying NOT NULL,
  "transfer_number" character varying NOT NULL,
  "operation_type_id" character varying NOT NULL,
  "operation_code" character varying NOT NULL,
  "origin_reference" character varying NULL,
  "source_location_id" character varying NOT NULL,
  "destination_location_id" character varying NOT NULL,
  "status" character varying NOT NULL,
  "priority" integer NULL,
  "reservation_method" character varying NOT NULL,
  "backorder_policy" character varying NOT NULL,
  "shipping_policy" character varying NOT NULL,
  "scheduled_at" timestamptz NULL,
  "deadline_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
  "backorder_of_id" character varying NULL,
  "return_of_id" character varying NULL,
  "chain_group_id" character varying NULL,
  "idempotency_key" character varying NULL,
  "note" character varying NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_stock_trfs_tid_number_org_id_ukey" UNIQUE ("transfer_number", "org_id"),
  CONSTRAINT "inventory_stock_transfers_backorder_of_id_fkey" FOREIGN KEY ("backorder_of_id") REFERENCES "inventory_stock_transfers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_transfers_destination_location_id_fkey" FOREIGN KEY ("destination_location_id") REFERENCES "inventory_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_transfers_operation_type_id_fkey" FOREIGN KEY ("operation_type_id") REFERENCES "inventory_stock_operation_types" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_transfers_source_location_id_fkey" FOREIGN KEY ("source_location_id") REFERENCES "inventory_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "invty_stock_trfs_backorder_of_id_idx" to table: "inventory_stock_transfers"
CREATE INDEX "invty_stock_trfs_backorder_of_id_idx" ON "inventory_stock_transfers" ("backorder_of_id");
-- Create index "invty_stock_trfs_chain_group_id_idx" to table: "inventory_stock_transfers"
CREATE INDEX "invty_stock_trfs_chain_group_id_idx" ON "inventory_stock_transfers" ("chain_group_id");
-- Create index "invty_stock_trfs_op_type_id_status_idx" to table: "inventory_stock_transfers"
CREATE INDEX "invty_stock_trfs_op_type_id_status_idx" ON "inventory_stock_transfers" ("operation_type_id", "status");
-- Create index "invty_stock_trfs_org_id_idem_key_idx" to table: "inventory_stock_transfers"
CREATE INDEX "invty_stock_trfs_org_id_idem_key_idx" ON "inventory_stock_transfers" ("org_id", "idempotency_key");
-- Create index "invty_stock_trfs_status_sched_at_idx" to table: "inventory_stock_transfers"
CREATE INDEX "invty_stock_trfs_status_sched_at_idx" ON "inventory_stock_transfers" ("status", "scheduled_at");
-- Create "inventory_stock_moves" table
CREATE TABLE "inventory_stock_moves" (
  "id" character varying NOT NULL,
  "transfer_id" character varying NULL,
  "sequence" integer NULL,
  "product_variant_id" character varying NOT NULL,
  "uom_id" character varying NULL,
  "demand_quantity" numeric NOT NULL,
  "base_demand_quantity" numeric NOT NULL,
  "source_location_id" character varying NOT NULL,
  "destination_location_id" character varying NOT NULL,
  "final_location_id" character varying NULL,
  "status" character varying NOT NULL,
  "priority" integer NULL,
  "scheduled_at" timestamptz NULL,
  "deadline_at" timestamptz NULL,
  "reservation_date" timestamptz NULL,
  "picked" boolean NULL,
  "origin_move_id" character varying NULL,
  "is_inventory_adjustment" boolean NULL,
  "scrap_id" character varying NULL,
  "valuation_value" numeric NULL,
  "remaining_quantity" numeric NULL,
  "remaining_value" numeric NULL,
  "currency_id" character varying NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_stock_moves_destination_location_id_fkey" FOREIGN KEY ("destination_location_id") REFERENCES "inventory_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_moves_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_moves_source_location_id_fkey" FOREIGN KEY ("source_location_id") REFERENCES "inventory_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_moves_transfer_id_fkey" FOREIGN KEY ("transfer_id") REFERENCES "inventory_stock_transfers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "invty_stock_moves_origin_move_id_idx" to table: "inventory_stock_moves"
CREATE INDEX "invty_stock_moves_origin_move_id_idx" ON "inventory_stock_moves" ("origin_move_id");
-- Create index "invty_stock_moves_pvar_id_status_idx" to table: "inventory_stock_moves"
CREATE INDEX "invty_stock_moves_pvar_id_status_idx" ON "inventory_stock_moves" ("product_variant_id", "status");
-- Create index "invty_stock_moves_status_sched_at_idx" to table: "inventory_stock_moves"
CREATE INDEX "invty_stock_moves_status_sched_at_idx" ON "inventory_stock_moves" ("status", "scheduled_at");
-- Create index "invty_stock_moves_trf_id_sequence_idx" to table: "inventory_stock_moves"
CREATE INDEX "invty_stock_moves_trf_id_sequence_idx" ON "inventory_stock_moves" ("transfer_id", "sequence");
-- Create "inventory_stock_move_dependencies" table
CREATE TABLE "inventory_stock_move_dependencies" (
  "id" character varying NOT NULL,
  "predecessor_move_id" character varying NOT NULL,
  "successor_move_id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_stock_mvdeps_tid_pred_succ_ukey" UNIQUE ("predecessor_move_id", "successor_move_id"),
  CONSTRAINT "inventory_stock_move_dependencies_predecessor_move_id_fkey" FOREIGN KEY ("predecessor_move_id") REFERENCES "inventory_stock_moves" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_move_dependencies_successor_move_id_fkey" FOREIGN KEY ("successor_move_id") REFERENCES "inventory_stock_moves" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "invty_stock_mvdeps_succ_move_id_idx" to table: "inventory_stock_move_dependencies"
CREATE INDEX "invty_stock_mvdeps_succ_move_id_idx" ON "inventory_stock_move_dependencies" ("successor_move_id");
-- Create "inventory_stock_move_lines" table
CREATE TABLE "inventory_stock_move_lines" (
  "id" character varying NOT NULL,
  "move_id" character varying NOT NULL,
  "transfer_id" character varying NULL,
  "product_variant_id" character varying NOT NULL,
  "uom_id" character varying NULL,
  "quantity" numeric NOT NULL,
  "base_quantity" numeric NOT NULL,
  "source_location_id" character varying NOT NULL,
  "destination_location_id" character varying NOT NULL,
  "lot_ref" character varying NOT NULL,
  "package_ref" character varying NOT NULL,
  "result_package_ref" character varying NOT NULL,
  "owner_ref" character varying NOT NULL,
  "picked" boolean NULL,
  "operation_at" timestamptz NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_stock_move_lines_destination_location_id_fkey" FOREIGN KEY ("destination_location_id") REFERENCES "inventory_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_move_lines_move_id_fkey" FOREIGN KEY ("move_id") REFERENCES "inventory_stock_moves" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_move_lines_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_move_lines_source_location_id_fkey" FOREIGN KEY ("source_location_id") REFERENCES "inventory_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_move_lines_transfer_id_fkey" FOREIGN KEY ("transfer_id") REFERENCES "inventory_stock_transfers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "invty_stock_mvlines_move_id_idx" to table: "inventory_stock_move_lines"
CREATE INDEX "invty_stock_mvlines_move_id_idx" ON "inventory_stock_move_lines" ("move_id");
-- Create index "invty_stock_mvlines_pvar_id_src_loc_idx" to table: "inventory_stock_move_lines"
CREATE INDEX "invty_stock_mvlines_pvar_id_src_loc_idx" ON "inventory_stock_move_lines" ("product_variant_id", "source_location_id");
-- Create index "invty_stock_mvlines_trf_id_idx" to table: "inventory_stock_move_lines"
CREATE INDEX "invty_stock_mvlines_trf_id_idx" ON "inventory_stock_move_lines" ("transfer_id");
-- Create "inventory_stock_product_configs" table
CREATE TABLE "inventory_stock_product_configs" (
  "id" character varying NOT NULL,
  "product_template_id" character varying NOT NULL,
  "inventory_uom_id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_stk_prod_cfgs_tid_ptpl_id_ukey" UNIQUE ("product_template_id", "org_id"),
  CONSTRAINT "inventory_stock_product_configs_product_template_id_fkey" FOREIGN KEY ("product_template_id") REFERENCES "inventory_product_templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "inventory_stock_quants" table
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
  CONSTRAINT "inventory_stock_quants_location_id_fkey" FOREIGN KEY ("location_id") REFERENCES "inventory_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_quants_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "invty_stock_quants_loc_id_pvar_id_idx" to table: "inventory_stock_quants"
CREATE INDEX "invty_stock_quants_loc_id_pvar_id_idx" ON "inventory_stock_quants" ("location_id", "product_variant_id");
-- Create index "invty_stock_quants_next_count_date_idx" to table: "inventory_stock_quants"
CREATE INDEX "invty_stock_quants_next_count_date_idx" ON "inventory_stock_quants" ("next_count_date");
-- Create index "invty_stock_quants_pvar_id_loc_id_idx" to table: "inventory_stock_quants"
CREATE INDEX "invty_stock_quants_pvar_id_loc_id_idx" ON "inventory_stock_quants" ("product_variant_id", "location_id");
-- Create "inventory_stock_scraps" table
CREATE TABLE "inventory_stock_scraps" (
  "id" character varying NOT NULL,
  "scrap_number" character varying NOT NULL,
  "origin_reference" character varying NULL,
  "transfer_id" character varying NULL,
  "product_variant_id" character varying NOT NULL,
  "base_uom_id" character varying NULL,
  "lot_ref" character varying NULL,
  "package_ref" character varying NULL,
  "owner_ref" character varying NULL,
  "source_location_id" character varying NOT NULL,
  "scrap_location_id" character varying NOT NULL,
  "quantity" numeric NOT NULL,
  "reason_code" character varying NULL,
  "reason" character varying NULL,
  "status" character varying NOT NULL,
  "move_id" character varying NULL,
  "completed_at" timestamptz NULL,
  "note" character varying NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_stock_scraps_number_org_id_ukey" UNIQUE ("scrap_number", "org_id"),
  CONSTRAINT "inventory_stock_scraps_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_scraps_scrap_location_id_fkey" FOREIGN KEY ("scrap_location_id") REFERENCES "inventory_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_scraps_source_location_id_fkey" FOREIGN KEY ("source_location_id") REFERENCES "inventory_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_scraps_transfer_id_fkey" FOREIGN KEY ("transfer_id") REFERENCES "inventory_stock_transfers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "invty_stock_scraps_pvar_id_idx" to table: "inventory_stock_scraps"
CREATE INDEX "invty_stock_scraps_pvar_id_idx" ON "inventory_stock_scraps" ("product_variant_id");
-- Create index "invty_stock_scraps_status_idx" to table: "inventory_stock_scraps"
CREATE INDEX "invty_stock_scraps_status_idx" ON "inventory_stock_scraps" ("status");
-- Create index "invty_stock_scraps_trf_id_idx" to table: "inventory_stock_scraps"
CREATE INDEX "invty_stock_scraps_trf_id_idx" ON "inventory_stock_scraps" ("transfer_id");
-- Create "inventory_warehouse_supply_relations" table
CREATE TABLE "inventory_warehouse_supply_relations" (
  "id" character varying NOT NULL,
  "source_warehouse_id" character varying NOT NULL,
  "destination_warehouse_id" character varying NOT NULL,
  "priority" integer NOT NULL,
  "is_default" boolean NOT NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_wh_supply_tid_src_dest_ukey" UNIQUE ("source_warehouse_id", "destination_warehouse_id"),
  CONSTRAINT "inventory_warehouse_supply_relations_destination_warehouse_id_f" FOREIGN KEY ("destination_warehouse_id") REFERENCES "inventory_warehouses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_warehouse_supply_relations_source_warehouse_id_fkey" FOREIGN KEY ("source_warehouse_id") REFERENCES "inventory_warehouses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "invty_wh_supply_dest_priority_idx" to table: "inventory_warehouse_supply_relations"
CREATE INDEX "invty_wh_supply_dest_priority_idx" ON "inventory_warehouse_supply_relations" ("destination_warehouse_id", "priority");
