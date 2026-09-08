-- Structured demand references and expiry on stock transfers.
--
-- A reservation held for another module could previously only be found again by the transfer id the
-- caller remembered. origin_reference exists, but it is free text naming an upstream document for a
-- human to read; nothing can resolve it. These three columns are the machine-readable counterpart:
--
--   source_type    what kind of demand, as '{module}_{concept}' — 'sales_fulfillment'
--   source_id      which one, on the transfer
--   source_item_id which line of it, on each move
--
-- The triple is what lets a caller release or reallocate exactly the hold it created, and lets a
-- partial dispense be attributed line by line rather than guessed at. All three are immutable and
-- carry no foreign key: the rows they name belong to another module, and a constraint across that
-- boundary would couple the two schemas' migrations.
--
-- expires_at is when a hold stops holding. Null, the ordinary case, means it does not lapse on a
-- timer — a transfer about to be validated needs no expiry. A value is what a customer-controlled
-- hold needs, or one unclaimed order keeps a kiosk slot out of the sellable pool indefinitely. The
-- sweep that acts on it releases stock and nothing else: expiring a hold is not cancelling the
-- demand behind it, and refunds nobody.
--
-- Both indexes serve that sweep and the lookup by demand; neither is unique, because one demand
-- legitimately produces several transfers over its life (a reservation, then its goods issue).

-- Modify "inventory_stock_transfers" table
ALTER TABLE "inventory_stock_transfers" ADD COLUMN IF NOT EXISTS "source_type" character varying NULL;
ALTER TABLE "inventory_stock_transfers" ADD COLUMN IF NOT EXISTS "source_id" character varying NULL;
ALTER TABLE "inventory_stock_transfers" ADD COLUMN IF NOT EXISTS "expires_at" timestamptz NULL;
-- Create index "invty_stock_trfs_source_idx" to table: "inventory_stock_transfers"
CREATE INDEX IF NOT EXISTS "invty_stock_trfs_source_idx" ON "inventory_stock_transfers" ("source_type", "source_id");
-- Create index "invty_stock_trfs_expires_at_idx" to table: "inventory_stock_transfers"
CREATE INDEX IF NOT EXISTS "invty_stock_trfs_expires_at_idx" ON "inventory_stock_transfers" ("expires_at");
-- Modify "inventory_stock_moves" table
ALTER TABLE "inventory_stock_moves" ADD COLUMN IF NOT EXISTS "source_item_id" character varying NULL;
