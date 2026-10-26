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
		('01M0CUR1QK4N7VZBX2M9TPE5RD', jsonb_build_object('en-US', 'Essential - Currency', 'vi-VN', 'Thiết yếu - Tiền tệ'), 'essential_currency', 'Currencies an amount may be denominated in', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M0CUR1QM6P9XB0Z4Q1VRG7TF', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M0CUR1QK4N7VZBX2M9TPE5RD', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0CUR1QN8R1ZD2B6S3XTJ9VG', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M0CUR1QK4N7VZBX2M9TPE5RD', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0CUR1QP0T3BF4D8V5ZWK1XH', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M0CUR1QK4N7VZBX2M9TPE5RD', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0CUR1QQ2V5DH6F0X7BYM3ZJ', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M0CUR1QK4N7VZBX2M9TPE5RD', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0CUR1QR4X7FK8H2Z9DAP5BK', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive a currency so it is out of the working set', '01M0CUR1QK4N7VZBX2M9TPE5RD', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- Languages: the catalogue is public to read (every page, the sign-in page included, renders from
-- it), so only the writes are actions. Tenant-wide reference data, hence tenant scope only.
DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M51EPZ8K0CA51CB8DX821SM3', jsonb_build_object('en-US', 'Essential - Language', 'vi-VN', 'Thiết yếu - Ngôn ngữ'), 'essential_language', 'Languages the application can be shown in', 'nikkierp', 'tenant', 'tenant', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M51EPZ8KZC4XAGSH6VMSQY2J', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M51EPZ8K0CA51CB8DX821SM3', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M51EPZ8K971CG6CDVQR2HM4C', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M51EPZ8K0CA51CB8DX821SM3', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M51EPZ8KAM7Q05W89H7KMFN4', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M51EPZ8K0CA51CB8DX821SM3', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
