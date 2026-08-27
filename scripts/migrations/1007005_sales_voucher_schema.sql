-- The voucher pair: a code, and the ledger of its uses (BR 26, 30, 32, 82).
--
-- A voucher code carries NO rules of its own. It is a credential pointing at a promotion program
-- (BR 25.1), and every condition, reward, compatibility rule and priority lives on that program.
-- That is why BR 27's long list of required voucher capabilities needs no voucher-side columns.
--
-- The redemption table exists because BR 82 says the usage counter alone is insufficient: a counter
-- cannot say WHICH orders consumed a voucher, cannot hold a use while an order is still a draft, and
-- cannot be audited when a customer disputes a discount.
--
-- THE UNIQUE INDEX ON (voucher_code_id, sales_order_id) IS THE CONCURRENCY CONTROL.
--
-- BR 30 requires that two tills redeeming the last use of a usage_limit=1 voucher at the same instant
-- cannot both succeed. The framework's repository cannot express a conditional increment - Update
-- takes a field map, and AffectedCount is hardcoded to 1 - so the counter cannot be the safe point.
-- This index can be, because Postgres enforces it regardless of what any application process
-- believes: the reservation row is written first, and a colliding write loses. usage_count is
-- maintained as a derived cache of these rows.
--
-- See D-43 for what this does NOT close: a usage_limit above 1 can still be over-reserved across
-- separate orders, since their rows do not collide with each other.

CREATE TABLE "sales_voucher_codes" ("id" character varying NOT NULL, "org_id" character varying NOT NULL, "code" character varying NOT NULL, "sales_promotion_program_id" character varying NOT NULL, "valid_from" timestamptz NULL, "valid_until" timestamptz NULL, "usage_limit" integer NULL, "usage_count" integer NOT NULL, "status" character varying NOT NULL, "is_archived" boolean NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NULL, "etag" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "sales_voucher_codes_code_ukey" UNIQUE ("code"), CONSTRAINT "sales_voucher_codes_sales_promotion_program_id_fkey" FOREIGN KEY ("sales_promotion_program_id") REFERENCES "sales_promotion_programs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

CREATE INDEX "sales_vouchers_tid_program_idx" ON "sales_voucher_codes" ("sales_promotion_program_id");

CREATE INDEX "sales_vouchers_tid_status_idx" ON "sales_voucher_codes" ("status");

CREATE INDEX "sales_vouchers_tid_validity_idx" ON "sales_voucher_codes" ("valid_from", "valid_until");

CREATE TABLE "sales_voucher_redemptions" ("id" character varying NOT NULL, "org_id" character varying NOT NULL, "voucher_code_id" character varying NOT NULL, "sales_order_id" character varying NOT NULL, "status" character varying NOT NULL, "reserved_at" timestamptz NULL, "redeemed_at" timestamptz NULL, "released_at" timestamptz NULL, "reversed_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NULL, "etag" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "sales_vouch_redms_tid_voucher_order_ukey" UNIQUE ("voucher_code_id", "sales_order_id"), CONSTRAINT "sales_voucher_redemptions_voucher_code_id_fkey" FOREIGN KEY ("voucher_code_id") REFERENCES "sales_voucher_codes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "sales_voucher_redemptions_sales_order_id_fkey" FOREIGN KEY ("sales_order_id") REFERENCES "sales_orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

CREATE INDEX "sales_vouch_redms_tid_voucher_idx" ON "sales_voucher_redemptions" ("voucher_code_id");

CREATE INDEX "sales_vouch_redms_tid_order_idx" ON "sales_voucher_redemptions" ("sales_order_id");

CREATE INDEX "sales_vouch_redms_tid_status_idx" ON "sales_voucher_redemptions" ("status");
