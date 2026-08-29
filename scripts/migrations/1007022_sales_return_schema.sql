-- Returns, return lines and refund legs.
--
-- A return is a document beside the order, never an edit of it: editing the order would restate a
-- transaction already paid for and reported on a VAT invoice.
--
-- The three consequences of a return -- goods back, money back, tax authority told -- fail
-- independently, so each carries its own status column. A return is commercially complete once the
-- first two are done; the fiscal status never gates the others.
--
-- No CHECK constraints anywhere in this framework: every invariant is enforced in the domain
-- service.

CREATE TABLE IF NOT EXISTS "sales_returns" (
	"id" character varying NOT NULL,
	"org_id" character varying NOT NULL,
	"return_number" character varying NOT NULL,
	"sales_order_id" character varying NOT NULL,
	"status" character varying NOT NULL DEFAULT 'draft',
	"inventory_return_status" character varying NOT NULL DEFAULT 'pending',
	"refund_status" character varying NOT NULL DEFAULT 'pending',
	"fiscal_adjustment_status" character varying NOT NULL DEFAULT 'pending',
	"reason" character varying NOT NULL,
	"inventory_disposition" character varying NULL,
	"refund_total" numeric NOT NULL DEFAULT 0,
	"inventory_reference" character varying NULL,
	"failure_reason" character varying NULL,
	"requested_at" timestamp with time zone NULL,
	"completed_at" timestamp with time zone NULL,
	"cancelled_at" timestamp with time zone NULL,
	"is_archived" boolean NOT NULL DEFAULT false,
	"created_at" timestamp with time zone NOT NULL DEFAULT NOW(),
	"updated_at" timestamp with time zone NULL,
	"etag" character varying NOT NULL,
	PRIMARY KEY ("id"),
	CONSTRAINT "sales_returns_return_number_key" UNIQUE ("return_number")
);

CREATE INDEX IF NOT EXISTS "sales_returns_tid_order" ON "sales_returns" ("sales_order_id");
CREATE INDEX IF NOT EXISTS "sales_returns_tid_status" ON "sales_returns" ("status");
CREATE INDEX IF NOT EXISTS "sales_returns_tid_fiscal" ON "sales_returns" ("fiscal_adjustment_status");

-- One order line appears at most once per return, so returning the same line twice means two
-- returns, each checked against what is still returnable.
CREATE TABLE IF NOT EXISTS "sales_return_lines" (
	"id" character varying NOT NULL,
	"org_id" character varying NOT NULL,
	"sales_return_id" character varying NOT NULL,
	"sales_order_line_id" character varying NOT NULL,
	"quantity" numeric NOT NULL,
	"refund_amount" numeric NOT NULL DEFAULT 0,
	"refund_tax_amount" numeric NOT NULL DEFAULT 0,
	"requires_inventory_return" boolean NOT NULL DEFAULT true,
	"is_archived" boolean NOT NULL DEFAULT false,
	"created_at" timestamp with time zone NOT NULL DEFAULT NOW(),
	"updated_at" timestamp with time zone NULL,
	"etag" character varying NOT NULL,
	PRIMARY KEY ("id"),
	CONSTRAINT "sales_return_lines_tid_uniq" UNIQUE ("sales_return_id", "sales_order_line_id")
);

CREATE INDEX IF NOT EXISTS "sales_return_lines_tid_return" ON "sales_return_lines" ("sales_return_id");
CREATE INDEX IF NOT EXISTS "sales_return_lines_tid_ordline" ON "sales_return_lines" ("sales_order_line_id");

-- Every refund leg names the payment it gives back: money returns by the route it arrived, and a
-- leg may not exceed its original payment.
CREATE TABLE IF NOT EXISTS "sales_refund_payments" (
	"id" character varying NOT NULL,
	"org_id" character varying NOT NULL,
	"sales_return_id" character varying NOT NULL,
	"original_sales_payment_id" character varying NOT NULL,
	"amount" numeric NOT NULL,
	"currency_code" character varying NOT NULL,
	"status" character varying NOT NULL DEFAULT 'pending',
	"provider_reference" character varying NULL,
	"failure_reason" character varying NULL,
	"completed_at" timestamp with time zone NULL,
	"is_archived" boolean NOT NULL DEFAULT false,
	"created_at" timestamp with time zone NOT NULL DEFAULT NOW(),
	"updated_at" timestamp with time zone NULL,
	"etag" character varying NOT NULL,
	PRIMARY KEY ("id")
);

CREATE INDEX IF NOT EXISTS "sales_refund_pays_tid_return" ON "sales_refund_payments" ("sales_return_id");
CREATE INDEX IF NOT EXISTS "sales_refund_pays_tid_orig" ON "sales_refund_payments" ("original_sales_payment_id");
CREATE INDEX IF NOT EXISTS "sales_refund_pays_tid_status" ON "sales_refund_payments" ("status");
