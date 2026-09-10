package app

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	itWarehouse "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/warehouse"
)

// NewPutawayRuleApplicationService is handed the composable default by the putaway rule onion.
func NewPutawayRuleApplicationService(base composable.CrudApplicationService) itWarehouse.PutawayRuleApplicationService {
	return &PutawayRuleApplicationServiceImpl{CrudApplicationService: base}
}

type PutawayRuleApplicationServiceImpl struct {
	composable.CrudApplicationService
}

// suggestPutawayResponse is what the lookup answers with. Both fields are empty when no rule
// applies, which is a normal answer rather than an error.
type suggestPutawayResponse struct {
	DestinationLocationId string `json:"destination_location_id"`
	MatchedRuleId         string `json:"matched_rule_id"`
}

// SuggestLocation answers where arriving goods should be put. Nothing is written: no quant
// moves, no move is created, nothing is reserved. Acting on the answer is the caller's next
// step, through the Stock movement operations.
func (this *PutawayRuleApplicationServiceImpl) SuggestLocation(
	ctx corectx.Context, query itWarehouse.SuggestPutawayLocationQuery,
) (*itWarehouse.SuggestPutawayLocationResult, error) {
	if _, cErrs := this.AssertAction(ctx, PermissionSuggestPutawayLocation, query); cErrs != nil {
		return anyFailure(cErrs, nil)
	}

	suggestion, err := services.SuggestPutawayLocation(ctx, services.PutawayContext{
		WarehouseId:       readStringField(query, "warehouse_id"),
		ArrivalLocationId: readStringField(query, "arrival_location_id"),
		ProductId:         readStringField(query, "product_id"),
		ProductCategoryId: readStringField(query, "product_category_id"),
		PackageTypeId:     readStringField(query, "package_type_id"),
	})
	if err != nil {
		return nil, err
	}
	response := suggestPutawayResponse{}
	if suggestion != nil {
		response.DestinationLocationId = suggestion.DestinationLocationId
		response.MatchedRuleId = suggestion.MatchedRuleId
	}
	return anyResult(response), nil
}
