-- Quotations and their lines (BR 87.1).
--
-- A QUOTATION IS NOT A DRAFT ORDER, and that is why these are tables rather than a status on
-- sales_orders. The numbering settles it: a quotation that never becomes an order would leave a hole
-- in the order sequence, and several quotations to one customer would burn several order numbers
-- before one was accepted. Fiscal and accounting systems read that sequence, so the holes are not
-- cosmetic. The two documents also mean different things - an order is a commitment, a quotation is
-- an offer that lapses - and `expired` is a legitimate resting state for one and meaningless for the
-- other.
--
-- THE LINES ARE DERIVED FROM sales_order_lines, not shared with them. Same shape where the shape
-- means the same thing, and deliberately missing where it does not: no fulfilled_quantity, no
-- returned_quantity, no requires_fulfillment, because a quotation moves no goods. One unit_price
-- where an order line has base and effective, because a quotation states the offer rather than
-- explaining how it was reached - the explanation is rebuilt when the engine runs again at
-- conversion.
--
-- converted_sales_order_id IS THE IDEMPOTENCY of the conversion: set once, at acceptance, so a second
-- accept finds it and returns that order rather than creating another. Two orders from one
-- acceptance is two deliveries and two invoices. It carries NO foreign key, so an order removed by a
-- retention sweep does not take its quotation's history with it.
--
-- valid_until is stored rather than derived from a settings window, because the window may change
-- and a quotation already sent must keep the deadline the customer was actually given.

CREATE TABLE "sales_quotations" ("id" character varying NOT NULL, "org_id" character varying NOT NULL, "quotation_number" character varying NOT NULL, "sales_channel_id" character varying NOT NULL, "sales_point_id" character varying NULL, "customer_reference" character varying NULL, "currency_code" character varying NOT NULL, "status" character varying NOT NULL, "valid_until" timestamptz NULL, "subtotal" numeric NULL, "discount_total" numeric NULL, "tax_total" numeric NULL, "grand_total" numeric NULL, "converted_sales_order_id" character varying NULL, "sent_at" timestamptz NULL, "accepted_at" timestamptz NULL, "cancelled_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NULL, "etag" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "sales_quotations_tid_number_ukey" UNIQUE ("quotation_number"), CONSTRAINT "sales_quotations_sales_channel_id_fkey" FOREIGN KEY ("sales_channel_id") REFERENCES "sales_channels" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);

CREATE INDEX "sales_quotations_tid_channel_idx" ON "sales_quotations" ("sales_channel_id");

CREATE INDEX "sales_quotations_tid_status_idx" ON "sales_quotations" ("status");

CREATE INDEX "sales_quotations_tid_order_idx" ON "sales_quotations" ("converted_sales_order_id");

CREATE INDEX "sales_quotations_tid_valid_idx" ON "sales_quotations" ("valid_until");

CREATE TABLE "sales_quotation_lines" ("id" character varying NOT NULL, "org_id" character varying NOT NULL, "sales_quotation_id" character varying NOT NULL, "line_number" integer NOT NULL, "product_variant_id" character varying NULL, "product_code_snapshot" character varying NULL, "product_name_snapshot" character varying NULL, "uom_id" character varying NULL, "quantity" numeric NOT NULL, "unit_price" numeric NULL, "discount_amount" numeric NULL, "net_amount" numeric NULL, "tax_amount" numeric NULL, "final_amount" numeric NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NULL, "etag" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "sales_quotation_lines_tid_quot_lineno_ukey" UNIQUE ("sales_quotation_id", "line_number"), CONSTRAINT "sales_quotation_lines_sales_quotation_id_fkey" FOREIGN KEY ("sales_quotation_id") REFERENCES "sales_quotations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

CREATE INDEX "sales_quotation_lines_tid_variant_idx" ON "sales_quotation_lines" ("product_variant_id");
