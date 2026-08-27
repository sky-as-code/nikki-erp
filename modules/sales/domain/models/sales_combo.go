package models

import (
	_ "embed"

	"github.com/shopspring/decimal"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/model"

	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

const (
	SalesComboSchemaName = "sales_combo"

	SalesComboFieldId           = "id"
	SalesComboFieldOrgId        = "org_id"
	SalesComboFieldCode         = "code"
	SalesComboFieldName         = "name"
	SalesComboFieldDescription  = "description"
	SalesComboFieldComboPrice   = "combo_price"
	SalesComboFieldValidFrom    = "valid_from"
	SalesComboFieldValidUntil   = "valid_until"
	SalesComboFieldReturnPolicy = "return_policy"
)

const (
	SalesComboComponentSchemaName = "sales_combo_component"

	SalesComboComponentFieldId               = "id"
	SalesComboComponentFieldOrgId            = "org_id"
	SalesComboComponentFieldSalesComboId     = "sales_combo_id"
	SalesComboComponentFieldProductVariantId = "product_variant_id"
	SalesComboComponentFieldQuantity         = "quantity"
	SalesComboComponentFieldUomId            = "uom_id"
	SalesComboComponentFieldIsRequired       = "is_required"
	SalesComboComponentFieldSelectionGroup   = "selection_group"

	SalesComboComponentEdgeSalesCombo = "sales_combo"
)

// ComboReturnPolicy is whether a customer may return part of a bundle (BR §19).
type ComboReturnPolicy string

const (
	// ComboReturnPolicyEntireOnly requires the whole bundle back. The default, and the restrictive
	// reading: a partial return of a bundle priced below its parts is a discount the customer keeps
	// on the items they did not bring back, so the business opts into that deliberately.
	ComboReturnPolicyEntireOnly = ComboReturnPolicy("entire_combo_only")

	// ComboReturnPolicyComponentAllowed permits returning individual components, refunded at their
	// allocated share rather than at their standalone price.
	ComboReturnPolicyComponentAllowed = ComboReturnPolicy("component_return_allowed")
)

//go:embed sales_combo.json
var salesComboSchemaJson string

func SalesComboSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesComboSchemaJson)
}

//go:embed sales_combo_component.json
var salesComboComponentSchemaJson string

func SalesComboComponentSchemaBuilder() *dmodel.ModelSchemaBuilder {
	return dmodel.ParseModelJson(salesComboComponentSchemaJson)
}

// SalesCombo is a bundle sold at a price of its own.
//
// combo_price is INDEPENDENT and never derived from the components (BR §15). That is the defining
// property: the business decides three items together are worth 48,000 whatever they cost apart, and
// a derived price would make the bundle drift every time a component was repriced. The per-component
// allocation is an output — for VAT, partial return and reporting — never an input.
type SalesCombo struct {
	basemodel.DynamicModelBase
}

func NewSalesCombo() *SalesCombo {
	return &SalesCombo{basemodel.NewDynamicModel()}
}

func NewSalesComboFrom(src dmodel.DynamicFields) *SalesCombo {
	return &SalesCombo{basemodel.NewDynamicModel(src)}
}

func (this SalesCombo) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this SalesCombo) GetCode() *string {
	return this.GetFieldData().GetString(SalesComboFieldCode)
}

func (this SalesCombo) GetName() *string {
	return this.GetFieldData().GetString(SalesComboFieldName)
}

func (this SalesCombo) GetComboPrice() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesComboFieldComboPrice)
}

func (this SalesCombo) GetValidFrom() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesComboFieldValidFrom)
}

func (this SalesCombo) GetValidUntil() *model.ModelDateTime {
	return this.GetFieldData().GetModelDateTime(SalesComboFieldValidUntil)
}

func (this SalesCombo) GetReturnPolicy() *string {
	return this.GetFieldData().GetString(SalesComboFieldReturnPolicy)
}

func (this SalesCombo) GetIsArchived() *bool {
	return this.GetFieldData().GetBool(basemodel.FieldIsArchived)
}

// AllowsComponentReturn reports whether part of this bundle may be returned on its own.
//
// An absent or unrecognised policy reads as the restrictive answer. A bundle whose policy could not
// be determined must not be the one where a customer keeps the bundle discount on the half they
// kept.
func (this SalesCombo) AllowsComponentReturn() bool {
	policy := this.GetReturnPolicy()
	return policy != nil && ComboReturnPolicy(*policy) == ComboReturnPolicyComponentAllowed
}

// SalesComboComponent is one product inside a bundle definition.
//
// Not to be confused with SalesOrderLineComponent: this is the DEFINITION of what a bundle contains,
// while that is what one particular sale of it actually included. The two are separate because a
// bundle can be redefined after a sale, and the sale must keep meaning what it meant.
type SalesComboComponent struct {
	basemodel.DynamicModelBase
}

func NewSalesComboComponent() *SalesComboComponent {
	return &SalesComboComponent{basemodel.NewDynamicModel()}
}

func NewSalesComboComponentFrom(src dmodel.DynamicFields) *SalesComboComponent {
	return &SalesComboComponent{basemodel.NewDynamicModel(src)}
}

func (this SalesComboComponent) GetId() *model.Id {
	return this.GetFieldData().GetModelId(basemodel.FieldId)
}

func (this SalesComboComponent) GetSalesComboId() *model.Id {
	return this.GetFieldData().GetModelId(SalesComboComponentFieldSalesComboId)
}

func (this SalesComboComponent) GetProductVariantId() *model.Id {
	return this.GetFieldData().GetModelId(SalesComboComponentFieldProductVariantId)
}

func (this SalesComboComponent) GetQuantity() *decimal.Decimal {
	return this.GetFieldData().GetDecimal(SalesComboComponentFieldQuantity)
}

func (this SalesComboComponent) GetUomId() *model.Id {
	return this.GetFieldData().GetModelId(SalesComboComponentFieldUomId)
}

func (this SalesComboComponent) GetIsRequired() *bool {
	return this.GetFieldData().GetBool(SalesComboComponentFieldIsRequired)
}

func (this SalesComboComponent) GetSelectionGroup() *string {
	return this.GetFieldData().GetString(SalesComboComponentFieldSelectionGroup)
}

// IsOptional reports whether the customer chooses whether to include this component.
//
// A nil is_required reads as required, matching the schema default: a component whose requiredness
// could not be determined should be included rather than silently dropped from what the customer
// paid for.
func (this SalesComboComponent) IsOptional() bool {
	required := this.GetIsRequired()
	return required != nil && !*required
}
