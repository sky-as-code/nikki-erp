-- IAM resources and actions for the IT Storage module.
--
-- The coreplatform tree carries a copy with a leading tenant_id column; ids and codes must match.
--
-- iam_resources.code is the SCHEMA NAME, byte-identical: the dynamic resource engine asserts
-- permissions using it as the resource code.
--
-- itstorage_space is system-managed: clients may only read it. itstorage_file is read and deleted
-- through the built-in routes and created by the streamed upload; its metadata is immutable, so it
-- has no update action.
--
-- The grants to the system "User" role (bare org-scoped read and create on itstorage_file) live in
-- 1001004_iam_authorize_seeds.sql, which creates that role and sorts after this file.
DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M5T2BJTACN4H60E8VA7FRE2R', jsonb_build_object('en-US', 'IT Storage - Space', 'vi-VN', 'Lưu trữ - Không gian lưu trữ'), 'itstorage_space', 'A system-managed folder of the storage tree', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M5T2BJTE29YK2T3SSE1QR94J', jsonb_build_object('en-US', 'IT Storage - File', 'vi-VN', 'Lưu trữ - Tập tin'), 'itstorage_file', 'An uploaded file and its stored object', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M5T2BJTGH4BWWPBKS8P3D55E', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M5T2BJTACN4H60E8VA7FRE2R', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M5T2BJTJE1Z13HT710JB1VSS', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M5T2BJTE29YK2T3SSE1QR94J', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M5T2BJTN4RWKKX74SCSEQVQ9', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M5T2BJTE29YK2T3SSE1QR94J', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M5T2BJTQXE2940KSWNEQ5BXW', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M5T2BJTE29YK2T3SSE1QR94J', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
