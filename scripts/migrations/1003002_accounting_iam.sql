-- IAM resources and actions for the Accounting module's Tax subsystem.
--
-- Every Accounting permission row lives here, in one file, so that a reviewer can see the module's
-- entire permission surface without opening several migrations.
--
-- The dynamic resource engine asserts permissions using the schema name as the resource code, so
-- these codes must stay byte-identical to the "accounting_tax*" schema names in
-- modules/accounting/domain/models. A code that drifts from its schema denies every request, with
-- nothing in the response pointing at the seed as the cause. That is also why the requirement's
-- shorthand ("accounting_taxgroup") is spelled here as the actual schema name
-- ("accounting_tax_group") instead.
--
-- Publish and withdraw are seeded as SEPARATE actions rather than folded into "update", because
-- they are materially different powers. Updating edits a draft that nothing has calculated with;
-- publishing makes a configuration binding on every subsequent transaction and freezes its material
-- fields forever (BR-TAX-ESS-SUP-002). A role that may fix a typo in a draft rate should not
-- thereby be able to put that rate into effect for the whole organization.
--
-- Only the five versioned resources carry them. A tax group or a jurisdiction has no lifecycle, so
-- seeding a publish action for one would grant a power that nothing checks.
--
-- Override and simulate hang off accounting_tax, per BR-TAX-ESS-053. Override is deliberately not a
-- kind of update: it changes the tax applied to a live transaction rather than the configuration,
-- and the requirement mandates both a distinct entitlement and a written reason.
--
-- Deliberately NO iam_entitlements rows. Unit of Measure grants the system "User" role a domain-wide
-- read so that any user can pick a unit while filling in an unrelated form. Tax configuration is not
-- comparable: rates, rules and mappings are commercially sensitive and legally consequential, so
-- access follows explicitly assigned roles. That is the same choice Sales, Products, Contacts and
-- Purchase made.

