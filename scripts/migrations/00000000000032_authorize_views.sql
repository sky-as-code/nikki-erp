DO $$
BEGIN
	IF 7 = (
		SELECT COUNT(*) FROM information_schema.tables 
		WHERE table_schema = 'public' AND (
			table_name = 'authz_entitlement_assignments' OR 
			table_name = 'ident_user_group_rel' OR
			table_name = 'authz_entitlements' OR
			table_name = 'authz_resources' OR
			table_name = 'authz_role_user' OR
			table_name = 'authz_role_suite_user' OR
			table_name = 'authz_role_rolesuite'
		)
	) THEN

		-- View for user effective entitlements (direct, roles, suites, and via groups)
		CREATE OR REPLACE VIEW authz_effective_user_entitlements AS
		WITH 
		-- 1. Direct assignments to user
		user_direct_assignments AS (
			SELECT
				assignment.subject_ref AS user_id,
				entitlement.action_expr,
				entitlement.resource_id,
				assignment.resource_name,
				entitlement.org_id,
				resource.scope_type,
				entitlement.action_id,
				assignment.scope_ref,
				assignment.action_name,
				'nikki_user' AS source
			FROM authz_entitlement_assignments assignment
			JOIN authz_entitlements entitlement ON assignment.entitlement_id = entitlement.id
			LEFT JOIN authz_resources resource ON entitlement.resource_id = resource.id
			WHERE assignment.subject_type = 'nikki_user'
		),
		-- 2. User's groups
		user_groups AS (
			SELECT user_id, group_id
			FROM ident_user_group_rel
		),
		-- 3. Assignments to groups the user belongs to
		group_assignments AS (
			SELECT
				user_group.user_id,
				entitlement.action_expr,
				entitlement.resource_id,
				assignment.resource_name,
				entitlement.org_id,
				resource.scope_type,
				entitlement.action_id,
				assignment.scope_ref,
				assignment.action_name,
				'nikki_group' AS source
			FROM authz_entitlement_assignments assignment
			JOIN authz_entitlements entitlement ON assignment.entitlement_id = entitlement.id
			LEFT JOIN authz_resources resource ON entitlement.resource_id = resource.id
			JOIN user_groups user_group ON assignment.subject_type = 'nikki_group' AND assignment.subject_ref = user_group.group_id
		),
		-- 4. User's roles (direct)
		user_roles AS (
			SELECT receiver_ref AS user_id, role_id
			FROM authz_role_user
			WHERE receiver_type = 'user'
		),
		-- 5. Assignments to roles directly assigned to user
		user_role_assignments AS (
			SELECT
				user_role.user_id,
				entitlement.action_expr,
				entitlement.resource_id,
				assignment.resource_name,
				entitlement.org_id,
				resource.scope_type,
				entitlement.action_id,
				assignment.scope_ref,
				assignment.action_name,
				'nikki_role' AS source
			FROM authz_entitlement_assignments assignment
			JOIN authz_entitlements entitlement ON assignment.entitlement_id = entitlement.id
			LEFT JOIN authz_resources resource ON entitlement.resource_id = resource.id
			JOIN user_roles user_role ON assignment.subject_type = 'nikki_role' AND assignment.subject_ref = user_role.role_id
		),
		-- 6. User's suites
		user_suites AS (
			SELECT receiver_ref AS user_id, role_suite_id
			FROM authz_role_suite_user
			WHERE receiver_type = 'user'
		),
		-- 7. Roles in user's suites
		suite_roles AS (
			SELECT us.user_id, rr.role_id
			FROM authz_role_rolesuite rr
			JOIN user_suites us ON rr.role_suite_id = us.role_suite_id
		),
		-- 8. Assignments to roles in user's suites
		user_suite_assignments AS (
			SELECT
				suite_role.user_id,
				entitlement.action_expr,
				entitlement.resource_id,
				assignment.resource_name,
				entitlement.org_id,
				resource.scope_type,
				entitlement.action_id,
				assignment.scope_ref,
				assignment.action_name,
				'nikki_suite' AS source
			FROM authz_entitlement_assignments assignment
			JOIN authz_entitlements entitlement ON assignment.entitlement_id = entitlement.id
			LEFT JOIN authz_resources resource ON entitlement.resource_id = resource.id
			JOIN suite_roles suite_role ON assignment.subject_type = 'nikki_role' AND assignment.subject_ref = suite_role.role_id
		),
		-- 9. Roles assigned to user's groups
		group_roles AS (
			SELECT ug.user_id, ru.role_id
			FROM authz_role_user ru
			JOIN user_groups ug ON ru.receiver_type = 'group' AND ru.receiver_ref = ug.group_id
		),
		-- 10. Assignments to roles assigned to user's groups
		group_role_assignments AS (
			SELECT
				group_role.user_id,
				entitlement.action_expr,
				entitlement.resource_id,
				assignment.resource_name,
				entitlement.org_id,
				resource.scope_type,
				entitlement.action_id,
				assignment.scope_ref,
				assignment.action_name,
				'nikki_group_role' AS source
			FROM authz_entitlement_assignments assignment
			JOIN authz_entitlements entitlement ON assignment.entitlement_id = entitlement.id
			LEFT JOIN authz_resources resource ON entitlement.resource_id = resource.id
			JOIN group_roles group_role ON assignment.subject_type = 'nikki_role' AND assignment.subject_ref = group_role.role_id
		),
		-- 11. Suites assigned to user's groups
		group_suites AS (
			SELECT ug.user_id, rsu.role_suite_id
			FROM authz_role_suite_user rsu
			JOIN user_groups ug ON rsu.receiver_type = 'group' AND rsu.receiver_ref = ug.group_id
		),
		-- 12. Roles in suites assigned to user's groups
		group_suite_roles AS (
			SELECT gs.user_id, rr.role_id
			FROM authz_role_rolesuite rr
			JOIN group_suites gs ON rr.role_suite_id = gs.role_suite_id
		),
		-- 13. Assignments to roles in suites assigned to user's groups
		group_suite_assignments AS (
			SELECT
				group_suite_role.user_id,
				entitlement.action_expr,
				entitlement.resource_id,
				assignment.resource_name,
				entitlement.org_id,
				resource.scope_type,
				entitlement.action_id,
				assignment.scope_ref,
				assignment.action_name,
				'nikki_group_suite' AS source
			FROM authz_entitlement_assignments assignment
			JOIN authz_entitlements entitlement ON assignment.entitlement_id = entitlement.id
			LEFT JOIN authz_resources resource ON entitlement.resource_id = resource.id
			JOIN group_suite_roles group_suite_role ON assignment.subject_type = 'nikki_role' AND assignment.subject_ref = group_suite_role.role_id
		)
		-- Final union of all sources for user
		SELECT * FROM user_direct_assignments
		UNION
		SELECT * FROM group_assignments
		UNION
		SELECT * FROM user_role_assignments
		UNION
		SELECT * FROM user_suite_assignments
		UNION
		SELECT * FROM group_role_assignments
		UNION
		SELECT * FROM group_suite_assignments;

		-- View for group effective entitlements (direct, roles, and suites)
		CREATE OR REPLACE VIEW authz_effective_group_entitlements AS
		WITH 
		-- 1. Direct assignments to group
		group_direct_assignments AS (
			SELECT
				assignment.subject_ref AS group_id,
				entitlement.action_expr,
				entitlement.resource_id,
				assignment.resource_name,
				entitlement.org_id,
				resource.scope_type,
				entitlement.action_id,
				assignment.scope_ref,
				assignment.action_name,
				'nikki_group' AS source
			FROM authz_entitlement_assignments assignment
			JOIN authz_entitlements entitlement ON assignment.entitlement_id = entitlement.id
			LEFT JOIN authz_resources resource ON entitlement.resource_id = resource.id
			WHERE assignment.subject_type = 'nikki_group'
		),
		-- 2. Group's roles (direct)
		group_roles AS (
			SELECT receiver_ref AS group_id, role_id
			FROM authz_role_user
			WHERE receiver_type = 'group'
		),
		-- 3. Assignments to roles directly assigned to group
		group_role_assignments AS (
			SELECT
				group_role.group_id,
				entitlement.action_expr,
				entitlement.resource_id,
				assignment.resource_name,
				entitlement.org_id,
				resource.scope_type,
				entitlement.action_id,
				assignment.scope_ref,
				assignment.action_name,
				'nikki_group_role' AS source
			FROM authz_entitlement_assignments assignment
			JOIN authz_entitlements entitlement ON assignment.entitlement_id = entitlement.id
			LEFT JOIN authz_resources resource ON entitlement.resource_id = resource.id
			JOIN group_roles group_role ON assignment.subject_type = 'nikki_role' AND assignment.subject_ref = group_role.role_id
		),
		-- 4. Group's suites
		group_suites AS (
			SELECT receiver_ref AS group_id, role_suite_id
			FROM authz_role_suite_user
			WHERE receiver_type = 'group'
		),
		-- 5. Roles in group's suites
		group_suite_roles AS (
			SELECT gs.group_id, rr.role_id
			FROM authz_role_rolesuite rr
			JOIN group_suites gs ON rr.role_suite_id = gs.role_suite_id
		),
		-- 6. Assignments to roles in group's suites
		group_suite_assignments AS (
			SELECT
				group_suite_role.group_id,
				entitlement.action_expr,
				entitlement.resource_id,
				assignment.resource_name,
				entitlement.org_id,
				resource.scope_type,
				entitlement.action_id,
				assignment.scope_ref,
				assignment.action_name,
				'nikki_group_suite' AS source
			FROM authz_entitlement_assignments assignment
			JOIN authz_entitlements entitlement ON assignment.entitlement_id = entitlement.id
			LEFT JOIN authz_resources resource ON entitlement.resource_id = resource.id
			JOIN group_suite_roles group_suite_role ON assignment.subject_type = 'nikki_role' AND assignment.subject_ref = group_suite_role.role_id
		)
		-- Final union of all sources for group
		SELECT * FROM group_direct_assignments
		UNION
		SELECT * FROM group_role_assignments
		UNION
		SELECT * FROM group_suite_assignments;

	END IF;
END $$;