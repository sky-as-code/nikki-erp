DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'ident_users'
	) THEN
		INSERT INTO "ident_users" ("id", "created_at", "display_name", "email", "etag", "is_owner", "status") VALUES
		('01JWNNJGS70Y07MBEV3AQ0M526', NOW(), 'System', 'system', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NULL, 'active'),
		('01JWNMZ36QHC7CQQ748H9NQ6J6', NOW(), 'Owner', 'owner', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, TRUE, 'active'),
		('01JWNXT3EY7FG47VDJTEPTDC98', NOW(), 'Thần sức mạnh bị xích', 'power@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NULL, 'locked'),
		('01JWNXXTF8958VVYAV33MVVMDN', NOW(), 'Gấu ngủ đông', 'bear@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NULL, 'archived'),
		('01JZQFDH0N51Q3BFQFMFFGSCSV', NOW(), 'đ. Test người dùng', 'd@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NULL, 'active'),
		('01JZQFF9QEXH71P2CG9Y9MY8MM', NOW(), 'Ơ. Test người dùng', 'ow@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NULL, 'archived'),
		('01JZQFFDKY8T4JB8R6NSY1331J', NOW(), 'Ă. Test người dùng', 'a@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NULL, 'active'),
		('01JZQFGVKZCTV7S310W0BDMWCS', NOW(), 'Ợ. Test người dùng', 'owj@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NULL, 'active'),
		('01JZQFY6EXRG0959Z95Y2EM3AM', NOW(), 'Ồ. Test người dùng', 'oh@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NULL, 'locked'),
		('01JZQFZFK6GM2D5X6MYHWH6FND', NOW(), 'Ớ. Test người dùng', 'ows@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NULL, 'active');
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'ident_groups'
	) THEN
		INSERT INTO "ident_groups" ("id", "name", "description", "etag", "created_at") VALUES
		('01JWNXBR5QJBH7PE9PQ9FW746V', 'Domain Users', 'Default group for all domain users', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, NOW());
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'ident_user_group_rel'
	) THEN
		INSERT INTO "ident_user_group_rel" ("user_id", "group_id") VALUES
		('01JWNXT3EY7FG47VDJTEPTDC98', '01JWNXBR5QJBH7PE9PQ9FW746V'),
		('01JWNXXTF8958VVYAV33MVVMDN', '01JWNXBR5QJBH7PE9PQ9FW746V');
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'ident_organizations'
	) THEN
		INSERT INTO "ident_organizations" ("id", "created_at", "display_name", "etag", "slug", "status") VALUES
		('01JWNY20G23KD4RV5VWYABQYHD', NOW(), 'My Company', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 'my-company', 'active'),
		('01K02G6J1CYAN9K8V4PAGSQ5Z8', NOW(), 'Old Company', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 'old-company', 'archived');
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'ident_user_org_rel'
	) THEN
		INSERT INTO "ident_user_org_rel" ("user_id", "org_id") VALUES
		('01JWNXT3EY7FG47VDJTEPTDC98', '01JWNY20G23KD4RV5VWYABQYHD'),
		('01JWNXXTF8958VVYAV33MVVMDN', '01JWNY20G23KD4RV5VWYABQYHD');
	END IF;
END $$;