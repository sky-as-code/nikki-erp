package services

import (
	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
)

// The currency rules follow Odoo's res.currency: rounding is the settled unit and decimal_places
// is derived from it, so a client cannot store a pair that disagrees.

const (
	currencyErrNamespace = "essential"

	// maxDerivedDecimalPlaces mirrors the schema's decimal_places range and rounding's scale.
	maxDerivedDecimalPlaces = 6
)

var defaultCurrencyRounding = decimal.RequireFromString("0.01")

// rejectDecimalPlaces refuses a client-written decimal_places: the server derives it from rounding.
func rejectDecimalPlaces(cmd dmodel.DynamicFields, vErrs *ft.ClientErrors) {
	if _, sent := cmd[models.CurrencyFieldDecimalPlaces]; sent {
		vErrs.Append(*ft.NewBusinessViolation(models.CurrencyFieldDecimalPlaces,
			ft.ErrorKey("currency.decimal_places_server_managed", currencyErrNamespace),
			"decimal_places is derived from rounding and cannot be set"))
	}
}

// assertRoundingPositive rejects a zero rounding, which the schema's range admits but which would
// make every amount a division by zero.
func assertRoundingPositive(cmd dmodel.DynamicFields, vErrs *ft.ClientErrors) {
	if _, sent := cmd[models.CurrencyFieldRounding]; !sent {
		return
	}
	rounding := cmd.GetDecimal(models.CurrencyFieldRounding)
	if rounding == nil || !rounding.IsPositive() {
		vErrs.Append(*ft.NewValidationError(models.CurrencyFieldRounding,
			ft.ErrorKey("currency.rounding_not_positive", currencyErrNamespace),
			"rounding must be greater than zero"))
	}
}

// fillDecimalPlacesOnCreate derives decimal_places from the rounding being created, or from the
// schema's default rounding when the client sent none.
func fillDecimalPlacesOnCreate(
	_ corectx.Context, entity *dynamicresource.DynamicEntity, _ *ft.ClientErrors,
) (*dynamicresource.DynamicEntity, error) {
	fields := entity.GetFieldData()
	rounding := fields.GetDecimal(models.CurrencyFieldRounding)
	if rounding == nil {
		rounding = &defaultCurrencyRounding
	}
	places := decimalPlacesFromRounding(*rounding)
	fields.SetInt32(models.CurrencyFieldDecimalPlaces, &places)
	return entity, nil
}

// fillDecimalPlacesOnUpdate re-derives decimal_places only when the update changes rounding.
func fillDecimalPlacesOnUpdate(
	_ corectx.Context, entity *dynamicresource.DynamicEntity, _ *ft.ClientErrors,
) (*dynamicresource.DynamicEntity, error) {
	fields := entity.GetFieldData()
	if rounding := fields.GetDecimal(models.CurrencyFieldRounding); rounding != nil {
		places := decimalPlacesFromRounding(*rounding)
		fields.SetInt32(models.CurrencyFieldDecimalPlaces, &places)
	}
	return entity, nil
}

// decimalPlacesFromRounding is Odoo's ceil(log10(1/rounding)), computed without floats: the
// smallest n for which rounding * 10^n reaches a whole unit. 0.01 → 2, 0.05 → 2, 0.25 → 1, 1 → 0.
func decimalPlacesFromRounding(rounding decimal.Decimal) int32 {
	if !rounding.IsPositive() {
		return defaultDecimalPlaces
	}
	one := decimal.NewFromInt(1)
	for places := int32(0); places < maxDerivedDecimalPlaces; places++ {
		if rounding.Shift(places).GreaterThanOrEqual(one) {
			return places
		}
	}
	return maxDerivedDecimalPlaces
}

// roundToCurrency rounds an amount to the nearest multiple of rounding, half away from zero, then
// trims it to decimal_places. Without a rounding it falls back to decimal_places alone.
func roundToCurrency(amount decimal.Decimal, rounding *decimal.Decimal, places int32) decimal.Decimal {
	if rounding == nil || !rounding.IsPositive() {
		return amount.Round(places)
	}
	return amount.Div(*rounding).Round(0).Mul(*rounding).Round(places)
}
