-- IAM for manual price overrides (SALES-039).
--
-- Two things are seeded here, and they sit on different resources.
--
-- `manual_discount` on SALES_ORDER is the power BR 87.4 gates: changing what a customer pays for
-- reasons outside the price list. It is deliberately NOT `update` — a role that may correct a
-- quantity should not thereby be able to discount a sale — and revoking takes the same permission as
-- granting, since withdrawing a discount changes the price just as granting one does.
--
-- `read` on SALES_MANUAL_DISCOUNT is the whole surface of the record itself. No create, no update,
-- no delete: rows are written by the grant operation and by nothing else, because a plain POST would
-- bypass every gate that makes an override auditable — the mandatory reason, the draft-only check,
-- and the audit entry recording both the old and the new price.

DO $$
BEGIN
	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_resources') THEN
		INSERT INTO "iam_resources" ("id","name","code","description","owner_type","max_scope","min_scope","created_at","etag") VALUES
		('01M3SALES00000000000000073', 'Sales Manual Discount', 'sales_manual_discount', 'An operator override of a sale price, with the reason it was granted', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_actions') THEN
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		('01M3SALES00000000000000074', 'Read', 'read', NULL, '01M3SALES00000000000000073', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;

		-- The power itself, on the ORDER, since that is what gets discounted. The resource id is
		-- written literally rather than looked up by code: the seed-coverage test parses these
		-- files without a database, so a SELECT-based insert would be invisible to it and an
		-- action could go unseeded without anything failing.
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		('01M3SALES00000000000000075', 'Manual discount', 'manual_discount', 'Change what a customer pays, for a stated reason outside the price list', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
