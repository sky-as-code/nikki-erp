-- IAM resource and actions for Essential's Currency master.
--
-- The dynamic resource engine asserts permissions using the schema name as the resource code, so
-- this code must stay identical to the "essential_currency" schema name. A code that drifts from
-- its schema denies every request, with nothing in the response pointing at the seed.
--
-- Currency gets set_archived but no lifecycle action of its own. Withdrawing a currency from use
-- is is_active, a plain field: amounts already recorded in it must stay readable either way, so
-- there is nothing for an action to do that a write does not already express.
--
-- Like Unit of Measure, the system role every user holds is granted domain-wide read. Any user
-- filling in a form that names an amount has to be able to pick the currency it is in, and a
-- currency list is not sensitive: it is the same public ISO 4217 data for every tenant.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M0CUR1QK4N7VZBX2M9TPE5RD', 'Currency', 'essential_currency', 'Currencies an amount may be denominated in', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M0CUR1QM6P9XB0Z4Q1VRG7TF', 'Create', 'create', NULL, '01M0CUR1QK4N7VZBX2M9TPE5RD', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0CUR1QN8R1ZD2B6S3XTJ9VG', 'Update', 'update', NULL, '01M0CUR1QK4N7VZBX2M9TPE5RD', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0CUR1QP0T3BF4D8V5ZWK1XH', 'Delete', 'delete', NULL, '01M0CUR1QK4N7VZBX2M9TPE5RD', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0CUR1QQ2V5DH6F0X7BYM3ZJ', 'Read', 'read', NULL, '01M0CUR1QK4N7VZBX2M9TPE5RD', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0CUR1QR4X7FK8H2Z9DAP5BK', 'Set archived status', 'set_archived', 'Archive a currency so it is out of the working set', '01M0CUR1QK4N7VZBX2M9TPE5RD', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_entitlements'
	) THEN
		INSERT INTO "iam_entitlements" (
			"id", "name", "description", "expression", "action_id", "resource_id", "role_id", "scope", "org_id", "org_unit_id", "is_archived", "created_at", "etag"
		) VALUES
		('01M0CUR1QS6Z9HM0K4B1FCR7DM', 'User - Read Currencies', 'Read currencies', 'read:essential_currency:domain', '01M0CUR1QQ2V5DH6F0X7BYM3ZJ', '01M0CUR1QK4N7VZBX2M9TPE5RD', '01KZJ5XRJDXSXZY0DKNNE6S086', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
