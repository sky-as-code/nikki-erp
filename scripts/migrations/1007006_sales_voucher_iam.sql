-- IAM resources and actions for the voucher pair (SALES-022).
--
-- Two resources with deliberately different permission surfaces.
--
-- A voucher CODE is operator-managed master data and gets the full five actions: somebody runs
-- campaigns, and running one means creating codes, correcting a validity window, withdrawing a
-- mispriced offer and reading the result.
--
-- A voucher REDEMPTION gets `read` alone. It is a ledger the system writes as orders move, and every
-- non-read power over it is a way to corrupt what it records: a client able to create one could forge
-- a discount's provenance, one able to update one could release a hold another order is relying on,
-- and one able to delete one would break the usage count that is derived from these rows. The same
-- reasoning already applies to sales_order_events, sales_order_adjustments and
-- sales_order_line_components.
--
-- The codes must stay byte-identical to the schema names. A code that drifts denies every request to
-- that resource, with nothing in the 403 pointing at this file as the cause.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M3SALES00000000000000048', 'Sales Voucher Code', 'sales_voucher_code', 'A credential a customer presents to activate a promotion program', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000049', 'Sales Voucher Redemption', 'sales_voucher_redemption', 'The ledger of which orders consumed which voucher codes', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Sales Voucher Code: full CRUD, an operator manages campaigns.
		('01M3SALES0000000000000004A', 'Create', 'create', NULL, '01M3SALES00000000000000048', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004B', 'Update', 'update', NULL, '01M3SALES00000000000000048', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004C', 'Delete', 'delete', NULL, '01M3SALES00000000000000048', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004D', 'Read', 'read', NULL, '01M3SALES00000000000000048', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004E', 'Set archived status', 'set_archived', 'Withdraw a voucher code, or bring a withdrawn one back', '01M3SALES00000000000000048', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Sales Voucher Redemption: read alone. See the note at the top of this file.
		('01M3SALES0000000000000004F', 'Read', 'read', NULL, '01M3SALES00000000000000049', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Applying a voucher is a power over the ORDER, not over the voucher code.
		--
		-- It reserves a use and changes what the basket costs; the code itself is untouched master
		-- data. Seating it on sales_order is what lets a till take a discount at the counter without
		-- also holding write permission over campaign configuration.
		('01M3SALES0000000000000004G', 'Apply voucher', 'apply_voucher', 'Apply a voucher code to a draft order, reserving one of its uses', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Confirm and cancel are separate powers from update, deliberately.
		--
		-- Confirming commits the business to a price and redeems vouchers; cancelling unwinds a sale
		-- and hands voucher uses back. A role that may correct a line should not thereby be able to
		-- do either, which is what folding them into `update` would grant.
		('01M3SALES0000000000000004H', 'Confirm', 'confirm', 'Commit a draft sales order: freeze its prices and redeem its vouchers', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004J', 'Cancel', 'cancel', 'Cancel a sales order that has not been paid or fulfilled', '01M3SALES0000000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
