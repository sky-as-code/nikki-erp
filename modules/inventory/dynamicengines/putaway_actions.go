package dynamicengines

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
)

// suggestPutawayResponse is what the lookup answers with. Both fields are empty when no rule
// applies, which is a normal answer rather than an error.
type suggestPutawayResponse struct {
	DestinationLocationId string `json:"destination_location_id"`
	MatchedRuleId         string `json:"matched_rule_id"`
}

// processSuggestPutawayLocation answers where arriving goods should be put. Nothing is written: no
// quant moves, no move is created, nothing is reserved. Acting on the answer is the caller's next
// step, through the Stock movement engine.
func processSuggestPutawayLocation(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	suggestion, err := services.SuggestPutawayLocation(ctx, services.PutawayContext{
		WarehouseId:       readStringField(input.Params, paramPutawayWarehouse),
		ArrivalLocationId: readStringField(input.Params, paramArrivalLocation),
		ProductId:         readStringField(input.Params, paramProductId),
		ProductCategoryId: readStringField(input.Params, paramProductCategory),
		PackageTypeId:     readStringField(input.Params, paramPackageType),
	})
	if err != nil {
		return nil, err
	}

	response := suggestPutawayResponse{}
	if suggestion != nil {
		response.DestinationLocationId = suggestion.DestinationLocationId
		response.MatchedRuleId = suggestion.MatchedRuleId
	}
	return &drif.ActionResult{Data: response, HasData: true}, nil
}
