-- IAM resources and actions for the Payment & Invoice module.
--
-- The dynamic resource engine asserts permissions using the schema name as the resource code, so
-- these codes must stay byte-identical to the "paymentinvoice_payment_method",
-- "paymentinvoice_payment_profile", "paymentinvoice_order", "paymentinvoice_transaction",
-- "paymentinvoice_invoice" and "paymentinvoice_invoice_line" schema names. A code that drifts from
-- its schema denies every request, with nothing in the response pointing at the seed.
--
-- Payment Order carries no create action. An order is not typed in: it comes into existence only
-- through create_payment, which generates its identifiers and asks the gateway for the payment
-- instrument in the same step. Seeding create would advertise a way to record an order that no
-- gateway knows about, which is a payment that can never be collected or reconciled.
--
-- create_payment and refund are separate actions rather than folded into one, and neither is
-- update. They are materially different powers: refund moves money back out of the business and
-- cannot be undone, while create_payment only asks for money. A role that may take payments should
-- not thereby be able to return them.
--
-- Transaction is read only. Its rows are the audit trail of what each gateway was asked and what
-- it answered, written by the payment flow and the callbacks; a row that a person could write or
-- edit would be worthless as evidence in a dispute, which is the only reason it is kept.
--
-- Invoice's issue is separate from update for the same reason as refund: issuing assigns a number
-- from a gap-free sequence and freezes the totals, and an issued invoice is an accounting document
-- rather than a draft someone may still correct.
--
-- Payment Profile is a resource of its own rather than a corner of Payment Method. A method says
-- what a payer may choose; a profile holds the merchant credentials the money settles into, and
-- reading one is not the same power as reading the other.
--
-- Deliberately no iam_entitlements rows. Payment and invoice data is not universally readable —
-- it contains customer identities, tax codes and amounts — so access follows explicitly assigned
-- roles rather than a blanket grant.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_resources'
	) THEN
		INSERT INTO "iam_resources" (
			"id", "name", "code", "description", "owner_type", "max_scope", "min_scope", "created_at", "etag"
		) VALUES
		('01M0PAY1SMY4Q2E7VC4DDZAPWF', 'Payment Order', 'paymentinvoice_order', 'A request to collect money through a payment gateway', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1V59JZP88YNHA9FEC4F', 'Payment Transaction', 'paymentinvoice_transaction', 'One payment or refund attempt recorded against an order', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY16EZ1JQ1TS64Q21TPE7', 'Invoice', 'paymentinvoice_invoice', 'An accounting document issued for a sale', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY13K1F11EGR7EET7DHAM', 'Invoice Line', 'paymentinvoice_invoice_line', 'One charged item of an invoice', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1PRF7Z8QK3M2N4T5V6W', 'Payment Profile', 'paymentinvoice_payment_profile', 'One merchant account at a gateway: the credentials a payment settles into', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1GEVS6GFS22NP6V3ZRW', 'Payment Method', 'paymentinvoice_payment_method', 'A configured way the business can be paid, naming the adapter that carries it out', 'nikkierp', 'domain', 'org', NOW(), (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;

	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		-- Payment Method. Reference data: adding a gateway account or withdrawing one is a row
		-- rather than a release, which is the whole point of it not being an enum.
		('01M0PAY1Z535174MEV75RNHAVA', 'Create', 'create', NULL, '01M0PAY1GEVS6GFS22NP6V3ZRW', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1PSDC0EYRH48TMMAA74', 'Read', 'read', NULL, '01M0PAY1GEVS6GFS22NP6V3ZRW', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1RSEQZJJHXZ2FMVG54N', 'Update', 'update', NULL, '01M0PAY1GEVS6GFS22NP6V3ZRW', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1QHHPQ1HBD5ABESXBKS', 'Delete', 'delete', NULL, '01M0PAY1GEVS6GFS22NP6V3ZRW', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY11YGG1SE82N91KSX4MX', 'Set archived status', 'set_archived', 'Archive a payment method so it is out of the working set', '01M0PAY1GEVS6GFS22NP6V3ZRW', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Payment Profile. The credentials a payment settles into, so read is a materially larger
		-- grant here than on a payment method: reading one profile hands back the merchant secrets in
		-- the clear, and so does a search that names the config field.
		('01M0PAY1PRF7Z8QK3M2N4T5V70', 'Create', 'create', NULL, '01M0PAY1PRF7Z8QK3M2N4T5V6W', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1PRF7Z8QK3M2N4T5V71', 'Read', 'read', 'Read a profile, including the gateway credentials it holds', '01M0PAY1PRF7Z8QK3M2N4T5V6W', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1PRF7Z8QK3M2N4T5V72', 'Update', 'update', NULL, '01M0PAY1PRF7Z8QK3M2N4T5V6W', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1PRF7Z8QK3M2N4T5V73', 'Delete', 'delete', NULL, '01M0PAY1PRF7Z8QK3M2N4T5V6W', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1PRF7Z8QK3M2N4T5V74', 'Set archived status', 'set_archived', 'Withdraw a merchant account from use without losing the payments it collected', '01M0PAY1PRF7Z8QK3M2N4T5V6W', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Payment Order. No create: see the note at the top of this file.
		('01M0PAY10AA5Y9QZ20ZKN9CQ89', 'Read', 'read', NULL, '01M0PAY1SMY4Q2E7VC4DDZAPWF', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1ZFBX5PXV06Z0C12PQH', 'Update', 'update', 'Correct the descriptive fields of an order; never its amount or status', '01M0PAY1SMY4Q2E7VC4DDZAPWF', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1431QXXE8YQ1T0EMF6A', 'Delete', 'delete', NULL, '01M0PAY1SMY4Q2E7VC4DDZAPWF', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1B6W0ZMV0FYA841TK27', 'Take payment', 'create_payment', 'Open an order and ask the gateway for the payment instrument', '01M0PAY1SMY4Q2E7VC4DDZAPWF', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY15MWQ2C0ZT4Y516XR83', 'Refund', 'refund', 'Return money for an order already paid', '01M0PAY1SMY4Q2E7VC4DDZAPWF', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1HAVGWD8TKTK7KBZASV', 'Clear terminal orders', 'remove_pos_orders', 'Clear the orders queued on one card terminal', '01M0PAY1SMY4Q2E7VC4DDZAPWF', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Payment Transaction. Read only: see the note at the top of this file.
		('01M0PAY1SK0G6D93AX36PAN8G6', 'Read', 'read', NULL, '01M0PAY1V59JZP88YNHA9FEC4F', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Invoice
		('01M0PAY1PSC5C510R71NK4S771', 'Create', 'create', NULL, '01M0PAY16EZ1JQ1TS64Q21TPE7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1FGC5MGT15EF15VYP82', 'Read', 'read', NULL, '01M0PAY16EZ1JQ1TS64Q21TPE7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1QC6AXW0XZ3YVDASK7G', 'Update', 'update', NULL, '01M0PAY16EZ1JQ1TS64Q21TPE7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1938DEYJVVE8T3MZG6C', 'Delete', 'delete', NULL, '01M0PAY16EZ1JQ1TS64Q21TPE7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1AVSMSG28BHSAVAJG2R', 'Set archived status', 'set_archived', 'Archive an invoice so it is out of the working set', '01M0PAY16EZ1JQ1TS64Q21TPE7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1DS35GY6SWYQ8YEEZV2', 'Issue', 'issue', 'Close a draft: assign its number, freeze its totals and stamp the issue date', '01M0PAY16EZ1JQ1TS64Q21TPE7', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),

		-- Invoice Line
		('01M0PAY11935H0BKW0TFNJNNHW', 'Create', 'create', NULL, '01M0PAY13K1F11EGR7EET7DHAM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY18KMV8VEXQ78F5QBHAN', 'Read', 'read', NULL, '01M0PAY13K1F11EGR7EET7DHAM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1PSKGBGKYZ9K7Q1BGJC', 'Update', 'update', NULL, '01M0PAY13K1F11EGR7EET7DHAM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text),
		('01M0PAY1VFM713F2EEJTEAT2B0', 'Delete', 'delete', NULL, '01M0PAY13K1F11EGR7EET7DHAM', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
