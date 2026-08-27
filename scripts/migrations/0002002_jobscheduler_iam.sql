-- IAM resources and actions for the Job Scheduler: the scheduled job, its executions and the
-- attempts within them. Every scheduler permission row lives here, in one file, so a reviewer can
-- see the module's entire permission surface without opening several migrations.
--
-- The dynamic resource engine asserts permissions using the schema name as the resource code, so
-- these codes must stay byte-identical to the "jobscheduler_job" / "jobscheduler_execution" /
-- "jobscheduler_attempt" schema names. A code that drifts from its schema denies every request,
-- with nothing in the response pointing at the seed as the cause.
--
-- Scope is "domain" for both max and min, unlike most modules which allow an org-level minimum.
-- A scheduled job is infrastructure owned by a module rather than by an organization: the tables
-- carry no tenant_id and no org_id, so an org-scoped grant would have nothing to narrow against
-- and would silently behave as a domain-wide one. Saying "domain" is the honest description of
-- what a grant here actually confers.
--
-- Deliberate omissions, each of which would otherwise look like something forgotten:
--
--   * jobscheduler_execution and jobscheduler_attempt get read ONLY. Both are written by the
--     scheduler's own workers inside the claim transaction, never through the API. Seeding create,
--     update or delete would advertise powers no endpoint offers, and a role granted them would
--     appear able to forge execution history.
--   * There is no set_archived on any of the three. None is archivable: a job's lifecycle is
--     is_enabled, and an execution's is its status. An archived-but-enabled job would be a
--     schedule withdrawn from view while still firing every minute.
--   * There is no "enable" or "disable" action. Both are an update of is_enabled, and the power
--     to edit a job's cron expression already implies the power to stop it firing; splitting them
--     would suggest a role could pause a job it cannot otherwise change.
--   * There is no "run now" action. The scheduler has no such endpoint in this scope, and seeding
--     the permission ahead of the capability would leave a grant that does nothing.
--
-- Deliberately NO iam_entitlements rows. Some modules grant the system "User" role a domain-wide
-- read so any user can reference their data while filling an unrelated form. Scheduler data is not
-- comparable: action_config carries the URL and headers a job calls with, which for a rest_api job
-- is effectively a description of an internal endpoint and how to reach it. Access follows
-- explicitly assigned roles.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M2JBSCH00000000000000001', 'Scheduled Job', 'jobscheduler_job', 'Recurring job registered by a module, with its cron schedule and action', 'nikkierp', 'domain', 'domain', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2JBSCH00000000000000002', 'Job Execution', 'jobscheduler_execution', 'One occurrence of a scheduled job, with the configuration it ran under', 'nikkierp', 'domain', 'domain', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2JBSCH00000000000000003', 'Job Attempt', 'jobscheduler_attempt', 'One actual run of an execution, including its outcome and lease', 'nikkierp', 'domain', 'domain', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Scheduled Job. Full CRUD: a job is registered, retuned and retired through the API.
		('01M2JBSCH00000000000000011', 'Create', 'create', 'Register a scheduled job', '01M2JBSCH00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2JBSCH00000000000000012', 'Read', 'read', NULL, '01M2JBSCH00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2JBSCH00000000000000013', 'Update', 'update', 'Change the schedule, action or retry settings of a job', '01M2JBSCH00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2JBSCH00000000000000014', 'Delete', 'delete', 'Remove a job. Its execution history is kept', '01M2JBSCH00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Job Execution. Read only: executions are materialized by the scheduler, never posted.
		('01M2JBSCH00000000000000021', 'Read', 'read', NULL, '01M2JBSCH00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Job Attempt. Read only, for the same reason.
		('01M2JBSCH00000000000000031', 'Read', 'read', NULL, '01M2JBSCH00000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END
$$;
