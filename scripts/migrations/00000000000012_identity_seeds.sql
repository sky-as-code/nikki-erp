DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'ident_users'
	) THEN
		INSERT INTO "ident_users" ("id", "created_at", "display_name", "email", "etag", "failed_login_attempts", "is_owner", "must_change_password", "password_hash", "password_changed_at", "status") VALUES
		('01JWNNJGS70Y07MBEV3AQ0M526', NOW(), 'System', 'system', EXTRACT(EPOCH FROM clock_timestamp()) * 1e9::bigint, 0, NULL, FALSE, '', NOW(), 'active'),
		('01JWNMZ36QHC7CQQ748H9NQ6J6', NOW(), 'Owner', 'owner', EXTRACT(EPOCH FROM clock_timestamp()) * 1e9::bigint, 0, TRUE, TRUE, '', NOW(), 'active'),
		('01JWNXT3EY7FG47VDJTEPTDC98', NOW(), 'Lạc Long Quân', 'dragon@domain.com', EXTRACT(EPOCH FROM clock_timestamp()) * 1e9::bigint, 0, NULL, FALSE, '', NOW(), 'active'),
		('01JWNXXTF8958VVYAV33MVVMDN', NOW(), 'Âu Cơ', 'fairy@domain.com', EXTRACT(EPOCH FROM clock_timestamp()) * 1e9::bigint, 0, NULL, FALSE, '', NOW(), 'active');
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'ident_groups'
	) THEN
		INSERT INTO "ident_groups" ("id", "name", "description", "etag", "created_at") VALUES
		('01JWNXBR5QJBH7PE9PQ9FW746V', 'Domain Users', 'Default group for all domain users', EXTRACT(EPOCH FROM clock_timestamp()) * 1e9::bigint, NOW());
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'ident_user_group'
	) THEN
		INSERT INTO "ident_user_group" ("user_id", "group_id") VALUES
		('01JWNXT3EY7FG47VDJTEPTDC98', '01JWNXBR5QJBH7PE9PQ9FW746V'),
		('01JWNXXTF8958VVYAV33MVVMDN', '01JWNXBR5QJBH7PE9PQ9FW746V');
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'ident_organizations'
	) THEN
		INSERT INTO "ident_organizations" ("id", "created_at", "display_name", "etag", "status", "slug") VALUES
		('01JWNY20G23KD4RV5VWYABQYHD', NOW(), 'My Company', EXTRACT(EPOCH FROM clock_timestamp()) * 1e9::bigint, 'active', 'my-company');
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_name = 'ident_user_org'
	) THEN
		INSERT INTO "ident_user_org" ("user_id", "org_id") VALUES
		('01JWNXT3EY7FG47VDJTEPTDC98', '01JWNY20G23KD4RV5VWYABQYHD'),
		('01JWNXXTF8958VVYAV33MVVMDN', '01JWNY20G23KD4RV5VWYABQYHD');
	END IF;
END $$;