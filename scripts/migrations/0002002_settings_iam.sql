-- IAM resources, actions and entitlements for the Settings module.
--
-- iam_resources.code is the SCHEMA NAME, byte-identical ("settings_record", not "SettingsRecord"
-- and not the table name "settings_records"). The dynamic resource engine asserts permissions
-- using the schema name as the resource code, so a code that drifts denies every request against
-- that resource, with nothing in the response pointing at this file as the cause.
--
-- iam_actions.code is the engine's ACTION NAME, lowercase. Both resources are plain CRUD at the
-- engine level; the behaviour that makes settings more than a key-value table (partial save, level
-- authorization, the enforcement fan-out) lives in the application services and rides on the same
-- read/update actions rather than adding codes of its own.
--
-- Neither resource gets set_archived: a settings row is either the owner's current value or it does
-- not exist. There is no state in which a setting is kept but out of use, so an archive action would
-- advertise a capability with no meaning behind it.
--
-- ENTITLEMENTS - the deliberate part.
--
-- The system User role every account holds is granted READ on settings_record and settings_schema,
-- and nothing else. Two reasons it must have read:
--
--   * A user cannot render their own preferences page without reading their own rows, and every
--     account has preferences. Withholding read here would break the profile screen for everyone.
--   * The settings page reads the schema to know what controls to draw, so read on
--     settings_schema is required for the same screen to render at all.
--
-- Read is safe to grant domain-wide because the *rows a caller sees are chosen by the service, not
-- by the grant*: UserPreferencesAppService resolves the owner from the request's own user id, so a
-- user holding this entitlement still reads only their own rows. The entitlement says "may use the
-- settings read path", not "may read every tenant's configuration".
--
-- Update is deliberately NOT granted to the User role. Tenant and organization configuration is
-- exactly what an ordinary account must not be able to rewrite, and a user changing their own
-- preferences goes through UserPreferencesAppService, which asserts no permission of its own
-- precisely because it can only ever write the caller's own rows. Granting update here would hand
-- every account the tenant-level write path as well, since both levels share one resource code.
-- A Tenant or Organization Administrator receives update through its own role, not through this
-- file.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M68990HNE4Z1PR28N6G1R1M2', 'Setting Schema', 'settings_schema', 'What a module declares it can be configured with, per level', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01MACPJ8N53VVRCDGNB7APWHPG', 'Setting Record', 'settings_record', 'One setting value held by one tenant, organization or user', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01MK4MWVJ323AVHSNC2PEQ8RVN', 'Create', 'create', NULL, '01M68990HNE4Z1PR28N6G1R1M2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01MVHMAN7HPNTT9HY8QZ0JDX7G', 'Read', 'read', NULL, '01M68990HNE4Z1PR28N6G1R1M2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01MZ8XS1ZB97BFSB2BY6GYKCV8', 'Update', 'update', NULL, '01M68990HNE4Z1PR28N6G1R1M2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M8AW2S19YCJJ8WGEZ7RE1KAZ', 'Delete', 'delete', NULL, '01M68990HNE4Z1PR28N6G1R1M2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M7SD4VK23A11JBHHEGTHQH5H', 'Create', 'create', NULL, '01MACPJ8N53VVRCDGNB7APWHPG', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01MFS8N4WX55Z4G0AXKB41SG81', 'Read', 'read', NULL, '01MACPJ8N53VVRCDGNB7APWHPG', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01MBA9WKZ5FN4SAN210Q820H8S', 'Update', 'update', NULL, '01MACPJ8N53VVRCDGNB7APWHPG', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01MNGK482D91SKJP082R2BAM10', 'Delete', 'delete', NULL, '01MACPJ8N53VVRCDGNB7APWHPG', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_entitlements'
	) THEN
		INSERT INTO "iam_entitlements" (
			"id", "name", "description", "expression", "action_id", "resource_id", "role_id", "scope", "org_id", "org_unit_id", "is_archived", "created_at", "etag"
		) VALUES
		('01MCJZ6P0ETGP6BE4TPQM20NEZ', 'User - Read Setting Schemas', 'Read the setting declarations needed to render the settings page', 'read:settings_schema:domain', '01MVHMAN7HPNTT9HY8QZ0JDX7G', '01M68990HNE4Z1PR28N6G1R1M2', '01KZJ5XRJDXSXZY0DKNNE6S086', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01MNJ1KF88AN0A9QWXXQWTY5ZN', 'User - Read Setting Records', 'Read own settings; the service resolves the owner from the request', 'read:settings_record:domain', '01MFS8N4WX55Z4G0AXKB41SG81', '01MACPJ8N53VVRCDGNB7APWHPG', '01KZJ5XRJDXSXZY0DKNNE6S086', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
