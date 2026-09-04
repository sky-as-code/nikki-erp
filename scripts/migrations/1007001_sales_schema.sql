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
  "adjusted_by_order_id" character varying NULL,
  "adjusts_order_id" character varying NULL,
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
  "exchange_of_return_id" character varying NULL,
  "external_reference" character varying NULL,
  "idempotency_key" character varying NULL,
  "confirmed_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
  "tax_snapshot" jsonb NULL,
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
  "uom_id" character varying NOT NULL,
  "ordered_quantity" numeric NOT NULL,
  "requires_fulfillment" boolean NOT NULL,
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
-- Create index "sales_payments_tid_bill_idx" to table: "sales_payments"
CREATE INDEX "sales_payments_tid_bill_idx" ON "sales_payments" ("sales_bill_id");
-- Create index "sales_payments_tid_method_idx" to table: "sales_payments"
CREATE INDEX "sales_payments_tid_method_idx" ON "sales_payments" ("payment_method_id");
-- Create index "sales_payments_tid_pmtorder_idx" to table: "sales_payments"
CREATE INDEX "sales_payments_tid_pmtorder_idx" ON "sales_payments" ("payment_order_id");
-- Create index "sales_payments_tid_status_idx" to table: "sales_payments"
CREATE INDEX "sales_payments_tid_status_idx" ON "sales_payments" ("status");
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
  CONSTRAINT "sales_pricelist_items_sales_pricelist_id_fkey" FOREIGN KEY ("sales_pricelist_id") REFERENCES "sales_pricelists" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
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
  CONSTRAINT "sales_quotation_lines_sales_quotation_id_fkey" FOREIGN KEY ("sales_quotation_id") REFERENCES "sales_quotations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
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
  "inventory_disposition" character varying NULL,
  "refund_total" numeric NOT NULL,
  "inventory_reference" character varying NULL,
  "failure_reason" character varying NULL,
  "requested_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
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
  "refund_amount" numeric NOT NULL,
  "refund_tax_amount" numeric NOT NULL,
  "requires_inventory_return" boolean NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_return_lines_tid_uniq_ukey" UNIQUE ("sales_return_id", "sales_order_line_id"),
  CONSTRAINT "sales_return_lines_sales_order_line_id_fkey" FOREIGN KEY ("sales_order_line_id") REFERENCES "sales_order_lines" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "sales_return_lines_sales_return_id_fkey" FOREIGN KEY ("sales_return_id") REFERENCES "sales_returns" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "sales_return_lines_tid_ordline_idx" to table: "sales_return_lines"
CREATE INDEX "sales_return_lines_tid_ordline_idx" ON "sales_return_lines" ("sales_order_line_id");
-- Create index "sales_return_lines_tid_return_idx" to table: "sales_return_lines"
CREATE INDEX "sales_return_lines_tid_return_idx" ON "sales_return_lines" ("sales_return_id");
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
