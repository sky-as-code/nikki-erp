-- IAM resources and actions for the Inventory Stock movement engine (stock BR §4.2.3 to §4.2.5).
--
-- The dynamic resource engine asserts permissions using the schema name as the resource code, so
-- these codes must stay byte-identical to the "inventory_stock_transfer" / "inventory_stock_move" /
-- "inventory_stock_move_line" / "inventory_stock_move_dependency" schema names. A code that drifts
-- from its schema denies every request, with nothing in the response pointing at the seed.
--
-- The six movement operations are seeded as SEPARATE actions rather than folded into "update",
-- because they are materially different powers. Validating a transfer moves real goods and cannot
-- be undone by an edit — the correction is a reverse transfer — while updating one changes a note.
-- A role that may do the second should not thereby be able to do the first.
--
-- Stock Move Line is read-only to clients, like the quant in 0005006: its rows are allocation
-- decisions written by the reservation engine, so its engine refuses create, update and delete and
-- no write actions are seeded. Seeding one would advertise a capability the engine rejects.
--
-- Deliberately no iam_entitlements rows, matching 0005006_inventory_stock_seeds.sql. Stock movement
-- is business data whose visibility follows explicitly assigned roles, and a blanket grant would
-- silently expose every transfer in the system.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M0B434KTKF2YBXPBBC2JH9XB', 'Stock Transfer', 'inventory_stock_transfer', 'The header of a stock transaction: what moves, between which locations, in what state', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4X4YENXT9PN8J703BQX4Q', 'Stock Move', 'inventory_stock_move', 'One line of demand within a transfer: this much of this variant, from here to there', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4WMQCNPQV5H6VRPCQHTGP', 'Stock Move Line', 'inventory_stock_move_line', 'The execution detail of a move: what was actually taken, and from which balance', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4JBZWQZ6PBAYTGAPKWX3G', 'Stock Move Dependency', 'inventory_stock_move_dependency', 'Ordering between the steps of a multi-step flow', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Stock Transfer: CRUD, then the six movement operations.
		('01M0B4JE67S9ZY40RV3Z3E87R2', 'Create', 'create', NULL, '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4PYXJ7JESHQXT054DFVKA', 'Update', 'update', NULL, '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4G9ZMPANTTK8156YFKQM8', 'Delete', 'delete', NULL, '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B449HKX7JRZ7E18CCWQAGN', 'Read', 'read', NULL, '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4Z9QV2KWX6J42M238624G', 'Confirm', 'confirm', 'Commit a draft transfer to its demand, reserving stock when the operation type says so', '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4ZGXXJN7450DGZB5QD4YJ', 'Reserve', 'reserve', 'Claim stock for the transfer''s moves without moving any of it', '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4XPZR2MZ1210Q7Q59YVGS', 'Unreserve', 'unreserve', 'Release the transfer''s claims, leaving on-hand quantities untouched', '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4TPK5A179SW4KMBCQT5HY', 'Validate', 'validate', 'Execute the transfer: the only operation that changes an on-hand balance, and one no edit can undo', '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4JHJFQXQF7T7WT9PJ6A5Z', 'Cancel', 'cancel', 'Abandon an unfinished transfer, releasing whatever it holds', '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Stock Move
		('01M0B4SPVX86PAWSQHJ62WARZ4', 'Create', 'create', NULL, '01M0B4X4YENXT9PN8J703BQX4Q', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B46CHAF22AG5P8MSV3APSZ', 'Update', 'update', NULL, '01M0B4X4YENXT9PN8J703BQX4Q', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4WV9PR3JJWM5V1RSFBTK2', 'Delete', 'delete', NULL, '01M0B4X4YENXT9PN8J703BQX4Q', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B4SBHVJWKG3EXFYP6CDZHY', 'Read', 'read', NULL, '01M0B4X4YENXT9PN8J703BQX4Q', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Stock Move Line. Read only: its rows are written by the reservation engine, and a
		-- client-written allocation would be a claim the balance itself knows nothing about.
		('01M0B4KB8JA48CQRTMF6J1B9WY', 'Read', 'read', NULL, '01M0B4WMQCNPQV5H6VRPCQHTGP', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Stock Move Dependency
		('01M0B4GJM613QQCEKG6NVT8675', 'Create', 'create', NULL, '01M0B4JBZWQZ6PBAYTGAPKWX3G', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B44FMEYRB12NBBJPDHH40T', 'Delete', 'delete', NULL, '01M0B4JBZWQZ6PBAYTGAPKWX3G', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0B407H0YTHMAN6ZHPANDZBA', 'Read', 'read', NULL, '01M0B4JBZWQZ6PBAYTGAPKWX3G', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;
END $$;
