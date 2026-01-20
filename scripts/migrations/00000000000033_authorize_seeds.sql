DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'authz_resources'
	) THEN
		INSERT INTO "authz_resources" ("id", "created_at", "name", "resource_type", "scope_type", "etag") VALUES
		('01KF326DEC4ND7PK91AZQY0TJ0', NOW(), 'AuthzAction', 'nikki_application', 'domain', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNX10DN9E2PVZTVPJNTWR7D', NOW(), 'AuthzEntitlement', 'nikki_application', 'domain', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNYWE9FBX2WTMYZMR9XHHX6', NOW(), 'AuthzResource', 'nikki_application', 'domain', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNYV4RQ1ZKWG8RE0RMFTVCM', NOW(), 'AuthzRole', 'nikki_application', 'domain', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNYW23X8CMREJ2Y9349BAE4', NOW(), 'AuthzRoleSuite', 'nikki_application', 'domain', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01KF32E4TSX9DVV9KVKBS7GQ0T', NOW(), 'AuthzGrantRequest', 'nikki_application', 'domain', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01KF328VKN38VN16RG17C8BECB', NOW(), 'AuthzRevokeRequest', 'nikki_application', 'domain', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNYQ2A04PPV4135QGXX7W0M', NOW(), 'IdentityUser', 'nikki_application', 'hierarchy', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNYQNN68RKN62TNG5K0CPCE', NOW(), 'IdentityGroup', 'nikki_application', 'org', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNYQTHN6JTRNWKJ1BMCYB80', NOW(), 'IdentityOrganization', 'nikki_application', 'domain', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNYRSKZ56EAE2QRFHJWRZPT', NOW(), 'IdentityHierarchyLevel', 'nikki_application', 'org', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWP5S7KJF8T3RSA8WDZVSZWA', NOW(), 'IdentityProfile', 'nikki_application', 'private', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'authz_actions'
	) THEN
		INSERT INTO "authz_actions" ("id", "name", "resource_id", "created_at", "created_by", "etag") VALUES
		-- AuthzAction
		('01JWNYMG1K2X4N8B3NTHQTDTD1', 'Create', '01KF326DEC4ND7PK91AZQY0TJ0', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNYKV61QB9B05AS52999QW2', 'Delete', '01KF326DEC4ND7PK91AZQY0TJ0', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNYJGBHZX2HH1Q7V1V7QNN3', 'Grant', '01KF326DEC4ND7PK91AZQY0TJ0', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNYJSCK13G33A1Z4BPY1C0A', 'Revoke', '01KF326DEC4ND7PK91AZQY0TJ0', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNYK975JE099W1NRAB68D9C', 'View', '01KF326DEC4ND7PK91AZQY0TJ0', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		-- AuthzEntitlement
		('01JWNYMG1K2X4N8B3NTHQMDMZB', 'Create', '01JWNX10DN9E2PVZTVPJNTWR7D', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNYKV61QB9B05AS52GCEPCR', 'Delete', '01JWNX10DN9E2PVZTVPJNTWR7D', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNYJGBHZX29Y3C7V1V7QNN3', 'Grant', '01JWNX10DN9E2PVZTVPJNTWR7D', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNYJSCK13G16P0Z4BPY1C0A', 'Revoke', '01JWNX10DN9E2PVZTVPJNTWR7D', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNYK975JE0PKC1NRAB68D9C', 'View', '01JWNX10DN9E2PVZTVPJNTWR7D', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		-- AuthzResource
		('01JWNYZ5EPJJMA3D367XMYEMM2', 'Create', '01JWNYWE9FBX2WTMYZMR9XHHX6', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNYZ8M3DXV2RNTP510CX9ZG', 'Delete', '01JWNYWE9FBX2WTMYZMR9XHHX6', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWY2HF8E72PQM8QHY0CHSVBT', 'Update', '01JWNYWE9FBX2WTMYZMR9XHHX6', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNYZEVSH78T2SH7WP47KDRM', 'View', '01JWNYWE9FBX2WTMYZMR9XHHX6', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		-- AuthzRole
		('01JWNZ14EZ00S2HWZD3Z7VANJK', 'Create', '01JWNYV4RQ1ZKWG8RE0RMFTVCM', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ177SB70PS1SSKMS676VA', 'Delete', '01JWNYV4RQ1ZKWG8RE0RMFTVCM', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ1A1MNC7X5AVVPM14EC3P', 'Update', '01JWNYV4RQ1ZKWG8RE0RMFTVCM', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ1D53FREVN8WX0Z7GZ1PS', 'View', '01JWNYV4RQ1ZKWG8RE0RMFTVCM', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		-- AuthzRoleSuite
		('01JWNZ29T8K173M5GA3HFXM1ME', 'Create', '01JWNYW23X8CMREJ2Y9349BAE4', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ2CMDCF51YV8KEW8QPQVZ', 'Delete', '01JWNYW23X8CMREJ2Y9349BAE4', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ2H9TPQSKEPTZ5KPHRE3H', 'Update', '01JWNYW23X8CMREJ2Y9349BAE4', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ2N37F8ZXC6MTC7QYNG6R', 'View', '01JWNYW23X8CMREJ2Y9349BAE4', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		-- AuthzGrantRequest
		('01JWNZ29T8K173M5GA3HF1TT5Y', 'Create', '01KF32E4TSX9DVV9KVKBS7GQ0T', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ2H9TPT1YY1TZ5KPHRE3H', 'Respond', '01KF32E4TSX9DVV9KVKBS7GQ0T', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ2CMDCF51YV895YY1PQVZ', 'Delete', '01KF32E4TSX9DVV9KVKBS7GQ0T', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ2N37F8ZXHIHI17QYNG6R', 'View', '01KF32E4TSX9DVV9KVKBS7GQ0T', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		-- AuthzRevokeRequest
		('01JWNZ29T8K173M5GA3HF911GT', 'Create', '01KF328VKN38VN16RG17C8BECB', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWN3TTKDCF51YV895YY1PQVZ', 'Delete', '01KF328VKN38VN16RG17C8BECB', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ2NTDTKZXHIHI17QYNG6R', 'View', '01KF328VKN38VN16RG17C8BECB', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		-- IdentityUser
		('01JWNZ3KA0ARGT9DAHQ1E6NZV0', 'Create', '01JWNYQ2A04PPV4135QGXX7W0M', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ3PTQXAWE8R3HDTYVAQEK', 'Delete', '01JWNYQ2A04PPV4135QGXX7W0M', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ3TQ7AVCFDTSS0VHXHCAB', 'Update', '01JWNYQ2A04PPV4135QGXX7W0M', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ3XVWDP73JGHRRBFAHQYJ', 'View', '01JWNYQ2A04PPV4135QGXX7W0M', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		-- IdentityGroup
		('01JWNZ4QY0ECRHAKR0ERQW97HW', 'Create', '01JWNYQNN68RKN62TNG5K0CPCE', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ4V0ZDTEEMZPRZF6282SP', 'Delete', '01JWNYQNN68RKN62TNG5K0CPCE', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ4Y4342HKE30Y4KE8MA8K', 'Update', '01JWNYQNN68RKN62TNG5K0CPCE', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ516R64X2S8A7STFXTP9B', 'View', '01JWNYQNN68RKN62TNG5K0CPCE', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		-- IdentityOrganization
		('01JWNZ5KW6WC643VXGKV1D0J64', 'Create', '01JWNYQTHN6JTRNWKJ1BMCYB80', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ5PZP72SZVA3GVHRZW3RG', 'Delete', '01JWNYQTHN6JTRNWKJ1BMCYB80', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ5SS046J9JVXS6WN316QB', 'Update', '01JWNYQTHN6JTRNWKJ1BMCYB80', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ5WJ1TM7P43W7FMENADTR', 'View', '01JWNYQTHN6JTRNWKJ1BMCYB80', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		-- IdentityHierarchyLevel
		('01JWNZ6NSG3ZWY82PEH1ERDZ5C', 'Create', '01JWNYRSKZ56EAE2QRFHJWRZPT', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ6SYC65GEMJJ6BRNTEXFC', 'Delete', '01JWNYRSKZ56EAE2QRFHJWRZPT', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ6XAZBQ8D11ETYGAN4N01', 'Update', '01JWNYRSKZ56EAE2QRFHJWRZPT', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNZ70QBW0B0KRMR5CNR56KX', 'View', '01JWNYRSKZ56EAE2QRFHJWRZPT', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWNYWE9FBX2WTMYZMR9HQQT1', 'Move', '01JWNYRSKZ56EAE2QRFHJWRZPT', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JX0PKTPPP5CN780TAPMK846J', 'ManageUsers', '01JWNYRSKZ56EAE2QRFHJWRZPT', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'authz_roles'
	) THEN
		INSERT INTO "authz_roles" ("id", "name", "description", "owner_type", "owner_ref", "is_requestable", "is_required_attachment", "is_required_comment", "created_at", "created_by", "etag") VALUES
		('01JWP72JJCDT4M0J8MSS51MN3T', 'Domain Administrator', 'Granted with all actions on all resources regardless of scope', 'user', '01JWNMZ36QHC7CQQ748H9NQ6J6', false, false, true, NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWP80E084MTYF2C882WNR6MJ', 'Identity module Readonly', 'Granted with view action on Users and Groups in Identity module, except with Owner user', 'user', '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWP80NHTHXSZDB1MZJXQ0MGQ', 'Identity module Administrator', 'Granted with all actions on all resources in Identity module, except with Owner user', 'user', '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWPB7TC3CG1EB567WYQCJM79', 'Identity module User Manager', 'Granted with all actions on on Users and Groups in Identity module, except with Owner user', 'user', '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWP80S5RXP8BD4YCY8ZHP7NZ', 'Authorize module Readonly', 'Granted with view action on all resources in Authorize module', 'user', '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWP80WR22SAG8Z7EYKDB00K6', 'Authorize module Administrator', 'Granted with all actions on all resources in Authorize module', 'user', '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWP810BRSH9GWCYQC463K012', 'Authorize module Moderator', 'Granted with all actions on Resource, Action, Role and Role Suite in Authorize module, but not allowed to delete the ones which are associated with an Entitlement', 'user', '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWP72RCDK8NVBJMZTWJK9R97', 'Org Administrator (My Company)', 'Granted with all actions on all resources regardless of hierarchy level in the organization My Company', 'user', '01JWNMZ36QHC7CQQ748H9NQ6J6', true, false, true, NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'authz_role_suites'
	) THEN
		INSERT INTO "authz_role_suites" ("id", "name", "description", "owner_type", "owner_ref", "is_requestable", "is_required_attachment", "is_required_comment", "created_at", "created_by", "etag") VALUES
		('01JWP9MVYX0K24R9H81SZEM7CE', 'Domain User Suite', 'Grant basic privileges to all users in the domain', 'user', '01JWNMZ36QHC7CQQ748H9NQ6J6', false, false, false, NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'authz_role_rolesuite'
	) THEN
		INSERT INTO "authz_role_rolesuite" ("role_id", "role_suite_id") VALUES
		-- Domain User Suite = Identity module Readonly + Authorize module Readonly
		('01JWP80E084MTYF2C882WNR6MJ', '01JWP9MVYX0K24R9H81SZEM7CE'),
		('01JWP80S5RXP8BD4YCY8ZHP7NZ', '01JWP9MVYX0K24R9H81SZEM7CE');
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'authz_role_suite_user'
	) THEN
		INSERT INTO "authz_role_suite_user" ("role_suite_id", "receiver_type", "receiver_ref", "approver_id") VALUES
		-- Domain User Suite => group Domain Users
		('01JWP9MVYX0K24R9H81SZEM7CE', 'group', '01JWNXBR5QJBH7PE9PQ9FW746V', '01JWNNJGS70Y07MBEV3AQ0M526');
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'authz_role_user'
	) THEN
		INSERT INTO "authz_role_user" ("role_id", "receiver_type", "receiver_ref", "approver_id") VALUES
		-- Identity module User Manager role => user 01JZQFDH0N51Q3BFQFMFFGSCSV
		('01JWPB7TC3CG1EB567WYQCJM79', 'user', '01JZQFDH0N51Q3BFQFMFFGSCSV', '01JWNNJGS70Y07MBEV3AQ0M526');
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'authz_entitlements'
	) THEN
		INSERT INTO "authz_entitlements" ("id", "name", "action_expr", "created_at", "created_by", "action_id", "resource_id", "etag") VALUES
		-- Domain Administrator: All actions on all resources
		('01JWP88N498RQS88TYVJ4Z20EX', 'Domain Administrator - All Permissions', '*:*', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', NULL, NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),

		-- Identity module Readonly: View actions on Identity resources
		('01JWP8EARV3B9A1HWFPMQZQ6HZ', 'Identity Readonly - View Users', 'IdentityUser:View', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', '01JWNZ3XVWDP73JGHRRBFAHQYJ', '01JWNYQ2A04PPV4135QGXX7W0M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWP8EFENXFNN17GSEJP0RCXZ', 'Identity Readonly - View Groups', 'IdentityGroup:View', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', '01JWNZ516R64X2S8A7STFXTP9B', '01JWNYQNN68RKN62TNG5K0CPCE', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),

		-- Identity module Administrator: All actions on Identity resources
		('01JWP8KSP3Q3YH6RKND552DWRR', 'Identity Admin - All User Actions', 'IdentityUser:*', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', NULL, '01JWNYQ2A04PPV4135QGXX7W0M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWP8KP39GKAH67FEAC7TZ631', 'Identity Admin - All Group Actions', 'IdentityGroup:*', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', NULL, '01JWNYQNN68RKN62TNG5K0CPCE', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWP8KK6K1M9WMP59BBAGEMB1', 'Identity Admin - All Organization Actions', 'IdentityOrganization:*', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', NULL, '01JWNYQTHN6JTRNWKJ1BMCYB80', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWP8KG15N26CRWNRM6F5CB29', 'Identity Admin - All HierarchyLevel Actions', 'IdentityHierarchyLevel:*', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', NULL, '01JWNYRSKZ56EAE2QRFHJWRZPT', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),

		-- Authorize module Readonly: View actions on Authorize resources
		('01JWPA3A4J2644C24V86419A2V', 'Authorize Readonly - View Entitlements', 'AuthzEntitlement:View', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', '01JWNYK975JE0PKC1NRAB68D9C', '01JWNX10DN9E2PVZTVPJNTWR7D', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),

		-- Authorize module Administrator: All actions on Authorize resources
		('01JWPA35MPHG33G77FKQNYJS21', 'Authorize Admin - All Resource Actions', 'AuthzResource:*', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', NULL, '01JWNYWE9FBX2WTMYZMR9XHHX6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWPA3232EYYN4HQMWBBV345B', 'Authorize Admin - All Role Actions', 'AuthzRole:*', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', NULL, '01JWNYV4RQ1ZKWG8RE0RMFTVCM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWPA2YP0HTKY290T7570N7QF', 'Authorize Admin - All RoleSuite Actions', 'AuthzRoleSuite:*', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', NULL, '01JWNYW23X8CMREJ2Y9349BAE4', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWPA2SEP8T4S2VKAKYDYME64', 'Authorize Admin - All Entitlement Actions', 'AuthzEntitlement:*', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', NULL, '01JWNX10DN9E2PVZTVPJNTWR7D', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),

		-- Authorize module Moderator: All actions on Resource, Action, Role and Role Suite (but not Entitlement)
		('01JWP8KCVWVYDSQ6C8SNDQD5F6', 'Authorize Moderator - Create Resources', 'AuthzResource:Create', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', '01JWNYZ5EPJJMA3D367XMYEMM2', '01JWNYWE9FBX2WTMYZMR9XHHX6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWP8KCVWVYDSQ6C8SNDQD5F7', 'Authorize Moderator - Update Resources', 'AuthzResource:Update', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', '01JWY2HF8E72PQM8QHY0CHSVBT', '01JWNYWE9FBX2WTMYZMR9XHHX6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWP8KCVWVYDSQ6C8SNDQD5F9', 'Authorize Moderator - View Resources', 'AuthzResource:View', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', '01JWNYZEVSH78T2SH7WP47KDRM', '01JWNYWE9FBX2WTMYZMR9XHHX6', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWP8K9QE8PW6BKZZ6EW9C9Z5', 'Authorize Moderator - Create Roles', 'AuthzRole:Create', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', '01JWNZ14EZ00S2HWZD3Z7VANJK', '01JWNYV4RQ1ZKWG8RE0RMFTVCM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWP8K9QE8PW6BKZZ6EW9C9Z6', 'Authorize Moderator - Update Roles', 'AuthzRole:Update', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', '01JWNZ1A1MNC7X5AVVPM14EC3P', '01JWNYV4RQ1ZKWG8RE0RMFTVCM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWP8K9QE8PW6BKZZ6EW9C9Z7', 'Authorize Moderator - View Roles', 'AuthzRole:View', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', '01JWNZ1D53FREVN8WX0Z7GZ1PS', '01JWNYV4RQ1ZKWG8RE0RMFTVCM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWP8K6DQKNR910S6CP90P24N', 'Authorize Moderator - Create RoleSuites', 'AuthzRoleSuite:Create', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', '01JWNZ29T8K173M5GA3HFXM1ME', '01JWNYW23X8CMREJ2Y9349BAE4', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWP8K6DQKNR910S6CP90P25N', 'Authorize Moderator - Update RoleSuites', 'AuthzRoleSuite:Update', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', '01JWNZ2H9TPQSKEPTZ5KPHRE3H', '01JWNYW23X8CMREJ2Y9349BAE4', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint),
		('01JWP8K6DQKNR910S6CP90P26N', 'Authorize Moderator - View RoleSuites', 'AuthzRoleSuite:View', NOW(), '01JWNNJGS70Y07MBEV3AQ0M526', '01JWNZ2N37F8ZXC6MTC7QYNG6R', '01JWNYW23X8CMREJ2Y9349BAE4', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'authz_entitlement_assignments'
	) THEN
		INSERT INTO "authz_entitlement_assignments" ("id", "entitlement_id", "subject_type", "subject_ref", "resolved_expr", "action_name", "resource_name", "scope_ref") VALUES
		-- Domain Administrator role gets all permissions
		('01JWP88N498RQS88TYVJ4Z20F0', '01JWP88N498RQS88TYVJ4Z20EX', 'nikki_role', '01JWP72JJCDT4M0J8MSS51MN3T', '01JWP72JJCDT4M0J8MSS51MN3T:*:*:*', NULL, NULL, NULL),
		
		-- Identity module Readonly role gets view permissions
		('01JWP8EARV3B9A1HWFPMQZQ6H1', '01JWP8EARV3B9A1HWFPMQZQ6HZ', 'nikki_role', '01JWP80E084MTYF2C882WNR6MJ', '01JWP80E084MTYF2C882WNR6MJ:01JWNY20G23KD4RV5VWYABQYH1:IdentityUser:View', 'View', 'IdentityUser', '01JWNY20G23KD4RV5VWYABQYH1'),
		('01JWP8EFENXFNN17GSEJP0RCX1', '01JWP8EFENXFNN17GSEJP0RCXZ', 'nikki_role', '01JWP80E084MTYF2C882WNR6MJ', '01JWP80E084MTYF2C882WNR6MJ:*:IdentityGroup:View', 'View', 'IdentityGroup', NULL),

		-- Identity module Administrator role gets all permissions on Identity resources
		('01JWP8KSP3Q3YH6RKND552DWR1', '01JWP8KSP3Q3YH6RKND552DWRR', 'nikki_role', '01JWP80NHTHXSZDB1MZJXQ0MGQ', '01JWP80NHTHXSZDB1MZJXQ0MGQ:01JWNY20G23KD4RV5VWYABKDT1:IdentityUser:*', NULL, 'IdentityUser', '01JWNY20G23KD4RV5VWYABKDT1'),
		('01JWP8KP39GKAH67FEAC7TZ632', '01JWP8KP39GKAH67FEAC7TZ631', 'nikki_role', '01JWP80NHTHXSZDB1MZJXQ0MGQ', '01JWP80NHTHXSZDB1MZJXQ0MGQ:*:IdentityGroup:*', NULL, 'IdentityGroup', NULL),
		('01JWP8KK6K1M9WMP59BBAGEMB2', '01JWP8KK6K1M9WMP59BBAGEMB1', 'nikki_role', '01JWP80NHTHXSZDB1MZJXQ0MGQ', '01JWP80NHTHXSZDB1MZJXQ0MGQ:*:IdentityOrganization:*', NULL, 'IdentityOrganization', NULL),
		('01JWP8KG15N26CRWNRM6F5CB30', '01JWP8KG15N26CRWNRM6F5CB29', 'nikki_role', '01JWP80NHTHXSZDB1MZJXQ0MGQ', '01JWP80NHTHXSZDB1MZJXQ0MGQ:*:IdentityHierarchyLevel:*', NULL, 'IdentityHierarchyLevel', NULL),

		-- Identity module User Manager role gets all permissions on Users and Groups
		('01JWPBJFYTYJJTM799RBTKFE21', '01JWP8KSP3Q3YH6RKND552DWRR', 'nikki_role', '01JWPB7TC3CG1EB567WYQCJM79', '01JWPB7TC3CG1EB567WYQCJM79:01JWNY20G23KD4RV5VWYABKDT1:IdentityUser:*', NULL, 'IdentityUser', '01JWNY20G23KD4RV5VWYABKDT1'),
		('01JWPBJPV5H1ST6H7N21CMZ9YO', '01JWP8KP39GKAH67FEAC7TZ631', 'nikki_role', '01JWPB7TC3CG1EB567WYQCJM79', '01JWPB7TC3CG1EB567WYQCJM79:*:IdentityGroup:*', NULL, 'IdentityGroup', NULL),

		-- Authorize module Readonly role gets view permissions on Authorize resources
		('01JWPA3M96PCSR4899SV91A8RQ', '01JWP8KCVWVYDSQ6C8SNDQD5F9', 'nikki_role', '01JWP80S5RXP8BD4YCY8ZHP7NZ', '01JWP80S5RXP8BD4YCY8ZHP7NZ:*:AuthzResource:View', 'View', 'AuthzResource', NULL),
		('01JWPA3DSPDS6NV8KGZAZZW3R3', '01JWP8K6DQKNR910S6CP90P26N', 'nikki_role', '01JWP80S5RXP8BD4YCY8ZHP7NZ', '01JWP80S5RXP8BD4YCY8ZHP7NZ:*:AuthzRoleSuite:View', 'View', 'AuthzRoleSuite', NULL),
		('01JWPA3A4J2644C24V86419A2W', '01JWPA3A4J2644C24V86419A2V', 'nikki_role', '01JWP80S5RXP8BD4YCY8ZHP7NZ', '01JWPA3A4J2644C24V86419A2W:*:AuthzEntitlement:View', 'View', 'AuthzEntitlement', NULL),

		-- Authorize module Administrator role gets all permissions on Authorize resources
		('01JWPA35MPHG33G77FKQNYJS2A', '01JWPA35MPHG33G77FKQNYJS21', 'nikki_role', '01JWP80WR22SAG8Z7EYKDB00K6', '01JWP80WR22SAG8Z7EYKDB00K6:*:AuthzResource:*', NULL, 'AuthzResource', NULL),
		('01JWPA3232EYYN4HQMWBBV34ZB', '01JWPA3232EYYN4HQMWBBV345B', 'nikki_role', '01JWP80WR22SAG8Z7EYKDB00K6', '01JWP80WR22SAG8Z7EYKDB00K6:*:AuthzRole:*', NULL, 'AuthzRole', NULL),
		('01JWPA2YP0HTKY290T7570NT3Z', '01JWPA2YP0HTKY290T7570N7QF', 'nikki_role', '01JWP80WR22SAG8Z7EYKDB00K6', '01JWP80WR22SAG8Z7EYKDB00K6:*:AuthzRoleSuite:*', NULL, 'AuthzRoleSuite', NULL),
		('01JWPA2SEP8T4S2VKAKYDYM2TS', '01JWPA2SEP8T4S2VKAKYDYME64', 'nikki_role', '01JWP80WR22SAG8Z7EYKDB00K6', '01JWP80WR22SAG8Z7EYKDB00K6:*:AuthzEntitlement:*', NULL, 'AuthzEntitlement', NULL),

		-- Authorize module Moderator role gets all permissions on Resource, Action, Role and Role Suite
		('01JWP8KCVWVYDSQ6C8SNDQDK55', '01JWP8KCVWVYDSQ6C8SNDQD5F6', 'nikki_role', '01JWP810BRSH9GWCYQC463K012', '01JWP810BRSH9GWCYQC463K012:*:AuthzResource:Create', 'Create', 'AuthzResource', NULL),
		('01JWP8KCVWVYDSQ6C8SNDQDDM2', '01JWP8KCVWVYDSQ6C8SNDQD5F7', 'nikki_role', '01JWP810BRSH9GWCYQC463K012', '01JWP810BRSH9GWCYQC463K012:*:AuthzResource:Update', 'Update', 'AuthzResource', NULL),
		('01JWP8KCVWVYDSQ6C8SNDQKDk1', '01JWP8KCVWVYDSQ6C8SNDQD5F9', 'nikki_role', '01JWP810BRSH9GWCYQC463K012', '01JWP810BRSH9GWCYQC463K012:*:AuthzResource:View', 'View', 'AuthzResource', NULL),
		('01JWP8K9QE8PW6BKZZ6EW9TDT2', '01JWP8K9QE8PW6BKZZ6EW9C9Z5', 'nikki_role', '01JWP810BRSH9GWCYQC463K012', '01JWP810BRSH9GWCYQC463K012:*:AuthzRole:Create', 'Create', 'AuthzRole', NULL),
		('01JWP8K9QE8PW6BKZZ6EW9YE1N', '01JWP8K9QE8PW6BKZZ6EW9C9Z6', 'nikki_role', '01JWP810BRSH9GWCYQC463K012', '01JWP810BRSH9GWCYQC463K012:*:AuthzRole:Update', 'Update', 'AuthzRole', NULL),
		('01JWP8K9QE8PW6BKZZ6EWT1VD7', '01JWP8K9QE8PW6BKZZ6EW9C9Z7', 'nikki_role', '01JWP810BRSH9GWCYQC463K012', '01JWP810BRSH9GWCYQC463K012:*:AuthzRole:View', 'View', 'AuthzRole', NULL),
		('01JWP8K6DQKNR910S6CP90GNNL', '01JWP8K6DQKNR910S6CP90P24N', 'nikki_role', '01JWP810BRSH9GWCYQC463K012', '01JWP810BRSH9GWCYQC463K012:*:AuthzRoleSuite:Create', 'Create', 'AuthzRoleSuite', NULL),
		('01JWP8K6DQKNR910S6CPDMBZT1', '01JWP8K6DQKNR910S6CP90P25N', 'nikki_role', '01JWP810BRSH9GWCYQC463K012', '01JWP810BRSH9GWCYQC463K012:*:AuthzRoleSuite:Update', 'Update', 'AuthzRoleSuite', NULL),
		('01JWP8K6DQKNR910S6CP9KKJQ2', '01JWP8K6DQKNR910S6CP90P26N', 'nikki_role', '01JWP810BRSH9GWCYQC463K012', '01JWP810BRSH9GWCYQC463K012:*:AuthzRoleSuite:View', 'View', 'AuthzRoleSuite', NULL),

		-- Seed for testing
		('01JWP88N498RQS88TDTYKD17B3', '01JWP88N498RQS88TYVJ4Z20EX', 'nikki_user', '01JWNMZ36QHC7CQQ748H9NQ6J6', '01JWNMZ36QHC7CQQ748H9NQ6J6:*:*:*', NULL, NULL, NULL),
		('01JWP88N498RQS88TDTYKD17B4', '01JWP8EARV3B9A1HWFPMQZQ6HZ', 'nikki_user', '01JZQFY6EXRG0959Z95Y2EM3AM', '01JZQFY6EXRG0959Z95Y2EM3AM:01JWNY20G23KD4RV5VWYABQYH1:IdentityUser:View', 'View', 'IdentityUser', '01JWNY20G23KD4RV5VWYABQYH1'),

		('01JWPA35MPHG33G77FKQGOPR34', '01JWPA35MPHG33G77FKQNYJS21', 'nikki_group', '01JWNXBR5QJBH7PE9PQ9FW746V', '01JWNXBR5QJBH7PE9PQ9FW746V:*:AuthzResource:*', NULL, 'AuthzResource', NULL);
	END IF;
END $$;