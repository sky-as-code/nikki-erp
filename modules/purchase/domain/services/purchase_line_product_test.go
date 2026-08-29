package services

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	itUom "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/uom"
	itExt "github.com/sky-as-code/nikki-erp/modules/purchase/interfaces/external"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// Stubs for the two ports, hand-written so each test states only what it is about.

type stubProducts struct {
	found        bool
	purchasable  bool
	archived     bool
	inventoryUom string
}

func (this *stubProducts) GetPurchasableProduct(
	_ corectx.Context, query itExt.GetPurchasableProductQuery,
) (*itExt.GetPurchasableProductResult, error) {
	if !this.found {
		return &itExt.GetPurchasableProductResult{}, nil
	}
	return &itExt.GetPurchasableProductResult{
		HasData: true,
		Data: itExt.GetPurchasableProductResultData{
			VariantId:      query.VariantId,
			TemplateId:     "01TEMPLATE",
			Purchasable:    this.purchasable,
			Archived:       this.archived,
			InventoryUomId: model.Id(this.inventoryUom),
		},
	}, nil
}

type stubUoms struct {
	found      bool
	archived   bool
	converted  string
	convertErr bool
}

func (this *stubUoms) GetUom(_ corectx.Context, query itExt.GetUomQuery) (*itExt.GetUomResult, error) {
	if !this.found {
		return &itExt.GetUomResult{}, nil
	}
	return &itExt.GetUomResult{
		HasData: true,
		Data: itUom.GetUomResultData{
			Id:         query.Id,
			IsArchived: this.archived,
		},
	}, nil
}

func (this *stubUoms) Convert(
	_ corectx.Context, _ itExt.ConvertQuantityQuery,
) (*itExt.ConvertQuantityResult, error) {
	if this.convertErr {
		// Essential reports incompatible unit categories as a client error, not a Go error.
		vErrs := ft.NewClientErrors()
		vErrs.Append(*ft.NewBusinessViolation("uom_id", "uom.different_category",
			"these units belong to different categories"))
		return &itExt.ConvertQuantityResult{ClientErrors: *vErrs}, nil
	}
	return &itExt.ConvertQuantityResult{
		HasData: true,
		Data:    itUom.ConvertQuantityResultData{Quantity: dec(this.converted)},
	}, nil
}

func usableProduct(inventoryUom string) *stubProducts {
	return &stubProducts{found: true, purchasable: true, inventoryUom: inventoryUom}
}

func usableUom(converted string) *stubUoms {
	return &stubUoms{found: true, converted: converted}
}

func productLineFor(variantId, uomId, quantity string) dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.PurchaseOrderLineFieldLineType:         string(models.PurchaseOrderLineTypeProduct),
		models.PurchaseOrderLineFieldProductVariantId: variantId,
		models.PurchaseOrderLineFieldUomId:            uomId,
		models.PurchaseOrderLineFieldQuantity:         dec(quantity),
		models.PurchaseOrderLineFieldUnitPrice:        dec("10"),
		models.PurchaseOrderLineFieldDiscountPercent:  dec("0"),
		models.PurchaseOrderLineFieldTaxAmount:        dec("0"),
	}
}

// The central rule: the line keeps what the buyer typed. "10 boxes" and "120 units" are the same
// goods but not the same request.
func TestTheOrderedQuantityAndUnitAreNeverOverwritten(t *testing.T) {
	validator := NewProductLineValidator(usableProduct("01UOM_UNIT"), usableUom("120"))
	line := productLineFor("01VARIANT", "01UOM_BOX", "10")
	vErrs := ft.NewClientErrors()

	require.NoError(t, prepared(validator)(nil, line, vErrs))

	require.Equal(t, 0, vErrs.Count())
	assert.True(t, decimalOf(line, models.PurchaseOrderLineFieldQuantity).Equal(dec("10")),
		"the ordered quantity must survive the conversion")
	assert.Equal(t, "01UOM_BOX", line[models.PurchaseOrderLineFieldUomId],
		"the ordered unit must survive the conversion")
	// The converted value lands here and nowhere else.
	assert.True(t, decimalOf(line, models.PurchaseOrderLineFieldInventoryQuantity).Equal(dec("120")))
}

