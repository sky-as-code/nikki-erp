package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/commerce/inventory/domain/models"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/usagecheck"
)

// Deleting a stock quant asks every dependant in Inventory's Dependants(). None of them holds a
// quant id today, so each answers "unused" through a table-less checker; these tests pin that the
// delete then goes through, and that one dependant failing to answer still refuses it.

func quantParams() dmodel.DynamicFields {
	return dmodel.DynamicFields{
		models.StockQuantFieldId: "QUANT-1",
		"org_id":                 "ORG-1",
	}
}

func TestAssertQuantDeletablePermitsWhenEveryDependantAnswersUnused(t *testing.T) {
	vErrs := ft.NewClientErrors()

	err := assertQuantDeletable(
		corectx.NewRequestContext(context.Background()),
		variantDispatcher(t,
			variantUsageReply{module: "sales", used: false},
			variantUsageReply{module: "purchase", used: false},
			variantUsageReply{module: "vendingmachine", used: false}),
		quantParams(), vErrs)

	require.NoError(t, err)
	assert.Zero(t, vErrs.Count())
}

// The regression this guards: a dependant with no checker for stock_quant answers with an error,
// which refuses every quant delete.
func TestAssertQuantDeletableRefusesWhenADependantHasNoChecker(t *testing.T) {
	vErrs := ft.NewClientErrors()

	err := assertQuantDeletable(
		corectx.NewRequestContext(context.Background()),
		variantDispatcher(t,
			variantUsageReply{module: "sales", used: false},
			variantUsageReply{module: "purchase", failure: "no checker for inventory.stock_quant"},
			variantUsageReply{module: "vendingmachine", used: false}),
		quantParams(), vErrs)

	require.NoError(t, err)
	require.Equal(t, 1, vErrs.Count())
	assert.Equal(t, usagecheck.ErrorKeyResourceUsedByModules, (*vErrs)[0].Key)
}
