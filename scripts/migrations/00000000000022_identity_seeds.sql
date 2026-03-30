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
		WHERE table_schema = 'public' AND table_name = 'ident_hierarchy_levels'
	) THEN
		INSERT INTO "ident_hierarchy_levels" ("id", "created_at", "etag", "name", "org_id", "parent_id", "updated_at") VALUES
		('01K1H8N3L0WX4Q6S8YRKT3D2A2', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, 'CEO', '01JWNY20G23KD4RV5VWYABQYHD', NULL, NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2A3', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, 'Director', '01JWNY20G23KD4RV5VWYABQYHD', '01K1H8N3L0WX4Q6S8YRKT3D2A2', NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2A4', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, 'Team Lead', '01JWNY20G23KD4RV5VWYABQYHD', '01K1H8N3L0WX4Q6S8YRKT3D2A3', NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2A5', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, 'VP Engineering', '01K1H7M2K9VW3P5R7XQJY2C1Z9', NULL, NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2A6', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, 'Engineering Manager', '01K1H7M2K9VW3P5R7XQJY2C1Z9', '01K1H8N3L0WX4Q6S8YRKT3D2A5', NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2C0', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, 'Legacy Manager', '01K02G6J1CYAN9K8V4PAGSQ5Z8', NULL, NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2C1', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, 'Support Specialist', '01K02G6J1CYAN9K8V4PAGSQ5Z8', '01K1H8N3L0WX4Q6S8YRKT3D2C0', NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2C2', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, 'Maintenance Technician', '01K02G6J1CYAN9K8V4PAGSQ5Z8', '01K1H8N3L0WX4Q6S8YRKT3D2C0', NULL);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'ident_users'
	) THEN
		INSERT INTO "ident_users" ("id", "avatar_url", "created_at", "display_name", "email", "etag", "hierarchy_id", "is_archived", "is_owner", "status", "updated_at") VALUES
		('01JWNNJGS70Y07MBEV3AQ0M526', NULL, NOW(), 'System', 'system@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NULL, FALSE, NULL, 'active', NULL),
		('01JWNMZ36QHC7CQQ748H9NQ6J6', NULL, NOW(), 'Admin Owner', 'owner@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NULL, FALSE, TRUE, 'active', NULL),
		('01JWNXT3EY7FG47VDJTEPTDC98', NULL, NOW(), 'Nguyễn Văn An', 'nguyen.van.an@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, '01K1H8N3L0WX4Q6S8YRKT3D2A2', FALSE, NULL, 'active', NULL),
		('01JWNXXTF8958VVYAV33MVVMDN', NULL, NOW(), 'Trần Thị Bình', 'tran.thi.binh@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, '01K1H8N3L0WX4Q6S8YRKT3D2A3', FALSE, NULL, 'active', NULL),
		('01JZQFDH0N51Q3BFQFMFFGSCSV', NULL, NOW(), 'Lê Văn Cường', 'le.van.cuong@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, '01K1H8N3L0WX4Q6S8YRKT3D2A4', FALSE, NULL, 'active', NULL),
		('01JZQFF9QEXH71P2CG9Y9MY8MM', NULL, NOW(), 'Phạm Thị Dung', 'pham.thi.dung@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, '01K1H8N3L0WX4Q6S8YRKT3D2A5', FALSE, NULL, 'locked', NULL),
		('01JZQFFDKY8T4JB8R6NSY1331J', NULL, NOW(), 'Hoàng Văn Em', 'hoang.van.em@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, '01K1H8N3L0WX4Q6S8YRKT3D2A6', FALSE, NULL, 'active', NULL),
		('01JZQFGVKZCTV7S310W0BDMWCS', NULL, NOW(), 'Đặng Thị Phương', 'dang.thi.phuong@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, '01K1H8N3L0WX4Q6S8YRKT3D2C0', FALSE, NULL, 'active', NULL),
		('01JZQFY6EXRG0959Z95Y2EM3AM', NULL, NOW(), 'Võ Văn Giang', 'vo.van.giang@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, '01K1H8N3L0WX4Q6S8YRKT3D2C1', FALSE, NULL, 'archived', NULL),
		('01JZQFZFK6GM2D5X6MYHWH6FND', NULL, NOW(), 'Bùi Thị Hoa', 'bui.thi.hoa@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, '01K1H8N3L0WX4Q6S8YRKT3D2C2', FALSE, NULL, 'active', NULL);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'ident_groups'
	) THEN
		INSERT INTO "ident_groups" ("id", "name", "description", "is_archived", "etag", "created_at", "updated_at") VALUES
		('01JWNXBR5QJBH7PE9PQ9FW746V', 'Domain Users', 'Default group for all domain users', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2A0', 'Administrators', 'System administrators group', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2A1', 'Project Managers', 'Project management team', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2B0', 'Legacy Support', 'Legacy system support team', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2B1', 'Archives Team', 'Archived content management', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2B2', 'Maintenance Group', 'System maintenance personnel', FALSE, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text, NOW(), NULL);
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'ident_user_group_rel'
	) THEN
		INSERT INTO "ident_user_group_rel" ("user_id", "group_id") VALUES
		('01JWNXT3EY7FG47VDJTEPTDC98', '01JWNXBR5QJBH7PE9PQ9FW746V'),
		('01JWNXXTF8958VVYAV33MVVMDN', '01JWNXBR5QJBH7PE9PQ9FW746V'),
		('01JZQFDH0N51Q3BFQFMFFGSCSV', '01K1H8N3L0WX4Q6S8YRKT3D2A0'),
		('01JZQFF9QEXH71P2CG9Y9MY8MM', '01K1H8N3L0WX4Q6S8YRKT3D2A1'),
		('01JZQFFDKY8T4JB8R6NSY1331J', '01K1H8N3L0WX4Q6S8YRKT3D2A0'),
		('01JZQFGVKZCTV7S310W0BDMWCS', '01K1H8N3L0WX4Q6S8YRKT3D2B0'),
		('01JZQFY6EXRG0959Z95Y2EM3AM', '01K1H8N3L0WX4Q6S8YRKT3D2B1'),
		('01JZQFZFK6GM2D5X6MYHWH6FND', '01K1H8N3L0WX4Q6S8YRKT3D2B2');
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'ident_user_org_rel'
	) THEN
		INSERT INTO "ident_user_org_rel" ("user_id", "org_id") VALUES
		('01JWNMZ36QHC7CQQ748H9NQ6J6', '01JWNY20G23KD4RV5VWYABQYHD'),
		('01JWNMZ36QHC7CQQ748H9NQ6J6', '01K02G6J1CYAN9K8V4PAGSQ5Z8'),
		('01JWNMZ36QHC7CQQ748H9NQ6J6', '01K1H7M2K9VW3P5R7XQJY2C1Z9'),
		('01JWNXT3EY7FG47VDJTEPTDC98', '01JWNY20G23KD4RV5VWYABQYHD'),
		('01JWNXXTF8958VVYAV33MVVMDN', '01JWNY20G23KD4RV5VWYABQYHD'),
		('01JZQFGVKZCTV7S310W0BDMWCS', '01K02G6J1CYAN9K8V4PAGSQ5Z8'),
		('01JZQFY6EXRG0959Z95Y2EM3AM', '01K02G6J1CYAN9K8V4PAGSQ5Z8'),
		('01JZQFZFK6GM2D5X6MYHWH6FND', '01K02G6J1CYAN9K8V4PAGSQ5Z8'),
		('01JZQFDH0N51Q3BFQFMFFGSCSV', '01JWNY20G23KD4RV5VWYABQYHD'),
		('01JZQFF9QEXH71P2CG9Y9MY8MM', '01K1H7M2K9VW3P5R7XQJY2C1Z9'),
		('01JZQFFDKY8T4JB8R6NSY1331J', '01K1H7M2K9VW3P5R7XQJY2C1Z9');
	END IF;
END $$;
