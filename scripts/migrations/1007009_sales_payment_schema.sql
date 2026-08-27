-- Payments against bills (BR 39-42, 75; CR 33/36/37/78).
--
-- THE UNIQUE ON (sales_bill_id, external_transaction_id) IS THE DEDUPE MECHANISM (D-29).
--
-- A gateway callback that arrives twice must not record the money twice, and a check-then-insert can
-- be raced by exactly the retry it is meant to catch. Postgres enforces this whatever any process
-- believes. It is a PARTIAL unique - null external_transaction_id is unconstrained - because a cash
-- payment has no provider id, and cash cannot arrive twice by retry anyway.
--
-- Only `captured` counts toward settling a bill. An authorization is a hold the provider may still
-- release; treating it as money in would settle a bill against funds that never arrived.
--
-- A refund is NOT a negative payment. It is its own resource (SALES-035) with its own lifecycle and
-- approval rules, which is why `amount` is constrained non-negative here.

CREATE TABLE "sales_payments" ("id" character varying NOT NULL, "org_id" character varying NOT NULL, "sales_bill_id" character varying NOT NULL, "payment_method_id" character varying NOT NULL, "payment_method_code_snapshot" character varying NULL, "amount" numeric NOT NULL, "currency_code" character varying NOT NULL, "status" character varying NOT NULL, "external_transaction_id" character varying NULL, "provider_reference" character varying NULL, "paid_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NULL, "etag" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "sales_payments_sales_bill_id_fkey" FOREIGN KEY ("sales_bill_id") REFERENCES "sales_bills" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

CREATE UNIQUE INDEX "sales_payments_tid_bill_extxn_ukey" ON "sales_payments" ("sales_bill_id", "external_transaction_id") WHERE "external_transaction_id" IS NOT NULL;

CREATE INDEX "sales_payments_tid_bill_idx" ON "sales_payments" ("sales_bill_id");

CREATE INDEX "sales_payments_tid_status_idx" ON "sales_payments" ("status");

CREATE INDEX "sales_payments_tid_method_idx" ON "sales_payments" ("payment_method_id");
