-- IAM resources and actions for the whole Contacts module: Party, Communication Channel and
-- Relationship.
--
-- Contacts asserted NO permissions at all before this. It carried two disagreeing sets of resource
-- code constants — PascalCase "ContactsParty" in constants/resources.go and snake_case
-- "contacts_party" on the entity — and a repo-wide search found neither set had a single call site.
-- The application service asserted nothing, and no contacts IAM migration existed. Converting the
-- module to the dynamic resource engine is what makes these rows load-bearing: the engine asserts
-- permission on every request using the schema name as the resource code.
--
-- So these codes must stay byte-identical to the "contacts_party" / "contacts_comm_channel" /
-- "contacts_relationship" schema names. A code that drifts from its schema denies every request,
-- with nothing in the response pointing at the seed.
--
-- Deliberately NO iam_entitlements rows. Unit of Measure grants the system "User" role a
-- domain-wide read so that any user can pick a unit while filling in an unrelated form. The contact
-- book is not comparable: a party record carries a supplier's legal name, tax registration number
-- and address, a communication channel carries a private phone number, and a vendor profile with
-- payment terms hangs off the party. A blanket grant would expose all of it to every authenticated
-- account, and nothing else in the test tree would notice. Access follows explicitly assigned
-- roles, which is the same choice Products made.
--
-- That decision is invisible until someone hits an unexpected 403, so it is pinned by an API test:
-- nikkierp/tests/api-rest/contacts/party/09_permissions.robot asserts that an account holding only
-- the system User role is refused, and that the same account can still read UoM.
--
-- All three resources get the same five actions. None of them has a lifecycle beyond archiving:
-- there is no state machine on a contact, so there is no operation to separate from update the way
-- Inventory separates validating a transfer from editing one.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M1C0N7ACTS00000000000001', jsonb_build_object('en-US', 'Contacts - Party', 'vi-VN', 'Danh bạ - Đối tác'), 'contacts_party', 'People and organizations the business deals with', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS00000000000002', jsonb_build_object('en-US', 'Contacts - Communication Channel', 'vi-VN', 'Danh bạ - Kênh liên lạc'), 'contacts_comm_channel', 'Ways of reaching a party: phone, email, postal address', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS00000000000003', jsonb_build_object('en-US', 'Contacts - Party Relationship', 'vi-VN', 'Danh bạ - Quan hệ đối tác'), 'contacts_relationship', 'Directed links between two parties', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000K', jsonb_build_object('en-US', 'Contacts - Vendor Profile', 'vi-VN', 'Danh bạ - Hồ sơ nhà cung cấp'), 'contacts_vendor_profile', 'Supplier-specific terms of a party: status, payment terms, lead time', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Party
		('01M1C0N7ACTS00000000000004', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M1C0N7ACTS00000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS00000000000005', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M1C0N7ACTS00000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS00000000000006', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M1C0N7ACTS00000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS00000000000007', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M1C0N7ACTS00000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS00000000000008', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive a party so it is hidden from the working set', '01M1C0N7ACTS00000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Communication Channel
		('01M1C0N7ACTS00000000000009', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M1C0N7ACTS00000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000A', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M1C0N7ACTS00000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000B', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M1C0N7ACTS00000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000C', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M1C0N7ACTS00000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000D', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive a communication channel', '01M1C0N7ACTS00000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Relationship
		('01M1C0N7ACTS0000000000000E', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M1C0N7ACTS00000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000F', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M1C0N7ACTS00000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000G', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M1C0N7ACTS00000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000H', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M1C0N7ACTS00000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000J', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive a party relationship', '01M1C0N7ACTS00000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Vendor Profile
		('01M1C0N7ACTS0000000000000M', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M1C0N7ACTS0000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000N', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M1C0N7ACTS0000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000P', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M1C0N7ACTS0000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000Q', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M1C0N7ACTS0000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000R', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive a vendor profile so it cannot be selected for new orders', '01M1C0N7ACTS0000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
