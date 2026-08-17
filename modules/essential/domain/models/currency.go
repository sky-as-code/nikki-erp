package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	CurrencySchemaName = "essential_currency"

	CurrencyFieldId            = basemodel.FieldId
	CurrencyFieldCode          = "code"
	CurrencyFieldNumericCode   = "numeric_code"
	CurrencyFieldName          = "name"
	CurrencyFieldSymbol        = "symbol"
	CurrencyFieldDecimalPlaces = "decimal_places"
	CurrencyFieldIsActive      = "is_active"
)

//go:embed currency.json
var currencySchemaJson string

func CurrencySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(currencySchemaJson)
}

// Currency is the set of currencies the system may denominate an amount in.
//
// It lives in Essential rather than in the module that happens to need it first, because more
// than one module records money and they must all agree on the same list: an invoice raised in
// one and a payment taken in another are the same currency or they cannot be reconciled.
type Currency struct {
	basemodel.DynamicModelBase
}

func NewCurrency() *Currency {
	return &Currency{basemodel.NewDynamicModel()}
}

func NewCurrencyFrom(src dmodel.DynamicFields) *Currency {
	return &Currency{basemodel.NewDynamicModel(src)}
}

func (this Currency) GetCode() *string {
	return this.GetFieldData().GetString(CurrencyFieldCode)
}

func (this *Currency) SetCode(v *string) {
	this.GetFieldData().SetString(CurrencyFieldCode, v)
}

func (this Currency) GetNumericCode() *string {
	return this.GetFieldData().GetString(CurrencyFieldNumericCode)
}

func (this *Currency) SetNumericCode(v *string) {
	this.GetFieldData().SetString(CurrencyFieldNumericCode, v)
}

func (this Currency) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(CurrencyFieldName)
}

func (this *Currency) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(CurrencyFieldName, v)
}

func (this Currency) GetSymbol() *string {
	return this.GetFieldData().GetString(CurrencyFieldSymbol)
}

func (this *Currency) SetSymbol(v *string) {
	this.GetFieldData().SetString(CurrencyFieldSymbol, v)
}

func (this Currency) GetDecimalPlaces() *int32 {
	return this.GetFieldData().GetInt32(CurrencyFieldDecimalPlaces)
}

func (this *Currency) SetDecimalPlaces(v *int32) {
	this.GetFieldData().SetInt32(CurrencyFieldDecimalPlaces, v)
}

func (this Currency) GetIsActive() *bool {
	return this.GetFieldData().GetBool(CurrencyFieldIsActive)
}

func (this *Currency) SetIsActive(v *bool) {
	this.GetFieldData().SetBool(CurrencyFieldIsActive, v)
}
