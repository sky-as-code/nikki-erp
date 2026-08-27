-- Exchange linkage on sales_orders (BR 87.5, SALES-041).
--
-- AN EXCHANGE IS A RETURN PLUS A NEW SALE, joined by this column — never a historical order edited to
-- name a different product. The original sale really did sell what it sold, and rewriting it would
-- restate a transaction the customer already paid for, invalidate any fiscal document issued against
-- it, and leave the returned goods unaccounted for.
--
-- NO FOREIGN KEY, deliberately, and for two reasons. sales_returns does not exist yet (Phase 6), so
-- there is nothing to point at; and even once it does, the constraint would be wrong here — a return
-- removed by a retention sweep must not take the replacement sale with it, and the two documents are
-- independently valid. This matches every other cross-document reference in Sales.
--
-- Nullable because nearly every order is an ordinary sale.

ALTER TABLE "sales_orders"
  ADD COLUMN IF NOT EXISTS "exchange_of_return_id" character varying NULL;

CREATE INDEX IF NOT EXISTS "sales_orders_tid_exchange_idx" ON "sales_orders" ("exchange_of_return_id");
