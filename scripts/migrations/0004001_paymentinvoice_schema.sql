-- Create "paymentinvoice_payment_methods" table
CREATE TABLE "paymentinvoice_payment_methods" (
  "id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "adapter_code" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "description" jsonb NULL,
  "currency_id" character varying NULL,
  "min_amount" numeric NULL,
  "max_amount" numeric NULL,
  "is_active" boolean NOT NULL,
  "config" jsonb NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "payinv_pay_methods_code_ukey" UNIQUE ("code")
);
-- Create index "payinv_pay_methods_adapter_idx" to table: "paymentinvoice_payment_methods"
CREATE INDEX "payinv_pay_methods_adapter_idx" ON "paymentinvoice_payment_methods" ("adapter_code");
-- Create index "payinv_pay_methods_is_active_idx" to table: "paymentinvoice_payment_methods"
CREATE INDEX "payinv_pay_methods_is_active_idx" ON "paymentinvoice_payment_methods" ("is_active");
-- Create "paymentinvoice_payment_profiles" table
CREATE TABLE "paymentinvoice_payment_profiles" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "method" character varying NOT NULL,
  "encrypted_config" character varying NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "payinv_pay_profiles_method_idx" to table: "paymentinvoice_payment_profiles"
CREATE INDEX "payinv_pay_profiles_method_idx" ON "paymentinvoice_payment_profiles" ("method");
-- Create index "payinv_pay_profiles_org_idx" to table: "paymentinvoice_payment_profiles"
CREATE INDEX "payinv_pay_profiles_org_idx" ON "paymentinvoice_payment_profiles" ("org_id");
-- Create "paymentinvoice_orders" table
CREATE TABLE "paymentinvoice_orders" (
  "id" character varying NOT NULL,
  "order_id" character varying NOT NULL,
  "order_code" character varying NOT NULL,
  "source" character varying NOT NULL,
  "status" character varying NOT NULL,
  "amount" numeric NOT NULL,
  "refund_amount" numeric NOT NULL,
  "currency_id" character varying NOT NULL,
  "payment_method_id" character varying NOT NULL,
  "payment_profile_id" character varying NULL,
  "content" character varying NULL,
  "return_url" character varying NULL,
  "last_sync_status" character varying NULL,
  "sync_logs" jsonb NULL,
  "metadata" jsonb NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "payinv_orders_order_code_ukey" UNIQUE ("order_code"),
  CONSTRAINT "payinv_orders_order_id_ukey" UNIQUE ("order_id"),
  CONSTRAINT "paymentinvoice_orders_payment_method_id_fkey" FOREIGN KEY ("payment_method_id") REFERENCES "paymentinvoice_payment_methods" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "payinv_orders_pay_method_idx" to table: "paymentinvoice_orders"
CREATE INDEX "payinv_orders_pay_method_idx" ON "paymentinvoice_orders" ("payment_method_id");
-- Create index "payinv_orders_status_created_idx" to table: "paymentinvoice_orders"
CREATE INDEX "payinv_orders_status_created_idx" ON "paymentinvoice_orders" ("status", "created_at");
-- Create index "payinv_orders_sync_status_idx" to table: "paymentinvoice_orders"
CREATE INDEX "payinv_orders_sync_status_idx" ON "paymentinvoice_orders" ("last_sync_status");
-- Create "paymentinvoice_invoices" table
CREATE TABLE "paymentinvoice_invoices" (
  "id" character varying NOT NULL,
  "number" character varying NULL,
  "status" character varying NOT NULL,
  "order_id" character varying NULL,
  "partner_name" character varying NOT NULL,
  "partner_tax_code" character varying NULL,
  "partner_address" character varying NULL,
  "currency_id" character varying NOT NULL,
  "subtotal_amount" numeric NOT NULL,
  "tax_amount" numeric NOT NULL,
  "total_amount" numeric NOT NULL,
  "issued_at" timestamptz NULL,
  "note" character varying NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "paymentinvoice_invoices_number_ukey" UNIQUE ("number"),
  CONSTRAINT "paymentinvoice_invoices_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "paymentinvoice_orders" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "payinv_invoices_issued_at_idx" to table: "paymentinvoice_invoices"
CREATE INDEX "payinv_invoices_issued_at_idx" ON "paymentinvoice_invoices" ("issued_at");
-- Create index "payinv_invoices_order_id_idx" to table: "paymentinvoice_invoices"
CREATE INDEX "payinv_invoices_order_id_idx" ON "paymentinvoice_invoices" ("order_id");
-- Create index "payinv_invoices_status_idx" to table: "paymentinvoice_invoices"
CREATE INDEX "payinv_invoices_status_idx" ON "paymentinvoice_invoices" ("status");
-- Create "paymentinvoice_invoice_lines" table
CREATE TABLE "paymentinvoice_invoice_lines" (
  "id" character varying NOT NULL,
  "invoice_id" character varying NOT NULL,
  "description" character varying NOT NULL,
  "quantity" integer NOT NULL,
  "unit_price" numeric NOT NULL,
  "tax_rate_percent" numeric NOT NULL,
  "amount" numeric NOT NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "paymentinvoice_invoice_lines_invoice_id_fkey" FOREIGN KEY ("invoice_id") REFERENCES "paymentinvoice_invoices" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "payinv_inv_lines_invoice_id_idx" to table: "paymentinvoice_invoice_lines"
CREATE INDEX "payinv_inv_lines_invoice_id_idx" ON "paymentinvoice_invoice_lines" ("invoice_id");
-- Create "paymentinvoice_transactions" table
CREATE TABLE "paymentinvoice_transactions" (
  "id" character varying NOT NULL,
  "order_id" character varying NOT NULL,
  "order_business_id" character varying NOT NULL,
  "status" character varying NOT NULL,
  "amount" numeric NOT NULL,
  "currency_id" character varying NOT NULL,
  "payment_method_id" character varying NOT NULL,
  "transaction_type" character varying NOT NULL,
  "content" character varying NULL,
  "ref_transaction_id" character varying NULL,
  "ref_payload" jsonb NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "paymentinvoice_transactions_order_id_fkey" FOREIGN KEY ("order_id") REFERENCES "paymentinvoice_orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "paymentinvoice_transactions_payment_method_id_fkey" FOREIGN KEY ("payment_method_id") REFERENCES "paymentinvoice_payment_methods" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "payinv_trans_order_biz_id_idx" to table: "paymentinvoice_transactions"
CREATE INDEX "payinv_trans_order_biz_id_idx" ON "paymentinvoice_transactions" ("order_business_id");
-- Create index "payinv_trans_order_id_idx" to table: "paymentinvoice_transactions"
CREATE INDEX "payinv_trans_order_id_idx" ON "paymentinvoice_transactions" ("order_id");
-- Create index "payinv_trans_ref_tran_id_idx" to table: "paymentinvoice_transactions"
CREATE INDEX "payinv_trans_ref_tran_id_idx" ON "paymentinvoice_transactions" ("ref_transaction_id");

