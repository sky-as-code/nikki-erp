-- Billing: settlement units, their allocations, and the lineage between bills (BR 33-38, 83).
--
-- A BILL IS NOT A VAT INVOICE. BR 33 forbids conflating them and BR 34 exists precisely because one
-- sale may need several settlement units and one legal document, or the reverse. Nothing in these
-- tables should ever grow an invoice number.
--
-- THE INVARIANT (BR 36): the sum of allocations across every bill of an order equals the order
-- amount EXACTLY. Not approximately - allocations come from dividing an order across bills, and
-- dividing rarely comes out even. The D-04 allocator assigns the whole rounding residual rather than
-- letting it vanish, which is what makes the equality checkable at all.
--
-- The unique on (sales_bill_id, sales_order_line_id) is what stops a split putting the same order
-- line on one bill twice. It is enforced by the database rather than by a check, because a split
-- writes several rows at once and a check between them can be raced.
--
-- NOTHING IS EVER DELETED (BR 83). A bill superseded by a split or a merge is marked cancelled and
-- kept, because the lineage rows point at it and an auditor tracing a payment must arrive somewhere.
-- sales_bill_relations therefore carries NO on-delete cascade: the trail has to outlive its subject.

CREATE TABLE "sales_bills" ("id" character varying NOT NULL, "org_id" character varying NOT NULL, "bill_number" character varying NOT NULL, "sales_order_id" character varying NOT NULL, "status" character varying NOT NULL, "payment_status" character varying NOT NULL, "currency_code" character varying NOT NULL, "subtotal" numeric NOT NULL, "discount_total" numeric NOT NULL, "tax_total" numeric NOT NULL, "total_amount" numeric NOT NULL, "settled_at" timestamptz NULL, "cancelled_at" timestamptz NULL, "is_archived" boolean NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NULL, "etag" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "sales_bills_bill_number_ukey" UNIQUE ("bill_number"), CONSTRAINT "sales_bills_sales_order_id_fkey" FOREIGN KEY ("sales_order_id") REFERENCES "sales_orders" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

CREATE TABLE "sales_bill_relations" ("id" character varying NOT NULL, "org_id" character varying NOT NULL, "source_bill_id" character varying NOT NULL, "target_bill_id" character varying NOT NULL, "relation_type" character varying NOT NULL, "created_at" timestamptz NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "sales_bill_rels_tid_pair_type_ukey" UNIQUE ("source_bill_id", "target_bill_id", "relation_type"), CONSTRAINT "sales_bill_relations_source_bill_id_fkey" FOREIGN KEY ("source_bill_id") REFERENCES "sales_bills" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, CONSTRAINT "sales_bill_relations_target_bill_id_fkey" FOREIGN KEY ("target_bill_id") REFERENCES "sales_bills" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION);

CREATE TABLE "sales_bill_lines" ("id" character varying NOT NULL, "org_id" character varying NOT NULL, "sales_bill_id" character varying NOT NULL, "sales_order_line_id" character varying NOT NULL, "quantity" numeric NOT NULL, "allocated_net_amount" numeric NOT NULL, "allocated_tax_amount" numeric NOT NULL, "allocated_total_amount" numeric NOT NULL, "created_at" timestamptz NOT NULL, "updated_at" timestamptz NULL, "etag" character varying NOT NULL, PRIMARY KEY ("id"), CONSTRAINT "sales_bill_lines_tid_bill_ordline_ukey" UNIQUE ("sales_bill_id", "sales_order_line_id"), CONSTRAINT "sales_bill_lines_sales_bill_id_fkey" FOREIGN KEY ("sales_bill_id") REFERENCES "sales_bills" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "sales_bill_lines_sales_order_line_id_fkey" FOREIGN KEY ("sales_order_line_id") REFERENCES "sales_order_lines" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);

CREATE INDEX "sales_bills_tid_order_idx" ON "sales_bills" ("sales_order_id");

CREATE INDEX "sales_bills_tid_status_idx" ON "sales_bills" ("status", "payment_status");

CREATE INDEX "sales_bill_lines_tid_bill_idx" ON "sales_bill_lines" ("sales_bill_id");

CREATE INDEX "sales_bill_lines_tid_ordline_idx" ON "sales_bill_lines" ("sales_order_line_id");

CREATE INDEX "sales_bill_rels_tid_source_idx" ON "sales_bill_relations" ("source_bill_id");

CREATE INDEX "sales_bill_rels_tid_target_idx" ON "sales_bill_relations" ("target_bill_id");
