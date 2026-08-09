DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		-- ('01KF326DEC4ND7PK91AZQY0TJ0', 'AuthzAction', 'authz_action', NULL, 'nikkierp', 'domain', 'domain', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- ('01JWNX10DN9E2PVZTVPJNTWR7D', 'AuthzEntitlement', 'iam_entitlement', NULL, 'nikkierp', 'domain', 'domain', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYWE9FBX2WTMYZMR9XHHX6', 'IAM Resource', 'iam_resource', NULL, 'nikkierp', 'domain', 'domain', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYV4RQ1ZKWG8RE0RMFTVCM', 'IAM Role', 'iam_role', NULL, 'nikkierp', 'domain', 'domain', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNX10DN9E2PVZTVPJNTWR7D', 'IAM Entitlement', 'iam_entitlement', NULL, 'nikkierp', 'domain', 'domain', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KF32E4TSX9DVV9KVKBS7GQ0T', 'IAM Grant Request', 'iam_grant_request', NULL, 'nikkierp', 'domain', 'domain', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYQ2A04PPV4135QGXX7W0M', 'IAM User', 'iam_user', NULL, 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYQNN68RKN62TNG5K0CPCE', 'IAM Group', 'iam_group', NULL, 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYQTHN6JTRNWKJ1BMCYB80', 'IAM Organization', 'iam_org', NULL, 'nikkierp', 'domain', 'domain', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYRSKZ56EAE2QRFHJWRZPT', 'IAM Organizational Unit', 'iam_orgunit', NULL, 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- AuthzAction
		-- ('01JWNYMG1K2X4N8B3NTHQTDTD1', 'Create', 'create', NULL, '01KF326DEC4ND7PK91AZQY0TJ0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- ('01JWNYKV61QB9B05AS52999QW2', 'Delete', 'delete', NULL, '01KF326DEC4ND7PK91AZQY0TJ0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- ('01JWNYJSCK13G33A1Z4BPY1C0A', 'Revoke', 'revoke', NULL, '01KF326DEC4ND7PK91AZQY0TJ0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- ('01JWNYK975JE099W1NRAB68D9C', 'Read', 'read', NULL, '01KF326DEC4ND7PK91AZQY0TJ0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- AuthzEntitlement
		-- ('01JWNYMG1K2X4N8B3NTHQMDMZB', 'Create', 'create', NULL, '01JWNX10DN9E2PVZTVPJNTWR7D', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- ('01JWNYKV61QB9B05AS52GCEPCR', 'Delete', 'delete', NULL, '01JWNX10DN9E2PVZTVPJNTWR7D', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- ('01JWNYJSCK13G16P0Z4BPY1C0A', 'Revoke', 'revoke', NULL, '01JWNX10DN9E2PVZTVPJNTWR7D', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- ('01JWNYK975JE0PKC1NRAB68D9C', 'Read', 'read', NULL, '01JWNX10DN9E2PVZTVPJNTWR7D', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- AuthzResource
		('01JWNYZ5EPJJMA3D367XMYEMM2', 'Create', 'create', NULL, '01JWNYWE9FBX2WTMYZMR9XHHX6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYZ8M3DXV2RNTP510CX9ZG', 'Delete', 'delete', NULL, '01JWNYWE9FBX2WTMYZMR9XHHX6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWY2HF8E72PQM8QHY0CHSVBT', 'Update', 'update', NULL, '01JWNYWE9FBX2WTMYZMR9XHHX6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYZEVSH78T2SH7WP47KDRM', 'Read', 'read', 'Read resources and their actions', '01JWNYWE9FBX2WTMYZMR9XHHX6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNHHJY19M5G5AT7ZQX9MWZSH', 'Manage actions', 'manage_actions', 'Create or delete actions of a resource', '01JWNYWE9FBX2WTMYZMR9XHHX6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- AuthzRole
		('01JWNZ14EZ00S2HWZD3Z7VANJK', 'Create', 'create', NULL, '01JWNYV4RQ1ZKWG8RE0RMFTVCM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ177SB70PS1SSKMS676VA', 'Delete', 'delete', NULL, '01JWNYV4RQ1ZKWG8RE0RMFTVCM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ1A1MNC7X5AVVPM14EC3P', 'Update', 'update', NULL, '01JWNYV4RQ1ZKWG8RE0RMFTVCM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ1D53FREVN8WX0Z7GZ1PS', 'Read', 'read', NULL, '01JWNYV4RQ1ZKWG8RE0RMFTVCM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ992AFRE8YEA70Z7GZ123', 'Manage entitlements', 'manage_entitlements', 'Create or delete entitlements of a role', '01JWNYV4RQ1ZKWG8RE0RMFTVCM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNV107H9R0LEARCH1V8QK7MA', 'Set archived status', 'set_archived', NULL, '01JWNYV4RQ1ZKWG8RE0RMFTVCM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- AuthzEntitlement
		('01JWNYMG1K2X4N8B3NTHQMDMZB', 'Create', 'create', NULL, '01JWNX10DN9E2PVZTVPJNTWR7D', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYKV61QB9B05AS52GCEPCR', 'Delete', 'delete', NULL, '01JWNX10DN9E2PVZTVPJNTWR7D', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNV107H9E4TUPD8H2P6QK7MA', 'Update', 'update', NULL, '01JWNX10DN9E2PVZTVPJNTWR7D', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYK975JE0PKC1NRAB68D9C', 'Read', 'read', NULL, '01JWNX10DN9E2PVZTVPJNTWR7D', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNV107H9MR0LES8H2P6QK7MA', 'Manage roles', 'manage_roles', 'Add or remove roles', '01JWNX10DN9E2PVZTVPJNTWR7D', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNV107H9ARCH1V8H2P6QK7MA', 'Set archived status', 'set_archived', NULL, '01JWNX10DN9E2PVZTVPJNTWR7D', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- AuthzGrantRequest
		('01JWNZ29T8K173M5GA3HF1TT5Y', 'Create', 'create', NULL, '01KF32E4TSX9DVV9KVKBS7GQ0T', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ2CMDCF51YV895YY1PQVZ', 'Delete', 'delete', NULL, '01KF32E4TSX9DVV9KVKBS7GQ0T', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNV107H9GRANTUPD8H2P6QK', 'Update', 'update', NULL, '01KF32E4TSX9DVV9KVKBS7GQ0T', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ2N37F8ZXHIHI17QYNG6R', 'Read', 'read', NULL, '01KF32E4TSX9DVV9KVKBS7GQ0T', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ2H9TPT1YY1TZ5KPHRE3H', 'Respond', 'respond', 'Respond grant or revoke requests despite not being role owners', '01KF32E4TSX9DVV9KVKBS7GQ0T', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- IamUser
		('01JWNZ3KA0ARGT9DAHQ1E6NZV0', 'Create', 'create', NULL, '01JWNYQ2A04PPV4135QGXX7W0M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ3PTQXAWE8R3HDTYVAQEK', 'Delete', 'delete', NULL, '01JWNYQ2A04PPV4135QGXX7W0M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ3TQ7AVCFDTSS0VHXHCAB', 'Update', 'update', NULL, '01JWNYQ2A04PPV4135QGXX7W0M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ3XVWDP73JGHRRBFAHQYJ', 'Read', 'read', NULL, '01JWNYQ2A04PPV4135QGXX7W0M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNV107H943VH4YJ2PFCQXH01', 'Set archived status', 'set_archived', NULL, '01JWNYQ2A04PPV4135QGXX7W0M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNRA5MGRAS5GNXKW3VBQZ001', 'Manage role assignments', 'manage_role_assignments', 'Assign or remove roles of a user', '01JWNYQ2A04PPV4135QGXX7W0M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- IdentityGroup
		('01JWNZ4QY0ECRHAKR0ERQW97HW', 'Create', 'create', NULL, '01JWNYQNN68RKN62TNG5K0CPCE', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ4V0ZDTEEMZPRZF6282SP', 'Delete', 'delete', NULL, '01JWNYQNN68RKN62TNG5K0CPCE', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ4Y4342HKE30Y4KE8MA8K', 'Update', 'update', NULL, '01JWNYQNN68RKN62TNG5K0CPCE', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ516R64X2S8A7STFXTP9B', 'Read', 'read', NULL, '01JWNYQNN68RKN62TNG5K0CPCE', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNHHJY19HMKE4JTCNZZ9WJ1T', 'Manage users', 'manage_users', 'Add or remove users', '01JWNYQNN68RKN62TNG5K0CPCE', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNRA5MGRAS5GNXKW3VBQZ002', 'Manage role assignments', 'manage_role_assignments', 'Assign or remove roles of a group', '01JWNYQNN68RKN62TNG5K0CPCE', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- IdentityOrganization
		('01JWNZ5KW6WC643VXGKV1D0J64', 'Create', 'create', NULL, '01JWNYQTHN6JTRNWKJ1BMCYB80', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ5PZP72SZVA3GVHRZW3RG', 'Delete', 'delete', NULL, '01JWNYQTHN6JTRNWKJ1BMCYB80', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ5SS046J9JVXS6WN316QB', 'Update', 'update', NULL, '01JWNYQTHN6JTRNWKJ1BMCYB80', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ5WJ1TM7P43W7FMENADTR', 'Read', 'read', NULL, '01JWNYQTHN6JTRNWKJ1BMCYB80', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNHHJY19B331PVTNDZVHA67P', 'Manage users', 'manage_users', 'Add or remove users', '01JWNYQTHN6JTRNWKJ1BMCYB80', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNV107H98GMY4KJDJVV0456J', 'Set archived status', 'set_archived', NULL, '01JWNYQTHN6JTRNWKJ1BMCYB80', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- IdentityOrganizationalUnit
		('01JWNZ6NSG3ZWY82PEH1ERDZ5C', 'Create', 'create', NULL, '01JWNYRSKZ56EAE2QRFHJWRZPT', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ6SYC65GEMJJ6BRNTEXFC', 'Delete', 'delete', NULL, '01JWNYRSKZ56EAE2QRFHJWRZPT', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ6XAZBQ8D11ETYGAN4N01', 'Update', 'update', NULL, '01JWNYRSKZ56EAE2QRFHJWRZPT', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ70QBW0B0KRMR5CNR56KX', 'Read', 'read', NULL, '01JWNYRSKZ56EAE2QRFHJWRZPT', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYWE9FBX2WTMYZMR9HQQT1', 'Move', 'move', NULL, '01JWNYRSKZ56EAE2QRFHJWRZPT', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JX0PKTPPP5CN780TAPMK846J', 'Manage users', 'manage_users', NULL, '01JWNYRSKZ56EAE2QRFHJWRZPT', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'iam_roles'
	) THEN
		INSERT INTO "iam_roles" (
			"id", "name", "description", "is_private", "owner_user_id", "is_requestable", "is_required_attachment",
			"is_required_comment", "is_archived", "created_at", "etag"
		) VALUES
		('01JWP72JJCDT4M0J8MSS51MN3T', 'Domain Administrator', 'Granted with all actions on all resources regardless of scope, except with Owner user', false, '01JWNMZ36QHC7CQQ748H9NQ6J6', false, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP80E084MTYF2C882WNR6MJ', 'Identity module Readonly', 'Granted with read action on Users and Groups in Identity module, except with Owner user', false, '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP80NHTHXSZDB1MZJXQ0MGQ', 'Identity module Administrator', 'Granted with all actions on all resources in Identity module, except with Owner user', false, '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWPB7TC3CG1EB567WYQCJM79', 'Identity module User Manager', 'Granted with all actions on on Users and Groups in Identity module, except with Owner user', false, '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP80S5RXP8BD4YCY8ZHP7NZ', 'Authorize module Readonly', 'Granted with read action on all resources in Authorize module', false, '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP80WR22SAG8Z7EYKDB00K6', 'Authorize module Administrator', 'Granted with all actions on all resources in Authorize module', false, '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP810BRSH9GWCYQC463K012', 'Authorize module Moderator', 'Granted with all actions on Resource, Action, and Role in Authorize module, but not allowed to delete the ones which are associated with an Entitlement', false, '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP72RCDK8NVBJMZTWJK9R97', 'Org Administrator (My Company)', 'Granted with all actions on all resources regardless of organizational units in the organization My Company', false, '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNHHJY194Y2F0KB34C4YVHC7', 'Org Administrator (requires membership)', 'Granted with all actions on all resources regardless of organizational units in the organization the user has membership', false, '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNHHJY19FFBFSXDJSAP7N6Z2', 'Sales Manager', 'Granted with all actions on all resources under org unit "Sales Department"', false, '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KZJ5XRJDXSXZY0DKNNE6S086', 'User', 'A system role that every user has it', false, '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);

	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_role_user_assignments'
	) THEN
		INSERT INTO "iam_role_user_assignments" (
			"id", "created_at", "role_id", "receiver_user_id", "approver_id", "role_request_id", "expires_at"
		) VALUES
		-- Preset user role assignments
		('01KNNFFNX1AHACN5ST2BM6VGYB', NOW(), '01JWP72JJCDT4M0J8MSS51MN3T', '01JWNXT3EY7FG47VDJTEPTDC98', '01JWNMZ36QHC7CQQ748H9NQ6J6', NULL, NULL),
		('01KNNFFNX11YQVW47NTAY1MVAB', NOW(), '01JWP80E084MTYF2C882WNR6MJ', '01JWNXXTF8958VVYAV33MVVMDN', '01JWNMZ36QHC7CQQ748H9NQ6J6', NULL, NULL),
		('01KNNFFNX1QEP3NRPM9BV8Z4SD', NOW(), '01JWP80WR22SAG8Z7EYKDB00K6', '01JZQFDH0N51Q3BFQFMFFGSCSV', '01JWNMZ36QHC7CQQ748H9NQ6J6', NULL, NULL),
		('01KNNFFNX1SRQBNB46Q23ZFK97', NOW(), '01JWP810BRSH9GWCYQC463K012', '01JZQFF9QEXH71P2CG9Y9MY8MM', '01JWNMZ36QHC7CQQ748H9NQ6J6', NULL, NULL),
		('01KNNFFNX19JAHGWDFXE6G84VF', NOW(), '01JWP72RCDK8NVBJMZTWJK9R97', '01JZQFFDKY8T4JB8R6NSY1331J', '01JWNMZ36QHC7CQQ748H9NQ6J6', NULL, NULL),
		('01KNNFFNX18B6NW6QW3ZEX3M2W', NOW(), '01KNHHJY194Y2F0KB34C4YVHC7', '01JZQFGVKZCTV7S310W0BDMWCS', '01JWNMZ36QHC7CQQ748H9NQ6J6', NULL, NULL);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'iam_entitlements'
	) THEN
		INSERT INTO "iam_entitlements" (
			"id", "name", "description", "expression", "action_id", "resource_id", "role_id", "scope", "org_id", "org_unit_id", "is_archived", "created_at", "etag"
		) VALUES
		-- Domain Administrator: All actions on all resources
		('01JWP88N498RQS88TYVJ4Z20EX', 'Domain Administrator - All Permissions', 'All permissions for domain administrator', '*:*:domain', NULL, NULL, '01JWP72JJCDT4M0J8MSS51MN3T', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		
		--  Org Administrator (requires membership): All actions on all resources under orgs which user has membership
		('01KNHHJY19XTT7QB7X9K9DJ2W7', 'Org Administrator (membership) - All Permissions', 'All permissions for org administrator (membership)', '*:*:org', NULL, NULL, '01KNHHJY194Y2F0KB34C4YVHC7', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		
		--  Org Administrator (org "My Company"): All actions on all resources under org "My Company"
		('01KNHHJY19M5407MRJSDTEG64Q', 'Org Administrator (My Company) - All Permissions', 'All permissions for org administrator of My Company', '*:*:org/01JWNY20G23KD4RV5VWYABQYHD', NULL, NULL, '01JWP72RCDK8NVBJMZTWJK9R97', 'org', '01JWNY20G23KD4RV5VWYABQYHD', NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Identity module Readonly: Read actions on Identity resources
		('01JWP8EARV3B9A1HWFPMQZQ6HZ', 'Identity Readonly - Read Users', 'Read users in Identity module', 'read:iam_user:domain', '01JWNZ3XVWDP73JGHRRBFAHQYJ', '01JWNYQ2A04PPV4135QGXX7W0M', '01JWP80E084MTYF2C882WNR6MJ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP8EFENXFNN17GSEJP0RCXZ', 'Identity Readonly - Read Groups', 'Read groups in Identity module', 'read:iam_group:domain', '01JWNZ516R64X2S8A7STFXTP9B', '01JWNYQNN68RKN62TNG5K0CPCE', '01JWP80E084MTYF2C882WNR6MJ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Identity module Administrator: All actions on Identity resources
		('01JWP8KSP3Q3YH6RKND552DWRR', 'Identity Admin - All User actions', 'All user actions in Identity module', '*:iam_user:domain', NULL, '01JWNYQ2A04PPV4135QGXX7W0M', '01JWP80NHTHXSZDB1MZJXQ0MGQ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP8KP39GKAH67FEAC7TZ631', 'Identity Admin - All Group actions', 'All group actions in Identity module', '*:iam_group:domain', NULL, '01JWNYQNN68RKN62TNG5K0CPCE', '01JWP80NHTHXSZDB1MZJXQ0MGQ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP8KK6K1M9WMP59BBAGEMB1', 'Identity Admin - All Organization actions', 'All organization actions in Identity module', '*:iam_org:domain', NULL, '01JWNYQTHN6JTRNWKJ1BMCYB80', '01JWP80NHTHXSZDB1MZJXQ0MGQ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP8KG15N26CRWNRM6F5CB29', 'Identity Admin - All Organizational Units actions', 'All organizational units actions in Identity module', '*:iam_orgunit:domain', NULL, '01JWNYRSKZ56EAE2QRFHJWRZPT', '01JWP80NHTHXSZDB1MZJXQ0MGQ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Identity module User Manager: All actions on Identity resources
		('01KNHHJY192QRXZCD7VP5HPPYG', 'Identity User Manager - All User actions', 'All user actions in Identity module', '*:iam_user:domain', NULL, '01JWNYQ2A04PPV4135QGXX7W0M', '01JWPB7TC3CG1EB567WYQCJM79', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNHHJY19KR6DK03ZBZ09KXPM', 'Identity User Manager - All Group actions', 'All group actions in Identity module', '*:iam_group:domain', NULL, '01JWNYQNN68RKN62TNG5K0CPCE', '01JWPB7TC3CG1EB567WYQCJM79', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		
		-- Authorize module Readonly: Read actions on Authorize resources
		('01JWPA3A4J2644C24V86419A2V', 'Authorize Readonly - Read Resources', 'Read resources in Authorize module', 'read:iam_resource:domain', '01JWNYZEVSH78T2SH7WP47KDRM', '01JWNYWE9FBX2WTMYZMR9XHHX6', '01JWP80S5RXP8BD4YCY8ZHP7NZ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNHHJY19ZDZFQDG71WJETFFH', 'Authorize Readonly - Read Roles', 'Read roles in Authorize module', 'read:iam_role:domain', '01JWNZ1D53FREVN8WX0Z7GZ1PS', '01JWNYV4RQ1ZKWG8RE0RMFTVCM', '01JWP80S5RXP8BD4YCY8ZHP7NZ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNV107H9ROENTREADH2P6QK', 'Authorize Readonly - Read Entitlements', 'Read entitlements in Authorize module', 'read:iam_entitlement:domain', '01JWNYK975JE0PKC1NRAB68D9C', '01JWNX10DN9E2PVZTVPJNTWR7D', '01JWP80S5RXP8BD4YCY8ZHP7NZ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNHHJY19TQPRN03XDDEKQ6PE', 'Authorize Readonly - Read Grant Request', 'Read grant requests in Authorize module', 'read:iam_grant_request:domain', '01JWNZ2N37F8ZXHIHI17QYNG6R', '01KF32E4TSX9DVV9KVKBS7GQ0T', '01JWP80S5RXP8BD4YCY8ZHP7NZ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Authorize module Administrator: All actions on Authorize resources
		('01JWPA35MPHG33G77FKQNYJS21', 'Authorize Admin - All Resource actions', 'All resource actions in Authorize module', '*:iam_resource:domain', NULL, '01JWNYWE9FBX2WTMYZMR9XHHX6', '01JWP80WR22SAG8Z7EYKDB00K6', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWPA3232EYYN4HQMWBBV345B', 'Authorize Admin - All Role actions', 'All role actions in Authorize module', '*:iam_role:domain', NULL, '01JWNYV4RQ1ZKWG8RE0RMFTVCM', '01JWP80WR22SAG8Z7EYKDB00K6', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNV107H9ADENTALLH2P6QK7', 'Authorize Admin - All Entitlement actions', 'All entitlement actions in Authorize module', '*:iam_entitlement:domain', NULL, '01JWNX10DN9E2PVZTVPJNTWR7D', '01JWP80WR22SAG8Z7EYKDB00K6', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JW197DCKNR9DD16MP9TYA89J', 'Authorize Admin - All Grant Request actions', 'All grant and revoke request actions in Authorize module', '*:iam_grant_request:domain', NULL, '01KF32E4TSX9DVV9KVKBS7GQ0T', '01JWP80WR22SAG8Z7EYKDB00K6', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Authorize module Moderator: All actions on Resource, Role except delete
		('01JWP8KCVWVYDSQ6C8SNDQD5F6', 'Authorize Moderator - Create Resources', 'Create resources in Authorize module', 'create:iam_resource:domain', '01JWNYZ5EPJJMA3D367XMYEMM2', '01JWNYWE9FBX2WTMYZMR9XHHX6', '01JWP810BRSH9GWCYQC463K012', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP8KCVWVYDSQ6C8SNDQD5F7', 'Authorize Moderator - Update Resources', 'Update resources in Authorize module', 'update:iam_resource:domain', '01JWY2HF8E72PQM8QHY0CHSVBT', '01JWNYWE9FBX2WTMYZMR9XHHX6', '01JWP810BRSH9GWCYQC463K012', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP8KCVWVYDSQ6C8SNDQD5F9', 'Authorize Moderator - Read Resources', 'Read resources in Authorize module', 'read:iam_resource:domain', '01JWNYZEVSH78T2SH7WP47KDRM', '01JWNYWE9FBX2WTMYZMR9XHHX6', '01JWP810BRSH9GWCYQC463K012', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP8K9QE8PW6BKZZ6EW9C9Z5', 'Authorize Moderator - Create Roles', 'Create roles in Authorize module', 'create:iam_role:domain', '01JWNZ14EZ00S2HWZD3Z7VANJK', '01JWNYV4RQ1ZKWG8RE0RMFTVCM', '01JWP810BRSH9GWCYQC463K012', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP8K9QE8PW6BKZZ6EW9C9Z6', 'Authorize Moderator - Update Roles', 'Update roles in Authorize module', 'update:iam_role:domain', '01JWNZ1A1MNC7X5AVVPM14EC3P', '01JWNYV4RQ1ZKWG8RE0RMFTVCM', '01JWP810BRSH9GWCYQC463K012', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP8K9QE8PW6BKZZ6EW9C9Z7', 'Authorize Moderator - Read Roles', 'Read roles in Authorize module', 'read:iam_role:domain', '01JWNZ1D53FREVN8WX0Z7GZ1PS', '01JWNYV4RQ1ZKWG8RE0RMFTVCM', '01JWP810BRSH9GWCYQC463K012', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Sales Manager: All actions on all resources under org unit "Sales Department"
		('01KNHHJY19X4XW1ZRWD5955TD0', 'Sales Managers - All permissions', 'All permissions in Sales Department', '*:*:orgunit/01K1H8N3L0WX4Q6S8YRKT3D2C1', NULL, NULL, '01KNHHJY19FFBFSXDJSAP7N6Z2', 'orgunit', NULL, '01K1H8N3L0WX4Q6S8YRKT3D2C1', false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;

	PERFORM iam_rebuild_all_user_perms();

END $$;