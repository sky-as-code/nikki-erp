package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	PaymentMethodSchemaName = "paymentinvoice_payment_method"

	PaymentMethodFieldId          = basemodel.FieldId
	PaymentMethodFieldCode        = "code"
	PaymentMethodFieldAdapterCode = "adapter_code"
	PaymentMethodFieldName        = "name"
	PaymentMethodFieldDescription = "description"
	PaymentMethodFieldCurrencyId  = "currency_id"
	PaymentMethodFieldMinAmount   = "min_amount"
	PaymentMethodFieldMaxAmount   = "max_amount"
	PaymentMethodFieldIsActive    = "is_active"
	PaymentMethodFieldConfig      = "config"
)

// The adapter codes this module ships an implementation for. A payment method row may name only
// one of these: an adapter_code with no adapter behind it is a method that accepts orders and can
// never collect them.
//
// These are adapter identifiers, not the payment methods themselves. The methods are rows, so a
// deployment adds one without a release; an adapter is code, so it cannot be.
const (
	AdapterCodeMomo   = "momo"
	AdapterCodeVietQr = "vietqr"
	AdapterCodeMpos   = "mpos"
)

//go:embed payment_method.json
var paymentMethodSchemaJson string

func PaymentMethodSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(paymentMethodSchemaJson)
}

// PaymentMethod is a way the business can be paid, as configured data rather than as code.
//
// The service it supersedes held this as a Go enum, which meant that offering an existing gateway
// under a second merchant account, or withdrawing one at short notice, was a code change and a
// deploy. Here it is a row: what remains in code is the adapter that knows a gateway's wire
// protocol, selected by adapter_code.
type PaymentMethod struct {
	basemodel.DynamicModelBase
}

func NewPaymentMethod() *PaymentMethod {
	return &PaymentMethod{basemodel.NewDynamicModel()}
}

func NewPaymentMethodFrom(src dmodel.DynamicFields) *PaymentMethod {
	return &PaymentMethod{basemodel.NewDynamicModel(src)}
}

func (this PaymentMethod) GetCode() *string {
	return this.GetFieldData().GetString(PaymentMethodFieldCode)
}

func (this *PaymentMethod) SetCode(v *string) {
	this.GetFieldData().SetString(PaymentMethodFieldCode, v)
}

func (this PaymentMethod) GetAdapterCode() *string {
	return this.GetFieldData().GetString(PaymentMethodFieldAdapterCode)
}

func (this *PaymentMethod) SetAdapterCode(v *string) {
	this.GetFieldData().SetString(PaymentMethodFieldAdapterCode, v)
}

func (this PaymentMethod) GetName() *model.LangJson {
	return this.GetFieldData().GetLangJson(PaymentMethodFieldName)
}

func (this *PaymentMethod) SetName(v *model.LangJson) {
	this.GetFieldData().SetLangJson(PaymentMethodFieldName, v)
}

func (this PaymentMethod) GetDescription() *model.LangJson {
	return this.GetFieldData().GetLangJson(PaymentMethodFieldDescription)
}

func (this *PaymentMethod) SetDescription(v *model.LangJson) {
	this.GetFieldData().SetLangJson(PaymentMethodFieldDescription, v)
}

func (this PaymentMethod) GetCurrencyId() *model.Id {
	return this.GetFieldData().GetModelId(PaymentMethodFieldCurrencyId)
}

func (this *PaymentMethod) SetCurrencyId(v *model.Id) {
	this.GetFieldData().SetModelId(PaymentMethodFieldCurrencyId, v)
}

func (this PaymentMethod) GetMinAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(PaymentMethodFieldMinAmount)
}

func (this *PaymentMethod) SetMinAmount(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(PaymentMethodFieldMinAmount, v)
}

func (this PaymentMethod) GetMaxAmount() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(PaymentMethodFieldMaxAmount)
}

func (this *PaymentMethod) SetMaxAmount(v *decimal.Decimal) {
	this.GetFieldData().SetDecimal(PaymentMethodFieldMaxAmount, v)
}

func (this PaymentMethod) GetIsActive() *bool {
	return this.GetFieldData().GetBool(PaymentMethodFieldIsActive)
}

func (this *PaymentMethod) SetIsActive(v *bool) {
	this.GetFieldData().SetBool(PaymentMethodFieldIsActive, v)
}

// GetConfig returns the adapter's non-secret per-method settings. It is reached through GetAny
// because a jsonmap has no typed accessor; only the owning adapter interprets what is in it.
func (this PaymentMethod) GetConfig() map[string]any {
	raw := this.GetFieldData().GetAny(PaymentMethodFieldConfig)
	if raw == nil {
		return nil
	}
	config, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	return config
}

func (this *PaymentMethod) SetConfig(v map[string]any) {
	this.GetFieldData().SetAny(PaymentMethodFieldConfig, v)
}
