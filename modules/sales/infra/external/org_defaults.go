package external

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	itCurrency "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/currency"
	invModels "github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"
)

type orgDefaultsLookup struct {
	currencySvc      itCurrency.CurrencyAppService
	operationTypeSvc itStock.StockOperationTypeDomainService
}

func (this *orgDefaultsLookup) CurrencyIdByCode(ctx corectx.Context, code string) (model.Id, error) {
	result, err := this.currencySvc.GetCurrencyByCode(ctx, itCurrency.GetCurrencyByCodeQuery{Code: code})
	if err != nil {
		return "", err
	}
	if result.ClientErrors.Count() > 0 || !result.HasData {
		return "", errors.Errorf("currency %s is not available: %v", code, result.ClientErrors.ToError())
	}
	return result.Data.Id, nil
}

func (this *orgDefaultsLookup) OperationTypeIdByCode(
	ctx corectx.Context, orgId model.Id, code string,
) (model.Id, error) {
	graph := &dmodel.SearchGraph{}
	graph.And(
		*dmodel.NewSearchNode().NewCondition(invModels.StockOperationTypeFieldCode, dmodel.Equals, code),
		*dmodel.NewSearchNode().NewCondition(basemodel.FieldOrgId, dmodel.Equals, string(orgId)),
	)
	result, err := this.operationTypeSvc.Search(ctx, dmodel.DynamicFields{
		"fields": []string{basemodel.FieldId},
		"graph":  graph,
		"page":   0,
		"size":   1,
	})
	if err != nil {
		return "", err
	}
	if result.ClientErrors.Count() > 0 {
		return "", result.ClientErrors.ToError()
	}
	if !result.HasData || len(result.Data.Items) == 0 {
		return "", errors.Errorf("organization %s has no operation type %s", orgId, code)
	}
	id, _ := result.Data.Items[0][basemodel.FieldId].(string)
	return model.Id(id), nil
}
