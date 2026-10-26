-- Storage space of every org seeded in 1001002_iam_identity_seeds.sql (those never pass through
-- initOrg). The standalone binary has no tenant, so there is no tenant root: each org space is a root
-- space and its path is its name. Every org created later
-- gets its space from the IT Storage org lifecycle (initOrg), not from SQL.
DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'itstorage_spaces'
	) THEN
		INSERT INTO "itstorage_spaces" (
			"id", "name", "parent_id", "path", "size_bytes", "owner_type", "owner_id", "managed_by_type", "created_at", "etag"
		) VALUES
		('01M5T2BJV2EW8XAN0FYYX9ASYK', '01JWNY20G23KD4RV5VWYABQYHD', NULL, '01JWNY20G23KD4RV5VWYABQYHD', 0, 'org', '01JWNY20G23KD4RV5VWYABQYHD', 'system', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M5T44NYYJCRBX1ABGJNQ6R7B', '01K02G6J1CYAN9K8V4PAGSQ5Z8', NULL, '01K02G6J1CYAN9K8V4PAGSQ5Z8', 0, 'org', '01K02G6J1CYAN9K8V4PAGSQ5Z8', 'system', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M5T44NYYFP79X1FB3T9HXWJ0', '01K1H7M2K9VW3P5R7XQJY2C1Z9', NULL, '01K1H7M2K9VW3P5R7XQJY2C1Z9', 0, 'org', '01K1H7M2K9VW3P5R7XQJY2C1Z9', 'system', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
