-- IAM resources and actions for the Phase 3 corrections (stock BR §4.2.7 to §4.2.10).
--
-- The dynamic resource engine asserts permissions using the schema name as the resource code, so
-- "inventory_stock_scrap" here must stay byte-identical to the StockScrapSchemaName constant. A
-- code that drifts from its schema denies every request, with nothing in the response pointing at
-- the seed.
--
-- Three groups of actions are seeded:
--
--   * Stock Scrap gets CRUD plus "do_scrap". The write actions exist here, unlike the quant's,
--     because a scrap is a document a user raises and may abandon while it is draft. What the
--     derived service constrains is when: a done scrap can be neither edited nor deleted.
--
--   * Stock Quant gains five counting actions, on a resource whose create/update/delete are
--     deliberately absent from 0005006 and stay absent. That is not a contradiction: the missing
--     CRUD stops a client setting a balance with no movement behind it, and none of these actions
--     touches on_hand_quantity. Enter and reset write count metadata; apply changes the balance
--     only by generating a movement.
--
--   * Stock Transfer gains "create_return". Raising a return commits the company to taking goods
--     back, which is a commercial decision rather than an edit to a shipping document.
--
-- They are separate actions rather than folded into "update" for the reason 0005008 gives: they
-- are materially different powers. Applying an inventory adjustment writes stock that no trade
-- movement explains, which is the most sensitive thing this module lets anyone do, and a role that
-- may fix a typo in a count note should not thereby be able to rewrite a balance.
--
-- Deliberately no iam_entitlements rows, matching 0005006 and 0005008. Stock data's visibility
-- follows explicitly assigned roles, and a blanket grant would silently expose every scrap and
-- every balance in the system.
--
-- Note there is no seed for the inventory-loss location, the scrap location, or the
-- INV_CORRECTION operation type that adjustments and scraps generate their movements through.
-- Those are per-org business data — locations and operation types are created through their own
-- APIs, as Phase 1 established by seeding none — so a global migration is the wrong place for
-- them. The services resolve them per org and report a business violation naming what is missing
-- when an org has not configured them yet.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M0C3SQRTAAAAAAAAAAAAAAAB', 'Stock Scrap', 'inventory_stock_scrap', 'A document that removes goods from usable stock by moving them to a scrap location', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Stock Scrap: CRUD, then the operation that executes the document.
		('01M0C3SQRTAAAAAAAAAAAAAAAC', 'Create', 'create', NULL, '01M0C3SQRTAAAAAAAAAAAAAAAB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0C3SQRTAAAAAAAAAAAAAAAD', 'Update', 'update', NULL, '01M0C3SQRTAAAAAAAAAAAAAAAB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0C3SQRTAAAAAAAAAAAAAAAE', 'Delete', 'delete', NULL, '01M0C3SQRTAAAAAAAAAAAAAAAB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0C3SQRTAAAAAAAAAAAAAAAF', 'Read', 'read', NULL, '01M0C3SQRTAAAAAAAAAAAAAAAB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0C3SQRTAAAAAAAAAAAAAAAG', 'Do Scrap', 'do_scrap', 'Execute the scrap: generate the movement that writes the goods off. Cannot be undone by an edit.', '01M0C3SQRTAAAAAAAAAAAAAAAB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;

	-- Physical inventory and cycle counting, on the existing Stock Quant resource from 0005006.
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) AND EXISTS (
		SELECT FROM "iam_resources" WHERE "code" = 'inventory_stock_quant'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M0C3SQRTAAAAAAAAAAAAAAAH', 'Enter Count', 'enter_count', 'Record what a physical count found. Does not change the balance.', (SELECT "id" FROM "iam_resources" WHERE "code" = 'inventory_stock_quant'), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0C3SQRTAAAAAAAAAAAAAAAJ', 'Reset Count', 'reset_count', 'Abandon a pending count. Does not change the balance.', (SELECT "id" FROM "iam_resources" WHERE "code" = 'inventory_stock_quant'), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0C3SQRTAAAAAAAAAAAAAAAK', 'Apply Adjustment', 'apply_adjustment', 'Turn a pending count into a real balance change by generating an adjustment movement.', (SELECT "id" FROM "iam_resources" WHERE "code" = 'inventory_stock_quant'), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0C3SQRTAAAAAAAAAAAAABA', 'Schedule Count', 'schedule_count', 'Set when a balance is next due to be counted.', (SELECT "id" FROM "iam_resources" WHERE "code" = 'inventory_stock_quant'), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0C3SQRTAAAAAAAAAAAAABB', 'Assign Counter', 'assign_counter', 'Name who is responsible for counting a balance.', (SELECT "id" FROM "iam_resources" WHERE "code" = 'inventory_stock_quant'), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;

	-- Create Reverse Transfer, on the existing Stock Transfer resource from 0005008.
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) AND EXISTS (
		SELECT FROM "iam_resources" WHERE "code" = 'inventory_stock_transfer'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M0C3SQRTAAAAAAAAAAAAABC', 'Create Return', 'create_return', 'Raise a reverse transfer for goods coming back. Never alters the original transfer.', (SELECT "id" FROM "iam_resources" WHERE "code" = 'inventory_stock_transfer'), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;
END
$$;