// A line already in the product's inventory unit needs no conversion, so the port is not called.
func TestNoConversionWhenTheUnitsAlreadyMatch(t *testing.T) {
	uoms := usableUom("999999") // would be obviously wrong if it were consulted
	validator := NewProductLineValidator(usableProduct("01UOM_UNIT"), uoms)
	line := productLineFor("01VARIANT", "01UOM_UNIT", "7")
	vErrs := ft.NewClientErrors()

	require.NoError(t, prepared(validator)(nil, line, vErrs))

	require.Equal(t, 0, vErrs.Count())
	assert.True(t, decimalOf(line, models.PurchaseOrderLineFieldInventoryQuantity).Equal(dec("7")))
}

// Ordering in litres a product counted in kilograms is not a conversion anyone can do, and storing
// the raw number would put a mass in a volume column.
func TestACrossCategoryUnitIsRefused(t *testing.T) {
	uoms := &stubUoms{found: true, convertErr: true}
	validator := NewProductLineValidator(usableProduct("01UOM_KG"), uoms)
	line := productLineFor("01VARIANT", "01UOM_LITRE", "5")
	vErrs := ft.NewClientErrors()

	require.NoError(t, prepared(validator)(nil, line, vErrs))

	require.Equal(t, 1, vErrs.Count())
	// Essential's own reason is carried through, so the caller sees which units disagreed.
	assert.Equal(t, "uom.different_category", (*vErrs)[0].Key)
}

// The three product refusals are deliberately distinct: a bad id, a product the business does not
// buy, and one it used to buy. Collapsing them would leave a buyer guessing which they hit.
func TestProductRefusalsAreDistinct(t *testing.T) {
	testCases := []struct {
		name    string
		product *stubProducts
		wantKey string
	}{
		{
			"no such product", &stubProducts{found: false},
			"purchase_order_line.product_not_found",
		},
		{
			"a real product the business does not buy",
			&stubProducts{found: true, purchasable: false},
			"purchase_order_line.product_not_purchasable",
		},
		{
			"a product it used to buy",
			&stubProducts{found: true, purchasable: true, archived: true},
			"purchase_order_line.product_archived",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			validator := NewProductLineValidator(testCase.product, usableUom("1"))
			vErrs := ft.NewClientErrors()

			require.NoError(t, prepared(validator)(
				nil, productLineFor("01VARIANT", "01UOM", "1"), vErrs))

			require.Equal(t, 1, vErrs.Count())
			assert.Equal(t, testCase.wantKey, (*vErrs)[0].Key)
		})
	}
}

// An archived unit is still resolvable so old lines read, but may not appear on something new.
func TestAnArchivedUnitIsRefusedOnANewLine(t *testing.T) {
	uoms := &stubUoms{found: true, archived: true, converted: "1"}
	validator := NewProductLineValidator(usableProduct("01UOM_UNIT"), uoms)
	vErrs := ft.NewClientErrors()

	require.NoError(t, prepared(validator)(nil, productLineFor("01VARIANT", "01UOM_OLD", "1"), vErrs))

	require.Equal(t, 1, vErrs.Count())
	assert.Equal(t, "purchase_order_line.uom_archived", (*vErrs)[0].Key)
}

// An unknown unit is refused outright, or a typo would produce a quantity expressed in nothing.
func TestAnUnknownUnitIsRefused(t *testing.T) {
	validator := NewProductLineValidator(usableProduct("01UOM_UNIT"), &stubUoms{found: false})
	vErrs := ft.NewClientErrors()

	require.NoError(t, prepared(validator)(nil, productLineFor("01VARIANT", "01NOPE", "1"), vErrs))

	require.Equal(t, 1, vErrs.Count())
	assert.Equal(t, "purchase_order_line.uom_not_found", (*vErrs)[0].Key)
}

