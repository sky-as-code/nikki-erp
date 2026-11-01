package external

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// OrgDefaultsLookupExtService resolves what a new organization's sales defaults name by code,
// since the ids differ per tenant and per organization.
type OrgDefaultsLookupExtService interface {
	CurrencyIdByCode(ctx corectx.Context, code string) (model.Id, error)
}
