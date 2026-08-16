-- Stock's settings for a product line: which unit its balances are counted in.
--
-- The unit is not a column on the product template, deliberately. What a balance means is Stock's
-- to decide, and putting it on the template would make Product the owner of a stock concept.
-- No quantity is stored here either: on-hand, reserved and available stay on the quant.
--
-- The inventory_uom_id is a plain reference to Essential's essential_uoms, with no foreign key.
-- A constraint across a module boundary would stop the two being deployed or migrated apart,
-- which is the same reason base_uom_id on inventory_stock_quants carries none.

-- Create "inventory_stock_product_configs" table
CREATE TABLE "inventory_stock_product_configs" (
  "id" character varying NOT NULL,
  "product_template_id" character varying NOT NULL,
  "inventory_uom_id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  -- One row per template. A second would give one product two inventory units and no way to say
  -- which its recorded balances are in.
  CONSTRAINT "invty_stk_prod_cfgs_tid_ptpl_id_ukey" UNIQUE ("product_template_id", "org_id"),
  CONSTRAINT "inventory_stock_product_configs_product_template_id_fkey" FOREIGN KEY ("product_template_id") REFERENCES "inventory_product_templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "invty_stk_prod_cfgs_inventory_uom_id_idx" to table: "inventory_stock_product_configs"
CREATE INDEX "invty_stk_prod_cfgs_inventory_uom_id_idx" ON "inventory_stock_product_configs" ("inventory_uom_id");
