-- IAM resources and actions for billing (SALES-024).
--
-- Three resources with deliberately different surfaces.
--
-- A BILL is operator-managed and gets full CRUD: somebody opens, corrects and cancels settlement
-- units. Its ALLOCATIONS get read alone - they are computed by split, merge and the initial bill,
-- and a client able to write one could make an order's bills stop summing to the order (BR 36),
-- which is the one invariant the whole billing model rests on. The LINEAGE is read alone for the
-- same reason as the audit trail: a client able to POST one could fabricate a paper trail showing a
-- payment settled a bill it never touched.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M3SALES0000000000000004K', 'Sales Bill', 'sales_bill', 'A settlement unit of a sale - never a VAT invoice (BR 33)', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004M', 'Sales Bill Line', 'sales_bill_line', 'One order line''s allocated share of one bill', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004N', 'Sales Bill Relation', 'sales_bill_relation', 'The lineage left behind by a bill split or merge', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Sales Bill: full CRUD, an operator manages settlement.
		('01M3SALES0000000000000004P', 'Create', 'create', NULL, '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004Q', 'Update', 'update', NULL, '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004R', 'Delete', 'delete', NULL, '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004S', 'Read', 'read', NULL, '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004T', 'Set archived status', 'set_archived', 'Archive a sales bill, or bring an archived one back', '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Allocations and lineage: read alone. See the note at the top of this file.
		('01M3SALES0000000000000004V', 'Read', 'read', NULL, '01M3SALES0000000000000004M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004W', 'Read', 'read', NULL, '01M3SALES0000000000000004N', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Split and merge are separate powers from update: they restructure settlement units and
		-- write lineage, and a role that may correct a bill's details need not be able to divide it.
		('01M3SALES0000000000000004X', 'Split', 'split', 'Divide one open bill into several', '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SALES0000000000000004Y', 'Merge', 'merge', 'Combine several open bills of one order into a single bill', '01M3SALES0000000000000004K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
