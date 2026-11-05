-- Seed data for the Payment Invoice module: the payment methods a deployment can be paid through.
--
-- These four mirror the methods coremart/scripts/migrations/2001003_vendingmachine_seed.sql
-- configures on its kiosks. That file writes them into vdmc_new_payments, which is the vending
-- module's own copy; this file writes the master records they are a copy OF, so that Sales and
-- anything else can name a payment method without reaching into the vending module's table.
--
-- Three of the four have an adapter in the gateway registry. "MBBANK" deliberately does not, and
-- is seeded active anyway: usability is not a property of the row, it also depends on which
-- adapters the running build ships, and a deployment with no mbbank adapter must report the method
-- as present-but-unusable rather than hide it. Seeding one such method is what makes that path
-- exercisable without editing data.
--
-- max_amount is EXCLUSIVE (paymentinvoice/domain/services/order_domservice.go:341) - an amount
-- equal to it is refused. The bounds below are therefore one VND above the intended ceiling.
--
-- Ids are fixed rather than generated and every insert is ON CONFLICT DO NOTHING, so running this
-- file twice changes nothing.


DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'paymentinvoice_payment_methods'
	) THEN
		INSERT INTO "paymentinvoice_payment_methods" (
			"id", "code", "adapter_code", "name", "description", "currency_id", "min_amount", "max_amount", "is_active", "config", "is_archived", "created_at", "updated_at", "etag"
		) VALUES
		('01M3PM00000000000000000001', 'MOMO', 'momo', 'Ví MoMo', 'Thanh toán ví điện tử, quét từ điện thoại của khách hàng', NULL, 1000, 20000001, true, NULL, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PM00000000000000000002', 'VIETQR', 'vietqr', 'Chuyển khoản VietQR', 'Chuyển khoản ngân hàng qua mã QR được tạo', NULL, 1000, 500000001, true, NULL, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PM00000000000000000003', 'MPOS', 'mpos', 'Máy POS thẻ', 'Thanh toán thẻ trên máy POS vật lý', NULL, 10000, 50000001, true, NULL, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3PM00000000000000000004', 'MBBANK', 'mbbank', 'Chuyển khoản MB Bank', 'Chuyển khoản ngân hàng qua MB Bank. Cố ý khởi tạo ở trạng thái hoạt động dù chưa có adapter mbbank đi kèm: đây là điều kiện khả dụng phụ thuộc triển khai được thể hiện ra, và khung nhìn hợp nhất phải báo phương thức này không khả dụng thay vì không tồn tại', NULL, 1000, 500000001, true, NULL, false, NOW(), NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;


-- Invoice numbering for every seeded organization, mirroring the Payment & Invoice org lifecycle
-- (OrgLifecycleAppServiceImpl.InitOrg): one paymentinvoice.invoice definition per org
-- (INV-<year>-NNNNNN, reset yearly), its id held in the org's "paymentinvoice" settings. No counter
-- rows: no invoice is seeded, so each org's first invoice of a year is INV-<year>-000001.
DO $$
DECLARE
	invoice_settings_schema_id varchar;
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'essential_sequence_definitions'
	) AND EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'settings_records'
	) THEN
		INSERT INTO "essential_sequence_definitions" (
			"id", "configuration_scope", "org_id", "owner_module", "code", "name", "description", "sub_scope",
			"scope_type", "format_owner", "algorithm", "algorithm_params", "prefix_template", "suffix_template", "padding",
			"increment", "initial_value", "reset_period", "timezone", "issuance_retention_days", "first_issued_at",
			"is_archived", "created_at", "updated_at", "created_by", "etag"
		) VALUES
		('01M3SQND000000000000000011', 'org', '01JWNY20G23KD4RV5VWYABQYHD', 'paymentinvoice', 'paymentinvoice.invoice', 'Số hóa đơn', NULL, NULL, 'org', 'sequence', 'increment', NULL, 'INV-%(range_year)s-', NULL, 6, 1, 1, 'yearly', 'Asia/Ho_Chi_Minh', 90, NULL, false, NOW(), NULL, NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SQND000000000000000021', 'org', '01K02G6J1CYAN9K8V4PAGSQ5Z8', 'paymentinvoice', 'paymentinvoice.invoice', 'Số hóa đơn', NULL, NULL, 'org', 'sequence', 'increment', NULL, 'INV-%(range_year)s-', NULL, 6, 1, 1, 'yearly', 'Asia/Ho_Chi_Minh', 90, NULL, false, NOW(), NULL, NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M3SQND000000000000000031', 'org', '01K1H7M2K9VW3P5R7XQJY2C1Z9', 'paymentinvoice', 'paymentinvoice.invoice', 'Số hóa đơn', NULL, NULL, 'org', 'sequence', 'increment', NULL, 'INV-%(range_year)s-', NULL, 6, 1, 1, 'yearly', 'Asia/Ho_Chi_Minh', 90, NULL, false, NOW(), NULL, NULL, (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT DO NOTHING;

		-- Seeds run before the first boot registers the module's org settings, and a settings record
		-- must name its declaration. Snapshot of paymentinvoice/domain/settings.OrgSettingsSchemaBuilder()
		-- .Build().ToModelJson(); startup registration matches on (module_key, level), keeps this id
		-- and refreshes the document.
		INSERT INTO "settings_schemas" ("id", "module_key", "level", "schema", "created_at")
		VALUES ('01M3SQNS000000000000000001', 'paymentinvoice', 'org',
			$paymentinvoice_org_schema${"fields":[{"data_type":{"max":30,"min":0,"type":"string"},"description":{"$ref":"settings_desc.invoice_number_sequence_id"},"label":{"$ref":"paymentinvoice.settings.invoiceNumberSequenceId"},"metadata":{"allow_override":true},"name":"invoice_number_sequence_id"}],"label":{"$ref":"paymentinvoice.settings.orgSettings"},"name":"paymentinvoice_org_settings"}$paymentinvoice_org_schema$::jsonb, NOW())
		ON CONFLICT DO NOTHING;

		SELECT "id" INTO STRICT invoice_settings_schema_id
		FROM "settings_schemas" WHERE "module_key" = 'paymentinvoice' AND "level" = 'org';

		INSERT INTO "settings_records" (
			"id", "schema_id", "module_key", "level", "owner_type", "owner_id",
			"name", "value", "allow_override", "created_at", "updated_at"
		) VALUES
		('01M3SQNR000000000000000011', invoice_settings_schema_id, 'paymentinvoice', 'org', 'org', '01JWNY20G23KD4RV5VWYABQYHD', 'invoice_number_sequence_id', jsonb_build_object('value', '01M3SQND000000000000000011'), NULL, NOW(), NULL),
		('01M3SQNR000000000000000021', invoice_settings_schema_id, 'paymentinvoice', 'org', 'org', '01K02G6J1CYAN9K8V4PAGSQ5Z8', 'invoice_number_sequence_id', jsonb_build_object('value', '01M3SQND000000000000000021'), NULL, NOW(), NULL),
		('01M3SQNR000000000000000031', invoice_settings_schema_id, 'paymentinvoice', 'org', 'org', '01K1H7M2K9VW3P5R7XQJY2C1Z9', 'invoice_number_sequence_id', jsonb_build_object('value', '01M3SQND000000000000000031'), NULL, NOW(), NULL)
		ON CONFLICT DO NOTHING;
	END IF;
END $$;