DO $$
BEGIN

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M4ACCT000000000000000001', 'Tax Jurisdiction', 'accounting_tax_jurisdiction', 'A territory in which a taxing authority levies tax', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000002', 'Tax Group', 'accounting_tax_group', 'A grouping of taxes for display and reporting; never a calculation formula', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000003', 'Tax Rounding Policy', 'accounting_tax_rounding_policy', 'How a computed tax amount is rounded, and whether per line or per document', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000004', 'Tax Product Classification', 'accounting_tax_product_classification', 'What a product is for tax purposes, which rules test rather than the product itself', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000005', 'Tax', 'accounting_tax', 'The stable business identity of a tax, as quoted on an invoice and in law', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000006', 'Tax Definition Version', 'accounting_tax_definition_version', 'Everything about a tax that affects determination or calculation, for one effective period', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000007', 'Tax Rate Version', 'accounting_tax_rate_version', 'The rate or fixed amount of a tax over one effective period', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000008', 'Tax Component', 'accounting_tax_component', 'One child tax inside a group tax, versioned with the definition that composes it', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000009', 'Tax Mapping', 'accounting_tax_mapping', 'A context-specific substitution of one tax for another', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000010', 'Tax Mapping Line', 'accounting_tax_mapping_line', 'One source-tax to target-tax substitution within a mapping', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000011', 'Tax Rule', 'accounting_tax_rule', 'A rule deciding which taxes apply to a transaction context', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000012', 'Tax Rule Condition', 'accounting_tax_rule_condition', 'One typed predicate on the tax context', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000013', 'Tax Rule Result', 'accounting_tax_rule_result', 'What a matching rule does to the candidate tax set', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Tax Jurisdiction
		('01M4ACCT000000000000000014', 'Create', 'create', NULL, '01M4ACCT000000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000015', 'Read', 'read', NULL, '01M4ACCT000000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000016', 'Update', 'update', NULL, '01M4ACCT000000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000017', 'Delete', 'delete', NULL, '01M4ACCT000000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000018', 'Set archived status', 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Group
		('01M4ACCT000000000000000019', 'Create', 'create', NULL, '01M4ACCT000000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000020', 'Read', 'read', NULL, '01M4ACCT000000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000021', 'Update', 'update', NULL, '01M4ACCT000000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000022', 'Delete', 'delete', NULL, '01M4ACCT000000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000023', 'Set archived status', 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Rounding Policy
		('01M4ACCT000000000000000024', 'Create', 'create', NULL, '01M4ACCT000000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000025', 'Read', 'read', NULL, '01M4ACCT000000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000026', 'Update', 'update', NULL, '01M4ACCT000000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000027', 'Delete', 'delete', NULL, '01M4ACCT000000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000028', 'Set archived status', 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000029', 'Publish', 'publish', 'Put a rounding policy into effect, freezing the fields that decide how tax is rounded', '01M4ACCT000000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000030', 'Withdraw', 'withdraw', 'Retire a rounding policy from new determination, leaving it readable for audit', '01M4ACCT000000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Product Classification
		('01M4ACCT000000000000000031', 'Create', 'create', NULL, '01M4ACCT000000000000000004', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000032', 'Read', 'read', NULL, '01M4ACCT000000000000000004', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000033', 'Update', 'update', NULL, '01M4ACCT000000000000000004', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000034', 'Delete', 'delete', NULL, '01M4ACCT000000000000000004', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000035', 'Set archived status', 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000004', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax
		('01M4ACCT000000000000000036', 'Create', 'create', NULL, '01M4ACCT000000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000037', 'Read', 'read', NULL, '01M4ACCT000000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000038', 'Update', 'update', NULL, '01M4ACCT000000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000039', 'Delete', 'delete', NULL, '01M4ACCT000000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000040', 'Set archived status', 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000041', 'Override tax', 'override', 'Substitute the determined tax set on a transaction; requires a reason and is recorded in the snapshot', '01M4ACCT000000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000042', 'Simulate tax', 'simulate', 'Run the tax engine against hypothetical inputs without creating any transaction', '01M4ACCT000000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Definition Version
		('01M4ACCT000000000000000043', 'Create', 'create', NULL, '01M4ACCT000000000000000006', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000044', 'Read', 'read', NULL, '01M4ACCT000000000000000006', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000045', 'Update', 'update', NULL, '01M4ACCT000000000000000006', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000046', 'Delete', 'delete', NULL, '01M4ACCT000000000000000006', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000047', 'Set archived status', 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000006', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000048', 'Publish', 'publish', 'Put a tax definition into effect, freezing every field that decides an amount', '01M4ACCT000000000000000006', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000049', 'Withdraw', 'withdraw', 'Retire a tax definition from new determination, leaving it readable for audit', '01M4ACCT000000000000000006', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Rate Version
		('01M4ACCT000000000000000050', 'Create', 'create', NULL, '01M4ACCT000000000000000007', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000051', 'Read', 'read', NULL, '01M4ACCT000000000000000007', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000052', 'Update', 'update', NULL, '01M4ACCT000000000000000007', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000053', 'Delete', 'delete', NULL, '01M4ACCT000000000000000007', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000054', 'Set archived status', 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000007', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000055', 'Publish', 'publish', 'Put a rate into effect; two published rates of one tax may never overlap in time', '01M4ACCT000000000000000007', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000056', 'Withdraw', 'withdraw', 'Retire a rate from new determination, leaving it readable for audit', '01M4ACCT000000000000000007', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Component
		('01M4ACCT000000000000000057', 'Create', 'create', NULL, '01M4ACCT000000000000000008', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000058', 'Read', 'read', NULL, '01M4ACCT000000000000000008', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000059', 'Update', 'update', NULL, '01M4ACCT000000000000000008', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000060', 'Delete', 'delete', NULL, '01M4ACCT000000000000000008', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000061', 'Set archived status', 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000008', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Mapping
		('01M4ACCT000000000000000062', 'Create', 'create', NULL, '01M4ACCT000000000000000009', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000063', 'Read', 'read', NULL, '01M4ACCT000000000000000009', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000064', 'Update', 'update', NULL, '01M4ACCT000000000000000009', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000065', 'Delete', 'delete', NULL, '01M4ACCT000000000000000009', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000066', 'Set archived status', 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000009', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000067', 'Publish', 'publish', 'Put a mapping into effect, freezing it and its lines', '01M4ACCT000000000000000009', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000068', 'Withdraw', 'withdraw', 'Retire a mapping from new determination, leaving it readable for audit', '01M4ACCT000000000000000009', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Mapping Line
		('01M4ACCT000000000000000069', 'Create', 'create', NULL, '01M4ACCT000000000000000010', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000070', 'Read', 'read', NULL, '01M4ACCT000000000000000010', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000071', 'Update', 'update', NULL, '01M4ACCT000000000000000010', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000072', 'Delete', 'delete', NULL, '01M4ACCT000000000000000010', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000073', 'Set archived status', 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000010', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Rule
		('01M4ACCT000000000000000074', 'Create', 'create', NULL, '01M4ACCT000000000000000011', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000075', 'Read', 'read', NULL, '01M4ACCT000000000000000011', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000076', 'Update', 'update', NULL, '01M4ACCT000000000000000011', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000077', 'Delete', 'delete', NULL, '01M4ACCT000000000000000011', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000078', 'Set archived status', 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000011', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000079', 'Publish', 'publish', 'Put a rule into effect, freezing its conditions, results and priority', '01M4ACCT000000000000000011', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000080', 'Withdraw', 'withdraw', 'Retire a rule from new determination, leaving it readable for audit', '01M4ACCT000000000000000011', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Rule Condition
		('01M4ACCT000000000000000081', 'Create', 'create', NULL, '01M4ACCT000000000000000012', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000082', 'Read', 'read', NULL, '01M4ACCT000000000000000012', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000083', 'Update', 'update', NULL, '01M4ACCT000000000000000012', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000084', 'Delete', 'delete', NULL, '01M4ACCT000000000000000012', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000085', 'Set archived status', 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000012', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Rule Result
		('01M4ACCT000000000000000086', 'Create', 'create', NULL, '01M4ACCT000000000000000013', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000087', 'Read', 'read', NULL, '01M4ACCT000000000000000013', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000088', 'Update', 'update', NULL, '01M4ACCT000000000000000013', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000089', 'Delete', 'delete', NULL, '01M4ACCT000000000000000013', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000090', 'Set archived status', 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000013', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
