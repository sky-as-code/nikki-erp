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
-- Create "purchase_vendor_product_prices" table
CREATE TABLE "purchase_vendor_product_prices" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "vendor_id" character varying NOT NULL,
  "product_template_id" character varying NOT NULL,
  "product_variant_id" character varying NULL,
  "purchase_uom_id" character varying NOT NULL,
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
  PRIMARY KEY ("id")
);
-- Create index "purch_vpp_tid_validity_idx" to table: "purchase_vendor_product_prices"
CREATE INDEX "purch_vpp_tid_validity_idx" ON "purchase_vendor_product_prices" ("valid_from", "valid_to");
-- Create index "purch_vpp_tid_variant_idx" to table: "purchase_vendor_product_prices"
CREATE INDEX "purch_vpp_tid_variant_idx" ON "purchase_vendor_product_prices" ("product_variant_id");
-- Create index "purch_vpp_tid_vendor_tmpl_idx" to table: "purchase_vendor_product_prices"
CREATE INDEX "purch_vpp_tid_vendor_tmpl_idx" ON "purchase_vendor_product_prices" ("vendor_id", "product_template_id");
-- Create "purchase_agreement_lines" table
CREATE TABLE "purchase_agreement_lines" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "purchase_agreement_id" character varying NOT NULL,
  "sequence" integer NOT NULL,
  "product_variant_id" character varying NOT NULL,
  "uom_id" character varying NOT NULL,
  "quantity" numeric NOT NULL,
  "unit_price" numeric NOT NULL,
  "description" character varying NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "purchase_agreement_lines_purchase_agreement_id_fkey" FOREIGN KEY ("purchase_agreement_id") REFERENCES "purchase_agreements" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
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
  CONSTRAINT "purchase_order_lines_purchase_order_id_fkey" FOREIGN KEY ("purchase_order_id") REFERENCES "purchase_orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "purch_ord_lines_tid_ord_id_seq_idx" to table: "purchase_order_lines"
CREATE INDEX "purch_ord_lines_tid_ord_id_seq_idx" ON "purchase_order_lines" ("purchase_order_id", "sequence");
-- Create index "purch_ord_lines_tid_pvar_id_idx" to table: "purchase_order_lines"
CREATE INDEX "purch_ord_lines_tid_pvar_id_idx" ON "purchase_order_lines" ("product_variant_id");
-- Create index "purch_ord_lines_tid_vpp_id_idx" to table: "purchase_order_lines"
CREATE INDEX "purch_ord_lines_tid_vpp_id_idx" ON "purchase_order_lines" ("vendor_product_price_id");
