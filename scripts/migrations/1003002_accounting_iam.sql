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
		('01M4ACCT000000000000000001', jsonb_build_object('en-US', 'Accounting - Tax Jurisdiction', 'vi-VN', 'Kế toán - Khu vực tài phán thuế'), 'accounting_tax_jurisdiction', 'A territory in which a taxing authority levies tax', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000002', jsonb_build_object('en-US', 'Accounting - Tax Group', 'vi-VN', 'Kế toán - Nhóm thuế'), 'accounting_tax_group', 'A grouping of taxes for display and reporting; never a calculation formula', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000003', jsonb_build_object('en-US', 'Accounting - Tax Rounding Policy', 'vi-VN', 'Kế toán - Chính sách làm tròn thuế'), 'accounting_tax_rounding_policy', 'How a computed tax amount is rounded, and whether per line or per document', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000004', jsonb_build_object('en-US', 'Accounting - Tax Product Classification', 'vi-VN', 'Kế toán - Phân loại thuế sản phẩm'), 'accounting_tax_product_classification', 'What a product is for tax purposes, which rules test rather than the product itself', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000005', jsonb_build_object('en-US', 'Accounting - Tax', 'vi-VN', 'Kế toán - Thuế'), 'accounting_tax', 'The stable business identity of a tax, as quoted on an invoice and in law', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000006', jsonb_build_object('en-US', 'Accounting - Tax Definition Version', 'vi-VN', 'Kế toán - Phiên bản định nghĩa thuế'), 'accounting_tax_definition_version', 'Everything about a tax that affects determination or calculation, for one effective period', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000007', jsonb_build_object('en-US', 'Accounting - Tax Rate Version', 'vi-VN', 'Kế toán - Phiên bản thuế suất'), 'accounting_tax_rate_version', 'The rate or fixed amount of a tax over one effective period', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000008', jsonb_build_object('en-US', 'Accounting - Tax Component', 'vi-VN', 'Kế toán - Thành phần thuế'), 'accounting_tax_component', 'One child tax inside a group tax, versioned with the definition that composes it', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000009', jsonb_build_object('en-US', 'Accounting - Tax Mapping', 'vi-VN', 'Kế toán - Ánh xạ thuế'), 'accounting_tax_mapping', 'A context-specific substitution of one tax for another', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000010', jsonb_build_object('en-US', 'Accounting - Tax Mapping Line', 'vi-VN', 'Kế toán - Dòng ánh xạ thuế'), 'accounting_tax_mapping_line', 'One source-tax to target-tax substitution within a mapping', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000011', jsonb_build_object('en-US', 'Accounting - Tax Rule', 'vi-VN', 'Kế toán - Quy tắc thuế'), 'accounting_tax_rule', 'A rule deciding which taxes apply to a transaction context', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000012', jsonb_build_object('en-US', 'Accounting - Tax Rule Condition', 'vi-VN', 'Kế toán - Điều kiện quy tắc thuế'), 'accounting_tax_rule_condition', 'One typed predicate on the tax context', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000013', jsonb_build_object('en-US', 'Accounting - Tax Rule Result', 'vi-VN', 'Kế toán - Kết quả quy tắc thuế'), 'accounting_tax_rule_result', 'What a matching rule does to the candidate tax set', 'nikkierp', 'tenant', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Tax Jurisdiction
		('01M4ACCT000000000000000014', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M4ACCT000000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000015', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M4ACCT000000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000016', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M4ACCT000000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000017', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M4ACCT000000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000018', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000001', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Group
		('01M4ACCT000000000000000019', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M4ACCT000000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000020', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M4ACCT000000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000021', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M4ACCT000000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000022', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M4ACCT000000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000023', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000002', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Rounding Policy
		('01M4ACCT000000000000000024', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M4ACCT000000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000025', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M4ACCT000000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000026', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M4ACCT000000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000027', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M4ACCT000000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000028', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000029', jsonb_build_object('en-US', 'Publish', 'vi-VN', 'Xuất bản'), 'publish', 'Put a rounding policy into effect, freezing the fields that decide how tax is rounded', '01M4ACCT000000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000030', jsonb_build_object('en-US', 'Withdraw', 'vi-VN', 'Ngừng áp dụng'), 'withdraw', 'Retire a rounding policy from new determination, leaving it readable for audit', '01M4ACCT000000000000000003', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Product Classification
		('01M4ACCT000000000000000031', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M4ACCT000000000000000004', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000032', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M4ACCT000000000000000004', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000033', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M4ACCT000000000000000004', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000034', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M4ACCT000000000000000004', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000035', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000004', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax
		('01M4ACCT000000000000000036', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M4ACCT000000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000037', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M4ACCT000000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000038', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M4ACCT000000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000039', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M4ACCT000000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000040', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000041', jsonb_build_object('en-US', 'Override tax', 'vi-VN', 'Ghi đè thuế'), 'override', 'Substitute the determined tax set on a transaction; requires a reason and is recorded in the snapshot', '01M4ACCT000000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000042', jsonb_build_object('en-US', 'Simulate tax', 'vi-VN', 'Mô phỏng thuế'), 'simulate', 'Run the tax engine against hypothetical inputs without creating any transaction', '01M4ACCT000000000000000005', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Definition Version
		('01M4ACCT000000000000000043', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M4ACCT000000000000000006', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000044', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M4ACCT000000000000000006', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000045', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M4ACCT000000000000000006', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000046', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M4ACCT000000000000000006', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000047', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000006', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000048', jsonb_build_object('en-US', 'Publish', 'vi-VN', 'Xuất bản'), 'publish', 'Put a tax definition into effect, freezing every field that decides an amount', '01M4ACCT000000000000000006', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000049', jsonb_build_object('en-US', 'Withdraw', 'vi-VN', 'Ngừng áp dụng'), 'withdraw', 'Retire a tax definition from new determination, leaving it readable for audit', '01M4ACCT000000000000000006', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Rate Version
		('01M4ACCT000000000000000050', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M4ACCT000000000000000007', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000051', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M4ACCT000000000000000007', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000052', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M4ACCT000000000000000007', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000053', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M4ACCT000000000000000007', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000054', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000007', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000055', jsonb_build_object('en-US', 'Publish', 'vi-VN', 'Xuất bản'), 'publish', 'Put a rate into effect; two published rates of one tax may never overlap in time', '01M4ACCT000000000000000007', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000056', jsonb_build_object('en-US', 'Withdraw', 'vi-VN', 'Ngừng áp dụng'), 'withdraw', 'Retire a rate from new determination, leaving it readable for audit', '01M4ACCT000000000000000007', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Component
		('01M4ACCT000000000000000057', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M4ACCT000000000000000008', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000058', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M4ACCT000000000000000008', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000059', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M4ACCT000000000000000008', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000060', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M4ACCT000000000000000008', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000061', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000008', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Mapping
		('01M4ACCT000000000000000062', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M4ACCT000000000000000009', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000063', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M4ACCT000000000000000009', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000064', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M4ACCT000000000000000009', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000065', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M4ACCT000000000000000009', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000066', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000009', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000067', jsonb_build_object('en-US', 'Publish', 'vi-VN', 'Xuất bản'), 'publish', 'Put a mapping into effect, freezing it and its lines', '01M4ACCT000000000000000009', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000068', jsonb_build_object('en-US', 'Withdraw', 'vi-VN', 'Ngừng áp dụng'), 'withdraw', 'Retire a mapping from new determination, leaving it readable for audit', '01M4ACCT000000000000000009', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Mapping Line
		('01M4ACCT000000000000000069', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M4ACCT000000000000000010', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000070', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M4ACCT000000000000000010', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000071', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M4ACCT000000000000000010', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000072', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M4ACCT000000000000000010', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000073', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000010', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Rule
		('01M4ACCT000000000000000074', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M4ACCT000000000000000011', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000075', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M4ACCT000000000000000011', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000076', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M4ACCT000000000000000011', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000077', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M4ACCT000000000000000011', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000078', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000011', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000079', jsonb_build_object('en-US', 'Publish', 'vi-VN', 'Xuất bản'), 'publish', 'Put a rule into effect, freezing its conditions, results and priority', '01M4ACCT000000000000000011', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000080', jsonb_build_object('en-US', 'Withdraw', 'vi-VN', 'Ngừng áp dụng'), 'withdraw', 'Retire a rule from new determination, leaving it readable for audit', '01M4ACCT000000000000000011', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Rule Condition
		('01M4ACCT000000000000000081', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M4ACCT000000000000000012', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000082', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M4ACCT000000000000000012', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000083', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M4ACCT000000000000000012', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000084', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M4ACCT000000000000000012', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000085', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000012', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		-- Tax Rule Result
		('01M4ACCT000000000000000086', jsonb_build_object('en-US', 'Create', 'vi-VN', 'Tạo'), 'create', NULL, '01M4ACCT000000000000000013', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000087', jsonb_build_object('en-US', 'Read', 'vi-VN', 'Xem'), 'read', NULL, '01M4ACCT000000000000000013', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000088', jsonb_build_object('en-US', 'Update', 'vi-VN', 'Cập nhật'), 'update', NULL, '01M4ACCT000000000000000013', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000089', jsonb_build_object('en-US', 'Delete', 'vi-VN', 'Xóa'), 'delete', NULL, '01M4ACCT000000000000000013', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M4ACCT000000000000000090', jsonb_build_object('en-US', 'Set archived status', 'vi-VN', 'Đặt trạng thái lưu trữ'), 'set_archived', 'Archive this configuration, or bring an archived one back', '01M4ACCT000000000000000013', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
