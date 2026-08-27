-- Fulfilment requests (BR 44, 45, 87.8).
--
-- SALES SENDS INTENT AND NEVER TOUCHES STOCK (BR 3.2, tested by BR 94.6). These tables record what
-- was asked of Inventory and what it answered. Inventory decides availability, warehouse, location
-- and the movements; nothing here is a stock quantity.
--
-- ONE ORDER HAS MANY REQUESTS (BR 45). Partial delivery, several warehouses and split shipment each
-- produce another, which is why this is a table rather than a fulfillment_reference column on the
-- order - a single column would have to be rewritten on the second shipment, losing the first.
--
-- ACCEPTED IS NOT COMPLETED. An accepted request has stock HELD; a completed one has goods MOVED.
-- Only completed requests feed sales_order_lines.fulfilled_quantity, because BR 7.3's failure -
-- money captured, goods not dispensed - lives exactly between the two.
--
-- Also adds sales_order_lines.requires_fulfillment: false for a service or a fee, which is what
-- lets an order of only such lines reach fulfillment_status 'not_required' rather than sitting
-- pending forever (D-14). It defaults TRUE because the two mistakes are not symmetric - a line
-- wrongly needing goods holds an order open until somebody notices, while one wrongly needing none
-- reports a sale as shipped that never was.

ALTER TABLE "sales_order_lines"
  ADD COLUMN IF NOT EXISTS "requires_fulfillment" boolean NOT NULL DEFAULT true;

CREATE TABLE "sales_fulfillment_requests" ("id" character varying NOT NULL, "org_id" character varying NOT NULL, "sales_order_id" character varying NOT NULL, "request_type" character varying NOT NULL, "status" character varying NOT NULL, "inventory_reference" character varying NULL, "failure_reason" character varying NULL, "requested_at" timestamptz NULL, "completed_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NULL, "etag" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "sales_fulfillment_requests_sales_order_id_fkey" FOREIGN KEY ("sales_order_id") REFERENCES "sales_orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

CREATE INDEX "sales_fulfil_reqs_tid_order_idx" ON "sales_fulfillment_requests" ("sales_order_id");

CREATE INDEX "sales_fulfil_reqs_tid_status_idx" ON "sales_fulfillment_requests" ("status");

CREATE INDEX "sales_fulfil_reqs_tid_invref_idx" ON "sales_fulfillment_requests" ("inventory_reference");

CREATE TABLE "sales_fulfillment_request_lines" ("id" character varying NOT NULL, "org_id" character varying NOT NULL, "sales_fulfillment_request_id" character varying NOT NULL, "sales_order_line_id" character varying NOT NULL, "quantity" numeric NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NULL, "etag" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "sales_fulfil_lines_tid_req_ordline_ukey" UNIQUE ("sales_fulfillment_request_id", "sales_order_line_id"), CONSTRAINT "sales_fulfil_req_lines_request_id_fkey" FOREIGN KEY ("sales_fulfillment_request_id") REFERENCES "sales_fulfillment_requests" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "sales_fulfillment_request_lines_sales_order_line_id_fkey" FOREIGN KEY ("sales_order_line_id") REFERENCES "sales_order_lines" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

CREATE INDEX "sales_fulfil_lines_tid_req_idx" ON "sales_fulfillment_request_lines" ("sales_fulfillment_request_id");

CREATE INDEX "sales_fulfil_lines_tid_ordline_idx" ON "sales_fulfillment_request_lines" ("sales_order_line_id");
