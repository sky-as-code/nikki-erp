package app

// Permission codes of the custom actions. Each must match a seeded iam_actions row; the resource
// code is the schema name, as for the built-in actions.

// Warehouse and location lifecycle. Only these two resources have suspend and resume, because
// only they have an operational state independent of archiving.
const (
	PermissionSuspend                = "suspend"
	PermissionResume                 = "resume"
	PermissionMoveLocation           = "move"
	PermissionConfigureIncomingFlow  = "configure_incoming_flow"
	PermissionConfigureOutgoingFlow  = "configure_outgoing_flow"
	PermissionSuggestPutawayLocation = "suggest_location"
)

// Counting, kept separate because they are different powers held by different people: a hand
// enters what they counted, a supervisor makes the count the balance. Applying an adjustment
// writes stock that no trade movement explains.
const (
	PermissionEnterCount      = "enter_count"
	PermissionResetCount      = "reset_count"
	PermissionApplyAdjustment = "apply_adjustment"
	PermissionScheduleCount   = "schedule_count"
	PermissionAssignCounter   = "assign_counter"
)

// One permission covers the product-facing stock reads: they are the same power sliced by which
// product you look at, and granting them separately would allow "may see a total but not the
// rows behind it".
const PermissionReadProductStock = "read_product_stock"

// The movement operations of a transfer.
const (
	PermissionConfirm                = "confirm"
	PermissionReserve                = "reserve"
	PermissionUnreserve              = "unreserve"
	PermissionValidate               = "validate"
	PermissionCancel                 = "cancel"
	PermissionCreateReturn           = "create_return"
	PermissionApplyFulfillmentResult = "apply_fulfillment_result"
)

const PermissionDoScrap = "do_scrap"

// Warehouse-level reservations. Reserve and check are collection-level operations on the
// reservation resource; release and consume act on one reservation. Protect has no permission
// because it is never reachable from a client: only the paying module's port calls it.
const (
	PermissionReserveWarehouseStock      = "reserve_warehouse_stock"
	PermissionCheckWarehouseAvailability = "check_warehouse_availability"
	PermissionReleaseReservation         = "release_reservation"
	PermissionConsumeReservation         = "consume_reservation"
)
