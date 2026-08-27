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
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_channels_code_ukey" UNIQUE ("code")
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
  "external_reference" character varying NULL,
  "idempotency_key" character varying NULL,
  "confirmed_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
  "cancelled_at" timestamptz NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_orders_order_number_ukey" UNIQUE ("order_number"),
  CONSTRAINT "sales_orders_sales_channel_id_fkey" FOREIGN KEY ("sales_channel_id") REFERENCES "sales_channels" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "sales_orders_sales_point_id_fkey" FOREIGN KEY ("sales_point_id") REFERENCES "sales_points" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
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
  "uom_id" character varying NOT NULL,
  "ordered_quantity" numeric NOT NULL,
  "fulfilled_quantity" numeric NOT NULL,
  "returned_quantity" numeric NOT NULL,
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
  CONSTRAINT "sales_order_lines_sales_order_id_fkey" FOREIGN KEY ("sales_order_id") REFERENCES "sales_orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_order_lines_tid_uom_idx" to table: "sales_order_lines"
CREATE INDEX "sales_order_lines_tid_uom_idx" ON "sales_order_lines" ("uom_id");
-- Create index "sales_order_lines_tid_variant_idx" to table: "sales_order_lines"
CREATE INDEX "sales_order_lines_tid_variant_idx" ON "sales_order_lines" ("product_variant_id");
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
  "uom_id" character varying NOT NULL,
  "allocated_net_amount" numeric NOT NULL,
  "allocated_tax_amount" numeric NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_ol_components_tid_line_seq_ukey" UNIQUE ("sales_order_line_id", "sequence"),
  CONSTRAINT "sales_order_line_components_sales_order_line_id_fkey" FOREIGN KEY ("sales_order_line_id") REFERENCES "sales_order_lines" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_ol_components_tid_variant_idx" to table: "sales_order_line_components"
CREATE INDEX "sales_ol_components_tid_variant_idx" ON "sales_order_line_components" ("product_variant_id");
-- Create index "sales_ol_components_tid_uom_idx" to table: "sales_order_line_components"
CREATE INDEX "sales_ol_components_tid_uom_idx" ON "sales_order_line_components" ("uom_id");
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
-- Create index "sales_order_adjs_tid_source_idx" to table: "sales_order_adjustments"
CREATE INDEX "sales_order_adjs_tid_source_idx" ON "sales_order_adjustments" ("source_type", "source_id");
-- Create index "sales_order_adjs_tid_return_idx" to table: "sales_order_adjustments"
CREATE INDEX "sales_order_adjs_tid_return_idx" ON "sales_order_adjustments" ("sales_return_id");
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
-- Create index "sales_order_evts_tid_order_idx" to table: "sales_order_events"
CREATE INDEX "sales_order_evts_tid_order_idx" ON "sales_order_events" ("sales_order_id");
-- Create index "sales_order_evts_tid_entity_idx" to table: "sales_order_events"
CREATE INDEX "sales_order_evts_tid_entity_idx" ON "sales_order_events" ("entity_type", "entity_id");
-- Create index "sales_order_evts_tid_actor_idx" to table: "sales_order_events"
CREATE INDEX "sales_order_evts_tid_actor_idx" ON "sales_order_events" ("actor_id");
-- Create "sales_pricelists" table
CREATE TABLE "sales_pricelists" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" character varying NOT NULL,
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
  CONSTRAINT "sales_pricelists_code_ukey" UNIQUE ("code"),
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
  "product_variant_id" character varying NOT NULL,
  "uom_id" character varying NOT NULL,
  "price" numeric NOT NULL,
  "min_quantity" numeric NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_pl_items_tid_list_variant_qty_ukey" UNIQUE ("sales_pricelist_id", "product_variant_id", "min_quantity"),
  CONSTRAINT "sales_pricelist_items_sales_pricelist_id_fkey" FOREIGN KEY ("sales_pricelist_id") REFERENCES "sales_pricelists" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_pl_items_tid_variant_idx" to table: "sales_pricelist_items"
CREATE INDEX "sales_pl_items_tid_variant_idx" ON "sales_pricelist_items" ("product_variant_id");
-- Create index "sales_pl_items_tid_uom_idx" to table: "sales_pricelist_items"
CREATE INDEX "sales_pl_items_tid_uom_idx" ON "sales_pricelist_items" ("uom_id");
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
  "uom_id" character varying NOT NULL,
  "is_required" boolean NOT NULL,
  "selection_group" character varying NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_combo_comps_tid_combo_variant_ukey" UNIQUE ("sales_combo_id", "product_variant_id"),
  CONSTRAINT "sales_combo_components_sales_combo_id_fkey" FOREIGN KEY ("sales_combo_id") REFERENCES "sales_combos" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_combo_comps_tid_variant_idx" to table: "sales_combo_components"
CREATE INDEX "sales_combo_comps_tid_variant_idx" ON "sales_combo_components" ("product_variant_id");
-- Create index "sales_combo_comps_tid_uom_idx" to table: "sales_combo_components"
CREATE INDEX "sales_combo_comps_tid_uom_idx" ON "sales_combo_components" ("uom_id");
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
-- Create index "sales_promo_progs_tid_validity_idx" to table: "sales_promotion_programs"
CREATE INDEX "sales_promo_progs_tid_validity_idx" ON "sales_promotion_programs" ("valid_from", "valid_until");
-- Create index "sales_promo_progs_tid_priority_idx" to table: "sales_promotion_programs"
CREATE INDEX "sales_promo_progs_tid_priority_idx" ON "sales_promotion_programs" ("priority");
-- Create index "sales_promo_progs_tid_excl_group_idx" to table: "sales_promotion_programs"
CREATE INDEX "sales_promo_progs_tid_excl_group_idx" ON "sales_promotion_programs" ("exclusive_group");
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
  CONSTRAINT "sales_promo_cond_groups_program_id_fkey" FOREIGN KEY ("sales_promotion_program_id") REFERENCES "sales_promotion_programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
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
