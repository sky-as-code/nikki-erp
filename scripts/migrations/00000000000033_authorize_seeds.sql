DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'authz_resources'
	) THEN
		INSERT INTO "authz_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		-- ('01KF326DEC4ND7PK91AZQY0TJ0', 'AuthzAction', 'authz_action', NULL, 'nikkierp', 'domain', 'domain', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- ('01JWNX10DN9E2PVZTVPJNTWR7D', 'AuthzEntitlement', 'authz_entitlement', NULL, 'nikkierp', 'domain', 'domain', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYWE9FBX2WTMYZMR9XHHX6', 'AuthzResource', 'authz_resource', NULL, 'nikkierp', 'domain', 'domain', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYV4RQ1ZKWG8RE0RMFTVCM', 'AuthzRole', 'authz_role', NULL, 'nikkierp', 'domain', 'domain', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYW23X8CMREJ2Y9349BAE4', 'AuthzRoleSuite', 'authz_role_suite', NULL, 'nikkierp', 'domain', 'domain', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KF32E4TSX9DVV9KVKBS7GQ0T', 'AuthzGrantRequest', 'authz_grant_request', NULL, 'nikkierp', 'domain', 'domain', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KF328VKN38VN16RG17C8BECB', 'AuthzRevokeRequest', 'authz_revoke_request', NULL, 'nikkierp', 'domain', 'domain', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYQ2A04PPV4135QGXX7W0M', 'IdentityUser', 'identity_user', NULL, 'nikkierp', 'domain', 'domain', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYQNN68RKN62TNG5K0CPCE', 'IdentityGroup', 'identity_group', NULL, 'nikkierp', 'org', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYQTHN6JTRNWKJ1BMCYB80', 'IdentityOrganization', 'identity_org', NULL, 'nikkierp', 'domain', 'domain', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYRSKZ56EAE2QRFHJWRZPT', 'IdentityOrganizationalUnit', 'identity_orgunit', NULL, 'nikkierp', 'org', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP5S7KJF8T3RSA8WDZVSZWA', 'IdentityProfile', 'identity_profile', NULL, 'nikkierp', 'private', 'private', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'authz_actions'
	) THEN
		INSERT INTO "authz_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- AuthzAction
		-- ('01JWNYMG1K2X4N8B3NTHQTDTD1', 'Create', 'create', NULL, '01KF326DEC4ND7PK91AZQY0TJ0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- ('01JWNYKV61QB9B05AS52999QW2', 'Delete', 'delete', NULL, '01KF326DEC4ND7PK91AZQY0TJ0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- ('01JWNYJSCK13G33A1Z4BPY1C0A', 'Revoke', 'revoke', NULL, '01KF326DEC4ND7PK91AZQY0TJ0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- ('01JWNYK975JE099W1NRAB68D9C', 'View', 'view', NULL, '01KF326DEC4ND7PK91AZQY0TJ0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- AuthzEntitlement
		-- ('01JWNYMG1K2X4N8B3NTHQMDMZB', 'Create', 'create', NULL, '01JWNX10DN9E2PVZTVPJNTWR7D', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- ('01JWNYKV61QB9B05AS52GCEPCR', 'Delete', 'delete', NULL, '01JWNX10DN9E2PVZTVPJNTWR7D', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- ('01JWNYJSCK13G16P0Z4BPY1C0A', 'Revoke', 'revoke', NULL, '01JWNX10DN9E2PVZTVPJNTWR7D', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- ('01JWNYK975JE0PKC1NRAB68D9C', 'View', 'view', NULL, '01JWNX10DN9E2PVZTVPJNTWR7D', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- AuthzResource
		('01JWNYZ5EPJJMA3D367XMYEMM2', 'Create', 'create', NULL, '01JWNYWE9FBX2WTMYZMR9XHHX6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYZ8M3DXV2RNTP510CX9ZG', 'Delete', 'delete', NULL, '01JWNYWE9FBX2WTMYZMR9XHHX6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWY2HF8E72PQM8QHY0CHSVBT', 'Update', 'update', NULL, '01JWNYWE9FBX2WTMYZMR9XHHX6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYZEVSH78T2SH7WP47KDRM', 'View', 'view', 'View resources and their actions', '01JWNYWE9FBX2WTMYZMR9XHHX6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNHHJY19M5G5AT7ZQX9MWZSH', 'Manage actions', 'manage_actions', 'Create or delete actions of a resource', '01JWNYWE9FBX2WTMYZMR9XHHX6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- AuthzRole
		('01JWNZ14EZ00S2HWZD3Z7VANJK', 'Create', 'create', NULL, '01JWNYV4RQ1ZKWG8RE0RMFTVCM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ177SB70PS1SSKMS676VA', 'Delete', 'delete', NULL, '01JWNYV4RQ1ZKWG8RE0RMFTVCM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ1A1MNC7X5AVVPM14EC3P', 'Update', 'update', NULL, '01JWNYV4RQ1ZKWG8RE0RMFTVCM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ1D53FREVN8WX0Z7GZ1PS', 'View', 'view', NULL, '01JWNYV4RQ1ZKWG8RE0RMFTVCM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ992AFRE8YEA70Z7GZ123', 'Manage entitlements', 'manage_entitlements', 'Create or delete entitlements of a role', '01JWNYV4RQ1ZKWG8RE0RMFTVCM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- AuthzGrantRequest
		('01JWNZ29T8K173M5GA3HF1TT5Y', 'Create', 'create', NULL, '01KF32E4TSX9DVV9KVKBS7GQ0T', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ2CMDCF51YV895YY1PQVZ', 'Delete', 'delete', NULL, '01KF32E4TSX9DVV9KVKBS7GQ0T', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ2N37F8ZXHIHI17QYNG6R', 'View', 'view', NULL, '01KF32E4TSX9DVV9KVKBS7GQ0T', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ2H9TPT1YY1TZ5KPHRE3H', 'Respond', 'respond', 'Respond a grant requests despite not being role owners', '01KF32E4TSX9DVV9KVKBS7GQ0T', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- AuthzRevokeRequest
		('01JWNZ29T8K173M5GA3HF911GT', 'Create', 'create', NULL, '01KF328VKN38VN16RG17C8BECB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWN3TTKDCF51YV895YY1PQVZ', 'Delete', 'delete', NULL, '01KF328VKN38VN16RG17C8BECB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ2NTDTKZXHIHI17QYNG6R', 'View', 'view', NULL, '01KF328VKN38VN16RG17C8BECB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNHHJY1926K9R1QWVZFQPY54', 'Respond', 'respond', 'Respond a revoke requests despite not being role owners', '01KF328VKN38VN16RG17C8BECB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- IdentityUser
		('01JWNZ3KA0ARGT9DAHQ1E6NZV0', 'Create', 'create', NULL, '01JWNYQ2A04PPV4135QGXX7W0M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ3PTQXAWE8R3HDTYVAQEK', 'Delete', 'delete', NULL, '01JWNYQ2A04PPV4135QGXX7W0M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ3TQ7AVCFDTSS0VHXHCAB', 'Update', 'update', NULL, '01JWNYQ2A04PPV4135QGXX7W0M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ3XVWDP73JGHRRBFAHQYJ', 'View', 'view', NULL, '01JWNYQ2A04PPV4135QGXX7W0M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- IdentityGroup
		('01JWNZ4QY0ECRHAKR0ERQW97HW', 'Create', 'create', NULL, '01JWNYQNN68RKN62TNG5K0CPCE', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ4V0ZDTEEMZPRZF6282SP', 'Delete', 'delete', NULL, '01JWNYQNN68RKN62TNG5K0CPCE', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ4Y4342HKE30Y4KE8MA8K', 'Update', 'update', NULL, '01JWNYQNN68RKN62TNG5K0CPCE', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ516R64X2S8A7STFXTP9B', 'View', 'view', NULL, '01JWNYQNN68RKN62TNG5K0CPCE', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNHHJY19HMKE4JTCNZZ9WJ1T', 'Manage users', 'manage_users', 'Add or remove users', '01JWNYQNN68RKN62TNG5K0CPCE', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- IdentityOrganization
		('01JWNZ5KW6WC643VXGKV1D0J64', 'Create', 'create', NULL, '01JWNYQTHN6JTRNWKJ1BMCYB80', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ5PZP72SZVA3GVHRZW3RG', 'Delete', 'delete', NULL, '01JWNYQTHN6JTRNWKJ1BMCYB80', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ5SS046J9JVXS6WN316QB', 'Update', 'update', NULL, '01JWNYQTHN6JTRNWKJ1BMCYB80', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ5WJ1TM7P43W7FMENADTR', 'View', 'view', NULL, '01JWNYQTHN6JTRNWKJ1BMCYB80', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNHHJY19B331PVTNDZVHA67P', 'Manage users', 'manage_users', 'Add or remove users', '01JWNYQTHN6JTRNWKJ1BMCYB80', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- IdentityOrganizationalUnit
		('01JWNZ6NSG3ZWY82PEH1ERDZ5C', 'Create', 'create', NULL, '01JWNYRSKZ56EAE2QRFHJWRZPT', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ6SYC65GEMJJ6BRNTEXFC', 'Delete', 'delete', NULL, '01JWNYRSKZ56EAE2QRFHJWRZPT', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ6XAZBQ8D11ETYGAN4N01', 'Update', 'update', NULL, '01JWNYRSKZ56EAE2QRFHJWRZPT', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNZ70QBW0B0KRMR5CNR56KX', 'View', 'view', NULL, '01JWNYRSKZ56EAE2QRFHJWRZPT', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWNYWE9FBX2WTMYZMR9HQQT1', 'Move', 'move', NULL, '01JWNYRSKZ56EAE2QRFHJWRZPT', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JX0PKTPPP5CN780TAPMK846J', 'Manage users', 'manage_users', NULL, '01JWNYRSKZ56EAE2QRFHJWRZPT', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'authz_roles'
	) THEN
		INSERT INTO "authz_roles" (
			"id", "name", "description", "owner_user_id", "is_requestable", "is_required_attachment",
			"is_required_comment", "is_archived", "created_at", "etag"
		) VALUES
		('01JWP72JJCDT4M0J8MSS51MN3T', 'Domain Administrator', 'Granted with all actions on all resources regardless of scope, except with Owner user', '01JWNMZ36QHC7CQQ748H9NQ6J6', false, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP80E084MTYF2C882WNR6MJ', 'Identity module Readonly', 'Granted with view action on Users and Groups in Identity module, except with Owner user', '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP80NHTHXSZDB1MZJXQ0MGQ', 'Identity module Administrator', 'Granted with all actions on all resources in Identity module, except with Owner user', '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWPB7TC3CG1EB567WYQCJM79', 'Identity module User Manager', 'Granted with all actions on on Users and Groups in Identity module, except with Owner user', '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP80S5RXP8BD4YCY8ZHP7NZ', 'Authorize module Readonly', 'Granted with view action on all resources in Authorize module', '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP80WR22SAG8Z7EYKDB00K6', 'Authorize module Administrator', 'Granted with all actions on all resources in Authorize module', '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP810BRSH9GWCYQC463K012', 'Authorize module Moderator', 'Granted with all actions on Resource, Action, Role and Role Suite in Authorize module, but not allowed to delete the ones which are associated with an Entitlement', '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP72RCDK8NVBJMZTWJK9R97', 'Org Administrator (My Company)', 'Granted with all actions on all resources regardless of organizational units in the organization My Company', '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNHHJY194Y2F0KB34C4YVHC7', 'Org Administrator (requires membership)', 'Granted with all actions on all resources regardless of organizational units in the organization the user has membership', '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNHHJY19FFBFSXDJSAP7N6Z2', 'Sales Manager', 'Granted with all actions on all resources under org unit "Sales Department"', '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'authz_entitlements'
	) THEN
		INSERT INTO "authz_entitlements" (
			"id", "name", "description", "expression", "action_id", "resource_id", "role_id", "scope", "org_id", "org_unit_id", "is_archived", "created_at", "etag"
		) VALUES
		-- Domain Administrator: All actions on all resources
		('01JWP88N498RQS88TYVJ4Z20EX', 'Domain Administrator - All Permissions', 'All permissions for domain administrator', '*:*:domain', NULL, NULL, '01JWP72JJCDT4M0J8MSS51MN3T', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		
		--  Org Administrator (requires membership): All actions on all resources under orgs which user has membership
		('01KNHHJY19XTT7QB7X9K9DJ2W7', 'Org Administrator (membership) - All Permissions', 'All permissions for org administrator (membership)', '*:*:org', NULL, NULL, '01KNHHJY194Y2F0KB34C4YVHC7', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		
		--  Org Administrator (org "My Company"): All actions on all resources under org "My Company"
		('01KNHHJY19M5407MRJSDTEG64Q', 'Org Administrator (My Company) - All Permissions', 'All permissions for org administrator of My Company', '*:*:org/01JWNY20G23KD4RV5VWYABQYHD', NULL, NULL, '01JWP72RCDK8NVBJMZTWJK9R97', 'org', '01JWNY20G23KD4RV5VWYABQYHD', NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Identity module Readonly: View actions on Identity resources
		('01JWP8EARV3B9A1HWFPMQZQ6HZ', 'Identity Readonly - View Users', 'View users in Identity module', 'view:user:domain', '01JWNZ3XVWDP73JGHRRBFAHQYJ', '01JWNYQ2A04PPV4135QGXX7W0M', '01JWP80E084MTYF2C882WNR6MJ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP8EFENXFNN17GSEJP0RCXZ', 'Identity Readonly - View Groups', 'View groups in Identity module', 'view:group:domain', '01JWNZ516R64X2S8A7STFXTP9B', '01JWNYQNN68RKN62TNG5K0CPCE', '01JWP80E084MTYF2C882WNR6MJ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Identity module Administrator: All actions on Identity resources
		('01JWP8KSP3Q3YH6RKND552DWRR', 'Identity Admin - All User actions', 'All user actions in Identity module', '*:identity_user:domain', NULL, '01JWNYQ2A04PPV4135QGXX7W0M', '01JWP80NHTHXSZDB1MZJXQ0MGQ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP8KP39GKAH67FEAC7TZ631', 'Identity Admin - All Group actions', 'All group actions in Identity module', '*:identity_group:domain', NULL, '01JWNYQNN68RKN62TNG5K0CPCE', '01JWP80NHTHXSZDB1MZJXQ0MGQ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP8KK6K1M9WMP59BBAGEMB1', 'Identity Admin - All Organization actions', 'All organization actions in Identity module', '*:identity_org:domain', NULL, '01JWNYQTHN6JTRNWKJ1BMCYB80', '01JWP80NHTHXSZDB1MZJXQ0MGQ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP8KG15N26CRWNRM6F5CB29', 'Identity Admin - All Organizational Units actions', 'All organizational units actions in Identity module', '*:identity_orgunit:domain', NULL, '01JWNYRSKZ56EAE2QRFHJWRZPT', '01JWP80NHTHXSZDB1MZJXQ0MGQ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Identity module User Manager: All actions on Identity resources
		('01KNHHJY192QRXZCD7VP5HPPYG', 'Identity User Manager - All User actions', 'All user actions in Identity module', '*:identity_user:domain', NULL, '01JWNYQ2A04PPV4135QGXX7W0M', '01JWPB7TC3CG1EB567WYQCJM79', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNHHJY19KR6DK03ZBZ09KXPM', 'Identity User Manager - All Group actions', 'All group actions in Identity module', '*:identity_group:domain', NULL, '01JWNYQ2A04PPV4135QGXX7W0M', '01JWPB7TC3CG1EB567WYQCJM79', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		
		-- Authorize module Readonly: View actions on Authorize resources
		('01JWPA3A4J2644C24V86419A2V', 'Authorize Readonly - View Resources', 'View resources in Authorize module', 'view:authz_resource:domain', '01JWNYZEVSH78T2SH7WP47KDRM', '01JWNYWE9FBX2WTMYZMR9XHHX6', '01JWP80S5RXP8BD4YCY8ZHP7NZ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNHHJY19ZDZFQDG71WJETFFH', 'Authorize Readonly - View Roles', 'View roles in Authorize module', 'view:authz_role:domain', '01JWNZ1D53FREVN8WX0Z7GZ1PS', '01JWNYV4RQ1ZKWG8RE0RMFTVCM', '01JWP80S5RXP8BD4YCY8ZHP7NZ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNHHJY19TQPRN03XDDEKQ6PE', 'Authorize Readonly - View Grant Request', 'View grant requests in Authorize module', 'view:authz_grant_request:domain', '01JWNZ2N37F8ZXHIHI17QYNG6R', '01KF32E4TSX9DVV9KVKBS7GQ0T', '01JWP80S5RXP8BD4YCY8ZHP7NZ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNHHJY19CTEAPCBJE02P4ZTQ', 'Authorize Readonly - View Revoke Request', 'View revoke requests in Authorize module', 'view:authz_revoke_request:domain', '01JWNZ2N37F8ZXHIHI17QYNG6R', '01KF328VKN38VN16RG17C8BECB', '01JWP80S5RXP8BD4YCY8ZHP7NZ', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Authorize module Administrator: All actions on Authorize resources
		('01JWPA35MPHG33G77FKQNYJS21', 'Authorize Admin - All Resource actions', 'All resource actions in Authorize module', '*:authz_resource:domain', NULL, '01JWNYWE9FBX2WTMYZMR9XHHX6', '01JWP80WR22SAG8Z7EYKDB00K6', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWPA3232EYYN4HQMWBBV345B', 'Authorize Admin - All Role actions', 'All role actions in Authorize module', '*:authz_role:domain', NULL, '01JWNYV4RQ1ZKWG8RE0RMFTVCM', '01JWP80WR22SAG8Z7EYKDB00K6', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JW197DCKNR9DD16MP9TYA89J', 'Authorize Admin - All Grant Request actions', 'All grant request actions in Authorize module', '*:authz_grant_request:domain', NULL, '01KF32E4TSX9DVV9KVKBS7GQ0T', '01JWP80WR22SAG8Z7EYKDB00K6', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01KNHHJY197GSXY5CX8VBMSFQ4', 'Authorize Admin - All Revoke Request actions', 'All revoke request actions in Authorize module', '*:authz_revoke_request:domain', NULL, '01JWNZ29T8K173M5GA3HF911GT', '01JWP80WR22SAG8Z7EYKDB00K6', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Authorize module Moderator: All actions on Resource, Role except delete
		('01JWP8KCVWVYDSQ6C8SNDQD5F6', 'Authorize Moderator - Create Resources', 'Create resources in Authorize module', 'create:authz_resource:domain', '01JWNYZ5EPJJMA3D367XMYEMM2', '01JWNYWE9FBX2WTMYZMR9XHHX6', '01JWP810BRSH9GWCYQC463K012', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP8KCVWVYDSQ6C8SNDQD5F7', 'Authorize Moderator - Update Resources', 'Update resources in Authorize module', 'update:authz_resource:domain', '01JWY2HF8E72PQM8QHY0CHSVBT', '01JWNYWE9FBX2WTMYZMR9XHHX6', '01JWP810BRSH9GWCYQC463K012', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP8KCVWVYDSQ6C8SNDQD5F9', 'Authorize Moderator - View Resources', 'View resources in Authorize module', 'view:authz_resource:domain', '01JWNYZEVSH78T2SH7WP47KDRM', '01JWNYWE9FBX2WTMYZMR9XHHX6', '01JWP810BRSH9GWCYQC463K012', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP8K9QE8PW6BKZZ6EW9C9Z5', 'Authorize Moderator - Create Roles', 'Create roles in Authorize module', 'create:authz_role:domain', '01JWNZ14EZ00S2HWZD3Z7VANJK', '01JWNYV4RQ1ZKWG8RE0RMFTVCM', '01JWP810BRSH9GWCYQC463K012', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP8K9QE8PW6BKZZ6EW9C9Z6', 'Authorize Moderator - Update Roles', 'Update roles in Authorize module', 'update:authz_role:domain', '01JWNZ1A1MNC7X5AVVPM14EC3P', '01JWNYV4RQ1ZKWG8RE0RMFTVCM', '01JWP810BRSH9GWCYQC463K012', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01JWP8K9QE8PW6BKZZ6EW9C9Z7', 'Authorize Moderator - View Roles', 'View roles in Authorize module', 'view:authz_role:domain', '01JWNZ1D53FREVN8WX0Z7GZ1PS', '01JWNYV4RQ1ZKWG8RE0RMFTVCM', '01JWP810BRSH9GWCYQC463K012', 'domain', NULL, NULL, false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Sales Manager: All actions on all resources under org unit "Sales Department"
		('01KNHHJY19X4XW1ZRWD5955TD0', 'Sales Managers - All permissions', 'All permissions in Sames Department', '*:*:orgunit/01K1H8N3L0WX4Q6S8YRKT3D2C1', NULL, NULL, '01JWP810BRSH9GWCYQC463K012', 'orgunit', NULL, '01K1H8N3L0WX4Q6S8YRKT3D2C1', false, NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text);
	END IF;

END $$;