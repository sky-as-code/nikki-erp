-- IAM resources and actions for the Job Scheduler: the scheduled job, its executions and the
-- attempts within them. Every scheduler permission row lives here, in one file, so a reviewer can
-- see the module's entire permission surface without opening several migrations.
--
-- The dynamic resource engine asserts permissions using the schema name as the resource code, so
-- these codes must stay byte-identical to the "jobscheduler_job" / "jobscheduler_execution" /
-- "jobscheduler_attempt" schema names. A code that drifts from its schema denies every request,
-- with nothing in the response pointing at the seed as the cause.
--
-- Scope is "tenant" for both max and min, unlike most modules which allow an org-level minimum.
-- A scheduled job is infrastructure owned by a module rather than by an organization: the tables
-- carry no tenant_id and no org_id, so an org-scoped grant would have nothing to narrow against
-- and would silently behave as a tenant-wide one. Saying "tenant" is the honest description of
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
		('01M2JBSCH00000000000000001', jsonb_build_object('en-US', 'Job Scheduler - Scheduled Job', 'vi-VN', 'Bộ lập lịch tác vụ - Tác vụ định kỳ'), 'jobscheduler_job', 'Recurring job registered by a module, with its cron schedule and action', 'nikkierp', 'tenant', 'tenant', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2JBSCH00000000000000002', jsonb_build_object('en-US', 'Job Scheduler - Job Execution', 'vi-VN', 'Bộ lập lịch tác vụ - Lượt thực thi tác vụ'), 'jobscheduler_execution', 'One occurrence of a scheduled job, with the configuration it ran under', 'nikkierp', 'tenant', 'tenant', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2JBSCH00000000000000003', jsonb_build_object('en-US', 'Job Scheduler - Job Attempt', 'vi-VN', 'Bộ lập lịch tác vụ - Lần chạy tác vụ'), 'jobscheduler_attempt', 'One actual run of an execution, including its outcome and lease', 'nikkierp', 'tenant', 'tenant', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Scheduled Job. Full CRUD: a job is registered, retuned and retired through the API.
		('01M2JBSCH00000000000000011', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', 'Register a scheduled job', '01M2JBSCH00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2JBSCH00000000000000012', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M2JBSCH00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2JBSCH00000000000000013', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', 'Change the schedule, action or retry settings of a job', '01M2JBSCH00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M2JBSCH00000000000000014', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', 'Remove a job. Its execution history is kept', '01M2JBSCH00000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Job Execution. Read only: executions are materialized by the scheduler, never posted.
		('01M2JBSCH00000000000000021', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M2JBSCH00000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Job Attempt. Read only, for the same reason.
		('01M2JBSCH00000000000000031', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M2JBSCH00000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END
$$;
