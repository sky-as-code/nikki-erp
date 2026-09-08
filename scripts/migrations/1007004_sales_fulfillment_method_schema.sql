-- Fulfillment methods: the policy catalogue a sale is handed over under.
--
-- A method answers "how do these goods reach the customer, and what happens when that fails" as
-- configuration rather than as code, so an operator can offer an anonymous kiosk sale that refunds
-- itself and an identified one that does not, without either being a deployment.
--
-- Two rules explain the shape of this file. First, a method is snapshotted onto the fulfillment
-- that uses it, so archiving one withdraws it from NEW orders and changes nothing that is already
-- running; that is why is_archived is a plain flag here and not a status, and why nothing cascades.
-- Second, the target of a fulfillment is not part of its method: "this kiosk" is a strategy for
-- picking the first target (initial_target_selection), after which the target lives on the
-- fulfillment and may move to another machine while the method stays exactly as it was.
--
-- default_fulfillment_method_id lands on both sales_channels and sales_points because an order
-- resolves its method in three steps -- what the order names, else the point's default, else the
-- channel's -- and a point must be able to narrow the channel's choice without a new table.
-- Neither column carries a foreign key onto sales_fulfillment_methods for the mapping table's
-- reason to carry one: these are nullable pointers cleared by an operator, and the archive guard in
-- the domain service is what stops a default being left dangling.

-- Create "sales_fulfillment_methods" table
CREATE TABLE "sales_fulfillment_methods" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "fulfillment_type" character varying NOT NULL,
  "initial_target_selection" character varying NOT NULL,
  "max_attempts" integer NULL,
  "failure_action" character varying NOT NULL,
  "allow_target_change" boolean NOT NULL,
  "allow_partial_fulfillment" boolean NOT NULL,
  "requires_authenticated_customer" boolean NOT NULL,
  "reservation_ttl_minutes" integer NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_fulfillment_methods_code_ukey" UNIQUE ("code")
);
-- Create index "sales_fulfil_methods_tid_type_arch_idx" to table: "sales_fulfillment_methods"
CREATE INDEX "sales_fulfil_methods_tid_type_arch_idx" ON "sales_fulfillment_methods" ("fulfillment_type", "is_archived");
-- Create "sales_channel_fulfillment_methods" table
CREATE TABLE "sales_channel_fulfillment_methods" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "sales_channel_id" character varying NOT NULL,
  "fulfillment_method_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "sales_chan_fulfil_methods_channel_fkey" FOREIGN KEY ("sales_channel_id") REFERENCES "sales_channels" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "sales_chan_fulfil_methods_method_fkey" FOREIGN KEY ("fulfillment_method_id") REFERENCES "sales_fulfillment_methods" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "sales_chan_fulfil_methods_tid_ch_me_ukey" to table: "sales_channel_fulfillment_methods"
CREATE UNIQUE INDEX "sales_chan_fulfil_methods_tid_ch_me_ukey" ON "sales_channel_fulfillment_methods" ("sales_channel_id", "fulfillment_method_id");
-- Modify "sales_channels" table
ALTER TABLE "sales_channels" ADD COLUMN IF NOT EXISTS "default_fulfillment_method_id" character varying NULL;
-- Modify "sales_points" table
ALTER TABLE "sales_points" ADD COLUMN IF NOT EXISTS "default_fulfillment_method_id" character varying NULL;
ALTER TABLE "sales_points" ADD COLUMN IF NOT EXISTS "fulfillment_enabled" boolean NOT NULL DEFAULT false;
ALTER TABLE "sales_points" ADD COLUMN IF NOT EXISTS "inventory_location_id" character varying NULL;
