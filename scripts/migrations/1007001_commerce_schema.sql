-- Create "inventory_integration_outbox" table
CREATE TABLE "inventory_integration_outbox" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "event_id" character varying NOT NULL,
  "aggregate_id" character varying NOT NULL,
  "event_type" character varying NOT NULL,
  "schema_version" character varying NOT NULL,
  "payload" jsonb NOT NULL,
  "occurred_at" timestamptz NOT NULL,
  "published_at" timestamptz NULL,
  "attempt_count" integer NULL,
  "last_error" character varying NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_outbox_eventid_ukey" UNIQUE ("event_id")
);
-- Create index "invty_outbox_aggregate_idx" to table: "inventory_integration_outbox"
CREATE INDEX "invty_outbox_aggregate_idx" ON "inventory_integration_outbox" ("aggregate_id");
-- Create index "invty_outbox_occurred_idx" to table: "inventory_integration_outbox"
CREATE INDEX "invty_outbox_occurred_idx" ON "inventory_integration_outbox" ("occurred_at");
-- Create index "invty_outbox_published_idx" to table: "inventory_integration_outbox"
CREATE INDEX "invty_outbox_published_idx" ON "inventory_integration_outbox" ("published_at");
-- Create index "invty_outbox_type_idx" to table: "inventory_integration_outbox"
CREATE INDEX "invty_outbox_type_idx" ON "inventory_integration_outbox" ("event_type");
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
-- Create "purchase_audit_events" table
CREATE TABLE "purchase_audit_events" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "entity_type" character varying NOT NULL,
  "entity_id" character varying NOT NULL,
  "action" character varying NOT NULL,
  "actor_id" character varying NULL,
  "from_status" character varying NULL,
  "to_status" character varying NULL,
  "reason" character varying NULL,
  "metadata" jsonb NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "purch_audit_evts_tid_entity_idx" to table: "purchase_audit_events"
CREATE INDEX "purch_audit_evts_tid_entity_idx" ON "purchase_audit_events" ("entity_type", "entity_id");
-- Create index "purch_audit_evts_tid_org_id_idx" to table: "purchase_audit_events"
CREATE INDEX "purch_audit_evts_tid_org_id_idx" ON "purchase_audit_events" ("org_id");
-- Create "purchase_configurations" table
CREATE TABLE "purchase_configurations" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "approval_mode" character varying NOT NULL,
  "approval_threshold" numeric NULL,
  "po_modification_policy" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "purch_configs_tid_org_id_ukey" UNIQUE ("org_id")
);
-- Create "purchase_sourcing_groups" table
CREATE TABLE "purchase_sourcing_groups" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "purch_srcgroups_tid_org_id_idx" to table: "purchase_sourcing_groups"
CREATE INDEX "purch_srcgroups_tid_org_id_idx" ON "purchase_sourcing_groups" ("org_id");
-- Create "sales_billing_instructions" table
CREATE TABLE "sales_billing_instructions" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_order_id" character varying NOT NULL,
  "bill_to_party_id" character varying NULL,
  "tax_id" character varying NULL,
  "legal_name" character varying NULL,
  "billing_address" character varying NULL,
  "billing_email" character varying NULL,
  "status" character varying NOT NULL,
  "source" character varying NOT NULL,
  "fetch_latest_party_details" boolean NULL,
  "snapshot_at" timestamptz NULL,
  "locked_at" timestamptz NULL,
  "submitted_at" timestamptz NULL,
  "issued_at" timestamptz NULL,
  "einvoice_reference" character varying NULL,
  "last_error_code" character varying NULL,
  "last_error_message" character varying NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "sales_billing_instrs_tid_billto_idx" to table: "sales_billing_instructions"
CREATE INDEX "sales_billing_instrs_tid_billto_idx" ON "sales_billing_instructions" ("bill_to_party_id");
-- Create index "sales_billing_instrs_tid_einvref_idx" to table: "sales_billing_instructions"
CREATE INDEX "sales_billing_instrs_tid_einvref_idx" ON "sales_billing_instructions" ("einvoice_reference");
-- Create index "sales_billing_instrs_tid_order_idx" to table: "sales_billing_instructions"
CREATE INDEX "sales_billing_instrs_tid_order_idx" ON "sales_billing_instructions" ("sales_order_id");
-- Create index "sales_billing_instrs_tid_status_idx" to table: "sales_billing_instructions"
CREATE INDEX "sales_billing_instrs_tid_status_idx" ON "sales_billing_instructions" ("status");
-- Create "sales_billing_issuance_attempts" table
CREATE TABLE "sales_billing_issuance_attempts" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "billing_instruction_id" character varying NOT NULL,
  "attempt_no" integer NOT NULL,
  "status" character varying NOT NULL,
  "started_at" timestamptz NOT NULL,
  "completed_at" timestamptz NULL,
  "provider_request_id" character varying NULL,
  "provider_invoice_reference" character varying NULL,
  "error_code" character varying NULL,
  "error_message" character varying NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "sales_billing_attempts_tid_instr_idx" to table: "sales_billing_issuance_attempts"
CREATE INDEX "sales_billing_attempts_tid_instr_idx" ON "sales_billing_issuance_attempts" ("billing_instruction_id");
-- Create index "sales_billing_attempts_tid_provreq_idx" to table: "sales_billing_issuance_attempts"
CREATE INDEX "sales_billing_attempts_tid_provreq_idx" ON "sales_billing_issuance_attempts" ("provider_request_id");
-- Create index "sales_billing_attempts_tid_status_idx" to table: "sales_billing_issuance_attempts"
CREATE INDEX "sales_billing_attempts_tid_status_idx" ON "sales_billing_issuance_attempts" ("status");
-- Create "sales_channel_payment_rel" table
CREATE TABLE "sales_channel_payment_rel" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_channel_id" character varying NOT NULL,
  "payment_method_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_channel_payment_rel_tid_chan_method_ukey" UNIQUE ("sales_channel_id", "payment_method_id")
);
-- Create "sales_integration_outbox" table
CREATE TABLE "sales_integration_outbox" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "event_id" character varying NOT NULL,
  "aggregate_id" character varying NOT NULL,
  "event_type" character varying NOT NULL,
  "schema_version" character varying NOT NULL,
  "payload" jsonb NOT NULL,
  "occurred_at" timestamptz NOT NULL,
  "published_at" timestamptz NULL,
  "attempt_count" integer NULL,
  "last_error" character varying NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_outbox_tid_eventid_ukey" UNIQUE ("event_id")
);
-- Create index "sales_outbox_tid_aggregate_idx" to table: "sales_integration_outbox"
CREATE INDEX "sales_outbox_tid_aggregate_idx" ON "sales_integration_outbox" ("aggregate_id");
-- Create index "sales_outbox_tid_occurred_idx" to table: "sales_integration_outbox"
CREATE INDEX "sales_outbox_tid_occurred_idx" ON "sales_integration_outbox" ("occurred_at");
-- Create index "sales_outbox_tid_published_idx" to table: "sales_integration_outbox"
CREATE INDEX "sales_outbox_tid_published_idx" ON "sales_integration_outbox" ("published_at");
-- Create index "sales_outbox_tid_type_idx" to table: "sales_integration_outbox"
CREATE INDEX "sales_outbox_tid_type_idx" ON "sales_integration_outbox" ("event_type");
-- Create "sales_order_events" table
CREATE TABLE "sales_order_events" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_order_id" character varying NOT NULL,
  "entity_type" character varying NOT NULL,
  "entity_id" character varying NOT NULL,
  "action" character varying NOT NULL,
  "actor_id" character varying NULL,
  "from_status" character varying NULL,
  "to_status" character varying NULL,
  "reason" character varying NULL,
  "metadata" jsonb NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "sales_order_evts_tid_actor_idx" to table: "sales_order_events"
CREATE INDEX "sales_order_evts_tid_actor_idx" ON "sales_order_events" ("actor_id");
-- Create index "sales_order_evts_tid_entity_idx" to table: "sales_order_events"
CREATE INDEX "sales_order_evts_tid_entity_idx" ON "sales_order_events" ("entity_type", "entity_id");
-- Create index "sales_order_evts_tid_order_idx" to table: "sales_order_events"
CREATE INDEX "sales_order_evts_tid_order_idx" ON "sales_order_events" ("sales_order_id");
-- Create "sales_promotion_compatibilities" table
CREATE TABLE "sales_promotion_compatibilities" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "program_a_id" character varying NOT NULL,
  "program_b_id" character varying NOT NULL,
  "compatibility" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_promo_compat_tid_pair_ukey" UNIQUE ("program_a_id", "program_b_id")
);
-- Create index "sales_promo_compat_tid_b_idx" to table: "sales_promotion_compatibilities"
CREATE INDEX "sales_promo_compat_tid_b_idx" ON "sales_promotion_compatibilities" ("program_b_id");
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
  CONSTRAINT "inventory_locations_parent_location_id_fkey" FOREIGN KEY ("parent_location_id") REFERENCES "inventory_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_locations_storage_category_id_fkey" FOREIGN KEY ("storage_category_id") REFERENCES "inventory_storage_categories" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_locations_warehouse_id_fkey" FOREIGN KEY ("warehouse_id") REFERENCES "inventory_warehouses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "invty_locs_barcode_idx" to table: "inventory_locations"
CREATE INDEX "invty_locs_barcode_idx" ON "inventory_locations" ("barcode");
-- Create index "invty_locs_parent_location_id_idx" to table: "inventory_locations"
CREATE INDEX "invty_locs_parent_location_id_idx" ON "inventory_locations" ("parent_location_id");
-- Create index "invty_locs_tid_code_org_wh_id_ukey_notnull" to table: "inventory_locations"
CREATE UNIQUE INDEX "invty_locs_tid_code_org_wh_id_ukey_notnull" ON "inventory_locations" ("code", "org_id", "warehouse_id") WHERE (warehouse_id IS NOT NULL);
-- Create index "invty_locs_tid_code_org_wh_id_ukey_null" to table: "inventory_locations"
CREATE UNIQUE INDEX "invty_locs_tid_code_org_wh_id_ukey_null" ON "inventory_locations" ("code", "org_id") WHERE (warehouse_id IS NULL);
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
  "color_hex" character varying NULL,
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
  "name" jsonb NOT NULL,
  "parent_category_id" character varying NOT NULL,
  "is_root" boolean NULL,
  "canonical_name" jsonb NULL,
  "display_order" integer NULL,
  "description" jsonb NULL,
  "source_system" character varying NOT NULL,
  "external_id" character varying NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_product_categories_parent_category_id_fkey" FOREIGN KEY ("parent_category_id") REFERENCES "inventory_product_categories" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "invty_prod_cats_tid_org_root_ukey" to table: "inventory_product_categories"
