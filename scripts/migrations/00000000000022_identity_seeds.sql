DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'ident_organizations'
	) THEN
		INSERT INTO "ident_organizations" ("id", "created_at", "display_name", "etag", "slug", "status") VALUES
		('01JWNY20G23KD4RV5VWYABQYHD', NOW(), 'My Company', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 'my-company', 'active'),
		('01K02G6J1CYAN9K8V4PAGSQ5Z8', NOW(), 'Old Company', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 'old-company', 'archived'),
		('01K1H7M2K9VW3P5R7XQJY2C1Z9', NOW(), 'Tech Solutions Ltd', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 'tech-solutions', 'active');
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'ident_users'
	) THEN
		INSERT INTO "ident_users" ("id", "created_at", "display_name", "email", "etag", "is_owner", "status") VALUES
		('01JWNNJGS70Y07MBEV3AQ0M526', NOW(), 'System', 'system@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NULL, 'active'),
		('01JWNMZ36QHC7CQQ748H9NQ6J6', NOW(), 'Admin Owner', 'owner@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, TRUE, 'active'),
		('01JWNXT3EY7FG47VDJTEPTDC98', NOW(), 'Nguyễn Văn An', 'nguyen.van.an@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NULL, 'active'),
		('01JWNXXTF8958VVYAV33MVVMDN', NOW(), 'Trần Thị Bình', 'tran.thi.binh@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NULL, 'active'),
		('01JZQFDH0N51Q3BFQFMFFGSCSV', NOW(), 'Lê Văn Cường', 'le.van.cuong@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NULL, 'active'),
		('01JZQFF9QEXH71P2CG9Y9MY8MM', NOW(), 'Phạm Thị Dung', 'pham.thi.dung@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NULL, 'locked'),
		('01JZQFFDKY8T4JB8R6NSY1331J', NOW(), 'Hoàng Văn Em', 'hoang.van.em@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NULL, 'active'),
		('01JZQFGVKZCTV7S310W0BDMWCS', NOW(), 'Đặng Thị Phương', 'dang.thi.phuong@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NULL, 'active'),
		('01JZQFY6EXRG0959Z95Y2EM3AM', NOW(), 'Võ Văn Giang', 'vo.van.giang@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NULL, 'archived'),
		('01JZQFZFK6GM2D5X6MYHWH6FND', NOW(), 'Bùi Thị Hoa', 'bui.thi.hoa@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NULL, 'active');
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'ident_groups'
	) THEN
		INSERT INTO "ident_groups" ("id", "name", "description", "etag", "created_at", "org_id") VALUES
		('01JWNXBR5QJBH7PE9PQ9FW746V', 'Domain Users', 'Default group for all domain users', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NOW(), '01JWNY20G23KD4RV5VWYABQYHD'),
		('01K1H8N3L0WX4Q6S8YRKT3D2A0', 'Administrators', 'System administrators group', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NOW(), '01JWNY20G23KD4RV5VWYABQYHD'),
		('01K1H8N3L0WX4Q6S8YRKT3D2A1', 'Project Managers', 'Project management team', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NOW(), '01K1H7M2K9VW3P5R7XQJY2C1Z9'),
		('01K1H8N3L0WX4Q6S8YRKT3D2B0', 'Legacy Support', 'Legacy system support team', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NOW(), '01K02G6J1CYAN9K8V4PAGSQ5Z8'),
		('01K1H8N3L0WX4Q6S8YRKT3D2B1', 'Archives Team', 'Archived content management', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NOW(), '01K02G6J1CYAN9K8V4PAGSQ5Z8'),
		('01K1H8N3L0WX4Q6S8YRKT3D2B2', 'Maintenance Group', 'System maintenance personnel', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NOW(), '01K02G6J1CYAN9K8V4PAGSQ5Z8');
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
		('01JWNXT3EY7FG47VDJTEPTDC98', '01JWNY20G23KD4RV5VWYABQYHD'),
		('01JWNXXTF8958VVYAV33MVVMDN', '01JWNY20G23KD4RV5VWYABQYHD'),
		('01JZQFGVKZCTV7S310W0BDMWCS', '01K02G6J1CYAN9K8V4PAGSQ5Z8'),
		('01JZQFY6EXRG0959Z95Y2EM3AM', '01K02G6J1CYAN9K8V4PAGSQ5Z8'),
		('01JZQFZFK6GM2D5X6MYHWH6FND', '01K02G6J1CYAN9K8V4PAGSQ5Z8'),
		('01JZQFDH0N51Q3BFQFMFFGSCSV', '01JWNY20G23KD4RV5VWYABQYHD'),
		('01JZQFF9QEXH71P2CG9Y9MY8MM', '01K1H7M2K9VW3P5R7XQJY2C1Z9'),
		('01JZQFFDKY8T4JB8R6NSY1331J', '01K1H7M2K9VW3P5R7XQJY2C1Z9');
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'ident_hierarchy_levels'
	) THEN
		INSERT INTO "ident_hierarchy_levels" ("id", "created_at", "etag", "name", "org_id", "parent_id") VALUES
		('01K1H8N3L0WX4Q6S8YRKT3D2A2', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 'CEO', '01JWNY20G23KD4RV5VWYABQYHD', NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2A3', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 'Director', '01JWNY20G23KD4RV5VWYABQYHD', '01K1H8N3L0WX4Q6S8YRKT3D2A2'),
		('01K1H8N3L0WX4Q6S8YRKT3D2A4', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 'Team Lead', '01JWNY20G23KD4RV5VWYABQYHD', '01K1H8N3L0WX4Q6S8YRKT3D2A3'),
		('01K1H8N3L0WX4Q6S8YRKT3D2A5', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 'VP Engineering', '01K1H7M2K9VW3P5R7XQJY2C1Z9', NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2A6', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 'Engineering Manager', '01K1H7M2K9VW3P5R7XQJY2C1Z9', '01K1H8N3L0WX4Q6S8YRKT3D2A5'),
		('01K1H8N3L0WX4Q6S8YRKT3D2C0', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 'Legacy Manager', '01K02G6J1CYAN9K8V4PAGSQ5Z8', NULL),
		('01K1H8N3L0WX4Q6S8YRKT3D2C1', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 'Support Specialist', '01K02G6J1CYAN9K8V4PAGSQ5Z8', '01K1H8N3L0WX4Q6S8YRKT3D2C0'),
		('01K1H8N3L0WX4Q6S8YRKT3D2C2', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 'Maintenance Technician', '01K02G6J1CYAN9K8V4PAGSQ5Z8', '01K1H8N3L0WX4Q6S8YRKT3D2C0');
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'ident_user_hierarchy_rel'
	) THEN
		INSERT INTO "ident_user_hierarchy_rel" ("user_id", "hierarchy_id") VALUES
		('01JWNXT3EY7FG47VDJTEPTDC98', '01K1H8N3L0WX4Q6S8YRKT3D2A2'),
		('01JWNXXTF8958VVYAV33MVVMDN', '01K1H8N3L0WX4Q6S8YRKT3D2A3'),
		('01JZQFDH0N51Q3BFQFMFFGSCSV', '01K1H8N3L0WX4Q6S8YRKT3D2A4'),
		('01JZQFF9QEXH71P2CG9Y9MY8MM', '01K1H8N3L0WX4Q6S8YRKT3D2A5'),
		('01JZQFFDKY8T4JB8R6NSY1331J', '01K1H8N3L0WX4Q6S8YRKT3D2A6'),
		('01JZQFGVKZCTV7S310W0BDMWCS', '01K1H8N3L0WX4Q6S8YRKT3D2C0'),
		('01JZQFY6EXRG0959Z95Y2EM3AM', '01K1H8N3L0WX4Q6S8YRKT3D2C1'),
		('01JZQFZFK6GM2D5X6MYHWH6FND', '01K1H8N3L0WX4Q6S8YRKT3D2C2');
	END IF;
END $$;