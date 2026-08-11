package services

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
)

const (
	priceTemplateId = "01TEMPLATE"
	priceVariantId  = "01VARIANT"
)

var priceToday = time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)

// BR §6.12 rule 2: a variant rule overrides the template's base rule.
func TestVariantPriceOverridesTemplatePrice(t *testing.T) {
	prices := []*models.ProductPrice{
		approvedTemplatePrice("10000", nil, nil),
		approvedVariantPrice("12500", nil, nil),
	}

	selected := SelectApplicablePrice(prices, priceVariantId, priceTemplateId, priceToday)

	require.NotNil(t, selected)
	assert.True(t, selected.Equal(decimal.RequireFromString("12500")),
		"the variant's own price must win over the template's")
}

// The template price is the base rule, used whenever the variant has none of its own.
func TestTemplatePriceAppliesWhenVariantHasNone(t *testing.T) {
	prices := []*models.ProductPrice{approvedTemplatePrice("10000", nil, nil)}

	selected := SelectApplicablePrice(prices, priceVariantId, priceTemplateId, priceToday)

	require.NotNil(t, selected)
	assert.True(t, selected.Equal(decimal.RequireFromString("10000")))
}

// A variant rule only wins while it applies. Once it expires the template's base price takes over
// again, rather than the product losing its price.
func TestExpiredVariantPriceFallsBackToTemplate(t *testing.T) {
	prices := []*models.ProductPrice{
		approvedTemplatePrice("10000", nil, nil),
		approvedVariantPrice("9000", date(2026, 1, 1), date(2026, 6, 30)),
	}

	selected := SelectApplicablePrice(prices, priceVariantId, priceTemplateId, priceToday)

	require.NotNil(t, selected)
	assert.True(t, selected.Equal(decimal.RequireFromString("10000")))
}

// This is the whole reason the result is nullable. A product with no applicable rule has no
// price, and the caller must be able to tell that from a price of zero — seeding a sale price
// from a missing one is how a product ends up being sold for nothing.
func TestNoApplicableRuleYieldsNilNotZero(t *testing.T) {
	testCases := []struct {
		name  string
		price *models.ProductPrice
	}{
		{"no rows at all", nil},
		{"draft is not yet in force", draftTemplatePrice("10000")},
		{"archived rule", archivedTemplatePrice("10000")},
		{"not yet effective", approvedTemplatePrice("10000", date(2099, 1, 1), nil)},
		{"already expired", approvedTemplatePrice("10000", nil, date(2020, 1, 1))},
		{"targets another product", approvedPriceFor("77000", "01OTHERVARIANT", "")},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var prices []*models.ProductPrice
			if testCase.price != nil {
				prices = []*models.ProductPrice{testCase.price}
			}

			assert.Nil(t, SelectApplicablePrice(prices, priceVariantId, priceTemplateId, priceToday),
				"an inapplicable rule must leave the product unpriced, never priced at zero")
		})
	}
}

// A scheduled price change is a second row, not an edit to the first, so the later one has to win
// once it comes into effect.
func TestLatestEffectiveFromWinsAtTheSameLevel(t *testing.T) {
	prices := []*models.ProductPrice{
		approvedTemplatePrice("10000", date(2026, 1, 1), nil),
		approvedTemplatePrice("11000", date(2026, 7, 1), nil),
		approvedTemplatePrice("12000", date(2099, 1, 1), nil), // not in force yet
	}

	selected := SelectApplicablePrice(prices, priceVariantId, priceTemplateId, priceToday)

	require.NotNil(t, selected)
	assert.True(t, selected.Equal(decimal.RequireFromString("11000")),
		"the most recently effective rule in force must win")
}

// A row with no start date is the standing price; a dated row is a deliberate later decision and
// supersedes it.
func TestDatedRuleBeatsTheStandingPrice(t *testing.T) {
	prices := []*models.ProductPrice{
		approvedTemplatePrice("10000", nil, nil),
		approvedTemplatePrice("8000", date(2026, 7, 1), nil),
	}

	selected := SelectApplicablePrice(prices, priceVariantId, priceTemplateId, priceToday)

	require.NotNil(t, selected)
	assert.True(t, selected.Equal(decimal.RequireFromString("8000")))
}

// effective_to names the last day the price applies, so it must still be in force throughout that
// day. Comparing against midnight would expire every price a day early.
func TestPriceAppliesThroughTheWholeOfItsLastDay(t *testing.T) {
	prices := []*models.ProductPrice{
		approvedTemplatePrice("10000", nil, date(2026, 8, 10)),
	}

	selected := SelectApplicablePrice(prices, priceVariantId, priceTemplateId, priceToday)

	require.NotNil(t, selected, "a price ending today is still in force today")
	assert.True(t, selected.Equal(decimal.RequireFromString("10000")))
}

func newPriceRow(amount string, status models.ProductPriceStatus) *models.ProductPrice {
	price := models.NewProductPrice()
	value := decimal.RequireFromString(amount)
	price.SetPrice(&value)
	price.SetStatus(&status)
	return price
}

func approvedPriceFor(amount string, variantId string, templateId string) *models.ProductPrice {
	price := newPriceRow(amount, models.ProductPriceStatusApproved)
	if variantId != "" {
		price.SetProductVariantId(strPtr(variantId))
	}
	if templateId != "" {
		price.SetProductTemplateId(strPtr(templateId))
	}
	return price
}

func approvedTemplatePrice(amount string, from *model.ModelDate, to *model.ModelDate) *models.ProductPrice {
	price := approvedPriceFor(amount, "", priceTemplateId)
	price.SetEffectiveFrom(from)
	price.SetEffectiveTo(to)
	return price
}

func approvedVariantPrice(amount string, from *model.ModelDate, to *model.ModelDate) *models.ProductPrice {
	price := approvedPriceFor(amount, priceVariantId, "")
	price.SetEffectiveFrom(from)
	price.SetEffectiveTo(to)
	return price
}

func draftTemplatePrice(amount string) *models.ProductPrice {
	price := newPriceRow(amount, models.ProductPriceStatusDraft)
	price.SetProductTemplateId(strPtr(priceTemplateId))
	return price
}

func archivedTemplatePrice(amount string) *models.ProductPrice {
	price := approvedPriceFor(amount, "", priceTemplateId)
	archived := true
	price.SetIsArchived(&archived)
	return price
}

func date(year int, month time.Month, day int) *model.ModelDate {
	value := model.ModelDate(time.Date(year, month, day, 0, 0, 0, 0, time.UTC))
	return &value
}
