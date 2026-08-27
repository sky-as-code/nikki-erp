-- Create "accounting_taxes" table
CREATE TABLE "accounting_taxes" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "tax_kind" character varying NOT NULL,
  "invoice_label" jsonb NULL,
  "description" character varying NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "acc_taxes_tid_code_ukey" UNIQUE ("code", "org_id")
);
-- Create "accounting_tax_jurisdictions" table
CREATE TABLE "accounting_tax_jurisdictions" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "country_code" character varying NOT NULL,
  "level" character varying NOT NULL,
  "parent_id" character varying NULL,
  "authority_name" character varying NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "acc_tx_juris_tid_code_ukey" UNIQUE ("code", "org_id"),
  CONSTRAINT "accounting_tax_jurisdictions_parent_id_fkey" FOREIGN KEY ("parent_id") REFERENCES "accounting_tax_jurisdictions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "acc_tx_juris_tid_parent_idx" to table: "accounting_tax_jurisdictions"
CREATE INDEX "acc_tx_juris_tid_parent_idx" ON "accounting_tax_jurisdictions" ("parent_id");
-- Create "accounting_tax_groups" table
CREATE TABLE "accounting_tax_groups" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "display_name" jsonb NULL,
  "description" character varying NULL,
  "display_sequence" integer NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "acc_tx_grp_tid_code_ukey" UNIQUE ("code", "org_id")
);
-- Create "accounting_tax_definition_versions" table
CREATE TABLE "accounting_tax_definition_versions" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "tax_id" character varying NOT NULL,
  "version_no" integer NOT NULL,
  "usage" character varying NOT NULL,
  "jurisdiction_id" character varying NOT NULL,
  "tax_group_id" character varying NOT NULL,
  "calculation_type" character varying NOT NULL,
  "tax_treatment" character varying NOT NULL,
  "price_inclusion_mode" character varying NOT NULL,
  "sequence" integer NOT NULL,
  "affect_subsequent_base" boolean NOT NULL,
  "base_affected_by_previous" boolean NOT NULL,
  "effective_from" date NOT NULL,
  "effective_to" date NULL,
  "legal_reference" character varying NULL,
  "supersedes_version_id" character varying NULL,
  "lifecycle_status" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "acc_tx_def_tid_tax_ver_ukey" UNIQUE ("tax_id", "version_no", "org_id"),
  CONSTRAINT "accounting_tax_definition_versions_jurisdiction_id_fkey" FOREIGN KEY ("jurisdiction_id") REFERENCES "accounting_tax_jurisdictions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "accounting_tax_definition_versions_supersedes_version_id_fkey" FOREIGN KEY ("supersedes_version_id") REFERENCES "accounting_tax_definition_versions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "accounting_tax_definition_versions_tax_group_id_fkey" FOREIGN KEY ("tax_group_id") REFERENCES "accounting_tax_groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "accounting_tax_definition_versions_tax_id_fkey" FOREIGN KEY ("tax_id") REFERENCES "accounting_taxes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "acc_tx_def_tid_eff_idx" to table: "accounting_tax_definition_versions"
CREATE INDEX "acc_tx_def_tid_eff_idx" ON "accounting_tax_definition_versions" ("tax_id", "effective_from", "effective_to");
-- Create "accounting_tax_components" table
CREATE TABLE "accounting_tax_components" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "parent_tax_definition_version_id" character varying NOT NULL,
  "component_tax_id" character varying NOT NULL,
  "sequence" integer NOT NULL,
  "affect_subsequent_base_override" boolean NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "acc_tx_comp_tid_par_child_ukey" UNIQUE ("parent_tax_definition_version_id", "component_tax_id", "org_id"),
  CONSTRAINT "accounting_tax_components_component_tax_id_fkey" FOREIGN KEY ("component_tax_id") REFERENCES "accounting_taxes" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "accounting_tax_components_parent_tax_definition_version_id_fkey" FOREIGN KEY ("parent_tax_definition_version_id") REFERENCES "accounting_tax_definition_versions" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "accounting_tax_mappings" table
CREATE TABLE "accounting_tax_mappings" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "version_no" integer NOT NULL,
  "priority" integer NOT NULL,
  "effective_from" date NOT NULL,
  "effective_to" date NULL,
  "supersedes_mapping_id" character varying NULL,
  "lifecycle_status" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "acc_tx_map_tid_code_ver_ukey" UNIQUE ("code", "version_no", "org_id"),
  CONSTRAINT "accounting_tax_mappings_supersedes_mapping_id_fkey" FOREIGN KEY ("supersedes_mapping_id") REFERENCES "accounting_tax_mappings" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "accounting_tax_mapping_lines" table
CREATE TABLE "accounting_tax_mapping_lines" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "tax_mapping_id" character varying NOT NULL,
  "source_tax_id" character varying NOT NULL,
  "target_tax_id" character varying NOT NULL,
  "sequence" integer NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "acc_tx_mapln_tid_map_src_ukey" UNIQUE ("tax_mapping_id", "source_tax_id", "org_id"),
  CONSTRAINT "accounting_tax_mapping_lines_source_tax_id_fkey" FOREIGN KEY ("source_tax_id") REFERENCES "accounting_taxes" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "accounting_tax_mapping_lines_target_tax_id_fkey" FOREIGN KEY ("target_tax_id") REFERENCES "accounting_taxes" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "accounting_tax_mapping_lines_tax_mapping_id_fkey" FOREIGN KEY ("tax_mapping_id") REFERENCES "accounting_tax_mappings" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create "accounting_tax_product_classifications" table
