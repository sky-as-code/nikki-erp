-- Stock movement engine: transfers, moves, move lines and move dependencies.
--
-- Phase 2 of Stock Management (ai-prompts/inventory/02-plans-phase2.md), covering BR §4.2.3 to
-- §4.2.5. Phase 1 shipped the balance core in 0005005/0005006; those files are already hashed in
-- atlas.sum and applied, so the movement tables go into a new pair rather than an edit.
--
-- Generated with `make gen-sql module=inventory` and split out of the full inventory DDL.
-- Statements are ordered by dependency: transfers, then moves, then the lines and dependencies
-- that reference them. The coremart copy is this same DDL plus the tenant_id column its apptrait
-- adds to every table.
--
-- Quantities are numeric, never floating point (BR §7.3): a balance is the running total of many
-- movements, and binary floating point drifts over those additions.
--
-- idempotency_key deliberately carries no unique constraint. A partial unique would also index the
-- rows where it IS NULL, which would permit only one un-validated transfer per org — and every
-- transfer starts un-validated. Replay is detected by the service reading the key back under the
-- row lock that validate already holds (BR §8.7).

-- Create "inventory_stock_transfers" table
CREATE TABLE "inventory_stock_transfers" (
  "id" character varying NOT NULL,
  "transfer_number" character varying NOT NULL,
  "operation_type_id" character varying NOT NULL,
  "operation_code" character varying NOT NULL,
  "origin_reference" character varying NULL,
  "source_location_id" character varying NOT NULL,
  "destination_location_id" character varying NOT NULL,
  "status" character varying NOT NULL,
  "priority" integer NULL,
  "reservation_method" character varying NOT NULL,
  "backorder_policy" character varying NOT NULL,
  "shipping_policy" character varying NOT NULL,
  "scheduled_at" timestamptz NULL,
  "deadline_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
  "backorder_of_id" character varying NULL,
  "return_of_id" character varying NULL,
  "chain_group_id" character varying NULL,
  "idempotency_key" character varying NULL,
  "note" character varying NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_stock_trfs_tid_number_org_id_ukey" UNIQUE ("transfer_number", "org_id"),
  CONSTRAINT "inventory_stock_transfers_operation_type_id_fkey" FOREIGN KEY ("operation_type_id") REFERENCES "inventory_stock_operation_types" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_transfers_source_location_id_fkey" FOREIGN KEY ("source_location_id") REFERENCES "inventory_stock_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_transfers_destination_location_id_fkey" FOREIGN KEY ("destination_location_id") REFERENCES "inventory_stock_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_transfers_backorder_of_id_fkey" FOREIGN KEY ("backorder_of_id") REFERENCES "inventory_stock_transfers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
CREATE INDEX "invty_stock_trfs_status_sched_at_idx" ON "inventory_stock_transfers" ("status", "scheduled_at");
CREATE INDEX "invty_stock_trfs_org_id_idem_key_idx" ON "inventory_stock_transfers" ("org_id", "idempotency_key");
CREATE INDEX "invty_stock_trfs_op_type_id_status_idx" ON "inventory_stock_transfers" ("operation_type_id", "status");
CREATE INDEX "invty_stock_trfs_backorder_of_id_idx" ON "inventory_stock_transfers" ("backorder_of_id");
CREATE INDEX "invty_stock_trfs_chain_group_id_idx" ON "inventory_stock_transfers" ("chain_group_id");

-- Create "inventory_stock_moves" table
CREATE TABLE "inventory_stock_moves" (
  "id" character varying NOT NULL,
  "transfer_id" character varying NULL,
  "sequence" integer NULL,
  "product_variant_id" character varying NOT NULL,
  "uom_id" character varying NULL,
  "demand_quantity" numeric NOT NULL,
  "base_demand_quantity" numeric NOT NULL,
  "source_location_id" character varying NOT NULL,
  "destination_location_id" character varying NOT NULL,
  "final_location_id" character varying NULL,
  "status" character varying NOT NULL,
  "priority" integer NULL,
  "scheduled_at" timestamptz NULL,
  "deadline_at" timestamptz NULL,
  "reservation_date" timestamptz NULL,
  "picked" boolean NULL,
  "origin_move_id" character varying NULL,
  "is_inventory_adjustment" boolean NULL,
  "scrap_id" character varying NULL,
  "valuation_value" numeric NULL,
  "remaining_quantity" numeric NULL,
  "remaining_value" numeric NULL,
  "currency_id" character varying NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_stock_moves_transfer_id_fkey" FOREIGN KEY ("transfer_id") REFERENCES "inventory_stock_transfers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_moves_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_moves_source_location_id_fkey" FOREIGN KEY ("source_location_id") REFERENCES "inventory_stock_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_moves_destination_location_id_fkey" FOREIGN KEY ("destination_location_id") REFERENCES "inventory_stock_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
CREATE INDEX "invty_stock_moves_trf_id_sequence_idx" ON "inventory_stock_moves" ("transfer_id", "sequence");
CREATE INDEX "invty_stock_moves_pvar_id_status_idx" ON "inventory_stock_moves" ("product_variant_id", "status");
CREATE INDEX "invty_stock_moves_status_sched_at_idx" ON "inventory_stock_moves" ("status", "scheduled_at");
CREATE INDEX "invty_stock_moves_origin_move_id_idx" ON "inventory_stock_moves" ("origin_move_id");

-- Create "inventory_stock_move_lines" table
CREATE TABLE "inventory_stock_move_lines" (
  "id" character varying NOT NULL,
  "move_id" character varying NOT NULL,
  "transfer_id" character varying NULL,
  "product_variant_id" character varying NOT NULL,
  "uom_id" character varying NULL,
  "quantity" numeric NOT NULL,
  "base_quantity" numeric NOT NULL,
  "source_location_id" character varying NOT NULL,
  "destination_location_id" character varying NOT NULL,
  "lot_ref" character varying NOT NULL,
  "package_ref" character varying NOT NULL,
  "result_package_ref" character varying NOT NULL,
  "owner_ref" character varying NOT NULL,
  "picked" boolean NULL,
  "operation_at" timestamptz NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_stock_move_lines_move_id_fkey" FOREIGN KEY ("move_id") REFERENCES "inventory_stock_moves" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_move_lines_transfer_id_fkey" FOREIGN KEY ("transfer_id") REFERENCES "inventory_stock_transfers" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_move_lines_product_variant_id_fkey" FOREIGN KEY ("product_variant_id") REFERENCES "inventory_product_variants" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_move_lines_source_location_id_fkey" FOREIGN KEY ("source_location_id") REFERENCES "inventory_stock_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_move_lines_destination_location_id_fkey" FOREIGN KEY ("destination_location_id") REFERENCES "inventory_stock_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
CREATE INDEX "invty_stock_mvlines_move_id_idx" ON "inventory_stock_move_lines" ("move_id");
CREATE INDEX "invty_stock_mvlines_trf_id_idx" ON "inventory_stock_move_lines" ("transfer_id");
CREATE INDEX "invty_stock_mvlines_pvar_id_src_loc_idx" ON "inventory_stock_move_lines" ("product_variant_id", "source_location_id");

-- Create "inventory_stock_move_dependencies" table
CREATE TABLE "inventory_stock_move_dependencies" (
  "id" character varying NOT NULL,
  "predecessor_move_id" character varying NOT NULL,
  "successor_move_id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "invty_stock_mvdeps_tid_pred_succ_ukey" UNIQUE ("predecessor_move_id", "successor_move_id"),
  CONSTRAINT "inventory_stock_move_dependencies_predecessor_move_id_fkey" FOREIGN KEY ("predecessor_move_id") REFERENCES "inventory_stock_moves" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "inventory_stock_move_dependencies_successor_move_id_fkey" FOREIGN KEY ("successor_move_id") REFERENCES "inventory_stock_moves" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
CREATE INDEX "invty_stock_mvdeps_succ_move_id_idx" ON "inventory_stock_move_dependencies" ("successor_move_id");
