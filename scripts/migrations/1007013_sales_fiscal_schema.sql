-- Fiscal document requests (BR 33, 46-50, 58, 77, 87.7).
--
-- A BILL IS NOT A VAT INVOICE (BR 33). This table is the join between the two, and the reason it is
-- a table rather than columns on sales_bills: one bill may need an original invoice and then several
-- adjustments as goods come back, so a bill has MANY fiscal requests. Columns on the bill would be
-- rewritten by the first adjustment, losing the original.
--
-- BUSINESS INTENT, NEVER A VENDOR COMMAND (BR 48, 49). `intent` says what commercially happened -
-- a sale was made, some of it came back - and the eInvoice provider decides which legal document
-- that requires (BR 50). There is no column here for a document type, a serial or a template,
-- because that is invoice law and BR 46 and BR 94.26 put it outside Sales.
--
-- PENDING IS NOT ISSUED (BR 77). The far side is a third-party network call, so `status` starts at
-- pending and moves to issued only on confirmed provider success. provider_reference - the only
-- durable link to the document, since the provider owns it and Sales holds no copy - is written at
-- that same moment and never regenerated.
--
-- IDEMPOTENCY IS AGAINST THE PROVIDER, not just this table (D-42). idempotency_key is unique here
-- AND travels to the provider: after a timeout Sales cannot tell whether the document was issued,
-- and only a key the provider recognises stops the retry issuing a second one. A duplicate VAT
-- invoice is a tax filing to correct, not a row to delete.
--
-- attempt_count and last_error exist because PROVIDER FAILURE IS NORMAL OPERATION rather than an
-- exception - a provider is unreachable, rate-limited or slow - and an operator needs to tell a
-- brief outage from a refusal in order to know whether to retry, correct the buyer information or
-- escalate.

CREATE TABLE "sales_fiscal_requests" ("id" character varying NOT NULL, "org_id" character varying NOT NULL, "sales_bill_id" character varying NOT NULL, "intent" character varying NOT NULL, "status" character varying NOT NULL, "idempotency_key" character varying NOT NULL, "provider_reference" character varying NULL, "attempt_count" integer NULL, "last_error" character varying NULL, "buyer_snapshot" jsonb NULL, "original_fiscal_request_id" character varying NULL, "requested_at" timestamptz NULL, "issued_at" timestamptz NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NULL, "etag" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "sales_fiscal_reqs_tid_idemkey_ukey" UNIQUE ("idempotency_key"), CONSTRAINT "sales_fiscal_requests_sales_bill_id_fkey" FOREIGN KEY ("sales_bill_id") REFERENCES "sales_bills" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

CREATE INDEX "sales_fiscal_reqs_tid_bill_idx" ON "sales_fiscal_requests" ("sales_bill_id");

CREATE INDEX "sales_fiscal_reqs_tid_status_idx" ON "sales_fiscal_requests" ("status");

CREATE INDEX "sales_fiscal_reqs_tid_provref_idx" ON "sales_fiscal_requests" ("provider_reference");

CREATE INDEX "sales_fiscal_reqs_tid_original_idx" ON "sales_fiscal_requests" ("original_fiscal_request_id");
