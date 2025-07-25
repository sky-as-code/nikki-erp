DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'core_enums'
	) THEN
		INSERT INTO "core_enums" ("id", "etag", "label", "value", "type") VALUES
			-- BEGIN User status
			('01JZK0R9WF30HMABN7XSW4YNFV', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint,
				'{
					"en_US": "Active",
					"vi_VN": "Đang hoạt động"
				}'::jsonb,
				'active',
				'ident_user_status'),
			('01JZK1458230BQ8C592CABM0RK', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint,
				'{
					"en_US": "Locked",
					"vi_VN": "Tạm khóa"
				}'::jsonb,
				'locked',
				'ident_user_status'),
			('01JZK15TKR71RH6PAB9ZRMKNHQ', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint,
				'{
					"en_US": "Archived",
					"vi_VN": "Ngưng hoạt động"
				}'::jsonb,
				'archived',
				'ident_user_status'),
			-- END User status

			-- BEGIN Organization status
			('01K02G37PCJAHSTC0AWG5JZ3X4', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint,
				'{
					"en_US": "Active",
					"vi_VN": "Đang hoạt động"
				}'::jsonb,
				'active',
				'ident_org_status'),
			('01K02G3CQXBQGD6C83WZ835JR2', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint,
				'{
					"en_US": "Archived",
					"vi_VN": "Ngưng hoạt động"
				}'::jsonb,
				'archived',
				'ident_org_status'),
			-- END Organization status

			-- BEGIN Dummy test status
			('01JZQF16QGN1YA1R6MKA0W7F0H', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint,
				'{
					"en_US": "d. Test status",
					"vi_VN": "đ. Test xì ta tớt"
				}'::jsonb,
				'd_test_status',
				'ident_user_status'),
			('01JZQF3DQ7R57B3TZNV417CZ3M', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint,
				'{
					"en_US": "ow. Test status",
					"vi_VN": "Ơ. Test xì ta tớt"
				}'::jsonb,
				'ow_test_status',
				'ident_user_status'),
			('01JZQEYPYFPCE26P46X437F2DM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint,
				'{
					"en_US": "A. Test status",
					"vi_VN": "Ă. Test xì ta tớt"
				}'::jsonb,
				'a_test_status',
				'ident_user_status'),
			('01JZQF66HR1KEAF528M3RH7A1K', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint,
				'{
					"en_US": "owj. Test status",
					"vi_VN": "Ợ. Test xì ta tớt"
				}'::jsonb,
				'owj_test_status',
				'ident_user_status'),
			('01JZQF2MM3CH282SRVXWHS0V1T', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint,
				'{
					"en_US": "oh. Test status",
					"vi_VN": "Ồ. Test xì ta tớt"
				}'::jsonb,
				'oh_test_status',
				'ident_user_status'),
			('01JZQF4JBFRYV9THX9ZCT2P19J', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint,
				'{
					"en_US": "ows. Test status",
					"vi_VN": "Ớ. Test xì ta tớt"
				}'::jsonb,
				'ows_test_status',
				'ident_user_status');
			-- END Dummy test status
	END IF;
	IF (
		SELECT COUNT(*) FROM information_schema.tables 
		WHERE (table_schema = 'public' AND table_name = 'ident_users')
			OR (table_schema = 'public' AND table_name = 'core_enums')
	) = 2 THEN
		INSERT INTO "ident_users" ("id", "created_at", "display_name", "email", "etag", "failed_login_attempts", "is_owner", "must_change_password", "password_hash", "password_changed_at", "status_id") VALUES
		('01JWNNJGS70Y07MBEV3AQ0M526', NOW(), 'System', 'system', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 0, NULL, FALSE, '', NOW(), 
			(SELECT "id" FROM "core_enums" WHERE "value" = 'active' AND "type" = 'ident_user_status')),
		('01JWNMZ36QHC7CQQ748H9NQ6J6', NOW(), 'Owner', 'owner', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 0, TRUE, TRUE, '', NOW(), 
			(SELECT "id" FROM "core_enums" WHERE "value" = 'active' AND "type" = 'ident_user_status')),
		('01JWNXT3EY7FG47VDJTEPTDC98', NOW(), 'Thần sức mạnh bị xích', 'power@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 0, NULL, FALSE, '', NOW(), 
			(SELECT "id" FROM "core_enums" WHERE "value" = 'locked' AND "type" = 'ident_user_status')),
		('01JWNXXTF8958VVYAV33MVVMDN', NOW(), 'Gấu ngủ đông', 'bear@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 0, NULL, FALSE, '', NOW(), 
			(SELECT "id" FROM "core_enums" WHERE "value" = 'archived' AND "type" = 'ident_user_status')),
		('01JZQFDH0N51Q3BFQFMFFGSCSV', NOW(), 'đ. Test người dùng', 'd@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 0, NULL, FALSE, '', NOW(), 
			(SELECT "id" FROM "core_enums" WHERE "value" = 'd_test_status' AND "type" = 'ident_user_status')),
		('01JZQFF9QEXH71P2CG9Y9MY8MM', NOW(), 'Ơ. Test người dùng', 'ow@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 0, NULL, FALSE, '', NOW(), 
			(SELECT "id" FROM "core_enums" WHERE "value" = 'ow_test_status' AND "type" = 'ident_user_status')),
		('01JZQFFDKY8T4JB8R6NSY1331J', NOW(), 'Ă. Test người dùng', 'a@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 0, NULL, FALSE, '', NOW(), 
			(SELECT "id" FROM "core_enums" WHERE "value" = 'a_test_status' AND "type" = 'ident_user_status')),
		('01JZQFGVKZCTV7S310W0BDMWCS', NOW(), 'Ợ. Test người dùng', 'owj@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 0, NULL, FALSE, '', NOW(), 
			(SELECT "id" FROM "core_enums" WHERE "value" = 'owj_test_status' AND "type" = 'ident_user_status')),
		('01JZQFY6EXRG0959Z95Y2EM3AM', NOW(), 'Ồ. Test người dùng', 'oh@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 0, NULL, FALSE, '', NOW(), 
			(SELECT "id" FROM "core_enums" WHERE "value" = 'oh_test_status' AND "type" = 'ident_user_status')),
		('01JZQFZFK6GM2D5X6MYHWH6FND', NOW(), 'Ớ. Test người dùng', 'ows@nikki.com', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 0, NULL, FALSE, '', NOW(), 
			(SELECT "id" FROM "core_enums" WHERE "value" = 'ows_test_status' AND "type" = 'ident_user_status'));
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

	IF (
		SELECT COUNT(*) FROM information_schema.tables 
		WHERE (table_schema = 'public' AND table_name = 'ident_organizations')
			OR (table_schema = 'public' AND table_name = 'core_enums')
	) = 2 THEN
		INSERT INTO "ident_organizations" ("id", "created_at", "display_name", "etag", "slug", "status_id") VALUES
		('01JWNY20G23KD4RV5VWYABQYHD', NOW(), 'My Company', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 'my-company',
			(SELECT "id" FROM "core_enums" WHERE "value" = 'active' AND "type" = 'ident_org_status')),
		('01K02G6J1CYAN9K8V4PAGSQ5Z8', NOW(), 'Old Company', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint, 'old-company',
			(SELECT "id" FROM "core_enums" WHERE "value" = 'archived' AND "type" = 'ident_org_status'));
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