-- Manual price overrides (BR 87.4).
--
-- THE BASE PRICE IS NEVER OVERWRITTEN. The line keeps what the catalogue and the pricelist gave it,
-- and the override rides on top as an adjustment. Overwriting sales_order_lines.base_unit_price
-- would be simpler and is wrong twice: BR 87.9's price explanation would show a chain starting from
-- a price the catalogue never charged, and an override would become indistinguishable from a
-- genuinely cheap product, so nobody could audit who had discounted what.
--
-- WHY A TABLE RATHER THAN A ROW IN sales_order_adjustments. Repricing DELETES the whole adjustment
-- chain and rewrites it from pricing-engine output, and confirming an order reprices it
-- unconditionally. An override written straight into the adjustments would therefore be erased
-- before the sale completed - silently, with the customer charged full price after being promised a
-- discount. These rows are ENGINE INPUT, replayed on every calculation, which is what makes them
-- survive.
--
-- discount_amount is POSITIVE, minimum zero. BR 87.4 authorises a discount; a negative value would
-- be a surcharge, and silently adding money to a customer's bill is the worst failure available in
-- this table. The engine caps it at what is actually owed, because discounting past the end is a
-- refund with its own money movement rather than a discount that overshot.
--
-- reason is MANDATORY and its minimum length is what enforces that. An override with no stated cause
-- is indistinguishable from an error, and it is the field an auditor asking why this customer paid
-- less than that one actually reads.
--
-- original_amount records what the price was when the operator decided. Stored rather than derived,
-- because repricing moves the surrounding numbers afterwards, so this is the only record of what
-- they were actually looking at.

CREATE TABLE "sales_manual_discounts" ("id" character varying NOT NULL, "org_id" character varying NOT NULL, "sales_order_id" character varying NOT NULL, "sales_order_line_id" character varying NULL, "discount_amount" numeric NOT NULL, "reason" character varying NOT NULL, "granted_by" character varying NULL, "original_amount" numeric NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NULL, "etag" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "sales_manual_discounts_sales_order_id_fkey" FOREIGN KEY ("sales_order_id") REFERENCES "sales_orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

CREATE INDEX "sales_manual_discs_tid_order_idx" ON "sales_manual_discounts" ("sales_order_id");

CREATE INDEX "sales_manual_discs_tid_line_idx" ON "sales_manual_discounts" ("sales_order_line_id");

CREATE INDEX "sales_manual_discs_tid_actor_idx" ON "sales_manual_discounts" ("granted_by");
