-- Unit of Measure (BR-UOM-ESS).
--
-- Replaces the earlier "essential_units" / "essential_unit_categories" pair, which modelled a
-- base_unit + integer multiplier rather than the reference-UoM + decimal factor the business
-- requirement specifies. Those tables carried development data only and are dropped outright;
-- the new tables are served by the dynamic resource engine as "essential_uom" / "essential_uomcat".

-- Drop the superseded tables. The FK from essential_units is dropped along with them.
DROP TABLE IF EXISTS "essential_units" CASCADE;
DROP TABLE IF EXISTS "essential_unit_categories" CASCADE;

-- Create "essential_uomcats" table
CREATE TABLE "essential_uomcats" (
  "id" character varying NOT NULL,
  "name" jsonb NOT NULL,
  -- The single Reference UoM of the category (BR-UOM-ESS-003). Nullable because a category is
  -- created before the UoM that becomes its reference exists. The FK is added below, once
  -- essential_uoms exists.
  "reference_uom_id" character varying NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL DEFAULT false,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_uomcats_name_org_id_ukey" UNIQUE ("name", "org_id")
);

-- Create "essential_uoms" table
CREATE TABLE "essential_uoms" (
  "id" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "symbol" character varying NOT NULL,
  -- NOT NULL: a UoM belongs to exactly one category and cannot exist outside one
  -- (BR-UOM-ESS-002). ON DELETE NO ACTION therefore blocks deleting a category that still
  -- has UoMs, rather than orphaning them.
  "category_id" character varying NOT NULL,
  -- 'reference' | 'bigger_equal' | 'smaller' (BR-UOM-ESS-009).
  "uom_type" character varying NOT NULL,
  -- Conversion factor relative to the category's Reference UoM (BR-UOM-ESS-008). numeric, not
  -- a float: BR-UOM-ESS-018 requires values such as 0.453592 to survive without drift.
  "factor" numeric(38, 12) NOT NULL,
  -- Rounding precision, 0 <= rounding < 1 (BR-UOM-ESS-017). Independent of factor.
  "rounding" numeric(38, 12) NOT NULL,
  "org_id" character varying NOT NULL,
  "is_archived" boolean NOT NULL DEFAULT false,
  "etag" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "essential_uoms_symbol_org_id_ukey" UNIQUE ("symbol", "org_id"),
  CONSTRAINT "essential_uoms_category_id_fkey" FOREIGN KEY ("category_id")
    REFERENCES "essential_uomcats" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);

-- The category -> reference UoM link closes the cycle, so it is added after both tables exist.
ALTER TABLE "essential_uomcats"
  ADD CONSTRAINT "essential_uomcats_reference_uom_id_fkey" FOREIGN KEY ("reference_uom_id")
    REFERENCES "essential_uoms" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- One Reference UoM per category (BR-UOM-ESS-005, UOM-ESS-INV-09). The application also
-- enforces this, but only the index makes it a true invariant under concurrent writes.
CREATE UNIQUE INDEX "essential_uoms_one_reference_per_category"
  ON "essential_uoms" ("category_id") WHERE "uom_type" = 'reference';

-- Listings and the category detail page both filter UoMs by their category.
CREATE INDEX "essential_uoms_category_id_idx" ON "essential_uoms" ("category_id");
