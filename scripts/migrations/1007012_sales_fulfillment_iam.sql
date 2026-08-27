-- IAM for fulfilment requests (SALES-029).
--
-- Both resources get `read` alone. Requests are raised by confirm and by the return workflow, and a
-- client able to write one could tell Inventory to move goods that no sale asked for - which is the
-- one thing BR 44's separation exists to prevent.

DO $$
BEGIN
	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_resources') THEN
		INSERT INTO "iam_resources" ("id","name","code","description","owner_type","max_scope","min_scope","created_at","etag") VALUES
		('01M3SALES00000000000000053', 'Sales Fulfilment Request', 'sales_fulfillment_request', 'One thing Sales asked Inventory to do for a sale', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000054', 'Sales Fulfilment Request Line', 'sales_fulfillment_request_line', 'One order line covered by a fulfilment request', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema='public' AND table_name='iam_actions') THEN
		INSERT INTO "iam_actions" ("id","name","code","description","resource_id","etag") VALUES
		('01M3SALES00000000000000055', 'Read', 'read', NULL, '01M3SALES00000000000000053', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES00000000000000056', 'Read', 'read', NULL, '01M3SALES00000000000000054', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
