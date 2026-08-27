-- The integration event outbox (BR 80, acceptance 94.34).
--
-- WHY A TABLE AND NOT A DIRECT PUBLISH. Publishing straight from a domain service makes the broker
-- part of the transaction without any of the guarantees of one, and both failures happen:
-- publish-then-commit announces a sale that then rolls back, commit-then-publish loses the
-- announcement of a sale that really happened. Neither is detectable afterwards. A row written in
-- the SAME transaction as the business change cannot disagree with it - either both landed or
-- neither did - and a background sweep then moves rows to the broker, turning an atomicity problem
-- into a delivery problem, which is the kind retrying actually fixes.
--
-- AN UNPUBLISHED ROW IS THE QUEUE. published_at NULL means waiting; there is deliberately no status
-- column, because a status and a timestamp that could disagree about the same fact is one more thing
-- to reconcile. The sweep's query is "where published_at IS NULL", which is what the index on it
-- serves.
--
-- DELIVERY IS AT-LEAST-ONCE, so event_id is UNIQUE and is what consumers deduplicate on. A sweep
-- that publishes and then fails before marking the row publishes again; the alternative, marking
-- first, loses events instead. Acceptance 94.34 requires consumers to tolerate the duplicate, and
-- event_id is what makes that possible rather than merely asked for.
--
-- NO FOREIGN KEYS AT ALL. An integration event must outlive the record it describes - a consumer
-- replaying the stream needs the cancellation of an order that has since been purged - and a cascade
-- would delete exactly the history somebody is replaying. Same reasoning as sales_order_events.
--
-- occurred_at is the BUSINESS time and published_at is the delivery time; they differ by however
-- long the row waited. Consumers order by the former, because ordering by the latter would deliver a
-- cancellation before the confirmation it cancels.

CREATE TABLE "sales_integration_outbox" ("id" character varying NOT NULL, "org_id" character varying NOT NULL, "event_id" character varying NOT NULL, "aggregate_id" character varying NOT NULL, "event_type" character varying NOT NULL, "schema_version" character varying NOT NULL, "payload" jsonb NOT NULL, "occurred_at" timestamptz NOT NULL, "published_at" timestamptz NULL, "attempt_count" integer NULL, "last_error" character varying NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NULL, "etag" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "sales_outbox_tid_eventid_ukey" UNIQUE ("event_id"));

CREATE INDEX "sales_outbox_tid_aggregate_idx" ON "sales_integration_outbox" ("aggregate_id");

CREATE INDEX "sales_outbox_tid_type_idx" ON "sales_integration_outbox" ("event_type");

CREATE INDEX "sales_outbox_tid_published_idx" ON "sales_integration_outbox" ("published_at");

CREATE INDEX "sales_outbox_tid_occurred_idx" ON "sales_integration_outbox" ("occurred_at");
