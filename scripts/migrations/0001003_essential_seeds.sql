-- IAM resources, actions and default entitlements for Unit of Measure (BR-UOM-ESS §13).
--
-- The dynamic resource engine asserts permissions using the schema name as the resource code,
-- so these codes must stay identical to the "essential_uom" / "essential_uomcat" schema names.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01KZJ8Q3ZK7XN4WVBM2PDHR5T0', 'Unit of Measure', 'essential_uom', 'Units of measure and their conversion factors', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZJ8Q3ZKB8F1TC6RJWNQ4XS2', 'UoM Category', 'essential_uomcat', 'Categories that bound unit-of-measure conversion', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Unit of Measure
		('01KZJ8Q3ZKC2M9YHTV5KFAP7W3', 'Create', 'create', NULL, '01KZJ8Q3ZK7XN4WVBM2PDHR5T0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZJ8Q3ZKD5R0XQGN8JBSE1Y4', 'Update', 'update', NULL, '01KZJ8Q3ZK7XN4WVBM2PDHR5T0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZJ8Q3ZKE7S3ZPKW1MCTF6Z5', 'Delete', 'delete', NULL, '01KZJ8Q3ZK7XN4WVBM2PDHR5T0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZJ8Q3ZKF9T5A1NX3PDVG8B6', 'Read', 'read', NULL, '01KZJ8Q3ZK7XN4WVBM2PDHR5T0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZJ8Q3ZKG1V7B2QY4RFWH9C7', 'Set archived status', 'set_archived', 'Archive a UoM so it cannot be used by new transactions', '01KZJ8Q3ZK7XN4WVBM2PDHR5T0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- UoM Category
		('01KZJ8Q3ZKH3W9C4RZ6SGXJ0D8', 'Create', 'create', NULL, '01KZJ8Q3ZKB8F1TC6RJWNQ4XS2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZJ8Q3ZKJ5X1D6S08THYK2E9', 'Update', 'update', NULL, '01KZJ8Q3ZKB8F1TC6RJWNQ4XS2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZJ8Q3ZKK7Y3E8T1AVJZL4F0', 'Delete', 'delete', NULL, '01KZJ8Q3ZKB8F1TC6RJWNQ4XS2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZJ8Q3ZKM9Z5F0V3BWK1M6G1', 'Read', 'read', NULL, '01KZJ8Q3ZKB8F1TC6RJWNQ4XS2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_entitlements'
	) THEN
		-- BR-UOM-ESS §13.3: the system role every user holds gets read on both resources, so
		-- that any user can pick a UoM in a transaction form without a bespoke grant.
		INSERT INTO "iam_entitlements" (
			"id", "name", "description", "expression", "action_id", "resource_id", "role_id", "scope", "org_id", "org_unit_id", "is_archived", "created_at", "etag"
		) VALUES
		('01KZJ8Q3ZKN1A7G2W5CXL3N8H2', 'User - Read Units of Measure', 'Read units of measure', 'read:essential_uom:domain', '01KZJ8Q3ZKF9T5A1NX3PDVG8B6', '01KZJ8Q3ZK7XN4WVBM2PDHR5T0', '01KZJ5XRJDXSXZY0DKNNE6S086', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZJ8Q3ZKP3B9H4X7DYM5P0J3', 'User - Read UoM Categories', 'Read unit-of-measure categories', 'read:essential_uomcat:domain', '01KZJ8Q3ZKM9Z5F0V3BWK1M6G1', '01KZJ8Q3ZKB8F1TC6RJWNQ4XS2', '01KZJ5XRJDXSXZY0DKNNE6S086', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;
END $$;