CREATE TABLE "accounting_tax_product_classifications" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "jurisdiction_id" character varying NULL,
  "external_code" character varying NULL,
  "description" character varying NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "acc_tx_class_tid_code_ukey" UNIQUE ("code", "org_id"),
  CONSTRAINT "accounting_tax_product_classifications_jurisdiction_id_fkey" FOREIGN KEY ("jurisdiction_id") REFERENCES "accounting_tax_jurisdictions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "accounting_tax_rate_versions" table
CREATE TABLE "accounting_tax_rate_versions" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "tax_id" character varying NOT NULL,
  "version_no" integer NOT NULL,
  "rate" numeric NULL,
  "fixed_amount" numeric NULL,
  "currency_code" character varying NULL,
  "rate_uom_id" character varying NULL,
  "effective_from" date NOT NULL,
  "effective_to" date NULL,
  "legal_reference" character varying NULL,
  "description" character varying NULL,
  "lifecycle_status" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "acc_tx_rate_tid_tax_ver_ukey" UNIQUE ("tax_id", "version_no", "org_id"),
  CONSTRAINT "accounting_tax_rate_versions_tax_id_fkey" FOREIGN KEY ("tax_id") REFERENCES "accounting_taxes" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "acc_tx_rate_tid_eff_idx" to table: "accounting_tax_rate_versions"
CREATE INDEX "acc_tx_rate_tid_eff_idx" ON "accounting_tax_rate_versions" ("tax_id", "effective_from", "effective_to");
-- Create "accounting_tax_rounding_policies" table
CREATE TABLE "accounting_tax_rounding_policies" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "jurisdiction_id" character varying NULL,
  "currency_code" character varying NOT NULL,
  "rounding_scope" character varying NOT NULL,
  "rounding_method" character varying NOT NULL,
  "rounding_increment" numeric NOT NULL,
  "precision" integer NULL,
  "version_no" integer NOT NULL,
  "effective_from" date NOT NULL,
  "effective_to" date NULL,
  "supersedes_policy_id" character varying NULL,
  "lifecycle_status" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "acc_tx_pol_tid_code_ver_ukey" UNIQUE ("code", "version_no", "org_id"),
  CONSTRAINT "accounting_tax_rounding_policies_jurisdiction_id_fkey" FOREIGN KEY ("jurisdiction_id") REFERENCES "accounting_tax_jurisdictions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "accounting_tax_rounding_policies_supersedes_policy_id_fkey" FOREIGN KEY ("supersedes_policy_id") REFERENCES "accounting_tax_rounding_policies" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "acc_tx_pol_tid_lookup_idx" to table: "accounting_tax_rounding_policies"
CREATE INDEX "acc_tx_pol_tid_lookup_idx" ON "accounting_tax_rounding_policies" ("currency_code", "effective_from", "effective_to");
-- Create "accounting_tax_rules" table
CREATE TABLE "accounting_tax_rules" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "code" character varying NOT NULL,
  "name" jsonb NOT NULL,
  "jurisdiction_id" character varying NULL,
  "priority" integer NOT NULL,
  "stop_processing" boolean NOT NULL,
  "effective_from" date NOT NULL,
  "effective_to" date NULL,
  "legal_reference" character varying NULL,
  "version_no" integer NOT NULL,
  "supersedes_rule_id" character varying NULL,
  "lifecycle_status" character varying NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "acc_tx_rule_tid_code_ver_ukey" UNIQUE ("code", "version_no", "org_id"),
  CONSTRAINT "accounting_tax_rules_jurisdiction_id_fkey" FOREIGN KEY ("jurisdiction_id") REFERENCES "accounting_tax_jurisdictions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "accounting_tax_rules_supersedes_rule_id_fkey" FOREIGN KEY ("supersedes_rule_id") REFERENCES "accounting_tax_rules" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "acc_tx_rule_tid_eval_idx" to table: "accounting_tax_rules"
CREATE INDEX "acc_tx_rule_tid_eval_idx" ON "accounting_tax_rules" ("priority", "effective_from", "effective_to");
-- Create "accounting_tax_rule_conditions" table
CREATE TABLE "accounting_tax_rule_conditions" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "tax_rule_id" character varying NOT NULL,
  "field_key" character varying NOT NULL,
  "operator" character varying NOT NULL,
  "value" jsonb NULL,
  "value_currency_code" character varying NULL,
  "sequence" integer NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "accounting_tax_rule_conditions_tax_rule_id_fkey" FOREIGN KEY ("tax_rule_id") REFERENCES "accounting_tax_rules" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "acc_tx_cond_tid_rule_idx" to table: "accounting_tax_rule_conditions"
CREATE INDEX "acc_tx_cond_tid_rule_idx" ON "accounting_tax_rule_conditions" ("tax_rule_id");
-- Create "accounting_tax_rule_results" table
CREATE TABLE "accounting_tax_rule_results" (
  "id" character varying NOT NULL,
  "org_id" character varying NOT NULL,
  "tax_rule_id" character varying NOT NULL,
  "action" character varying NOT NULL,
  "tax_id" character varying NULL,
  "tax_mapping_id" character varying NULL,
  "tax_treatment" character varying NULL,
  "sequence" integer NOT NULL,
  "is_archived" boolean NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NULL,
  "etag" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "accounting_tax_rule_results_tax_id_fkey" FOREIGN KEY ("tax_id") REFERENCES "accounting_taxes" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "accounting_tax_rule_results_tax_mapping_id_fkey" FOREIGN KEY ("tax_mapping_id") REFERENCES "accounting_tax_mappings" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "accounting_tax_rule_results_tax_rule_id_fkey" FOREIGN KEY ("tax_rule_id") REFERENCES "accounting_tax_rules" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "acc_tx_res_tid_rule_idx" to table: "accounting_tax_rule_results"
CREATE INDEX "acc_tx_res_tid_rule_idx" ON "accounting_tax_rule_results" ("tax_rule_id");
