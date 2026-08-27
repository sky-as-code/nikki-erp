-- Create "jobscheduler_jobs" table
CREATE TABLE "jobscheduler_jobs" (
  "id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "job_type" character varying NOT NULL,
  "module_name" character varying NOT NULL,
  "job_key" character varying NOT NULL,
  "action_type" character varying NOT NULL,
  "action_config" jsonb NOT NULL,
  "cron_expression" character varying NOT NULL,
  "effective_from" timestamptz NULL,
  "effective_until" timestamptz NULL,
  "is_enabled" boolean NULL,
  "max_attempts" integer NULL,
  "retry_interval_seconds" integer NULL,
  "concurrency_policy" character varying NULL,
  "misfire_policy" character varying NULL,
  "next_run_at" timestamptz NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "jobsched_jobs_module_job_key_ukey" UNIQUE ("module_name", "job_key")
);
-- Create index "jobsched_jobs_enabled_next_run_idx" to table: "jobscheduler_jobs"
CREATE INDEX "jobsched_jobs_enabled_next_run_idx" ON "jobscheduler_jobs" ("is_enabled", "next_run_at");
-- Create index "jobsched_jobs_module_name_idx" to table: "jobscheduler_jobs"
CREATE INDEX "jobsched_jobs_module_name_idx" ON "jobscheduler_jobs" ("module_name");
-- Create "jobscheduler_executions" table
CREATE TABLE "jobscheduler_executions" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "job_id" character varying NULL,
  "execution_key" character varying NOT NULL,
  "scheduled_for" timestamptz NOT NULL,
  "next_occurrence_at" timestamptz NULL,
  "status" character varying NOT NULL,
  "available_at" timestamptz NOT NULL,
  "started_at" timestamptz NULL,
  "finished_at" timestamptz NULL,
  "attempt_count" integer NULL,
  "job_snapshot" jsonb NOT NULL,
  "failure_code" character varying NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "jobscheduler_executions_execution_key_ukey" UNIQUE ("execution_key"),
  CONSTRAINT "jobscheduler_executions_job_id_fkey" FOREIGN KEY ("job_id") REFERENCES "jobscheduler_jobs" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "jobsched_execs_job_scheduled_idx" to table: "jobscheduler_executions"
CREATE INDEX "jobsched_execs_job_scheduled_idx" ON "jobscheduler_executions" ("job_id", "scheduled_for");
-- Create index "jobsched_execs_status_available_idx" to table: "jobscheduler_executions"
CREATE INDEX "jobsched_execs_status_available_idx" ON "jobscheduler_executions" ("status", "available_at");
-- Create "jobscheduler_attempts" table
CREATE TABLE "jobscheduler_attempts" (
  "id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "execution_id" character varying NOT NULL,
  "attempt_number" integer NOT NULL,
  "status" character varying NOT NULL,
  "instance_id" character varying NULL,
  "started_at" timestamptz NULL,
  "finished_at" timestamptz NULL,
  "duration_ms" bigint NULL,
  "next_retry_at" timestamptz NULL,
  "lease_expires_at" timestamptz NULL,
  "error_code" character varying NULL,
  "error_message" character varying NULL,
  "http_status_code" integer NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "jobsched_attempts_exec_number_ukey" UNIQUE ("execution_id", "attempt_number"),
  CONSTRAINT "jobscheduler_attempts_execution_id_fkey" FOREIGN KEY ("execution_id") REFERENCES "jobscheduler_executions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "jobsched_attempts_status_lease_idx" to table: "jobscheduler_attempts"
CREATE INDEX "jobsched_attempts_status_lease_idx" ON "jobscheduler_attempts" ("status", "lease_expires_at");
