package external

import (
	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itCurrency "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/currency"
)

type orgDefaultsLookup struct {
	currencySvc itCurrency.CurrencyAppService
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
