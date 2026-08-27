-- IAM for payments (SALES-027/028).
--
-- The resource gets `read` alone. Money is recorded through the record_payment operation, which
-- applies six gates a plain POST would bypass - including the two that ask paymentinvoice whether
-- the method may be used at all (CR 33). A writable payment resource would let a client record money
-- that no gateway ever took.
--
-- `pay` and `settle` hang off the BILL, because that is what they change.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M3SALES0000000000000004Z', 'Sales Payment', 'sales_payment', 'One movement of money against a bill', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M3SALES00000000000000050', 'Read', 'read', NULL, '01M3SALES0000000000000004Z', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000051', 'Record payment', 'pay', 'Take a payment against an open bill', '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000052', 'Settle', 'settle', 'Close a bill whose money is fully in', '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
