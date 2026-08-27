-- Seed data for the Accounting module: the Vietnam VAT baseline.
--
-- Every rate, threshold and date below is DATA, not code. BR-TAX-ESS-037 and AC-TAX-32 require that
-- Vietnamese VAT be configurable without touching the engine, so nothing in modules/accounting
-- contains the numbers 0, 5, 8 or 10 as a tax rate; the calculator reads whatever these rows say.
--
-- Scope is the credit/invoice method (phuong phap khau tru) only. The direct-on-revenue method is
-- out of V1 (BR-TAX-ESS-SUP-028), so this file does not pretend to cover it.
--
-- Three things here are worth explaining, because each looks like a mistake until it does not:
--
--   * VAT 0% is seeded as a percentage tax with a rate of 0, NOT as calculation_type 'none'.
--     A zero-rated export is taxed, at zero, and the supplier may still reclaim input tax on it.
--     Modelling it as "no calculation" would erase the fact that a rate applied at all, and with it
--     the distinction from an exemption (BR-TAX-ESS-SUP-015, AC-TAX-15).
--
--   * The 2% reduction (Resolution 204/2025/QH15) is a RULE that removes VN_VAT_10 and adds
--     VN_VAT_8, not an edit to the 10% rate. That is what lets it expire on 2026-12-31 by itself,
--     with no deployment and no data change (AC-TAX-33) — and what keeps the pre-reduction rate
--     intact for any transaction dated before it began.
--
--   * Both the 10% rate version and the 8% one are published and both are "in force" today. They do
--     not overlap for a single tax: they belong to two different taxes, and the rule decides which
--     tax applies to a given line. TAX-SUP-INV-06 forbids two published rates OF ONE TAX
--     overlapping, which is not this.
--
-- Every id is fixed and every insert is ON CONFLICT DO NOTHING, so running this file twice changes
-- nothing.


DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'accounting_tax_jurisdictions'
	) THEN
		INSERT INTO "accounting_tax_jurisdictions" (
			"id", "org_id", "code", "name", "country_code", "level", "parent_id", "authority_name", "is_archived", "created_at", "updated_at", "etag"
		) VALUES
		('01M4ACCTVN0000000000000001', '01KNV107H8P0V2CB1R9T9ZNMFF', 'VN', '{"en-US": "Vietnam", "vi-VN": "Viet Nam"}', 'VN', 'country', NULL, 'Tong cuc Thue', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'accounting_tax_groups'
	) THEN
		INSERT INTO "accounting_tax_groups" (
			"id", "org_id", "code", "name", "display_name", "description", "display_sequence", "is_archived", "created_at", "updated_at", "etag"
		) VALUES
		('01M4ACCTVN0000000000000002', '01KNV107H8P0V2CB1R9T9ZNMFF', 'VAT', '{"en-US": "Value Added Tax", "vi-VN": "Thue gia tri gia tang"}', '{"en-US": "VAT", "vi-VN": "GTGT"}', 'Vietnamese value added tax', 10, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'accounting_tax_rounding_policies'
	) THEN
		INSERT INTO "accounting_tax_rounding_policies" (
			"id", "org_id", "code", "name", "jurisdiction_id", "currency_code", "rounding_scope", "rounding_method", "rounding_increment", "precision", "version_no", "effective_from", "effective_to", "supersedes_policy_id", "lifecycle_status", "is_archived", "created_at", "updated_at", "etag"
		) VALUES
		('01M4ACCTVN0000000000000003', '01KNV107H8P0V2CB1R9T9ZNMFF', 'VN-VND-DOC', '{"en-US": "Vietnam VND document rounding", "vi-VN": "Lam tron VND theo chung tu"}', '01M4ACCTVN0000000000000001', 'VND', 'document', 'half_up', 1, 0, 1, '2025-07-01', NULL, NULL, 'published', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'accounting_tax_product_classifications'
	) THEN
		INSERT INTO "accounting_tax_product_classifications" (
			"id", "org_id", "code", "name", "jurisdiction_id", "external_code", "description", "is_archived", "created_at", "updated_at", "etag"
		) VALUES
		('01M4ACCTVN0000000000000010', '01KNV107H8P0V2CB1R9T9ZNMFF', 'VN_STANDARD', '{"en-US": "Standard rated", "vi-VN": "Thue suat pho thong"}', '01M4ACCTVN0000000000000001', NULL, 'Goods and services at the standard VAT rate', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000011', '01KNV107H8P0V2CB1R9T9ZNMFF', 'VN_VAT_REDUCTION_ELIGIBLE', '{"en-US": "Eligible for VAT reduction", "vi-VN": "Duoc giam thue GTGT"}', '01M4ACCTVN0000000000000001', NULL, 'Goods and services the 2% reduction applies to, per Resolution 204/2025/QH15', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000012', '01KNV107H8P0V2CB1R9T9ZNMFF', 'VN_ESSENTIAL', '{"en-US": "Essential goods", "vi-VN": "Hang hoa thiet yeu"}', '01M4ACCTVN0000000000000001', NULL, 'Goods and services taxed at the reduced statutory rate', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000013', '01KNV107H8P0V2CB1R9T9ZNMFF', 'VN_EXPORT', '{"en-US": "Export", "vi-VN": "Xuat khau"}', '01M4ACCTVN0000000000000001', NULL, 'Exported goods and services, zero rated', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000014', '01KNV107H8P0V2CB1R9T9ZNMFF', 'VN_EXEMPT', '{"en-US": "Exempt", "vi-VN": "Khong chiu thue"}', '01M4ACCTVN0000000000000001', NULL, 'Supplies outside the scope of VAT', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'accounting_taxes'
	) THEN
		INSERT INTO "accounting_taxes" (
			"id", "org_id", "code", "name", "tax_kind", "invoice_label", "description", "is_archived", "created_at", "updated_at", "etag"
		) VALUES
		('01M4ACCTVN0000000000000020', '01KNV107H8P0V2CB1R9T9ZNMFF', 'VN_VAT_10', '{"en-US": "VAT 10%", "vi-VN": "GTGT 10%"}', 'vat', '{"en-US": "VAT 10%", "vi-VN": "GTGT 10%"}', 'Standard Vietnamese VAT rate', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000021', '01KNV107H8P0V2CB1R9T9ZNMFF', 'VN_VAT_8', '{"en-US": "VAT 8%", "vi-VN": "GTGT 8%"}', 'vat', '{"en-US": "VAT 8%", "vi-VN": "GTGT 8%"}', 'Temporarily reduced VAT rate under Resolution 204/2025/QH15', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000022', '01KNV107H8P0V2CB1R9T9ZNMFF', 'VN_VAT_5', '{"en-US": "VAT 5%", "vi-VN": "GTGT 5%"}', 'vat', '{"en-US": "VAT 5%", "vi-VN": "GTGT 5%"}', 'Reduced statutory VAT rate', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000023', '01KNV107H8P0V2CB1R9T9ZNMFF', 'VN_VAT_0', '{"en-US": "VAT 0%", "vi-VN": "GTGT 0%"}', 'vat', '{"en-US": "VAT 0%", "vi-VN": "GTGT 0%"}', 'Zero rated: taxed at 0%, which is not the same as exempt', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000024', '01KNV107H8P0V2CB1R9T9ZNMFF', 'VN_VAT_EXEMPT', '{"en-US": "VAT exempt", "vi-VN": "Khong chiu thue GTGT"}', 'vat', '{"en-US": "VAT exempt", "vi-VN": "Khong chiu thue GTGT"}', 'Exempt supplies: no VAT and no input tax recovery', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'accounting_tax_definition_versions'
	) THEN
		INSERT INTO "accounting_tax_definition_versions" (
			"id", "org_id", "tax_id", "version_no", "usage", "jurisdiction_id", "tax_group_id", "calculation_type", "tax_treatment", "price_inclusion_mode", "sequence", "affect_subsequent_base", "base_affected_by_previous", "effective_from", "effective_to", "legal_reference", "supersedes_version_id", "lifecycle_status", "is_archived", "created_at", "updated_at", "etag"
		) VALUES
		('01M4ACCTVN0000000000000030', '01KNV107H8P0V2CB1R9T9ZNMFF', '01M4ACCTVN0000000000000020', 1, 'sale', '01M4ACCTVN0000000000000001', '01M4ACCTVN0000000000000002', 'percentage', 'taxable', 'inherit', 10, false, false, '2025-07-01', NULL, 'Luat Thue GTGT 48/2024/QH15', NULL, 'published', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000031', '01KNV107H8P0V2CB1R9T9ZNMFF', '01M4ACCTVN0000000000000021', 1, 'sale', '01M4ACCTVN0000000000000001', '01M4ACCTVN0000000000000002', 'percentage', 'taxable', 'inherit', 10, false, false, '2025-07-01', NULL, 'Luat Thue GTGT 48/2024/QH15', NULL, 'published', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000032', '01KNV107H8P0V2CB1R9T9ZNMFF', '01M4ACCTVN0000000000000022', 1, 'sale', '01M4ACCTVN0000000000000001', '01M4ACCTVN0000000000000002', 'percentage', 'taxable', 'inherit', 10, false, false, '2025-07-01', NULL, 'Luat Thue GTGT 48/2024/QH15', NULL, 'published', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000033', '01KNV107H8P0V2CB1R9T9ZNMFF', '01M4ACCTVN0000000000000023', 1, 'sale', '01M4ACCTVN0000000000000001', '01M4ACCTVN0000000000000002', 'percentage', 'zero_rated', 'inherit', 10, false, false, '2025-07-01', NULL, 'Luat Thue GTGT 48/2024/QH15', NULL, 'published', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000034', '01KNV107H8P0V2CB1R9T9ZNMFF', '01M4ACCTVN0000000000000024', 1, 'sale', '01M4ACCTVN0000000000000001', '01M4ACCTVN0000000000000002', 'none', 'exempt', 'inherit', 10, false, false, '2025-07-01', NULL, 'Luat Thue GTGT 48/2024/QH15', NULL, 'published', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'accounting_tax_rate_versions'
	) THEN
		INSERT INTO "accounting_tax_rate_versions" (
			"id", "org_id", "tax_id", "version_no", "rate", "fixed_amount", "currency_code", "rate_uom_id", "effective_from", "effective_to", "legal_reference", "description", "lifecycle_status", "is_archived", "created_at", "updated_at", "etag"
		) VALUES
		('01M4ACCTVN0000000000000040', '01KNV107H8P0V2CB1R9T9ZNMFF', '01M4ACCTVN0000000000000020', 1, '10', NULL, NULL, NULL, '2025-07-01', NULL, 'Luat Thue GTGT 48/2024/QH15', NULL, 'published', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000041', '01KNV107H8P0V2CB1R9T9ZNMFF', '01M4ACCTVN0000000000000021', 1, '8', NULL, NULL, NULL, '2025-07-01', NULL, 'Nghi quyet 204/2025/QH15', NULL, 'published', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000042', '01KNV107H8P0V2CB1R9T9ZNMFF', '01M4ACCTVN0000000000000022', 1, '5', NULL, NULL, NULL, '2025-07-01', NULL, 'Luat Thue GTGT 48/2024/QH15', NULL, 'published', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000043', '01KNV107H8P0V2CB1R9T9ZNMFF', '01M4ACCTVN0000000000000023', 1, '0', NULL, NULL, NULL, '2025-07-01', NULL, 'Luat Thue GTGT 48/2024/QH15', NULL, 'published', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'accounting_tax_rules'
	) THEN
		INSERT INTO "accounting_tax_rules" (
			"id", "org_id", "code", "name", "jurisdiction_id", "priority", "stop_processing", "effective_from", "effective_to", "legal_reference", "version_no", "supersedes_rule_id", "lifecycle_status", "is_archived", "created_at", "updated_at", "etag"
		) VALUES
		('01M4ACCTVN0000000000000051', '01KNV107H8P0V2CB1R9T9ZNMFF', 'VN-EXPORT-ZERO-RATED', '{"en-US": "Exports are zero rated", "vi-VN": "Xuat khau huong thue suat 0%"}', '01M4ACCTVN0000000000000001', 10, true, '2025-07-01', NULL, 'Luat Thue GTGT 48/2024/QH15', 1, NULL, 'published', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000052', '01KNV107H8P0V2CB1R9T9ZNMFF', 'VN-EXEMPT-SUPPLY', '{"en-US": "Exempt supplies carry no VAT", "vi-VN": "Hang khong chiu thue GTGT"}', '01M4ACCTVN0000000000000001', 20, true, '2025-07-01', NULL, 'Luat Thue GTGT 48/2024/QH15', 1, NULL, 'published', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000050', '01KNV107H8P0V2CB1R9T9ZNMFF', 'VN-VAT-REDUCTION-2025', '{"en-US": "2% VAT reduction", "vi-VN": "Giam 2% thue GTGT"}', '01M4ACCTVN0000000000000001', 30, false, '2025-07-01', '2026-12-31', 'Nghi quyet 204/2025/QH15', 1, NULL, 'published', false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'accounting_tax_rule_conditions'
	) THEN
		INSERT INTO "accounting_tax_rule_conditions" (
			"id", "org_id", "tax_rule_id", "field_key", "operator", "value", "value_currency_code", "sequence", "is_archived", "created_at", "updated_at", "etag"
		) VALUES
		('01M4ACCTVN0000000000000060', '01KNV107H8P0V2CB1R9T9ZNMFF', '01M4ACCTVN0000000000000051', 'product_tax_classification', 'eq', '["VN_EXPORT"]', NULL, 1, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000061', '01KNV107H8P0V2CB1R9T9ZNMFF', '01M4ACCTVN0000000000000052', 'product_tax_classification', 'eq', '["VN_EXEMPT"]', NULL, 1, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000062', '01KNV107H8P0V2CB1R9T9ZNMFF', '01M4ACCTVN0000000000000050', 'product_tax_classification', 'eq', '["VN_VAT_REDUCTION_ELIGIBLE"]', NULL, 1, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'accounting_tax_rule_results'
	) THEN
		INSERT INTO "accounting_tax_rule_results" (
			"id", "org_id", "tax_rule_id", "action", "tax_id", "tax_mapping_id", "tax_treatment", "sequence", "is_archived", "created_at", "updated_at", "etag"
		) VALUES
		('01M4ACCTVN0000000000000070', '01KNV107H8P0V2CB1R9T9ZNMFF', '01M4ACCTVN0000000000000051', 'remove_tax', '01M4ACCTVN0000000000000020', NULL, NULL, 1, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000071', '01KNV107H8P0V2CB1R9T9ZNMFF', '01M4ACCTVN0000000000000051', 'add_tax', '01M4ACCTVN0000000000000023', NULL, NULL, 2, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000072', '01KNV107H8P0V2CB1R9T9ZNMFF', '01M4ACCTVN0000000000000052', 'remove_tax', '01M4ACCTVN0000000000000020', NULL, NULL, 1, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000073', '01KNV107H8P0V2CB1R9T9ZNMFF', '01M4ACCTVN0000000000000052', 'add_tax', '01M4ACCTVN0000000000000024', NULL, NULL, 2, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000074', '01KNV107H8P0V2CB1R9T9ZNMFF', '01M4ACCTVN0000000000000050', 'remove_tax', '01M4ACCTVN0000000000000020', NULL, NULL, 1, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCTVN0000000000000075', '01KNV107H8P0V2CB1R9T9ZNMFF', '01M4ACCTVN0000000000000050', 'add_tax', '01M4ACCTVN0000000000000021', NULL, NULL, 2, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

END $$;
