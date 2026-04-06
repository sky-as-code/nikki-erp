DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'ident_organizations'
	) THEN
		INSERT INTO "ident_organizations" ("id", "address", "display_name", "legal_name", "phone_number", "slug", "etag", "created_at", "updated_at") VALUES
		('01JWNY20G23KD4RV5VWYABQYHD', NULL, 'My Company', NULL, NULL, 'my-company', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K02G6J1CYAN9K8V4PAGSQ5Z8', NULL, 'Old Company', NULL, NULL, 'old-company', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H7M2K9VW3P5R7XQJY2C1Z9', NULL, 'Tech Solutions Ltd', NULL, NULL, 'tech-solutions', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'ident_org_units'
	) THEN
		INSERT INTO "ident_org_units" ("id", "name", "description", "path", "parent_id", "etag", "created_at", "updated_at", "org_id") VALUES
		('01K1H8N3L0WX4Q6S8YRKT3D2A2', 'CEO', NULL, ARRAY['01JWNY20G23KD4RV5VWYABQYHD', '01K1H8N3L0WX4Q6S8YRKT3D2A2']::varchar[], NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL, '01JWNY20G23KD4RV5VWYABQYHD'),
		('01K1H8N3L0WX4Q6S8YRKT3D2A3', 'Director', NULL, ARRAY['01JWNY20G23KD4RV5VWYABQYHD', '01K1H8N3L0WX4Q6S8YRKT3D2A2', '01K1H8N3L0WX4Q6S8YRKT3D2A3']::varchar[], '01K1H8N3L0WX4Q6S8YRKT3D2A2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL, '01JWNY20G23KD4RV5VWYABQYHD'),
		('01K1H8N3L0WX4Q6S8YRKT3D2A4', 'Team Lead', NULL, ARRAY['01JWNY20G23KD4RV5VWYABQYHD', '01K1H8N3L0WX4Q6S8YRKT3D2A2', '01K1H8N3L0WX4Q6S8YRKT3D2A3', '01K1H8N3L0WX4Q6S8YRKT3D2A4']::varchar[], '01K1H8N3L0WX4Q6S8YRKT3D2A3', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL, '01JWNY20G23KD4RV5VWYABQYHD'),
		('01K1H8N3L0WX4Q6S8YRKT3D2A5', 'VP Engineering', NULL, ARRAY['01K1H7M2K9VW3P5R7XQJY2C1Z9', '01K1H8N3L0WX4Q6S8YRKT3D2A5']::varchar[], NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL, '01K1H7M2K9VW3P5R7XQJY2C1Z9'),
		('01K1H8N3L0WX4Q6S8YRKT3D2A6', 'Engineering Manager', NULL, ARRAY['01K1H7M2K9VW3P5R7XQJY2C1Z9', '01K1H8N3L0WX4Q6S8YRKT3D2A5', '01K1H8N3L0WX4Q6S8YRKT3D2A6']::varchar[], '01K1H8N3L0WX4Q6S8YRKT3D2A5', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL, '01K1H7M2K9VW3P5R7XQJY2C1Z9'),
		('01K1H8N3L0WX4Q6S8YRKT3D2C0', 'Legacy Manager', NULL, ARRAY['01K02G6J1CYAN9K8V4PAGSQ5Z8', '01K1H8N3L0WX4Q6S8YRKT3D2C0']::varchar[], NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL, '01K02G6J1CYAN9K8V4PAGSQ5Z8'),
		('01K1H8N3L0WX4Q6S8YRKT3D2C1', 'Support Specialist', NULL, ARRAY['01K02G6J1CYAN9K8V4PAGSQ5Z8', '01K1H8N3L0WX4Q6S8YRKT3D2C0', '01K1H8N3L0WX4Q6S8YRKT3D2C1']::varchar[], '01K1H8N3L0WX4Q6S8YRKT3D2C0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL, '01K02G6J1CYAN9K8V4PAGSQ5Z8'),
		('01K1H8N3L0WX4Q6S8YRKT3D2C2', 'Maintenance Technician', NULL, ARRAY['01K02G6J1CYAN9K8V4PAGSQ5Z8', '01K1H8N3L0WX4Q6S8YRKT3D2C0', '01K1H8N3L0WX4Q6S8YRKT3D2C2']::varchar[], '01K1H8N3L0WX4Q6S8YRKT3D2C0', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL, '01K02G6J1CYAN9K8V4PAGSQ5Z8');
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'ident_users'
	) THEN
		INSERT INTO "ident_users" ("id", "avatar_url", "display_name", "email", "status", "is_owner", "org_unit_id", "is_archived", "etag", "created_at", "updated_at") VALUES
		('01JWNNJGS70Y07MBEV3AQ0M526', NULL, 'System', 'system@nikki.com', 'active', NULL, NULL, FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01JWNMZ36QHC7CQQ748H9NQ6J6', NULL, 'Admin Owner', 'owner@nikki.com', 'active', TRUE, NULL, FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01JWNXT3EY7FG47VDJTEPTDC98', NULL, 'Nguyễn Văn An', 'nguyen.van.an@nikki.com', 'active', NULL, '01K1H8N3L0WX4Q6S8YRKT3D2A2', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01JWNXXTF8958VVYAV33MVVMDN', NULL, 'Trần Thị Bình', 'tran.thi.binh@nikki.com', 'active', NULL, '01K1H8N3L0WX4Q6S8YRKT3D2A3', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01JZQFDH0N51Q3BFQFMFFGSCSV', NULL, 'Lê Văn Cường', 'le.van.cuong@nikki.com', 'active', NULL, '01K1H8N3L0WX4Q6S8YRKT3D2A4', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01JZQFF9QEXH71P2CG9Y9MY8MM', NULL, 'Phạm Thị Dung', 'pham.thi.dung@nikki.com', 'locked', NULL, '01K1H8N3L0WX4Q6S8YRKT3D2A5', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01JZQFFDKY8T4JB8R6NSY1331J', NULL, 'Hoàng Văn Em', 'hoang.van.em@nikki.com', 'active', NULL, '01K1H8N3L0WX4Q6S8YRKT3D2A6', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01JZQFGVKZCTV7S310W0BDMWCS', NULL, 'Đặng Thị Phương', 'dang.thi.phuong@nikki.com', 'active', NULL, '01K1H8N3L0WX4Q6S8YRKT3D2C0', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01JZQFY6EXRG0959Z95Y2EM3AM', NULL, 'Võ Văn Giang', 'vo.van.giang@nikki.com', 'archived', NULL, '01K1H8N3L0WX4Q6S8YRKT3D2C1', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01JZQFZFK6GM2D5X6MYHWH6FND', NULL, 'Bùi Thị Hoa', 'bui.thi.hoa@nikki.com', 'active', NULL, '01K1H8N3L0WX4Q6S8YRKT3D2C2', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'ident_groups'
	) THEN
		INSERT INTO "ident_groups" ("id", "name", "description", "owner_id", "is_archived", "etag", "created_at", "updated_at") VALUES
		('01JWNXBR5QJBH7PE9PQ9FW746V', 'Domain Users', 'Default group for all domain users', '01JWNMZ36QHC7CQQ748H9NQ6J6', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2A0', 'Administrators', 'System administrators group', '01JWNMZ36QHC7CQQ748H9NQ6J6', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2A1', 'Project Managers', 'Project management team', '01JWNMZ36QHC7CQQ748H9NQ6J6', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2B0', 'Legacy Support', 'Legacy system support team', '01JWNMZ36QHC7CQQ748H9NQ6J6', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2B1', 'Archives Team', 'Archived content management', '01JWNMZ36QHC7CQQ748H9NQ6J6', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2B2', 'Maintenance Group', 'System maintenance personnel', '01JWNMZ36QHC7CQQ748H9NQ6J6', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'ident_group_user_rel'
	) THEN
		INSERT INTO "ident_group_user_rel" ("group_id", "user_id") VALUES
		('01JWNXBR5QJBH7PE9PQ9FW746V', '01JWNXT3EY7FG47VDJTEPTDC98'),
		('01JWNXBR5QJBH7PE9PQ9FW746V', '01JWNXXTF8958VVYAV33MVVMDN'),
		('01K1H8N3L0WX4Q6S8YRKT3D2A0', '01JZQFDH0N51Q3BFQFMFFGSCSV'),
		('01K1H8N3L0WX4Q6S8YRKT3D2A1', '01JZQFF9QEXH71P2CG9Y9MY8MM'),
		('01K1H8N3L0WX4Q6S8YRKT3D2A0', '01JZQFFDKY8T4JB8R6NSY1331J'),
		('01K1H8N3L0WX4Q6S8YRKT3D2B0', '01JZQFGVKZCTV7S310W0BDMWCS'),
		('01K1H8N3L0WX4Q6S8YRKT3D2B1', '01JZQFY6EXRG0959Z95Y2EM3AM'),
		('01K1H8N3L0WX4Q6S8YRKT3D2B2', '01JZQFZFK6GM2D5X6MYHWH6FND');
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'ident_org_user_rel'
	) THEN
		INSERT INTO "ident_org_user_rel" ("org_id", "user_id") VALUES
		('01JWNY20G23KD4RV5VWYABQYHD', '01JWNMZ36QHC7CQQ748H9NQ6J6'),
		('01K02G6J1CYAN9K8V4PAGSQ5Z8', '01JWNMZ36QHC7CQQ748H9NQ6J6'),
		('01K1H7M2K9VW3P5R7XQJY2C1Z9', '01JWNMZ36QHC7CQQ748H9NQ6J6'),
		('01JWNY20G23KD4RV5VWYABQYHD', '01JWNXT3EY7FG47VDJTEPTDC98'),
		('01JWNY20G23KD4RV5VWYABQYHD', '01JWNXXTF8958VVYAV33MVVMDN'),
		('01K02G6J1CYAN9K8V4PAGSQ5Z8', '01JZQFGVKZCTV7S310W0BDMWCS'),
		('01K02G6J1CYAN9K8V4PAGSQ5Z8', '01JZQFY6EXRG0959Z95Y2EM3AM'),
		('01K02G6J1CYAN9K8V4PAGSQ5Z8', '01JZQFZFK6GM2D5X6MYHWH6FND'),
		('01JWNY20G23KD4RV5VWYABQYHD', '01JZQFDH0N51Q3BFQFMFFGSCSV'),
		('01K1H7M2K9VW3P5R7XQJY2C1Z9', '01JZQFF9QEXH71P2CG9Y9MY8MM'),
		('01K1H7M2K9VW3P5R7XQJY2C1Z9', '01JZQFFDKY8T4JB8R6NSY1331J');
	END IF;
END $$;
