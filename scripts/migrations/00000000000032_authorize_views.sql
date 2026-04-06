DO $$
BEGIN
	IF (
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = 'public'
			AND table_name IN (
				'ident_users',
				'ident_groups',
				'ident_group_user_rel',
				'authz_role_assignments',
				'authz_entitlement_role_rel',
				'authz_entitlements',
				'authz_actions',
				'authz_resources',
				'authz_roles'
			)
	) = 9 THEN
		CREATE OR REPLACE VIEW authz_effective_user_entitlements AS
		WITH
		user_role_entitlements AS (
			SELECT
				ra.receiver_user_id AS user_id,
				e.id AS entitlement_id,
				e.scope AS entitlement_scope,
				r.org_id AS entitlement_org_id,
				e.org_unit_id AS entitlement_org_unit_id,
				a.name AS action_code,
				res.name AS resource_code,
				r.id AS role_id,
				ra.id AS role_assignment_id,
				'user_role'::text AS source
			FROM authz_role_assignments ra
			JOIN ident_users u ON u.id = ra.receiver_user_id AND u.is_archived = FALSE
			JOIN authz_roles r ON r.id = ra.role_id AND r.is_archived = FALSE
			JOIN authz_entitlement_role_rel er ON er.role_id = ra.role_id
			JOIN authz_entitlements e ON e.id = er.entitlement_id AND e.is_archived = FALSE
			JOIN authz_actions a ON a.id = e.action_id
			JOIN authz_resources res ON res.id = a.resource_id
			WHERE (ra.expires_at IS NULL OR ra.expires_at > NOW())
		),
		group_role_entitlements AS (
			SELECT
				gur.user_id,
				e.id AS entitlement_id,
				e.scope AS entitlement_scope,
				r.org_id AS entitlement_org_id,
				e.org_unit_id AS entitlement_org_unit_id,
				a.code AS action_code,
				res.code AS resource_code,
				r.id AS role_id,
				ra.id AS role_assignment_id,
				'group_role'::text AS source
			FROM authz_role_assignments ra
			JOIN ident_groups g ON g.id = ra.receiver_group_id AND g.is_archived = FALSE
			JOIN ident_group_user_rel gur ON gur.group_id = ra.receiver_group_id
			JOIN ident_users u ON u.id = gur.user_id AND u.is_archived = FALSE
			JOIN authz_roles r ON r.id = ra.role_id AND r.is_archived = FALSE
			JOIN authz_entitlement_role_rel er ON er.role_id = ra.role_id
			JOIN authz_entitlements e ON e.id = er.entitlement_id AND e.is_archived = FALSE
			JOIN authz_actions a ON a.id = e.action_id
			JOIN authz_resources res ON res.id = a.resource_id
			WHERE (ra.expires_at IS NULL OR ra.expires_at > NOW())
		)
		SELECT DISTINCT ON (user_id, entitlement_id, role_assignment_id)
			user_id,
			entitlement_id,
			entitlement_scope,
			entitlement_org_id,
			entitlement_org_unit_id,
			action_code,
			resource_code,
			role_id,
			role_assignment_id,
			source
		FROM (
			SELECT * FROM user_role_entitlements
			UNION ALL
			SELECT * FROM group_role_entitlements
		) AS combined
		ORDER BY
			user_id,
			entitlement_id,
			role_assignment_id,
			CASE source WHEN 'user_role' THEN 0 ELSE 1 END; -- Force DISTINCT ON to keep user_role when there are duplicated trio: user_id, entitlement_id, role_assignment_id

		CREATE OR REPLACE VIEW authz_effective_group_entitlements AS
		SELECT
			ra.receiver_group_id AS group_id,
			e.id AS entitlement_id,
			e.scope AS entitlement_scope,
			e.org_id AS entitlement_org_id,
			e.org_unit_id AS entitlement_org_unit_id,
			a.code AS action_code,
			res.code AS resource_code,
			r.id AS role_id,
			ra.id AS role_assignment_id,
			'group_role'::text AS source
		FROM authz_role_assignments ra
		JOIN ident_groups g ON g.id = ra.receiver_group_id AND g.is_archived = FALSE
		JOIN authz_roles r ON r.id = ra.role_id AND r.is_archived = FALSE
		JOIN authz_entitlement_role_rel er ON er.role_id = ra.role_id
		JOIN authz_entitlements e ON e.id = er.entitlement_id AND e.is_archived = FALSE
		JOIN authz_actions a ON a.id = e.action_id
		WHERE (ra.expires_at IS NULL OR ra.expires_at > NOW());
	END IF;
END $$;