// Ordinary rather than erroneous states: each would be a false refusal if the validator insisted on
// a complete product-and-unit pair.
func TestTheLegitimateAbsences(t *testing.T) {
	t.Run("a section buys nothing and needs no product", func(t *testing.T) {
		validator := NewProductLineValidator(&stubProducts{found: false}, &stubUoms{found: false})
		line := dmodel.DynamicFields{
			models.PurchaseOrderLineFieldLineType: string(models.PurchaseOrderLineTypeSection),
		}
		vErrs := ft.NewClientErrors()

		require.NoError(t, prepared(validator)(nil, line, vErrs))

		assert.Equal(t, 0, vErrs.Count())
		assert.True(t, decimalOf(line, models.PurchaseOrderLineFieldInventoryQuantity).IsZero())
	})

	t.Run("a freight charge is a priced line with no product", func(t *testing.T) {
		validator := NewProductLineValidator(&stubProducts{found: false}, &stubUoms{found: false})
		line := productLineFor("", "", "3")
		vErrs := ft.NewClientErrors()

		require.NoError(t, prepared(validator)(nil, line, vErrs))

		assert.Equal(t, 0, vErrs.Count())
		assert.True(t, decimalOf(line, models.PurchaseOrderLineFieldInventoryQuantity).Equal(dec("3")))
	})

	t.Run("a service has no stock configuration", func(t *testing.T) {
		// InventoryUomId empty: the product exists and is purchasable but nothing counts its stock.
		validator := NewProductLineValidator(usableProduct(""), usableUom("999"))
		line := productLineFor("01VARIANT", "01UOM_HOUR", "8")
		vErrs := ft.NewClientErrors()

		require.NoError(t, prepared(validator)(nil, line, vErrs))

		assert.Equal(t, 0, vErrs.Count())
		assert.True(t, decimalOf(line, models.PurchaseOrderLineFieldInventoryQuantity).Equal(dec("8")),
			"with no inventory unit to convert to, the ordered quantity stands")
	})
}

// inventory_quantity is required_for_create, so every path through the validator must leave one, or
// the schema refuses with a message about a missing field rather than the real problem.
func TestEveryAcceptedPathFillsInventoryQuantity(t *testing.T) {
	for _, testCase := range []struct {
		name string
		line dmodel.DynamicFields
	}{
		{"note", dmodel.DynamicFields{
			models.PurchaseOrderLineFieldLineType: string(models.PurchaseOrderLineTypeNote)}},
		{"freight", productLineFor("", "", "2")},
		{"product", productLineFor("01VARIANT", "01UOM_BOX", "4")},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			validator := NewProductLineValidator(usableProduct("01UOM_UNIT"), usableUom("48"))
			vErrs := ft.NewClientErrors()

			require.NoError(t, prepared(validator)(nil, testCase.line, vErrs))
			require.Equal(t, 0, vErrs.Count())

			value, present := testCase.line[models.PurchaseOrderLineFieldInventoryQuantity]
			require.True(t, present, "inventory_quantity is required_for_create")
			_, isDecimal := value.(decimal.Decimal)
			assert.True(t, isDecimal, "inventory_quantity must be a decimal, got %T", value)
		})
	}
}

// prepared adapts PrepareLine's two return values to the single one require.NoError takes. The
// template id is asserted by its own test below rather than at every call site.
func prepared(validator *ProductLineValidator) func(
	corectx.Context, dmodel.DynamicFields, *ft.ClientErrors) error {
	return func(ctx corectx.Context, line dmodel.DynamicFields, vErrs *ft.ClientErrors) error {
		_, err := validator.PrepareLine(ctx, line, vErrs)
		return err
	}
}

// Pricing needs the template id, and PrepareLine already has it because Inventory answered for it
// while checking purchasability. Returning it saves a second cross-module read inside the write
// transaction.
func TestPrepareLineReturnsTheTemplateIdForPricing(t *testing.T) {
	validator := NewProductLineValidator(usableProduct("01UOM_UNIT"), usableUom("1"))
	vErrs := ft.NewClientErrors()

	templateId, err := validator.PrepareLine(
		nil, productLineFor("01VARIANT", "01UOM_UNIT", "1"), vErrs)

	require.NoError(t, err)
	require.Equal(t, 0, vErrs.Count())
	assert.Equal(t, "01TEMPLATE", templateId,
		"the variant's template is what a vendor price is recorded against")
}

// A line with nothing to price against reports no template, so the caller skips resolution rather
// than searching for quotes on an empty id.
func TestPrepareLineReportsNoTemplateForALineWithNoProduct(t *testing.T) {
	validator := NewProductLineValidator(usableProduct("01UOM_UNIT"), usableUom("1"))

	freeText := productLineFor("", "01UOM_UNIT", "1")
	templateId, err := validator.PrepareLine(nil, freeText, ft.NewClientErrors())
	require.NoError(t, err)
	assert.Empty(t, templateId, "a free-text charge has no product to price")

	section := dmodel.DynamicFields{
		models.PurchaseOrderLineFieldLineType: string(models.PurchaseOrderLineTypeSection),
	}
	templateId, err = validator.PrepareLine(nil, section, ft.NewClientErrors())
	require.NoError(t, err)
	assert.Empty(t, templateId, "a section line buys nothing")
}
