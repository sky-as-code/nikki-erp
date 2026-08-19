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
		('01M1C0N7ACTS00000000000001', 'Party', 'contacts_party', 'People and organizations the business deals with', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS00000000000002', 'Communication Channel', 'contacts_comm_channel', 'Ways of reaching a party: phone, email, postal address', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS00000000000003', 'Party Relationship', 'contacts_relationship', 'Directed links between two parties', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000K', 'Vendor Profile', 'contacts_vendor_profile', 'Supplier-specific terms of a party: status, payment terms, lead time', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Party
		('01M1C0N7ACTS00000000000004', 'Create', 'create', NULL, '01M1C0N7ACTS00000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS00000000000005', 'Update', 'update', NULL, '01M1C0N7ACTS00000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS00000000000006', 'Delete', 'delete', NULL, '01M1C0N7ACTS00000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS00000000000007', 'Read', 'read', NULL, '01M1C0N7ACTS00000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS00000000000008', 'Set archived status', 'set_archived', 'Archive a party so it is hidden from the working set', '01M1C0N7ACTS00000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Communication Channel
		('01M1C0N7ACTS00000000000009', 'Create', 'create', NULL, '01M1C0N7ACTS00000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000A', 'Update', 'update', NULL, '01M1C0N7ACTS00000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000B', 'Delete', 'delete', NULL, '01M1C0N7ACTS00000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000C', 'Read', 'read', NULL, '01M1C0N7ACTS00000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000D', 'Set archived status', 'set_archived', 'Archive a communication channel', '01M1C0N7ACTS00000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Relationship
		('01M1C0N7ACTS0000000000000E', 'Create', 'create', NULL, '01M1C0N7ACTS00000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000F', 'Update', 'update', NULL, '01M1C0N7ACTS00000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000G', 'Delete', 'delete', NULL, '01M1C0N7ACTS00000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000H', 'Read', 'read', NULL, '01M1C0N7ACTS00000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000J', 'Set archived status', 'set_archived', 'Archive a party relationship', '01M1C0N7ACTS00000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Vendor Profile
		('01M1C0N7ACTS0000000000000M', 'Create', 'create', NULL, '01M1C0N7ACTS0000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000N', 'Update', 'update', NULL, '01M1C0N7ACTS0000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000P', 'Delete', 'delete', NULL, '01M1C0N7ACTS0000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000Q', 'Read', 'read', NULL, '01M1C0N7ACTS0000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M1C0N7ACTS0000000000000R', 'Set archived status', 'set_archived', 'Archive a vendor profile so it cannot be selected for new orders', '01M1C0N7ACTS0000000000000K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
