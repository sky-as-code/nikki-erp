-- Rebuild functions for the provenance-row cache created in 0002005.
--
-- Three changes against the 0002003 versions:
--   * one row per GRANT PATH (source_kind/source_id), never one per (user, ent);
--   * no LEFT JOIN on org membership - that fan-out multiplied every direct-role
--     row by the user's org count and made org membership look like grant provenance;
--   * expires_at is copied onto the row so an expired grant self-excludes at READ
--     time, instead of surviving until some unrelated event happens to rebuild.

CREATE OR REPLACE FUNCTION iam_calc_user_perm(
	p_user_id varchar DEFAULT NULL,
	p_role_id varchar DEFAULT NULL
)
RETURNS TABLE (
	-- The order must exactly match the INSERT column list below.
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
	group_membership_id       varchar,
	org_unit_id               varchar,
	source_kind               varchar,
	source_id                 varchar,
	expires_at                timestamptz
)
LANGUAGE sql
STABLE
AS $$
WITH role_entitlements AS (

	-- Roles assigned directly to the user
	SELECT
		ra.receiver_user_id AS user_id,
		e.id                AS ent_id,
		e.expression        AS ent_expression,
		a.id                AS action_id,
		res.id              AS resource_id,
		res.code            AS resource_code,
		NULL::varchar       AS role_group_assignment_id,
		ra.id               AS role_user_assignment_id,
		e.scope,
		e.org_id,
		NULL::varchar       AS group_membership_id,
		e.org_unit_id,
		'direct'::varchar   AS source_kind,
		ra.id               AS source_id,
		ra.expires_at
	FROM iam_role_user_assignments ra
		JOIN iam_roles r ON r.id = ra.role_id AND r.is_archived = FALSE
		JOIN iam_entitlements e ON e.role_id = r.id AND e.is_archived = FALSE
		LEFT JOIN iam_actions a ON a.id = e.action_id AND a.resource_id = e.resource_id
		LEFT JOIN iam_resources res ON res.id = a.resource_id
	WHERE (p_user_id IS NULL OR ra.receiver_user_id = p_user_id)
		AND (p_role_id IS NULL OR r.id = p_role_id)
		AND ra.receiver_user_id IS NOT NULL

	UNION ALL

	-- Roles assigned to a group the user belongs to
	SELECT
		gur.user_id     AS user_id,
		e.id            AS ent_id,
		e.expression    AS ent_expression,
		a.id            AS action_id,
		res.id          AS resource_id,
		res.code        AS resource_code,
		ra.id           AS role_group_assignment_id,
		NULL::varchar   AS role_user_assignment_id,
		e.scope,
		e.org_id,
		gur.id          AS group_membership_id,
		e.org_unit_id,
		'group'::varchar AS source_kind,
		ra.id            AS source_id,
		ra.expires_at
	FROM iam_group_user_rel gur
		JOIN iam_role_group_assignments ra ON ra.receiver_group_id = gur.group_id
		JOIN iam_roles r ON r.id = ra.role_id AND r.is_archived = FALSE
		JOIN iam_entitlements e ON e.role_id = r.id AND e.is_archived = FALSE
		LEFT JOIN iam_actions a ON a.id = e.action_id AND a.resource_id = e.resource_id
		LEFT JOIN iam_resources res ON res.id = a.resource_id
	WHERE (p_user_id IS NULL OR gur.user_id = p_user_id)
		AND (p_role_id IS NULL OR r.id = p_role_id)
)
SELECT * FROM role_entitlements re;
$$;


-- Column list shared by every INSERT below. Written out rather than relying on
-- SELECT * so that a future column added to the table cannot silently shift the
-- mapping between the function's result and the table.
CREATE OR REPLACE FUNCTION iam_rebuild_user_perm(p_user_id varchar)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
	DELETE FROM iam_user_permissions WHERE user_id = p_user_id;

	INSERT INTO iam_user_permissions (
		user_id, ent_id, ent_expression, action_id, resource_id, resource_code,
		role_group_assignment_id, role_user_assignment_id, scope, org_id,
		group_membership_id, org_unit_id, source_kind, source_id, expires_at
	)
	SELECT * FROM iam_calc_user_perm(p_user_id, NULL)
	ON CONFLICT DO NOTHING;
END $$;


-- Rebuild every holder of one role, in a single round trip regardless of how many
-- holders there are. This is what makes "add an entitlement to a role" reach the
-- people who already hold it - the event the old design never handled.
CREATE OR REPLACE FUNCTION iam_rebuild_perms_for_role(p_role_id varchar)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
	-- Delete by entitlement, not by user: only the rows this role is responsible
	-- for go away, so a user's grants from other roles are untouched.
	DELETE FROM iam_user_permissions up
	WHERE up.ent_id IN (SELECT e.id FROM iam_entitlements e WHERE e.role_id = p_role_id);

	INSERT INTO iam_user_permissions (
		user_id, ent_id, ent_expression, action_id, resource_id, resource_code,
		role_group_assignment_id, role_user_assignment_id, scope, org_id,
		group_membership_id, org_unit_id, source_kind, source_id, expires_at
	)
	SELECT * FROM iam_calc_user_perm(NULL, p_role_id)
	ON CONFLICT DO NOTHING;
END $$;


-- Offline tooling only (after a migration or a seed change). Batched per user so
-- it never takes the exclusive lock TRUNCATE needs, and never leaves a window in
-- which every non-owner request sees zero permissions.
CREATE OR REPLACE FUNCTION iam_rebuild_all_user_perms()
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
	v_user_id varchar;
BEGIN
	FOR v_user_id IN
		SELECT id FROM iam_users ORDER BY id
	LOOP
		PERFORM iam_rebuild_user_perm(v_user_id);
	END LOOP;
END $$;


-- Housekeeping: long-expired rows are already invisible to every read because of
-- the expiry predicate, so this only reclaims space. Bounded so it can be run on a
-- live system without a long lock.
CREATE OR REPLACE FUNCTION iam_sweep_expired_perms(p_limit integer DEFAULT 10000)
RETURNS integer
LANGUAGE plpgsql
AS $$
DECLARE
	v_deleted integer;
BEGIN
	WITH doomed AS (
		SELECT user_id, ent_id, source_kind, source_id
		FROM iam_user_permissions
		WHERE expires_at IS NOT NULL AND expires_at < NOW()
		LIMIT p_limit
	)
	DELETE FROM iam_user_permissions up
	USING doomed d
	WHERE up.user_id = d.user_id AND up.ent_id = d.ent_id
		AND up.source_kind = d.source_kind AND up.source_id = d.source_id;

	GET DIAGNOSTICS v_deleted = ROW_COUNT;
	RETURN v_deleted;
END $$;


CREATE OR REPLACE FUNCTION iam_delete_user_password_stores()
RETURNS TRIGGER AS $$
BEGIN
	DELETE FROM "iam_password_stores"
	WHERE "principal_type" = 'nikkiuser'
	  AND "principal_id" = OLD."id";
	RETURN OLD;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS iam_users_delete_password_stores ON "iam_users";

-- AFTER DELETE, not BEFORE: the credential rows should go only once the user row has
-- actually been removed, so a delete that rolls back leaves the credentials intact.
CREATE TRIGGER iam_users_delete_password_stores
	AFTER DELETE ON "iam_users"
	FOR EACH ROW
	EXECUTE FUNCTION iam_delete_user_password_stores();
