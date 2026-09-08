-- The permission for recording what an executor physically handed over.
--
-- Only one new action is seeded. The other three demand-addressed routes — reserve_for_source,
-- release_for_source and reallocate_reservation — deliberately answer to the existing `reserve` and
-- `unreserve` permissions: they are the same powers over the same stock, differing only in whether
-- the caller names the transfer or the demand it was raised for, and giving them separate codes
-- would let a role hold one without the other by accident.
--
-- apply_fulfillment_result is separate from `validate` because the callers are different in kind.
-- Validate is an operator saying "ship this document"; this is a service relaying what a machine
-- managed to dispense. Granting a vending service `validate` would let it complete any transfer in
-- the organization, which is far more than reporting its own result requires.

DO $$
BEGIN
	IF EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'iam_actions'
	) THEN
		INSERT INTO "iam_actions" ("id", "name", "code", "description", "resource_id", "etag") VALUES
		('01M3NVFRES0000000000000001', 'Apply fulfillment result', 'apply_fulfillment_result', 'Record what an executor physically handed over against a reservation: consume what left, release the hold on what did not', '01M0B434KTKF2YBXPBBC2JH9XB', (EXTRACT(EPOCH FROM clock_timestamp()) * 1e9)::bigint::text)
		ON CONFLICT ("id") DO NOTHING;
	END IF;
END $$;