CREATE UNIQUE INDEX "invty_prod_cats_tid_org_root_ukey" ON "inventory_product_categories" ("org_id", "is_root") WHERE (is_root IS NOT NULL);
-- Create index "invty_prod_cats_tid_src_sys_ext_id_ukey" to table: "inventory_product_categories"
CREATE UNIQUE INDEX "invty_prod_cats_tid_src_sys_ext_id_ukey" ON "inventory_product_categories" ("source_system", "external_id") WHERE (external_id IS NOT NULL);
-- Create "inventory_product_templates" table
CREATE TABLE "inventory_product_templates" (
  "id" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "short_name" jsonb NULL,
  "product_type" character varying NOT NULL,
  "is_combo" boolean NULL,
  "category_id" character varying NOT NULL,
  "brand_id" character varying NULL,
  "uom_id" character varying NULL,
  "sale_ok" boolean NULL,
  "purchase_ok" boolean NULL,
  "description" jsonb NULL,
  "sales_description" jsonb NULL,
  "purchase_description" jsonb NULL,
  "preview_image_file_id" character varying NULL,
  "base_sales_price" numeric NULL,
  "default_weight" numeric NULL,
  "default_length" numeric NULL,
  "default_width" numeric NULL,
  "default_height" numeric NULL,
  "status" character varying NOT NULL,
  "source_system" character varying NOT NULL,
  "external_id" character varying NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_product_templates_brand_id_fkey" FOREIGN KEY ("brand_id") REFERENCES "inventory_brands" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_product_templates_category_id_fkey" FOREIGN KEY ("category_id") REFERENCES "inventory_product_categories" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_product_templates_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "essential_uoms" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "invty_prod_tpls_tid_src_sys_ext_id_ukey" to table: "inventory_product_templates"
CREATE UNIQUE INDEX "invty_prod_tpls_tid_src_sys_ext_id_ukey" ON "inventory_product_templates" ("source_system", "external_id") WHERE (external_id IS NOT NULL);
-- Create "inventory_product_template_attributes" table
CREATE TABLE "inventory_product_template_attributes" (
  "id" character varying NOT NULL,
  "product_template_id" character varying NOT NULL,
  "attribute_id" character varying NOT NULL,
  "sequence" integer NULL,
  "is_enabled" boolean NOT NULL,
  "application_revision" integer NOT NULL,
  "variant_creation_snapshot" character varying NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_prod_tpl_attrs_tid_ptpl_id_attr_id_rev_ukey" UNIQUE ("product_template_id", "attribute_id", "application_revision"),
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
  "is_selected" boolean NOT NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_prod_tpl_attr_vals_tid_tattr_id_aval_id_ukey" UNIQUE ("template_attribute_id", "attribute_value_id"),
  CONSTRAINT "inventory_product_template_attribute_values_attribute_value_id_" FOREIGN KEY ("attribute_value_id") REFERENCES "inventory_product_attribute_values" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_product_template_attribute_values_template_attribute_" FOREIGN KEY ("template_attribute_id") REFERENCES "inventory_product_template_attributes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "inventory_product_template_combination_exclusions" table
CREATE TABLE "inventory_product_template_combination_exclusions" (
  "id" character varying NOT NULL,
  "product_template_id" character varying NOT NULL,
  "left_template_attribute_value_id" character varying NOT NULL,
  "right_template_attribute_value_id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_prod_tpl_comb_excls_tid_ptpl_id_left_right_ukey" UNIQUE ("product_template_id", "left_template_attribute_value_id", "right_template_attribute_value_id"),
  CONSTRAINT "inventory_product_template_combination_exclusions_left_template" FOREIGN KEY ("left_template_attribute_value_id") REFERENCES "inventory_product_template_attribute_values" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_product_template_combination_exclusions_product_templ" FOREIGN KEY ("product_template_id") REFERENCES "inventory_product_templates" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "inventory_product_template_combination_exclusions_right_templat" FOREIGN KEY ("right_template_attribute_value_id") REFERENCES "inventory_product_template_attribute_values" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "inventory_product_variants" table
CREATE TABLE "inventory_product_variants" (
  "id" character varying NOT NULL,
  "product_template_id" character varying NOT NULL,
  "variant_kind" character varying NOT NULL,
  "base_variant_id" character varying NULL,
  "combination_key" character varying NOT NULL,
  "sku" character varying NULL,
  "primary_barcode" character varying NULL,
  "preview_image_file_id" character varying NULL,
  "cost" numeric NULL,
  "weight" numeric NULL,
  "length" numeric NULL,
  "width" numeric NULL,
  "height" numeric NULL,
  "status" character varying NOT NULL,
  "archive_source" character varying NULL,
  "source_system" character varying NOT NULL,
  "external_id" character varying NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_prod_variants_tid_ptpl_id_kind_comb_key_ukey" UNIQUE ("product_template_id", "variant_kind", "combination_key"),
  CONSTRAINT "inventory_product_variants_base_variant_id_fkey" FOREIGN KEY ("base_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "inventory_product_variants_product_template_id_fkey" FOREIGN KEY ("product_template_id") REFERENCES "inventory_product_templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "invty_prod_variants_tid_org_barcode_ukey" to table: "inventory_product_variants"
CREATE UNIQUE INDEX "invty_prod_variants_tid_org_barcode_ukey" ON "inventory_product_variants" ("org_id", "primary_barcode") WHERE (primary_barcode IS NOT NULL);
-- Create index "invty_prod_variants_tid_org_sku_ukey" to table: "inventory_product_variants"
CREATE UNIQUE INDEX "invty_prod_variants_tid_org_sku_ukey" ON "inventory_product_variants" ("org_id", "sku") WHERE (sku IS NOT NULL);
-- Create index "invty_prod_variants_tid_src_sys_ext_id_ukey" to table: "inventory_product_variants"
CREATE UNIQUE INDEX "invty_prod_variants_tid_src_sys_ext_id_ukey" ON "inventory_product_variants" ("source_system", "external_id") WHERE (external_id IS NOT NULL);
-- Create "inventory_product_variant_attribute_values" table
CREATE TABLE "inventory_product_variant_attribute_values" (
  "id" character varying NOT NULL,
  "product_variant_id" character varying NOT NULL,
  "template_attribute_value_id" character varying NOT NULL,
  "sales_price_extra" numeric NULL,
  "org_id" character varying NOT NULL,
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
  "source_type" character varying NULL,
  "source_id" character varying NULL,
  "expires_at" timestamptz NULL,
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
-- Create index "invty_stock_trfs_expires_at_idx" to table: "inventory_stock_transfers"
CREATE INDEX "invty_stock_trfs_expires_at_idx" ON "inventory_stock_transfers" ("expires_at");
-- Create index "invty_stock_trfs_op_type_id_status_idx" to table: "inventory_stock_transfers"
CREATE INDEX "invty_stock_trfs_op_type_id_status_idx" ON "inventory_stock_transfers" ("operation_type_id", "status");
-- Create index "invty_stock_trfs_org_id_idem_key_idx" to table: "inventory_stock_transfers"
CREATE INDEX "invty_stock_trfs_org_id_idem_key_idx" ON "inventory_stock_transfers" ("org_id", "idempotency_key");
-- Create index "invty_stock_trfs_source_idx" to table: "inventory_stock_transfers"
CREATE INDEX "invty_stock_trfs_source_idx" ON "inventory_stock_transfers" ("source_type", "source_id");
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
  "source_item_id" character varying NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_stock_moves_destination_location_id_fkey" FOREIGN KEY ("destination_location_id") REFERENCES "inventory_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_moves_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_moves_source_location_id_fkey" FOREIGN KEY ("source_location_id") REFERENCES "inventory_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_moves_transfer_id_fkey" FOREIGN KEY ("transfer_id") REFERENCES "inventory_stock_transfers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_moves_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "essential_uoms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
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
  CONSTRAINT "inventory_stock_move_lines_transfer_id_fkey" FOREIGN KEY ("transfer_id") REFERENCES "inventory_stock_transfers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_move_lines_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "essential_uoms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
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
  "inventory_uom_id" character varying NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_stk_prod_cfgs_tid_ptpl_id_ukey" UNIQUE ("product_template_id", "org_id"),
  CONSTRAINT "inventory_stock_product_configs_inventory_uom_id_fkey" FOREIGN KEY ("inventory_uom_id") REFERENCES "essential_uoms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
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
  CONSTRAINT "inventory_stock_quants_base_uom_id_fkey" FOREIGN KEY ("base_uom_id") REFERENCES "essential_uoms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "inventory_stock_quants_location_id_fkey" FOREIGN KEY ("location_id") REFERENCES "inventory_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_quants_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "invty_stock_quants_loc_id_pvar_id_idx" to table: "inventory_stock_quants"
CREATE INDEX "invty_stock_quants_loc_id_pvar_id_idx" ON "inventory_stock_quants" ("location_id", "product_variant_id");
-- Create index "invty_stock_quants_next_count_date_idx" to table: "inventory_stock_quants"
CREATE INDEX "invty_stock_quants_next_count_date_idx" ON "inventory_stock_quants" ("next_count_date");
-- Create index "invty_stock_quants_pvar_id_loc_id_idx" to table: "inventory_stock_quants"
CREATE INDEX "invty_stock_quants_pvar_id_loc_id_idx" ON "inventory_stock_quants" ("product_variant_id", "location_id");
-- Create "inventory_stock_reservations" table
CREATE TABLE "inventory_stock_reservations" (
  "id" character varying NOT NULL,
  "warehouse_id" character varying NOT NULL,
  "product_variant_id" character varying NOT NULL,
  "base_uom_id" character varying NOT NULL,
  "quantity" numeric NOT NULL,
  "consumed_quantity" numeric NOT NULL,
  "released_quantity" numeric NOT NULL,
  "reserved_until" timestamptz NULL,
  "status" character varying NOT NULL,
  "source_module" character varying NOT NULL,
  "source_type" character varying NOT NULL,
  "source_id" character varying NOT NULL,
  "source_line_id" character varying NOT NULL,
  "source_revision" integer NOT NULL,
  "idempotency_key" character varying NOT NULL,
  "request_fingerprint" character varying NOT NULL,
  "release_reason" character varying NULL,
  "released_at" timestamptz NULL,
  "expiry_recorded_at" timestamptz NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_stock_rsv_source_line_rev_org_ukey" UNIQUE ("source_module", "source_type", "source_id", "source_line_id", "source_revision", "org_id"),
  CONSTRAINT "inventory_stock_reservations_base_uom_id_fkey" FOREIGN KEY ("base_uom_id") REFERENCES "essential_uoms" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "inventory_stock_reservations_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "inventory_stock_reservations_warehouse_id_fkey" FOREIGN KEY ("warehouse_id") REFERENCES "inventory_warehouses" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT
);
-- Create index "invty_stock_rsv_org_idem_key_idx" to table: "inventory_stock_reservations"
CREATE INDEX "invty_stock_rsv_org_idem_key_idx" ON "inventory_stock_reservations" ("org_id", "idempotency_key");
-- Create index "invty_stock_rsv_scope_status_until_idx" to table: "inventory_stock_reservations"
CREATE INDEX "invty_stock_rsv_scope_status_until_idx" ON "inventory_stock_reservations" ("org_id", "warehouse_id", "product_variant_id", "status", "reserved_until");
-- Create index "invty_stock_rsv_source_idx" to table: "inventory_stock_reservations"
CREATE INDEX "invty_stock_rsv_source_idx" ON "inventory_stock_reservations" ("source_module", "source_type", "source_id");
-- Create index "invty_stock_rsv_until_id_idx" to table: "inventory_stock_reservations"
CREATE INDEX "invty_stock_rsv_until_id_idx" ON "inventory_stock_reservations" ("reserved_until", "id");
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
  CONSTRAINT "inventory_stock_scraps_base_uom_id_fkey" FOREIGN KEY ("base_uom_id") REFERENCES "essential_uoms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
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
-- Create "inventory_warehouse_product_guards" table
CREATE TABLE "inventory_warehouse_product_guards" (
  "id" character varying NOT NULL,
  "warehouse_id" character varying NOT NULL,
  "product_variant_id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_wh_prod_guards_wh_pvar_org_ukey" UNIQUE ("warehouse_id", "product_variant_id", "org_id"),
  CONSTRAINT "inventory_warehouse_product_guards_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "inventory_warehouse_product_guards_warehouse_id_fkey" FOREIGN KEY ("warehouse_id") REFERENCES "inventory_warehouses" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
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
-- Create "purchase_agreements" table
CREATE TABLE "purchase_agreements" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "reference" character varying NULL,
  "agreement_type" character varying NOT NULL,
  "status" character varying NOT NULL,
  "vendor_id" character varying NULL,
  "buyer_id" character varying NOT NULL,
  "currency_id" character varying NULL,
  "start_date" date NULL,
  "end_date" date NULL,
  "description" character varying NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "purch_agreements_tid_code_org_id_ukey" UNIQUE ("code", "org_id")
);
-- Create index "purch_agreements_tid_org_id_idx" to table: "purchase_agreements"
CREATE INDEX "purch_agreements_tid_org_id_idx" ON "purchase_agreements" ("org_id");
-- Create index "purch_agreements_tid_status_idx" to table: "purchase_agreements"
CREATE INDEX "purch_agreements_tid_status_idx" ON "purchase_agreements" ("status");
-- Create index "purch_agreements_tid_vendor_id_idx" to table: "purchase_agreements"
CREATE INDEX "purch_agreements_tid_vendor_id_idx" ON "purchase_agreements" ("vendor_id");
-- Create "purchase_agreement_lines" table
CREATE TABLE "purchase_agreement_lines" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "purchase_agreement_id" character varying NOT NULL,
  "sequence" integer NOT NULL,
  "product_variant_id" character varying NOT NULL,
  "uom_id" character varying NULL,
  "quantity" numeric NOT NULL,
  "unit_price" numeric NOT NULL,
  "description" character varying NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "purchase_agreement_lines_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "purchase_agreement_lines_purchase_agreement_id_fkey" FOREIGN KEY ("purchase_agreement_id") REFERENCES "purchase_agreements" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "purchase_agreement_lines_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "essential_uoms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "purch_agr_lines_tid_agr_id_seq_idx" to table: "purchase_agreement_lines"
CREATE INDEX "purch_agr_lines_tid_agr_id_seq_idx" ON "purchase_agreement_lines" ("purchase_agreement_id", "sequence");
-- Create index "purch_agr_lines_tid_pvar_id_idx" to table: "purchase_agreement_lines"
CREATE INDEX "purch_agr_lines_tid_pvar_id_idx" ON "purchase_agreement_lines" ("product_variant_id");
-- Create "purchase_orders" table
CREATE TABLE "purchase_orders" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "status" character varying NOT NULL,
  "vendor_id" character varying NOT NULL,
  "vendor_reference" character varying NULL,
  "source_reference" character varying NULL,
  "buyer_id" character varying NOT NULL,
  "currency_id" character varying NULL,
  "order_deadline" timestamptz NULL,
  "expected_arrival" timestamptz NULL,
  "confirmed_at" timestamptz NULL,
  "agreement_id" character varying NULL,
  "sourcing_group_id" character varying NULL,
  "priority" character varying NOT NULL,
  "terms_conditions" character varying NULL,
  "is_locked" boolean NOT NULL,
  "vendor_acknowledged" boolean NOT NULL,
  "untaxed_amount" numeric NOT NULL,
  "tax_amount" numeric NOT NULL,
  "total_amount" numeric NOT NULL,
  "approval_required" boolean NOT NULL,
  "approved_by" character varying NULL,
  "approved_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "purch_orders_tid_code_org_id_ukey" UNIQUE ("code", "org_id")
);
-- Create index "purch_orders_tid_agreement_id_idx" to table: "purchase_orders"
CREATE INDEX "purch_orders_tid_agreement_id_idx" ON "purchase_orders" ("agreement_id");
-- Create index "purch_orders_tid_org_id_status_idx" to table: "purchase_orders"
CREATE INDEX "purch_orders_tid_org_id_status_idx" ON "purchase_orders" ("org_id", "status");
-- Create index "purch_orders_tid_srcgroup_id_idx" to table: "purchase_orders"
CREATE INDEX "purch_orders_tid_srcgroup_id_idx" ON "purchase_orders" ("sourcing_group_id");
-- Create index "purch_orders_tid_total_amount_idx" to table: "purchase_orders"
CREATE INDEX "purch_orders_tid_total_amount_idx" ON "purchase_orders" ("total_amount");
-- Create index "purch_orders_tid_vendor_id_idx" to table: "purchase_orders"
CREATE INDEX "purch_orders_tid_vendor_id_idx" ON "purchase_orders" ("vendor_id");
-- Create "purchase_order_lines" table
CREATE TABLE "purchase_order_lines" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "purchase_order_id" character varying NOT NULL,
  "sequence" integer NOT NULL,
  "line_type" character varying NOT NULL,
  "product_variant_id" character varying NULL,
  "description" character varying NULL,
  "quantity" numeric NOT NULL,
  "uom_id" character varying NULL,
  "inventory_quantity" numeric NOT NULL,
  "unit_price" numeric NOT NULL,
  "vendor_product_price_id" character varying NULL,
  "resolved_unit_price" numeric NULL,
  "discount_percent" numeric NOT NULL,
  "expected_arrival" timestamptz NULL,
  "subtotal" numeric NOT NULL,
  "tax_amount" numeric NOT NULL,
  "total" numeric NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "purchase_order_lines_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "purchase_order_lines_purchase_order_id_fkey" FOREIGN KEY ("purchase_order_id") REFERENCES "purchase_orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "purchase_order_lines_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "essential_uoms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "purch_ord_lines_tid_ord_id_seq_idx" to table: "purchase_order_lines"
CREATE INDEX "purch_ord_lines_tid_ord_id_seq_idx" ON "purchase_order_lines" ("purchase_order_id", "sequence");
-- Create index "purch_ord_lines_tid_pvar_id_idx" to table: "purchase_order_lines"
CREATE INDEX "purch_ord_lines_tid_pvar_id_idx" ON "purchase_order_lines" ("product_variant_id");
-- Create index "purch_ord_lines_tid_vpp_id_idx" to table: "purchase_order_lines"
CREATE INDEX "purch_ord_lines_tid_vpp_id_idx" ON "purchase_order_lines" ("vendor_product_price_id");
-- Create "purchase_vendor_product_prices" table
CREATE TABLE "purchase_vendor_product_prices" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "vendor_id" character varying NOT NULL,
  "product_template_id" character varying NOT NULL,
  "product_variant_id" character varying NULL,
  "purchase_uom_id" character varying NULL,
  "currency_id" character varying NOT NULL,
  "min_quantity" numeric NOT NULL,
  "unit_price" numeric NOT NULL,
  "valid_from" timestamptz NULL,
  "valid_to" timestamptz NULL,
  "lead_time_days" integer NOT NULL,
  "sequence" integer NOT NULL,
  "vendor_product_code" character varying NULL,
  "vendor_product_name" character varying NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "purchase_vendor_product_prices_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "purchase_vendor_product_prices_purchase_uom_id_fkey" FOREIGN KEY ("purchase_uom_id") REFERENCES "essential_uoms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "purch_vpp_tid_validity_idx" to table: "purchase_vendor_product_prices"
CREATE INDEX "purch_vpp_tid_validity_idx" ON "purchase_vendor_product_prices" ("valid_from", "valid_to");
-- Create index "purch_vpp_tid_variant_idx" to table: "purchase_vendor_product_prices"
CREATE INDEX "purch_vpp_tid_variant_idx" ON "purchase_vendor_product_prices" ("product_variant_id");
-- Create index "purch_vpp_tid_vendor_tmpl_idx" to table: "purchase_vendor_product_prices"
CREATE INDEX "purch_vpp_tid_vendor_tmpl_idx" ON "purchase_vendor_product_prices" ("vendor_id", "product_template_id");
-- Create "sales_channels" table
CREATE TABLE "sales_channels" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "managed_by_module" character varying NULL,
  "status" character varying NOT NULL,
  "is_system" boolean NOT NULL,
  "default_fulfillment_method_id" character varying NULL,
  "auto_confirm_order" boolean NOT NULL,
  "auto_confirm_refund" boolean NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_channels_code_org_id_ukey" UNIQUE ("code", "org_id")
);
-- Create "sales_points" table
CREATE TABLE "sales_points" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_channel_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "code" character varying NULL,
  "external_reference_id" character varying NULL,
  "external_reference_type" character varying NULL,
  "status" character varying NOT NULL,
  "default_fulfillment_method_id" character varying NULL,
  "fulfillment_enabled" boolean NOT NULL,
  "inventory_location_id" character varying NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_points_sales_channel_id_fkey" FOREIGN KEY ("sales_channel_id") REFERENCES "sales_channels" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "sales_points_tid_channel_code_ukey" to table: "sales_points"
CREATE UNIQUE INDEX "sales_points_tid_channel_code_ukey" ON "sales_points" ("sales_channel_id", "code") WHERE (code IS NOT NULL);
-- Create index "sales_points_tid_channel_ext_ref_ukey" to table: "sales_points"
CREATE UNIQUE INDEX "sales_points_tid_channel_ext_ref_ukey" ON "sales_points" ("sales_channel_id", "external_reference_id") WHERE (external_reference_id IS NOT NULL);
-- Create index "sales_points_tid_channel_status_idx" to table: "sales_points"
CREATE INDEX "sales_points_tid_channel_status_idx" ON "sales_points" ("sales_channel_id", "status");
-- Create "sales_orders" table
CREATE TABLE "sales_orders" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "order_number" character varying NOT NULL,
  "sales_channel_id" character varying NOT NULL,
  "sales_point_id" character varying NOT NULL,
  "customer_reference" character varying NULL,
  "adjusted_by_order_id" character varying NULL,
  "adjusts_order_id" character varying NULL,
  "requested_fulfillment_method_id" character varying NULL,
  "requested_target_outlet_id" character varying NULL,
  "customer_identity_mode" character varying NOT NULL,
  "sold_to_party_id" character varying NULL,
  "bill_to_party_id" character varying NULL,
  "payer_party_id" character varying NULL,
  "crm_opportunity_reference" character varying NULL,
  "currency_code" character varying NOT NULL,
  "status" character varying NOT NULL,
  "payment_status" character varying NOT NULL,
  "fulfillment_status" character varying NOT NULL,
  "invoice_status" character varying NOT NULL,
  "subtotal" numeric NOT NULL,
  "discount_total" numeric NOT NULL,
  "tax_total" numeric NOT NULL,
  "grand_total" numeric NOT NULL,
  "estimated_total_price" numeric NULL,
  "exchange_of_return_id" character varying NULL,
  "external_reference" character varying NULL,
  "idempotency_key" character varying NULL,
  "stage_version" integer NULL,
  "initial_bill_id" character varying NULL,
  "confirmed_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
  "tax_snapshot" jsonb NULL,
  "cancelled_at" timestamptz NULL,
  "auto_confirm_order" boolean NOT NULL,
  "auto_confirm_refund" boolean NOT NULL,
  "confirmation_note" character varying NULL,
  "cancellation_note" character varying NULL,
  "refund_note" character varying NULL,
  "valid_until" timestamptz NULL,
  "expired_at" timestamptz NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_orders_order_number_ukey" UNIQUE ("order_number"),
  CONSTRAINT "sales_orders_sales_channel_id_fkey" FOREIGN KEY ("sales_channel_id") REFERENCES "sales_channels" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "sales_orders_sales_point_id_fkey" FOREIGN KEY ("sales_point_id") REFERENCES "sales_points" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "sales_orders_tid_billto_idx" to table: "sales_orders"
CREATE INDEX "sales_orders_tid_billto_idx" ON "sales_orders" ("bill_to_party_id");
-- Create index "sales_orders_tid_channel_extref_ukey" to table: "sales_orders"
CREATE UNIQUE INDEX "sales_orders_tid_channel_extref_ukey" ON "sales_orders" ("sales_channel_id", "external_reference") WHERE (external_reference IS NOT NULL);
-- Create index "sales_orders_tid_channel_idem_ukey" to table: "sales_orders"
CREATE UNIQUE INDEX "sales_orders_tid_channel_idem_ukey" ON "sales_orders" ("sales_channel_id", "idempotency_key") WHERE (idempotency_key IS NOT NULL);
-- Create index "sales_orders_tid_channel_status_idx" to table: "sales_orders"
CREATE INDEX "sales_orders_tid_channel_status_idx" ON "sales_orders" ("sales_channel_id", "status");
-- Create index "sales_orders_tid_customer_idx" to table: "sales_orders"
CREATE INDEX "sales_orders_tid_customer_idx" ON "sales_orders" ("customer_reference");
-- Create index "sales_orders_tid_point_status_idx" to table: "sales_orders"
CREATE INDEX "sales_orders_tid_point_status_idx" ON "sales_orders" ("sales_point_id", "status");
-- Create "sales_bills" table
CREATE TABLE "sales_bills" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "bill_number" character varying NOT NULL,
  "sales_order_id" character varying NOT NULL,
  "status" character varying NOT NULL,
  "payment_status" character varying NOT NULL,
  "currency_code" character varying NOT NULL,
  "subtotal" numeric NOT NULL,
  "discount_total" numeric NOT NULL,
  "tax_total" numeric NOT NULL,
  "total_amount" numeric NOT NULL,
  "settled_at" timestamptz NULL,
  "cancelled_at" timestamptz NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_bills_bill_number_ukey" UNIQUE ("bill_number"),
  CONSTRAINT "sales_bills_sales_order_id_fkey" FOREIGN KEY ("sales_order_id") REFERENCES "sales_orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_bills_tid_order_idx" to table: "sales_bills"
CREATE INDEX "sales_bills_tid_order_idx" ON "sales_bills" ("sales_order_id");
-- Create index "sales_bills_tid_status_idx" to table: "sales_bills"
CREATE INDEX "sales_bills_tid_status_idx" ON "sales_bills" ("status", "payment_status");
-- Create "sales_order_lines" table
CREATE TABLE "sales_order_lines" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_order_id" character varying NOT NULL,
  "line_number" integer NOT NULL,
  "line_type" character varying NOT NULL,
  "product_variant_id" character varying NULL,
  "product_code_snapshot" character varying NULL,
  "product_name_snapshot" character varying NULL,
  "uom_id" character varying NULL,
  "ordered_quantity" numeric NOT NULL,
  "requires_fulfillment" boolean NOT NULL,
  "fulfilled_quantity" numeric NOT NULL,
  "returned_quantity" numeric NOT NULL,
  "estimated_price" numeric NULL,
  "base_unit_price" numeric NOT NULL,
  "effective_unit_price" numeric NOT NULL,
  "gross_amount" numeric NOT NULL,
  "discount_amount" numeric NOT NULL,
  "net_amount" numeric NOT NULL,
  "tax_rate_snapshot" numeric NOT NULL,
  "tax_amount" numeric NOT NULL,
  "final_amount" numeric NOT NULL,
  "pricing_source" character varying NOT NULL,
  "source_promotion_program_id" character varying NULL,
  "sales_combo_id" character varying NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_order_lines_tid_order_lineno_ukey" UNIQUE ("sales_order_id", "line_number"),
  CONSTRAINT "sales_order_lines_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "sales_order_lines_sales_order_id_fkey" FOREIGN KEY ("sales_order_id") REFERENCES "sales_orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sales_order_lines_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "essential_uoms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "sales_order_lines_tid_uom_idx" to table: "sales_order_lines"
CREATE INDEX "sales_order_lines_tid_uom_idx" ON "sales_order_lines" ("uom_id");
-- Create index "sales_order_lines_tid_variant_idx" to table: "sales_order_lines"
CREATE INDEX "sales_order_lines_tid_variant_idx" ON "sales_order_lines" ("product_variant_id");
-- Create "sales_bill_lines" table
CREATE TABLE "sales_bill_lines" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_bill_id" character varying NOT NULL,
  "sales_order_line_id" character varying NOT NULL,
  "quantity" numeric NOT NULL,
  "allocated_net_amount" numeric NOT NULL,
  "allocated_tax_amount" numeric NOT NULL,
  "allocated_total_amount" numeric NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_bill_lines_tid_bill_ordline_ukey" UNIQUE ("sales_bill_id", "sales_order_line_id"),
  CONSTRAINT "sales_bill_lines_sales_bill_id_fkey" FOREIGN KEY ("sales_bill_id") REFERENCES "sales_bills" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sales_bill_lines_sales_order_line_id_fkey" FOREIGN KEY ("sales_order_line_id") REFERENCES "sales_order_lines" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_bill_lines_tid_bill_idx" to table: "sales_bill_lines"
CREATE INDEX "sales_bill_lines_tid_bill_idx" ON "sales_bill_lines" ("sales_bill_id");
-- Create index "sales_bill_lines_tid_ordline_idx" to table: "sales_bill_lines"
CREATE INDEX "sales_bill_lines_tid_ordline_idx" ON "sales_bill_lines" ("sales_order_line_id");
-- Create "sales_bill_relations" table
CREATE TABLE "sales_bill_relations" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "source_bill_id" character varying NOT NULL,
  "target_bill_id" character varying NOT NULL,
  "relation_type" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_bill_rels_tid_pair_type_ukey" UNIQUE ("source_bill_id", "target_bill_id", "relation_type"),
  CONSTRAINT "sales_bill_relations_source_bill_id_fkey" FOREIGN KEY ("source_bill_id") REFERENCES "sales_bills" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "sales_bill_relations_target_bill_id_fkey" FOREIGN KEY ("target_bill_id") REFERENCES "sales_bills" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "sales_bill_rels_tid_source_idx" to table: "sales_bill_relations"
CREATE INDEX "sales_bill_rels_tid_source_idx" ON "sales_bill_relations" ("source_bill_id");
-- Create index "sales_bill_rels_tid_target_idx" to table: "sales_bill_relations"
CREATE INDEX "sales_bill_rels_tid_target_idx" ON "sales_bill_relations" ("target_bill_id");
-- Create "sales_fulfillment_methods" table
CREATE TABLE "sales_fulfillment_methods" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "fulfillment_type" character varying NOT NULL,
  "initial_target_selection" character varying NOT NULL,
  "max_attempts" integer NULL,
  "failure_action" character varying NOT NULL,
  "allow_target_change" boolean NOT NULL,
  "allow_partial_fulfillment" boolean NOT NULL,
  "requires_authenticated_customer" boolean NOT NULL,
  "reservation_ttl_minutes" integer NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_fulfillment_methods_code_org_id_ukey" UNIQUE ("code", "org_id")
);
-- Create index "sales_fulfil_methods_tid_type_arch_idx" to table: "sales_fulfillment_methods"
CREATE INDEX "sales_fulfil_methods_tid_type_arch_idx" ON "sales_fulfillment_methods" ("fulfillment_type", "is_archived");
-- Create "sales_channel_fulfillment_methods" table
CREATE TABLE "sales_channel_fulfillment_methods" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_channel_id" character varying NOT NULL,
  "fulfillment_method_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_chan_fulfil_methods_tid_ch_me_ukey" UNIQUE ("sales_channel_id", "fulfillment_method_id"),
  CONSTRAINT "sales_channel_fulfillment_methods_fulfillment_method_id_fkey" FOREIGN KEY ("fulfillment_method_id") REFERENCES "sales_fulfillment_methods" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "sales_channel_fulfillment_methods_sales_channel_id_fkey" FOREIGN KEY ("sales_channel_id") REFERENCES "sales_channels" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "sales_combos" table
CREATE TABLE "sales_combos" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "combo_price" numeric NOT NULL,
  "valid_from" timestamptz NULL,
  "valid_until" timestamptz NULL,
  "return_policy" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_combos_code_ukey" UNIQUE ("code")
);
-- Create index "sales_combos_tid_validity_idx" to table: "sales_combos"
CREATE INDEX "sales_combos_tid_validity_idx" ON "sales_combos" ("valid_from", "valid_until");
-- Create "sales_combo_components" table
CREATE TABLE "sales_combo_components" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_combo_id" character varying NOT NULL,
  "product_variant_id" character varying NOT NULL,
  "quantity" numeric NOT NULL,
  "uom_id" character varying NULL,
  "is_required" boolean NOT NULL,
  "selection_group" character varying NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_combo_comps_tid_combo_variant_ukey" UNIQUE ("sales_combo_id", "product_variant_id"),
  CONSTRAINT "sales_combo_components_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "sales_combo_components_sales_combo_id_fkey" FOREIGN KEY ("sales_combo_id") REFERENCES "sales_combos" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sales_combo_components_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "essential_uoms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "sales_combo_comps_tid_uom_idx" to table: "sales_combo_components"
CREATE INDEX "sales_combo_comps_tid_uom_idx" ON "sales_combo_components" ("uom_id");
-- Create index "sales_combo_comps_tid_variant_idx" to table: "sales_combo_components"
CREATE INDEX "sales_combo_comps_tid_variant_idx" ON "sales_combo_components" ("product_variant_id");
-- Create "sales_fiscal_requests" table
CREATE TABLE "sales_fiscal_requests" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_bill_id" character varying NOT NULL,
  "intent" character varying NOT NULL,
  "status" character varying NOT NULL,
  "idempotency_key" character varying NOT NULL,
  "provider_reference" character varying NULL,
  "attempt_count" integer NULL,
  "last_error" character varying NULL,
  "buyer_snapshot" jsonb NULL,
  "original_fiscal_request_id" character varying NULL,
  "requested_at" timestamptz NULL,
  "issued_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_fiscal_reqs_tid_idemkey_ukey" UNIQUE ("idempotency_key"),
  CONSTRAINT "sales_fiscal_requests_sales_bill_id_fkey" FOREIGN KEY ("sales_bill_id") REFERENCES "sales_bills" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_fiscal_reqs_tid_bill_idx" to table: "sales_fiscal_requests"
CREATE INDEX "sales_fiscal_reqs_tid_bill_idx" ON "sales_fiscal_requests" ("sales_bill_id");
-- Create index "sales_fiscal_reqs_tid_original_idx" to table: "sales_fiscal_requests"
CREATE INDEX "sales_fiscal_reqs_tid_original_idx" ON "sales_fiscal_requests" ("original_fiscal_request_id");
-- Create index "sales_fiscal_reqs_tid_provref_idx" to table: "sales_fiscal_requests"
CREATE INDEX "sales_fiscal_reqs_tid_provref_idx" ON "sales_fiscal_requests" ("provider_reference");
-- Create index "sales_fiscal_reqs_tid_status_idx" to table: "sales_fiscal_requests"
CREATE INDEX "sales_fiscal_reqs_tid_status_idx" ON "sales_fiscal_requests" ("status");
-- Create "sales_order_fulfillments" table
CREATE TABLE "sales_order_fulfillments" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_order_id" character varying NOT NULL,
  "fulfillment_method_id" character varying NOT NULL,
  "fulfillment_type" character varying NOT NULL,
  "target_outlet_id" character varying NULL,
  "fulfillment_status" character varying NOT NULL,
  "max_attempts" integer NULL,
  "failure_action" character varying NOT NULL,
  "allow_target_change" boolean NOT NULL,
  "allow_partial_fulfillment" boolean NOT NULL,
  "reservation_ttl_minutes" integer NULL,
  "reservation_expires_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_order_fulfillments_fulfillment_method_id_fkey" FOREIGN KEY ("fulfillment_method_id") REFERENCES "sales_fulfillment_methods" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "sales_order_fulfillments_sales_order_id_fkey" FOREIGN KEY ("sales_order_id") REFERENCES "sales_orders" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "sales_order_fulfillments_target_outlet_id_fkey" FOREIGN KEY ("target_outlet_id") REFERENCES "sales_points" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "sales_order_fulfil_tid_order_idx" to table: "sales_order_fulfillments"
CREATE INDEX "sales_order_fulfil_tid_order_idx" ON "sales_order_fulfillments" ("sales_order_id");
-- Create index "sales_order_fulfil_tid_status_expiry_idx" to table: "sales_order_fulfillments"
CREATE INDEX "sales_order_fulfil_tid_status_expiry_idx" ON "sales_order_fulfillments" ("fulfillment_status", "reservation_expires_at");
-- Create index "sales_order_fulfil_tid_target_status_idx" to table: "sales_order_fulfillments"
CREATE INDEX "sales_order_fulfil_tid_target_status_idx" ON "sales_order_fulfillments" ("target_outlet_id", "fulfillment_status");
-- Create "sales_fulfillment_attempts" table
CREATE TABLE "sales_fulfillment_attempts" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "fulfillment_id" character varying NOT NULL,
  "attempt_no" integer NOT NULL,
  "executor_outlet_id" character varying NOT NULL,
  "attempt_status" character varying NOT NULL,
  "external_correlation_id" character varying NULL,
  "result_event_id" character varying NULL,
  "result_payload_hash" character varying NULL,
  "inventory_result_ref" character varying NULL,
  "started_at" timestamptz NOT NULL,
  "completed_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_fulfil_attempts_tid_fulfil_no_ukey" UNIQUE ("fulfillment_id", "attempt_no"),
  CONSTRAINT "sales_fulfillment_attempts_executor_outlet_id_fkey" FOREIGN KEY ("executor_outlet_id") REFERENCES "sales_points" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "sales_fulfillment_attempts_fulfillment_id_fkey" FOREIGN KEY ("fulfillment_id") REFERENCES "sales_order_fulfillments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "sales_fulfil_attempts_tid_event_ukey" to table: "sales_fulfillment_attempts"
CREATE UNIQUE INDEX "sales_fulfil_attempts_tid_event_ukey" ON "sales_fulfillment_attempts" ("org_id", "result_event_id") WHERE (result_event_id IS NOT NULL);
-- Create index "sales_fulfil_attempts_tid_fulfil_idx" to table: "sales_fulfillment_attempts"
CREATE INDEX "sales_fulfil_attempts_tid_fulfil_idx" ON "sales_fulfillment_attempts" ("fulfillment_id", "attempt_no");
-- Create index "sales_fulfil_attempts_tid_status_idx" to table: "sales_fulfillment_attempts"
CREATE INDEX "sales_fulfil_attempts_tid_status_idx" ON "sales_fulfillment_attempts" ("attempt_status");
-- Create "sales_order_fulfillment_items" table
CREATE TABLE "sales_order_fulfillment_items" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "fulfillment_id" character varying NOT NULL,
  "sales_order_line_id" character varying NOT NULL,
  "product_variant_id" character varying NOT NULL,
  "uom_id" character varying NULL,
  "ordered_qty" numeric NOT NULL,
  "fulfilled_qty" numeric NOT NULL,
  "refunded_qty" numeric NOT NULL,
  "source_location_id" character varying NULL,
  "inventory_reservation_ref" character varying NULL,
  "item_status" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_order_fulfillment_items_fulfillment_id_fkey" FOREIGN KEY ("fulfillment_id") REFERENCES "sales_order_fulfillments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "sales_order_fulfillment_items_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "sales_order_fulfillment_items_sales_order_line_id_fkey" FOREIGN KEY ("sales_order_line_id") REFERENCES "sales_order_lines" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "sales_order_fulfillment_items_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "essential_uoms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "sales_fulfil_items_tid_fulfil_idx" to table: "sales_order_fulfillment_items"
CREATE INDEX "sales_fulfil_items_tid_fulfil_idx" ON "sales_order_fulfillment_items" ("fulfillment_id");
-- Create index "sales_fulfil_items_tid_line_idx" to table: "sales_order_fulfillment_items"
CREATE INDEX "sales_fulfil_items_tid_line_idx" ON "sales_order_fulfillment_items" ("sales_order_line_id");
-- Create "sales_fulfillment_attempt_items" table
CREATE TABLE "sales_fulfillment_attempt_items" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "attempt_id" character varying NOT NULL,
  "fulfillment_item_id" character varying NOT NULL,
  "attempted_qty" numeric NOT NULL,
  "dispensed_qty" numeric NOT NULL,
  "failed_qty" numeric NOT NULL,
  "item_result" character varying NOT NULL,
  "failure_code" character varying NULL,
  "failure_message" character varying NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_attempt_items_tid_att_item_ukey" UNIQUE ("attempt_id", "fulfillment_item_id"),
  CONSTRAINT "sales_fulfillment_attempt_items_attempt_id_fkey" FOREIGN KEY ("attempt_id") REFERENCES "sales_fulfillment_attempts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "sales_fulfillment_attempt_items_fulfillment_item_id_fkey" FOREIGN KEY ("fulfillment_item_id") REFERENCES "sales_order_fulfillment_items" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "sales_attempt_items_tid_attempt_idx" to table: "sales_fulfillment_attempt_items"
CREATE INDEX "sales_attempt_items_tid_attempt_idx" ON "sales_fulfillment_attempt_items" ("attempt_id");
-- Create index "sales_attempt_items_tid_fulfil_item_idx" to table: "sales_fulfillment_attempt_items"
CREATE INDEX "sales_attempt_items_tid_fulfil_item_idx" ON "sales_fulfillment_attempt_items" ("fulfillment_item_id");
-- Create "sales_fulfillment_requests" table
CREATE TABLE "sales_fulfillment_requests" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_order_id" character varying NOT NULL,
  "request_type" character varying NOT NULL,
  "status" character varying NOT NULL,
  "inventory_reference" character varying NULL,
  "failure_reason" character varying NULL,
  "requested_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_fulfillment_requests_sales_order_id_fkey" FOREIGN KEY ("sales_order_id") REFERENCES "sales_orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_fulfil_reqs_tid_invref_idx" to table: "sales_fulfillment_requests"
CREATE INDEX "sales_fulfil_reqs_tid_invref_idx" ON "sales_fulfillment_requests" ("inventory_reference");
-- Create index "sales_fulfil_reqs_tid_order_idx" to table: "sales_fulfillment_requests"
CREATE INDEX "sales_fulfil_reqs_tid_order_idx" ON "sales_fulfillment_requests" ("sales_order_id");
-- Create index "sales_fulfil_reqs_tid_status_idx" to table: "sales_fulfillment_requests"
CREATE INDEX "sales_fulfil_reqs_tid_status_idx" ON "sales_fulfillment_requests" ("status");
-- Create "sales_fulfillment_request_lines" table
CREATE TABLE "sales_fulfillment_request_lines" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_fulfillment_request_id" character varying NOT NULL,
  "sales_order_line_id" character varying NOT NULL,
  "quantity" numeric NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_fulfil_lines_tid_req_ordline_ukey" UNIQUE ("sales_fulfillment_request_id", "sales_order_line_id"),
  CONSTRAINT "sales_fulfillment_request_lines_sales_fulfillment_request_id_fk" FOREIGN KEY ("sales_fulfillment_request_id") REFERENCES "sales_fulfillment_requests" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sales_fulfillment_request_lines_sales_order_line_id_fkey" FOREIGN KEY ("sales_order_line_id") REFERENCES "sales_order_lines" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_fulfil_lines_tid_ordline_idx" to table: "sales_fulfillment_request_lines"
CREATE INDEX "sales_fulfil_lines_tid_ordline_idx" ON "sales_fulfillment_request_lines" ("sales_order_line_id");
-- Create index "sales_fulfil_lines_tid_req_idx" to table: "sales_fulfillment_request_lines"
CREATE INDEX "sales_fulfil_lines_tid_req_idx" ON "sales_fulfillment_request_lines" ("sales_fulfillment_request_id");
-- Create "sales_fulfillment_target_changes" table
CREATE TABLE "sales_fulfillment_target_changes" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "fulfillment_id" character varying NOT NULL,
  "from_outlet_id" character varying NULL,
  "to_outlet_id" character varying NOT NULL,
  "inventory_operation_ref" character varying NULL,
  "reason" character varying NOT NULL,
  "actor_type" character varying NOT NULL,
  "actor_id" character varying NULL,
  "changed_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_fulfillment_target_changes_from_outlet_id_fkey" FOREIGN KEY ("from_outlet_id") REFERENCES "sales_points" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "sales_fulfillment_target_changes_fulfillment_id_fkey" FOREIGN KEY ("fulfillment_id") REFERENCES "sales_order_fulfillments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "sales_fulfillment_target_changes_to_outlet_id_fkey" FOREIGN KEY ("to_outlet_id") REFERENCES "sales_points" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "sales_target_changes_tid_fulfil_idx" to table: "sales_fulfillment_target_changes"
CREATE INDEX "sales_target_changes_tid_fulfil_idx" ON "sales_fulfillment_target_changes" ("fulfillment_id", "changed_at");
-- Create "sales_manual_discounts" table
CREATE TABLE "sales_manual_discounts" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_order_id" character varying NOT NULL,
  "sales_order_line_id" character varying NULL,
  "discount_amount" numeric NOT NULL,
  "reason" character varying NOT NULL,
  "granted_by" character varying NULL,
  "original_amount" numeric NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_manual_discounts_sales_order_id_fkey" FOREIGN KEY ("sales_order_id") REFERENCES "sales_orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_manual_discs_tid_actor_idx" to table: "sales_manual_discounts"
CREATE INDEX "sales_manual_discs_tid_actor_idx" ON "sales_manual_discounts" ("granted_by");
-- Create index "sales_manual_discs_tid_line_idx" to table: "sales_manual_discounts"
CREATE INDEX "sales_manual_discs_tid_line_idx" ON "sales_manual_discounts" ("sales_order_line_id");
-- Create index "sales_manual_discs_tid_order_idx" to table: "sales_manual_discounts"
CREATE INDEX "sales_manual_discs_tid_order_idx" ON "sales_manual_discounts" ("sales_order_id");
-- Create "sales_order_adjustments" table
CREATE TABLE "sales_order_adjustments" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_order_id" character varying NOT NULL,
  "sales_order_line_id" character varying NULL,
  "sequence" integer NOT NULL,
  "adjustment_type" character varying NOT NULL,
  "source_type" character varying NULL,
  "source_id" character varying NULL,
  "description" character varying NULL,
  "base_amount" numeric NOT NULL,
  "adjustment_amount" numeric NOT NULL,
  "sales_return_id" character varying NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_order_adjs_tid_order_seq_ukey" UNIQUE ("sales_order_id", "sequence"),
  CONSTRAINT "sales_order_adjustments_sales_order_id_fkey" FOREIGN KEY ("sales_order_id") REFERENCES "sales_orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_order_adjs_tid_line_idx" to table: "sales_order_adjustments"
CREATE INDEX "sales_order_adjs_tid_line_idx" ON "sales_order_adjustments" ("sales_order_line_id");
-- Create index "sales_order_adjs_tid_return_idx" to table: "sales_order_adjustments"
CREATE INDEX "sales_order_adjs_tid_return_idx" ON "sales_order_adjustments" ("sales_return_id");
-- Create index "sales_order_adjs_tid_source_idx" to table: "sales_order_adjustments"
CREATE INDEX "sales_order_adjs_tid_source_idx" ON "sales_order_adjustments" ("source_type", "source_id");
-- Create "sales_order_line_allocations" table
CREATE TABLE "sales_order_line_allocations" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_order_line_id" character varying NOT NULL,
  "source_location_id" character varying NOT NULL,
  "quantity" numeric NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_ol_alloc_tid_line_location_ukey" UNIQUE ("sales_order_line_id", "source_location_id"),
  CONSTRAINT "sales_order_line_allocations_sales_order_line_id_fkey" FOREIGN KEY ("sales_order_line_id") REFERENCES "sales_order_lines" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_ol_alloc_tid_location_idx" to table: "sales_order_line_allocations"
CREATE INDEX "sales_ol_alloc_tid_location_idx" ON "sales_order_line_allocations" ("source_location_id");
-- Create "sales_order_line_components" table
CREATE TABLE "sales_order_line_components" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_order_line_id" character varying NOT NULL,
  "sequence" integer NOT NULL,
  "product_variant_id" character varying NOT NULL,
  "product_code_snapshot" character varying NULL,
  "product_name_snapshot" character varying NULL,
  "quantity" numeric NOT NULL,
  "uom_id" character varying NULL,
  "allocated_net_amount" numeric NOT NULL,
  "allocated_tax_amount" numeric NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_ol_components_tid_line_seq_ukey" UNIQUE ("sales_order_line_id", "sequence"),
  CONSTRAINT "sales_order_line_components_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "sales_order_line_components_sales_order_line_id_fkey" FOREIGN KEY ("sales_order_line_id") REFERENCES "sales_order_lines" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sales_order_line_components_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "essential_uoms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "sales_ol_components_tid_uom_idx" to table: "sales_order_line_components"
CREATE INDEX "sales_ol_components_tid_uom_idx" ON "sales_order_line_components" ("uom_id");
-- Create index "sales_ol_components_tid_variant_idx" to table: "sales_order_line_components"
CREATE INDEX "sales_ol_components_tid_variant_idx" ON "sales_order_line_components" ("product_variant_id");
-- Create "sales_payments" table
CREATE TABLE "sales_payments" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_bill_id" character varying NOT NULL,
  "payment_method_id" character varying NOT NULL,
  "payment_method_code_snapshot" character varying NULL,
  "amount" numeric NOT NULL,
  "currency_code" character varying NOT NULL,
  "status" character varying NOT NULL,
  "qr_code_url" character varying NULL,
  "pay_url" character varying NULL,
  "idempotency_key" character varying NULL,
  "external_transaction_id" character varying NULL,
  "provider_reference" character varying NULL,
  "payment_order_id" character varying NULL,
  "paid_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_payments_sales_bill_id_fkey" FOREIGN KEY ("sales_bill_id") REFERENCES "sales_bills" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_payments_tid_bill_extxn_ukey" to table: "sales_payments"
CREATE UNIQUE INDEX "sales_payments_tid_bill_extxn_ukey" ON "sales_payments" ("sales_bill_id", "external_transaction_id") WHERE (external_transaction_id IS NOT NULL);
-- Create index "sales_payments_tid_bill_idem_ukey" to table: "sales_payments"
CREATE UNIQUE INDEX "sales_payments_tid_bill_idem_ukey" ON "sales_payments" ("sales_bill_id", "idempotency_key") WHERE (idempotency_key IS NOT NULL);
-- Create index "sales_payments_tid_bill_idx" to table: "sales_payments"
CREATE INDEX "sales_payments_tid_bill_idx" ON "sales_payments" ("sales_bill_id");
-- Create index "sales_payments_tid_method_idx" to table: "sales_payments"
CREATE INDEX "sales_payments_tid_method_idx" ON "sales_payments" ("payment_method_id");
-- Create index "sales_payments_tid_pmtorder_idx" to table: "sales_payments"
CREATE INDEX "sales_payments_tid_pmtorder_idx" ON "sales_payments" ("payment_order_id");
-- Create index "sales_payments_tid_status_idx" to table: "sales_payments"
CREATE INDEX "sales_payments_tid_status_idx" ON "sales_payments" ("status");
-- Create "sales_point_payment_rel" table
CREATE TABLE "sales_point_payment_rel" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_point_id" character varying NOT NULL,
  "payment_method_id" character varying NOT NULL,
  "payment_profile_id" character varying NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_point_payment_rel_tid_point_method_ukey" UNIQUE ("sales_point_id", "payment_method_id"),
  CONSTRAINT "sales_point_payment_rel_sales_point_id_fkey" FOREIGN KEY ("sales_point_id") REFERENCES "sales_points" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "sales_pricelists" table
CREATE TABLE "sales_pricelists" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "currency_id" character varying NOT NULL,
  "is_default" boolean NOT NULL,
  "sales_channel_id" character varying NULL,
  "sales_point_id" character varying NULL,
  "valid_from" timestamptz NULL,
  "valid_until" timestamptz NULL,
  "priority" integer NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_pricelists_code_org_id_ukey" UNIQUE ("code", "org_id"),
  CONSTRAINT "sales_pricelists_sales_channel_id_fkey" FOREIGN KEY ("sales_channel_id") REFERENCES "sales_channels" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "sales_pricelists_sales_point_id_fkey" FOREIGN KEY ("sales_point_id") REFERENCES "sales_points" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "sales_pricelists_tid_channel_idx" to table: "sales_pricelists"
CREATE INDEX "sales_pricelists_tid_channel_idx" ON "sales_pricelists" ("sales_channel_id");
-- Create index "sales_pricelists_tid_point_idx" to table: "sales_pricelists"
CREATE INDEX "sales_pricelists_tid_point_idx" ON "sales_pricelists" ("sales_point_id");
-- Create index "sales_pricelists_tid_validity_idx" to table: "sales_pricelists"
CREATE INDEX "sales_pricelists_tid_validity_idx" ON "sales_pricelists" ("valid_from", "valid_until");
-- Create "sales_pricelist_items" table
CREATE TABLE "sales_pricelist_items" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_pricelist_id" character varying NOT NULL,
  "product_variant_id" character varying NULL,
  "uom_id" character varying NULL,
  "price" numeric NULL,
  "min_quantity" numeric NOT NULL,
  "applies_to" character varying NOT NULL,
  "product_template_id" character varying NULL,
  "product_category_id" character varying NULL,
  "valid_from" timestamptz NULL,
  "valid_to" timestamptz NULL,
  "sequence" integer NOT NULL,
  "calculation_method" character varying NOT NULL,
  "discount_percent" numeric NULL,
  "base_price_source" character varying NULL,
  "base_pricelist_id" character varying NULL,
  "surcharge_amount" numeric NULL,
  "rounding_increment" numeric NULL,
  "minimum_margin" numeric NULL,
  "maximum_margin" numeric NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_pricelist_items_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "sales_pricelist_items_sales_pricelist_id_fkey" FOREIGN KEY ("sales_pricelist_id") REFERENCES "sales_pricelists" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sales_pricelist_items_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "essential_uoms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "sales_pl_items_tid_cat_idx" to table: "sales_pricelist_items"
CREATE INDEX "sales_pl_items_tid_cat_idx" ON "sales_pricelist_items" ("product_category_id");
-- Create index "sales_pl_items_tid_tmpl_idx" to table: "sales_pricelist_items"
CREATE INDEX "sales_pl_items_tid_tmpl_idx" ON "sales_pricelist_items" ("product_template_id");
-- Create index "sales_pl_items_tid_uom_idx" to table: "sales_pricelist_items"
CREATE INDEX "sales_pl_items_tid_uom_idx" ON "sales_pricelist_items" ("uom_id");
-- Create index "sales_pl_items_tid_validity_idx" to table: "sales_pricelist_items"
CREATE INDEX "sales_pl_items_tid_validity_idx" ON "sales_pricelist_items" ("valid_from", "valid_to");
-- Create index "sales_pl_items_tid_variant_idx" to table: "sales_pricelist_items"
CREATE INDEX "sales_pl_items_tid_variant_idx" ON "sales_pricelist_items" ("product_variant_id");
-- Create "sales_promotion_programs" table
CREATE TABLE "sales_promotion_programs" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" character varying NOT NULL,
  "activation_type" character varying NOT NULL,
  "priority" integer NOT NULL,
  "valid_from" timestamptz NULL,
  "valid_until" timestamptz NULL,
  "stack_policy" character varying NOT NULL,
  "exclusive_group" character varying NULL,
  "usage_limit" integer NULL,
  "usage_limit_per_customer" integer NULL,
  "return_behavior" character varying NOT NULL,
  "restore_on_full_return" boolean NOT NULL,
  "restore_on_partial_return" boolean NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_promotion_programs_code_ukey" UNIQUE ("code")
);
-- Create index "sales_promo_progs_tid_activation_idx" to table: "sales_promotion_programs"
CREATE INDEX "sales_promo_progs_tid_activation_idx" ON "sales_promotion_programs" ("activation_type");
-- Create index "sales_promo_progs_tid_excl_group_idx" to table: "sales_promotion_programs"
CREATE INDEX "sales_promo_progs_tid_excl_group_idx" ON "sales_promotion_programs" ("exclusive_group");
-- Create index "sales_promo_progs_tid_priority_idx" to table: "sales_promotion_programs"
CREATE INDEX "sales_promo_progs_tid_priority_idx" ON "sales_promotion_programs" ("priority");
-- Create index "sales_promo_progs_tid_validity_idx" to table: "sales_promotion_programs"
CREATE INDEX "sales_promo_progs_tid_validity_idx" ON "sales_promotion_programs" ("valid_from", "valid_until");
-- Create "sales_promotion_condition_groups" table
CREATE TABLE "sales_promotion_condition_groups" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_promotion_program_id" character varying NOT NULL,
  "sequence" integer NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_promo_cgroups_tid_prog_seq_ukey" UNIQUE ("sales_promotion_program_id", "sequence"),
  CONSTRAINT "sales_promotion_condition_groups_sales_promotion_program_id_fke" FOREIGN KEY ("sales_promotion_program_id") REFERENCES "sales_promotion_programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "sales_promotion_conditions" table
CREATE TABLE "sales_promotion_conditions" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "group_id" character varying NOT NULL,
  "condition_type" character varying NOT NULL,
  "operator" character varying NOT NULL,
  "value_text" character varying NULL,
  "value_decimal" numeric NULL,
  "value_from" numeric NULL,
  "value_to" numeric NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_promotion_conditions_group_id_fkey" FOREIGN KEY ("group_id") REFERENCES "sales_promotion_condition_groups" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_promo_conds_tid_group_idx" to table: "sales_promotion_conditions"
CREATE INDEX "sales_promo_conds_tid_group_idx" ON "sales_promotion_conditions" ("group_id");
-- Create index "sales_promo_conds_tid_type_idx" to table: "sales_promotion_conditions"
CREATE INDEX "sales_promo_conds_tid_type_idx" ON "sales_promotion_conditions" ("condition_type");
-- Create "sales_promotion_condition_targets" table
CREATE TABLE "sales_promotion_condition_targets" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "condition_id" character varying NOT NULL,
  "target_type" character varying NOT NULL,
  "target_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_promo_ctargets_tid_cond_target_ukey" UNIQUE ("condition_id", "target_type", "target_id"),
  CONSTRAINT "sales_promotion_condition_targets_condition_id_fkey" FOREIGN KEY ("condition_id") REFERENCES "sales_promotion_conditions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_promo_ctargets_tid_target_idx" to table: "sales_promotion_condition_targets"
CREATE INDEX "sales_promo_ctargets_tid_target_idx" ON "sales_promotion_condition_targets" ("target_type", "target_id");
-- Create "sales_promotion_rewards" table
CREATE TABLE "sales_promotion_rewards" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_promotion_program_id" character varying NOT NULL,
  "sequence" integer NOT NULL,
  "reward_type" character varying NOT NULL,
  "value" numeric NOT NULL,
  "target_scope" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_promo_rewards_tid_prog_seq_ukey" UNIQUE ("sales_promotion_program_id", "sequence"),
  CONSTRAINT "sales_promotion_rewards_sales_promotion_program_id_fkey" FOREIGN KEY ("sales_promotion_program_id") REFERENCES "sales_promotion_programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "sales_quotations" table
CREATE TABLE "sales_quotations" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "quotation_number" character varying NOT NULL,
  "sales_channel_id" character varying NOT NULL,
  "sales_point_id" character varying NULL,
  "customer_reference" character varying NULL,
  "currency_code" character varying NOT NULL,
  "status" character varying NOT NULL,
  "valid_until" timestamptz NULL,
  "subtotal" numeric NULL,
  "discount_total" numeric NULL,
  "tax_total" numeric NULL,
  "grand_total" numeric NULL,
  "converted_sales_order_id" character varying NULL,
  "sent_at" timestamptz NULL,
  "accepted_at" timestamptz NULL,
  "cancelled_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_quotations_tid_number_ukey" UNIQUE ("quotation_number"),
  CONSTRAINT "sales_quotations_sales_channel_id_fkey" FOREIGN KEY ("sales_channel_id") REFERENCES "sales_channels" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "sales_quotations_tid_channel_idx" to table: "sales_quotations"
CREATE INDEX "sales_quotations_tid_channel_idx" ON "sales_quotations" ("sales_channel_id");
-- Create index "sales_quotations_tid_order_idx" to table: "sales_quotations"
CREATE INDEX "sales_quotations_tid_order_idx" ON "sales_quotations" ("converted_sales_order_id");
-- Create index "sales_quotations_tid_status_idx" to table: "sales_quotations"
CREATE INDEX "sales_quotations_tid_status_idx" ON "sales_quotations" ("status");
-- Create index "sales_quotations_tid_valid_idx" to table: "sales_quotations"
CREATE INDEX "sales_quotations_tid_valid_idx" ON "sales_quotations" ("valid_until");
-- Create "sales_quotation_lines" table
CREATE TABLE "sales_quotation_lines" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_quotation_id" character varying NOT NULL,
  "line_number" integer NOT NULL,
  "product_variant_id" character varying NULL,
  "product_code_snapshot" character varying NULL,
  "product_name_snapshot" character varying NULL,
  "uom_id" character varying NULL,
  "quantity" numeric NOT NULL,
  "unit_price" numeric NULL,
  "discount_amount" numeric NULL,
  "net_amount" numeric NULL,
  "tax_amount" numeric NULL,
  "final_amount" numeric NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_quotation_lines_tid_quot_lineno_ukey" UNIQUE ("sales_quotation_id", "line_number"),
  CONSTRAINT "sales_quotation_lines_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE RESTRICT,
  CONSTRAINT "sales_quotation_lines_sales_quotation_id_fkey" FOREIGN KEY ("sales_quotation_id") REFERENCES "sales_quotations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sales_quotation_lines_uom_id_fkey" FOREIGN KEY ("uom_id") REFERENCES "essential_uoms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "sales_quotation_lines_tid_variant_idx" to table: "sales_quotation_lines"
CREATE INDEX "sales_quotation_lines_tid_variant_idx" ON "sales_quotation_lines" ("product_variant_id");
-- Create "sales_returns" table
CREATE TABLE "sales_returns" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "return_number" character varying NOT NULL,
  "sales_order_id" character varying NOT NULL,
  "status" character varying NOT NULL,
  "inventory_return_status" character varying NOT NULL,
  "refund_status" character varying NOT NULL,
  "fiscal_adjustment_status" character varying NOT NULL,
  "reason" character varying NOT NULL,
  "refund_reason" character varying NOT NULL,
  "return_type" character varying NOT NULL,
  "inventory_disposition" character varying NULL,
  "refund_total" numeric NOT NULL,
  "inventory_reference" character varying NULL,
  "failure_reason" character varying NULL,
  "requested_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
  "confirmation_note" character varying NULL,
  "confirmed_at" timestamptz NULL,
  "locked_refund_amount" numeric NULL,
  "cancelled_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_returns_return_number_ukey" UNIQUE ("return_number"),
  CONSTRAINT "sales_returns_sales_order_id_fkey" FOREIGN KEY ("sales_order_id") REFERENCES "sales_orders" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "sales_returns_tid_fiscal_idx" to table: "sales_returns"
CREATE INDEX "sales_returns_tid_fiscal_idx" ON "sales_returns" ("fiscal_adjustment_status");
-- Create index "sales_returns_tid_order_idx" to table: "sales_returns"
CREATE INDEX "sales_returns_tid_order_idx" ON "sales_returns" ("sales_order_id");
-- Create index "sales_returns_tid_status_idx" to table: "sales_returns"
CREATE INDEX "sales_returns_tid_status_idx" ON "sales_returns" ("status");
-- Create "sales_refund_payments" table
CREATE TABLE "sales_refund_payments" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_return_id" character varying NOT NULL,
  "original_sales_payment_id" character varying NOT NULL,
  "amount" numeric NOT NULL,
  "currency_code" character varying NOT NULL,
  "status" character varying NOT NULL,
  "provider_reference" character varying NULL,
  "failure_reason" character varying NULL,
  "completed_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_refund_payments_original_sales_payment_id_fkey" FOREIGN KEY ("original_sales_payment_id") REFERENCES "sales_payments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "sales_refund_payments_sales_return_id_fkey" FOREIGN KEY ("sales_return_id") REFERENCES "sales_returns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_refund_pays_tid_orig_idx" to table: "sales_refund_payments"
CREATE INDEX "sales_refund_pays_tid_orig_idx" ON "sales_refund_payments" ("original_sales_payment_id");
-- Create index "sales_refund_pays_tid_return_idx" to table: "sales_refund_payments"
CREATE INDEX "sales_refund_pays_tid_return_idx" ON "sales_refund_payments" ("sales_return_id");
-- Create index "sales_refund_pays_tid_status_idx" to table: "sales_refund_payments"
CREATE INDEX "sales_refund_pays_tid_status_idx" ON "sales_refund_payments" ("status");
-- Create "sales_return_lines" table
CREATE TABLE "sales_return_lines" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_return_id" character varying NOT NULL,
  "sales_order_line_id" character varying NOT NULL,
  "quantity" numeric NOT NULL,
  "fulfillment_id" character varying NULL,
  "fulfillment_item_id" character varying NULL,
  "requested_qty" numeric NULL,
  "refunded_qty" numeric NULL,
  "refund_amount" numeric NOT NULL,
  "refund_tax_amount" numeric NOT NULL,
  "requires_inventory_return" boolean NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_return_lines_sales_order_line_id_fkey" FOREIGN KEY ("sales_order_line_id") REFERENCES "sales_order_lines" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "sales_return_lines_sales_return_id_fkey" FOREIGN KEY ("sales_return_id") REFERENCES "sales_returns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_return_lines_tid_ordline_idx" to table: "sales_return_lines"
CREATE INDEX "sales_return_lines_tid_ordline_idx" ON "sales_return_lines" ("sales_order_line_id");
-- Create index "sales_return_lines_tid_return_idx" to table: "sales_return_lines"
CREATE INDEX "sales_return_lines_tid_return_idx" ON "sales_return_lines" ("sales_return_id");
-- Create index "sales_return_lines_tid_uniq_ukey_notnull" to table: "sales_return_lines"
CREATE UNIQUE INDEX "sales_return_lines_tid_uniq_ukey_notnull" ON "sales_return_lines" ("sales_return_id", "sales_order_line_id", "fulfillment_item_id") WHERE (fulfillment_item_id IS NOT NULL);
-- Create index "sales_return_lines_tid_uniq_ukey_null" to table: "sales_return_lines"
CREATE UNIQUE INDEX "sales_return_lines_tid_uniq_ukey_null" ON "sales_return_lines" ("sales_return_id", "sales_order_line_id") WHERE (fulfillment_item_id IS NULL);
-- Create "sales_voucher_codes" table
CREATE TABLE "sales_voucher_codes" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "sales_promotion_program_id" character varying NOT NULL,
  "valid_from" timestamptz NULL,
  "valid_until" timestamptz NULL,
  "usage_limit" integer NULL,
  "usage_count" integer NOT NULL,
  "status" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_voucher_codes_code_ukey" UNIQUE ("code"),
  CONSTRAINT "sales_voucher_codes_sales_promotion_program_id_fkey" FOREIGN KEY ("sales_promotion_program_id") REFERENCES "sales_promotion_programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_vouchers_tid_program_idx" to table: "sales_voucher_codes"
CREATE INDEX "sales_vouchers_tid_program_idx" ON "sales_voucher_codes" ("sales_promotion_program_id");
-- Create index "sales_vouchers_tid_status_idx" to table: "sales_voucher_codes"
CREATE INDEX "sales_vouchers_tid_status_idx" ON "sales_voucher_codes" ("status");
-- Create index "sales_vouchers_tid_validity_idx" to table: "sales_voucher_codes"
CREATE INDEX "sales_vouchers_tid_validity_idx" ON "sales_voucher_codes" ("valid_from", "valid_until");
-- Create "sales_voucher_redemptions" table
CREATE TABLE "sales_voucher_redemptions" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "voucher_code_id" character varying NOT NULL,
  "sales_order_id" character varying NOT NULL,
  "status" character varying NOT NULL,
  "reserved_at" timestamptz NULL,
  "redeemed_at" timestamptz NULL,
  "released_at" timestamptz NULL,
  "reversed_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_vouch_redms_tid_voucher_order_ukey" UNIQUE ("voucher_code_id", "sales_order_id"),
  CONSTRAINT "sales_voucher_redemptions_sales_order_id_fkey" FOREIGN KEY ("sales_order_id") REFERENCES "sales_orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "sales_voucher_redemptions_voucher_code_id_fkey" FOREIGN KEY ("voucher_code_id") REFERENCES "sales_voucher_codes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_vouch_redms_tid_order_idx" to table: "sales_voucher_redemptions"
CREATE INDEX "sales_vouch_redms_tid_order_idx" ON "sales_voucher_redemptions" ("sales_order_id");
-- Create index "sales_vouch_redms_tid_status_idx" to table: "sales_voucher_redemptions"
CREATE INDEX "sales_vouch_redms_tid_status_idx" ON "sales_voucher_redemptions" ("status");
-- Create index "sales_vouch_redms_tid_voucher_idx" to table: "sales_voucher_redemptions"
CREATE INDEX "sales_vouch_redms_tid_voucher_idx" ON "sales_voucher_redemptions" ("voucher_code_id");
