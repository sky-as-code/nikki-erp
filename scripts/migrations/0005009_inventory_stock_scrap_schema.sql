-- Stock corrections: the scrap document.
--
-- Phase 3 of Stock Management (ai-prompts/inventory/04-plan-phase3.md), covering BR §4.2.9.
-- Phases 1 and 2 shipped in 0005005/0005006 and 0005007/0005008; those files are already hashed
-- in atlas.sum, so the scrap table goes into a new pair rather than an edit to them.
--
-- Generated with `make gen-sql module=inventory` and split out of the full inventory DDL. The
-- coremart copy is this same DDL plus the tenant_id column its apptrait adds to every table, and
-- was generated from coremart/ so the index names are budgeted against the longer tenant-prefixed
-- form (see coremart/modules/vending_machine/domain/index_name_length_test.go).
--
-- Physical inventory and cycle counting need no DDL at all: their fields already exist on
-- inventory_stock_quants from 0005005, which is why this phase adds one table rather than four.
-- A count lives on the balance it describes (BR §4.2.7.2, decision F2), and applying it generates
-- an ordinary movement.
--
-- There is no is_archived column, per BR §4.2.9.2. A draft scrap is deleted outright and a done
-- scrap is permanent: the movement it generated is real, so hiding the document would leave that
-- movement unexplained. A scrap made in error is corrected by a reverse movement.
--
-- quantity is numeric, never floating point (BR §7.3).

-- Create "inventory_stock_scraps" table
CREATE TABLE "inventory_stock_scraps" ("id" character varying NOT NULL, "scrap_number" character varying NOT NULL, "origin_reference" character varying NULL, "transfer_id" character varying NULL, "product_variant_id" character varying NOT NULL, "base_uom_id" character varying NOT NULL, "lot_ref" character varying NULL, "package_ref" character varying NULL, "owner_ref" character varying NULL, "source_location_id" character varying NOT NULL, "scrap_location_id" character varying NOT NULL, "quantity" numeric NOT NULL, "reason_code" character varying NULL, "reason" character varying NULL, "status" character varying NOT NULL, "move_id" character varying NULL, "completed_at" timestamptz NULL, "note" character varying NULL, "org_id" character varying NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NULL, "etag" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "invty_stock_scraps_number_org_id_ukey" UNIQUE ("scrap_number", "org_id"), CONSTRAINT "inventory_stock_scraps_transfer_id_fkey" FOREIGN KEY ("transfer_id") REFERENCES "inventory_stock_transfers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "inventory_stock_scraps_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "inventory_stock_scraps_source_location_id_fkey" FOREIGN KEY ("source_location_id") REFERENCES "inventory_stock_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "inventory_stock_scraps_scrap_location_id_fkey" FOREIGN KEY ("scrap_location_id") REFERENCES "inventory_stock_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);
CREATE INDEX "invty_stock_scraps_status_idx" ON "inventory_stock_scraps" ("status");
CREATE INDEX "invty_stock_scraps_pvar_id_idx" ON "inventory_stock_scraps" ("product_variant_id");
CREATE INDEX "invty_stock_scraps_trf_id_idx" ON "inventory_stock_scraps" ("transfer_id");
