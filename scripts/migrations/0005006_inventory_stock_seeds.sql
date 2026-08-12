-- IAM resources and actions for the Inventory Stock submodule (stock BR §4.2.1, §4.2.2).
--
-- The dynamic resource engine asserts permissions using the schema name as the resource code, so
-- these codes must stay identical to the "inventory_stock_location" /
-- "inventory_stock_operation_type" / "inventory_stock_quant" schema names. A code that drifts
-- from its schema denies every request, with nothing in the response pointing at the seed.
--
-- Stock Quant is read-only by design: it is current state, not a document, and its balances are
-- the running total of completed movements. Its engine refuses create, update and delete, so no
-- write actions are seeded for it — seeding one would advertise a capability the engine rejects.
--
-- Deliberately no iam_entitlements rows, matching 0005003_inventory_product_seeds.sql. Unit of
-- Measure grants the system "User" role a domain-wide read so that any user can pick a unit while
-- filling in an unrelated form. Stock is business data whose visibility follows explicitly
-- assigned roles, so there is no blanket grant here.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M0B2Q5R34KTKF2YBXPBBC2JH', 'Stock Location', 'inventory_stock_location', 'Where stock is held, and the counterparties movements run against', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5R9XBX4YENXT9PN8J70', 'Stock Operation Type', 'inventory_stock_operation_type', 'Policy applied when processing a stock transfer', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5R3BQX4QWMQCNPQV5H6', 'Stock Quant', 'inventory_stock_quant', 'Current stock balance at one product, location, lot, package and owner', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Stock Location
		('01M0B2Q5RVRPCQHTGPJBZWQZ6P', 'Create', 'create', NULL, '01M0B2Q5R34KTKF2YBXPBBC2JH', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5RBAYTGAPKWX3GJE67S', 'Update', 'update', NULL, '01M0B2Q5R34KTKF2YBXPBBC2JH', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5R9ZY40RV3Z3E87R2PY', 'Delete', 'delete', NULL, '01M0B2Q5R34KTKF2YBXPBBC2JH', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5RXJ7JESHQXT054DFVK', 'Read', 'read', NULL, '01M0B2Q5R34KTKF2YBXPBBC2JH', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5RAG9ZMPANTTK8156YF', 'Set archived status', 'set_archived', 'Archive a stock location so new transfers cannot use it', '01M0B2Q5R34KTKF2YBXPBBC2JH', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Stock Operation Type
		('01M0B2Q5RKQM849HKX7JRZ7E18', 'Create', 'create', NULL, '01M0B2Q5R9XBX4YENXT9PN8J70', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5RCCWQAGNZ9QV2KWX6J', 'Update', 'update', NULL, '01M0B2Q5R9XBX4YENXT9PN8J70', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5R42M238624GZGXXJN7', 'Delete', 'delete', NULL, '01M0B2Q5R9XBX4YENXT9PN8J70', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5R450DGZB5QD4YJXPZR', 'Read', 'read', NULL, '01M0B2Q5R9XBX4YENXT9PN8J70', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B2Q5R2MZ1210Q7Q59YVGST', 'Set archived status', 'set_archived', 'Archive an operation type so new transfers cannot use it, while historical transfers still resolve it', '01M0B2Q5R9XBX4YENXT9PN8J70', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Stock Quant. Read only: balances change through an adjustment, transfer or scrap,
		-- each of which records the movement that justifies the new number (AC-STOCK-002).
		('01M0B2Q5RPK5A179SW4KMBCQT5', 'Read', 'read', NULL, '01M0B2Q5R3BQX4QWMQCNPQV5H6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;
END $$;
