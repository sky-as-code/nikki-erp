package promotion

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
)

// SalesVoucherDomainService owns normalising, uniqueness, type derivation, batch generation,
// revocation and the delete guard (BR-V01..V06, BR-L03, BR-L04).
type SalesVoucherDomainService interface {
	dynamicresource.CrudDomainService

	GenerateBatch(ctx corectx.Context, cmd GenerateVouchersCommand) (*GenerateVouchersResult, error)
	Revoke(ctx corectx.Context, voucherId, note string) (*RecordResult, error)
}

// SalesVoucherApplicationService is the authorized voucher surface.
type SalesVoucherApplicationService interface {
	dynamicresource.CrudApplicationService

	GenerateBatch(ctx corectx.Context, cmd GenerateVouchersCommand) (*GenerateVouchersResult, error)
	Revoke(ctx corectx.Context, cmd LifecycleCommand) (*RecordResult, error)
}

// The generate_batch body: {org_id, promotion_id, quantity, prefix?, valid_from?, expires_at?,
// customer_id?, usage_limit?}; it answers {count, vouchers: [{id, code}]}. revoke takes {note?}.
const (
	FieldVoucherQuantity = "quantity"
	FieldVoucherPrefix   = "prefix"
	FieldVoucherNote     = "note"
)

type (
	GenerateVouchersCommand = dynamicresource.CreateCommand
	GenerateVouchersResult  = dyn.OpResult[map[string]any]
)
