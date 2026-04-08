CREATE OR REPLACE FUNCTION authz_calc_user_perm(p_user_id varchar DEFAULT NULL)
RETURNS TABLE (
	-- The order must exactly match CREATE TABLE "authz_user_permissions"
	user_id                   varchar,
	ent_id                    varchar,
	ent_expression            varchar,
	action_id                 varchar,
	resource_id               varchar,
	resource_code             varchar,
	role_group_assignment_id  varchar,
	role_user_assignment_id   varchar,
	scope                     varchar,
	org_id                    varchar,
	org_membership_id         varchar,
	group_membership_id       varchar,
	org_unit_id               varchar
)
LANGUAGE sql
AS $$
WITH role_entitlements AS (

	-- Direct user roles
	SELECT
		ra.receiver_user_id AS user_id,
		e.id                AS ent_id,
		e.expression        AS ent_expression,
		a.id                AS action_id,
		res.id              AS resource_id,
		res.code            AS resource_code,
		NULL                AS role_group_assignment_id,
		ra.id               AS role_user_assignment_id,
		e.scope,
		e.org_id,
		our.id               AS org_membership_id,
		NULL                 AS group_membership_id,
		e.org_unit_id
	FROM authz_role_user_assignments ra
		JOIN authz_roles r ON r.id = ra.role_id AND r.is_archived = FALSE
		JOIN authz_entitlements e ON e.role_id = r.id AND e.is_archived = FALSE
		LEFT JOIN authz_actions a ON a.id = e.action_id AND a.resource_id = e.resource_id
		LEFT JOIN authz_resources res ON res.id = a.resource_id
		LEFT JOIN ident_org_user_rel our ON our.user_id = ra.receiver_user_id
	WHERE (p_user_id IS NULL OR ra.receiver_user_id = p_user_id)
		AND ra.receiver_user_id IS NOT NULL
		AND (ra.expires_at IS NULL OR ra.expires_at > NOW())

	UNION ALL

	-- Group roles exploded to users
	SELECT
		gur.user_id     AS user_id,
		e.id            AS ent_id,
		e.expression    AS ent_expression,
		a.id            AS action_id,
		res.id          AS resource_id,
		res.code        AS resource_code,
		ra.id           AS role_group_assignment_id,
		NULL            AS role_user_assignment_id,
		e.scope,
		e.org_id,
		NULL            AS org_membership_id,
		gur.id          AS group_membership_id,
		e.org_unit_id
	FROM ident_group_user_rel gur
		JOIN authz_role_group_assignments ra ON ra.receiver_group_id = gur.group_id AND (ra.expires_at IS NULL OR ra.expires_at > NOW())
		JOIN authz_roles r ON r.id = ra.role_id AND r.is_archived = FALSE
		JOIN authz_entitlements e ON e.role_id = r.id AND e.is_archived = FALSE
		LEFT JOIN authz_actions a ON a.id = e.action_id AND a.resource_id = e.resource_id
		LEFT JOIN authz_resources res ON res.id = a.resource_id
	WHERE (p_user_id IS NULL OR gur.user_id = p_user_id)
)
SELECT * FROM role_entitlements re;
$$;


CREATE OR REPLACE FUNCTION authz_rebuild_user_perm(p_user_id varchar)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
	DELETE FROM authz_user_permissions WHERE user_id = p_user_id;

	INSERT INTO authz_user_permissions
		SELECT * FROM authz_calc_user_perm(p_user_id)
		ON CONFLICT DO NOTHING;
END $$;


CREATE OR REPLACE FUNCTION authz_rebuild_all_user_perms()
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
	TRUNCATE TABLE authz_user_permissions;

	INSERT INTO authz_user_permissions
		SELECT * FROM authz_calc_user_perm(NULL)
		ON CONFLICT DO NOTHING;
END $$;

