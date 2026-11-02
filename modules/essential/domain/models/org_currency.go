package models

import (
	_ "embed"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	OrgCurrencySchemaName = "essential_org_currency"

	OrgCurrencyFieldOrgId      = basemodel.FieldOrgId
	OrgCurrencyFieldCurrencyId = "currency_id"
	OrgCurrencyFieldIsActive   = "is_active"
)

//go:embed org_currency.json
var orgCurrencySchemaJson string

func OrgCurrencySchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(orgCurrencySchemaJson)
}

// OrgCurrency is one organization's activation of a tenant-wide currency. The catalogue entry is shared by
// every organization of the tenant; whether an organization uses it is this row, so one
// organization switching a currency off leaves the others untouched.
type OrgCurrency struct {
	basemodel.DynamicModelBase
}

func NewOrgCurrencyFrom(src dmodel.DynamicFields) *OrgCurrency {
	return &OrgCurrency{basemodel.NewDynamicModel(src)}
}
