package settings

import (
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
)

// QuotaAppService is the read other modules consume to learn how much of a resource an owner may
// use. It asserts no permission: it is a cross-module port, and the consuming module has already
// authorized the operation the quota limits.
type QuotaAppService interface {
	// GetQuota answers HasData=false when no quota row exists for the owner; the caller decides
	// whether that means "unlimited" or "nothing allowed".
	GetQuota(ctx corectx.Context, query GetQuotaQuery) (*GetQuotaResult, error)
}

type GetQuotaQuery struct {
	Name      string   `json:"name"`
	OwnerType string   `json:"owner_type"`
	OwnerId   model.Id `json:"owner_id"`
}

type GetQuotaResultData struct {
	Id        model.Id `json:"id"`
	Name      string   `json:"name"`
	OwnerType string   `json:"owner_type"`
	OwnerId   model.Id `json:"owner_id"`
	// ItStorageCapacityMibytes is nil when the quota sets no storage cap.
	ItStorageCapacityMibytes *int64 `json:"itstorage_capacity_mibytes"`
}

type GetQuotaResult = dyn.OpResult[GetQuotaResultData]
