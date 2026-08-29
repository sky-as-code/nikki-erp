-- IAM for returns, return lines and refund legs.
--
-- `process_return` is its own action rather than `update` because it moves money out of the
-- business: a role that may correct a return's reason must not thereby be able to release its
-- refund.
--
-- sales_return_line and sales_refund_payment are read-only to clients. Lines are written by
-- create_return, which caps each quantity at what is still returnable and prices it from the
-- historical amounts; refund legs are written by the return workflow, which caps each at what its
-- original payment captured. A writable row either way is an outflow with no matching inflow.

DO $$
BEGIN
	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_resources') THEN
		INSERT INTO "iam_resources" ("id","name","code","description","owner_type","max_scope","min_scope","created_at","etag") VALUES
		('01M3SALES00000000000000076', 'Sales Return', 'sales_return', 'A customer sending goods back, and the refund that settles it', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000080', 'Sales Return Line', 'sales_return_line', 'One order line coming back, in whole or in part', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000082', 'Sales Refund Payment', 'sales_refund_payment', 'One leg of a refund, against one original payment', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_actions') THEN
		-- The return document: readable, creatable, correctable, and archivable.
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		('01M3SALES00000000000000077', 'Read', 'read', NULL, '01M3SALES00000000000000076', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000078', 'Create', 'create', 'Record that a customer wants to send something back', '01M3SALES00000000000000076', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000079', 'Update', 'update', 'Correct a return, or cancel one before anything has moved', '01M3SALES00000000000000076', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- The power that moves money and goods. Separate from `update` on purpose: this is the one
		-- that cannot be undone once it starts.
		('01M3SALES0000000000000007A', 'Process return', 'process_return', 'Send the goods back, refund the customer and adjust the VAT invoice', '01M3SALES00000000000000076', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Read only; see the header.
		('01M3SALES00000000000000081', 'Read', 'read', NULL, '01M3SALES00000000000000080', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000083', 'Read', 'read', NULL, '01M3SALES00000000000000082', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
