-- Permissions for the product-stock integration.
--
-- Two things are added. The new configuration resource, which holds a product line's inventory
-- unit; and one read action on the existing Stock Quant resource, covering the summaries a
-- product page shows.
--
-- The reads are one permission rather than six. They are the same power — seeing how much stock a
-- product has — sliced only by which product is being looked at, so separating them would let an
-- administrator grant a nonsensical combination such as "may see a template's total but not the
-- variants behind it". Granting it lets a user look; it lets them change nothing, because every
-- action behind it is a read.

DO $$
BEGIN
	-- ---------------------------------------------------------------------------------------
	-- Stock's settings for a product line (CR §11.4).
	-- ---------------------------------------------------------------------------------------
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M2A7QK3P5NXCW9VBDT4RGZH2', 'Stock Product Configuration', 'inventory_stock_product_config', 'Stock settings of a product line, currently the unit its balances are counted in', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M2A7QK3PB8YFHV5JQW6E2SND', 'Create', 'create', 'Set the unit a product line''s stock is counted in', '01M2A7QK3P5NXCW9VBDT4RGZH2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2A7QK3PC4TZR7KDXA9HM3VE', 'Read', 'read', NULL, '01M2A7QK3P5NXCW9VBDT4RGZH2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Changing the unit after stock has moved would reinterpret every quantity ever recorded,
		-- so the engine refuses it outright. This permission covers the case where it is still
		-- allowed: a product that has never been used.
		('01M2A7QK3PDGX2W8N4VJQ5F7BK', 'Update', 'update', 'Change the unit, which the engine permits only while the product has no stock or stock history', '01M2A7QK3P5NXCW9VBDT4RGZH2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2A7QK3PEH6RQY3TZB8KCW9M', 'Delete', 'delete', NULL, '01M2A7QK3P5NXCW9VBDT4RGZH2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	-- ---------------------------------------------------------------------------------------
	-- The product-facing stock reads, on the existing Stock Quant resource.
	-- ---------------------------------------------------------------------------------------
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) AND EXISTS (
		SELECT FROM "iam_resources" WHERE "code" = 'inventory_stock_quant'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag")
		SELECT '01M2A7QK3PF9J5MXV2NQ8YT4HR', 'Read Product Stock', 'read_product_stock', 'See how much stock a product has, where it sits and what would be stranded by archiving it. Reads only.', "id", (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text
		FROM "iam_resources" WHERE "code" = 'inventory_stock_quant'
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
