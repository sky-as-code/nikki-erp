-- IAM resource and actions of the Audit Log module.
--
-- iam_resources.code is the SCHEMA NAME, byte-identical: the dynamic resource engine asserts
-- permissions using it as the resource code.
--
-- auditlog_action is append-only: entries are written by the core/audit.AuditSink port, never
-- through the API, so "read" is its only action. The plain search and get behind it are for
-- administrators; the per-record history (GET /v1/auditlog/actions/history) is authorized by
-- read access on the target record instead and needs no grant here.
--
-- No role is granted anything here: the Domain and Org Administrator roles already hold
-- "*:*:tenant" / "*:*:org" (1001004_iam_authorize_seeds.sql), which covers read:auditlog_action.
--
-- The coreplatform tree carries a copy with a leading tenant_id column; ids and codes must match.
DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M9AD1TAC0000000000000001', jsonb_build_object('en-US', 'Audit Log - Action', 'vi-VN', 'Nhật ký hệ thống - Nhật ký hành động'), 'auditlog_action', 'One audited action on a record: who did it, when, and which fields changed', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M9AD1TAC0000000000000011', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', 'Read audit entries', '01M9AD1TAC0000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
