DO $$
BEGIN
	IF 2 = (
		SELECT COUNT(*) FROM information_schema.tables 
		WHERE table_schema = 'public' AND (
			table_name = 'authz_entitlements' OR table_name = 'ident_user_group'
		)
	) THEN

		CREATE OR REPLACE VIEW authz_effective_user_entitlements AS
		WITH user_direct_entitlements AS (
			SELECT
				e.subject_ref AS user_id,
				e.action_expr,
				e.resource_id,
				e.scope_ref,
				'user' AS source
			FROM authz_entitlements e
			WHERE e.subject_type = 'user'
		),
		user_roles AS (
			SELECT receiver_ref AS user_id, role_id
			FROM authz_role_user
			WHERE receiver_type = 'user'
		),
		user_role_entitlements AS (
			SELECT
				ur.user_id,
				e.action_expr,
				e.resource_id,
				e.scope_ref,
				'user_role' AS source
			FROM authz_entitlements e
			JOIN user_roles ur ON e.subject_type = 'role' AND e.subject_ref = ur.role_id
		),
		user_suites AS (
			SELECT receiver_ref AS user_id, role_suite_id
			FROM authz_role_suite_user
			WHERE receiver_type = 'user'
		),
		suite_roles AS (
			SELECT us.user_id, rr.role_id
			FROM authz_role_rolesuite rr
			JOIN user_suites us ON rr.role_suite_id = us.role_suite_id
		),
		user_suite_entitlements AS (
			SELECT
				sr.user_id,
				e.action_expr,
				e.resource_id,
				e.scope_ref,
				'user_suite' AS source
			FROM authz_entitlements e
			JOIN suite_roles sr ON e.subject_type = 'role' AND e.subject_ref = sr.role_id
		),
		user_groups AS (
			SELECT user_id, group_id
			FROM ident_user_group
		),
		group_roles AS (
			SELECT ug.user_id, ru.role_id
			FROM authz_role_user ru
			JOIN user_groups ug ON ru.receiver_type = 'group' AND ru.receiver_ref = ug.group_id
		),
		group_role_entitlements AS (
			SELECT
				gr.user_id,
				e.action_expr,
				e.resource_id,
				e.scope_ref,
				'group_role' AS source
			FROM authz_entitlements e
			JOIN group_roles gr ON e.subject_type = 'role' AND e.subject_ref = gr.role_id
		),
		group_suites AS (
			SELECT ug.user_id, rsu.role_suite_id
			FROM authz_role_suite_user rsu
			JOIN user_groups ug ON rsu.receiver_type = 'group' AND rsu.receiver_ref = ug.group_id
		),
		group_suite_roles AS (
			SELECT gs.user_id, rr.role_id
			FROM authz_role_rolesuite rr
			JOIN group_suites gs ON rr.role_suite_id = gs.role_suite_id
		),
		group_suite_entitlements AS (
			SELECT
				gsr.user_id,
				e.action_expr,
				e.resource_id,
				e.scope_ref,
				'group_suite' AS source
			FROM authz_entitlements e
			JOIN group_suite_roles gsr ON e.subject_type = 'role' AND e.subject_ref = gsr.role_id
		)
		SELECT * FROM user_direct_entitlements
		UNION
		SELECT * FROM user_role_entitlements
		UNION
		SELECT * FROM user_suite_entitlements
		UNION
		SELECT * FROM group_role_entitlements
		UNION
		SELECT * FROM group_suite_entitlements;

	END IF;
END $$;
