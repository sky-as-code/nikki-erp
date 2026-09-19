package services

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	itCurrency "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/currency"
)

// defaultDecimalPlaces is used only when a currency record carries no decimal_places at all.
//
// The schema declares the field required with a default of 2, so this is unreachable through the
// engine; it exists so that a row written by a migration that forgot the column rounds to the
// commonest precision rather than to zero. Rounding money to zero places silently discards cents.
const defaultDecimalPlaces = 2

// NewCurrencyDomainService is handed the composable default by the currency onion. The CRUD is
// the default's; what this type adds is the money capability other modules depend on.
func NewCurrencyDomainService(base composable.CrudDomainService) itCurrency.CurrencyDomainService {
	return &CurrencyDomainServiceImpl{CrudDomainService: base}
}

type CurrencyDomainServiceImpl struct {
	composable.CrudDomainService
}

// GetCurrency fetches a single currency, so that a consuming module can validate a currency
// reference it holds without reaching into Essential's repositories.
func (this *CurrencyDomainServiceImpl) GetCurrency(
	ctx corectx.Context, query itCurrency.GetCurrencyQuery,
) (*itCurrency.GetCurrencyResult, error) {
	found, err := this.loadCurrency(ctx, query.Id)
	if err != nil {
		return nil, errors.Wrap(err, "get currency")
	}
	if found == nil {
		// Absence is a legitimate answer to a lookup, not a violation for the caller to render;
		// the caller decides what a missing currency means in its own context.
		return &itCurrency.GetCurrencyResult{HasData: false}, nil
	}

	return &itCurrency.GetCurrencyResult{
		Data:    toResultData(found),
		HasData: true,
	}, nil
}

// Round rounds an amount to the currency's decimal_places.
//
// A currency that cannot be found is a violation rather than a silent pass-through: returning the
// unrounded amount would put more fractional digits into a total than the currency has, and the
// error would surface later as a penny that does not reconcile.
func (this *CurrencyDomainServiceImpl) Round(
	ctx corectx.Context, query itCurrency.RoundQuery,
) (*itCurrency.RoundResult, error) {
	vErrs := &ft.ClientErrors{}

	found, err := this.loadCurrency(ctx, query.CurrencyId)
	if err != nil {
		return nil, errors.Wrap(err, "round amount")
	}
	if found == nil {
		vErrs.Append(*ft.NewBusinessViolation(models.CurrencyFieldId, "currency.not_found",
			"the currency does not exist"))
		return &itCurrency.RoundResult{ClientErrors: *vErrs}, nil
	}

	return &itCurrency.RoundResult{
		Data:    itCurrency.RoundResultData{Amount: query.Amount.Round(decimalPlacesOf(found))},
		HasData: true,
	}, nil
}

// AssertUsable reports whether a currency may be chosen for a NEW amount.
//
// The three checks are separate violations rather than one, because they are three different
// situations for whoever reads the error: a currency that was never seeded, one the business has
// withdrawn from use, and one archived out of the working set.
func (this *CurrencyDomainServiceImpl) AssertUsable(
	ctx corectx.Context, query itCurrency.AssertUsableQuery,
) (*itCurrency.AssertUsableResult, error) {
	vErrs := &ft.ClientErrors{}
	field := query.Field
	if field == "" {
		field = models.CurrencyFieldId
	}

	found, err := this.loadCurrency(ctx, query.Id)
	if err != nil {
		return nil, errors.Wrap(err, "assert currency usable")
	}

	switch {
	case found == nil:
		vErrs.Append(*ft.NewBusinessViolation(field, "currency.not_found",
			"the currency does not exist"))
	case !util.ValueOrZeroOf(found.GetIsActive()):
		// Inactive is not archived. A currency withdrawn from use stays readable so that amounts
		// already recorded in it still resolve; it simply cannot be chosen again.
		vErrs.Append(*ft.NewBusinessViolation(field, "currency.not_active",
			"the currency is not active"))
	case util.ValueOrZeroOf(found.IsArchived()):
		vErrs.Append(*ft.NewBusinessViolation(field, "currency.archived",
			"the currency is archived"))
	}

	if vErrs.Count() > 0 {
		return &itCurrency.AssertUsableResult{ClientErrors: *vErrs}, nil
	}
	return &itCurrency.AssertUsableResult{HasData: true}, nil
}

// loadCurrency fetches one currency by id, returning nil when it does not exist.
// GetCurrencyByCode resolves an alphabetic code, for callers holding a code rather than an id --
// the `default_currency` org setting chiefly, which stores free text.
func (this *CurrencyDomainServiceImpl) GetCurrencyByCode(
	ctx corectx.Context, query itCurrency.GetCurrencyByCodeQuery,
) (*itCurrency.GetCurrencyByCodeResult, error) {
	found, err := this.loadCurrencyBy(ctx, dmodel.DynamicFields{models.CurrencyFieldCode: query.Code})
	if err != nil {
		return nil, errors.Wrap(err, "get currency by code")
	}
	if found == nil {
		return &itCurrency.GetCurrencyByCodeResult{HasData: false}, nil
	}

	return &itCurrency.GetCurrencyByCodeResult{
		Data:    toResultData(found),
		HasData: true,
	}, nil
}

func (this *CurrencyDomainServiceImpl) loadCurrency(
	ctx corectx.Context, currencyId model.Id,
) (*models.Currency, error) {
	return this.loadCurrencyBy(ctx, dmodel.DynamicFields{models.CurrencyFieldId: currencyId})
}

// loadCurrencyBy reads at most one currency. Both callers filter on a unique column, so a filter
// that matches at all matches exactly once.
func (this *CurrencyDomainServiceImpl) loadCurrencyBy(
	ctx corectx.Context, filter dmodel.DynamicFields,
) (*models.Currency, error) {
	found, err := this.Repository().GetOne(ctx, dyn.RepoGetOneParam{Filter: filter})
	if err != nil {
		return nil, errors.Wrap(err, "loadCurrencyBy")
	}
	if !found.HasData {
		return nil, nil
	}
	return models.NewCurrencyFrom(found.Data), nil
}

func toResultData(src *models.Currency) itCurrency.GetCurrencyResultData {
	return itCurrency.GetCurrencyResultData{
		Id:            *src.GetId(),
		Code:          util.ValueOrZeroOf(src.GetCode()),
		Symbol:        util.ValueOrZeroOf(src.GetSymbol()),
		DecimalPlaces: decimalPlacesOf(src),
		IsActive:      util.ValueOrZeroOf(src.GetIsActive()),
		IsArchived:    util.ValueOrZeroOf(src.IsArchived()),
	}
}

// decimalPlacesOf reads a currency's precision, falling back only when the column is absent.
//
// The zero value is not used as the fallback: decimal_places of 0 is meaningful (VND is quoted in
// whole dong), so a missing column read as zero would silently round cents off a USD amount.
func decimalPlacesOf(src *models.Currency) int32 {
	if places := src.GetDecimalPlaces(); places != nil {
		return *places
	}
	return defaultDecimalPlaces
}
