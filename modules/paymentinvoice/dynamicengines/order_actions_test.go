package dynamicengines

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// The actions declare no ParamSchema — their params are a mix of order fields and method-specific
// input that no single schema describes — so the request body is checked here instead. These tests
// cover what that check has to get right.

// An amount that cannot be read must be refused, never defaulted. Falling back to zero would take
// a malformed request and turn it into a free order that the gateway happily accepts.
func TestAMalformedAmountIsRefusedNotDefaultedToZero(t *testing.T) {
	for name, value := range map[string]any{
		"absent":       nil,
		"empty":        "",
		"not_a_number": "abc",
		"a_list":       []any{1},
		"an_object":    map[string]any{"amount": 1},
		"a_bool":       true,
	} {
		params := dmodel.DynamicFields{paramPaymentMethodId: "01METHOD0000000000000000000"}
		if value != nil {
			params[paramAmount] = value
		}

		_, vErrs := buildCreatePaymentCommand(params)

		assert.Equal(t, 1, vErrs.Count(), name)
	}
}

// An exact amount survives JSON as a string, because float64 cannot hold every decimal. Parsing
// it as a float and rounding would collect something other than what was asked for.
func TestAnAmountKeepsItsExactValueThroughAString(t *testing.T) {
	cmd, vErrs := buildCreatePaymentCommand(dmodel.DynamicFields{
		paramPaymentMethodId: "01METHOD0000000000000000000",
		paramAmount:          "150000.75",
	})

	assert.Zero(t, vErrs.Count())
	assert.Equal(t, "150000.75", cmd.Amount.String())
}

func TestTheNumericJsonShapesAreAccepted(t *testing.T) {
	for name, value := range map[string]any{
		"float64": float64(150000),
		"int":     150000,
		"int64":   int64(150000),
		"decimal": decimal.RequireFromString("150000"),
	} {
		cmd, vErrs := buildCreatePaymentCommand(dmodel.DynamicFields{
			paramPaymentMethodId: "01METHOD0000000000000000000",
			paramAmount:          value,
		})

		assert.Zero(t, vErrs.Count(), name)
		assert.Equal(t, "150000", cmd.Amount.String(), name)
	}
}

func TestAPaymentMethodIsRequired(t *testing.T) {
	_, vErrs := buildCreatePaymentCommand(dmodel.DynamicFields{paramAmount: "150000"})

	assert.Equal(t, 1, vErrs.Count())
}

// pos_id is accepted at the top level as well as inside metadata, because that is where the old
// service's callers put it and the vending machines still do. Either spelling has to reach the
// adapter, or every card-terminal payment from an un-updated caller fails validation.
func TestATopLevelPosIdIsFoldedIntoTheMetadata(t *testing.T) {
	cmd, vErrs := buildCreatePaymentCommand(dmodel.DynamicFields{
		paramPaymentMethodId: "01METHOD0000000000000000000",
		paramAmount:          "150000",
		paramPosId:           "POS-01",
	})

	assert.Zero(t, vErrs.Count())
	assert.Equal(t, "POS-01", cmd.Metadata[models.OrderMetaPosId])
}

func TestAMetadataPosIdIsPassedThrough(t *testing.T) {
	cmd, vErrs := buildCreatePaymentCommand(dmodel.DynamicFields{
		paramPaymentMethodId: "01METHOD0000000000000000000",
		paramAmount:          "150000",
		paramMetadata:        map[string]any{models.OrderMetaPosId: "POS-02"},
	})

	assert.Zero(t, vErrs.Count())
	assert.Equal(t, "POS-02", cmd.Metadata[models.OrderMetaPosId])
}

// Keys the adapter declares beyond pos_id must survive: the metadata map exists precisely so an
// adapter can require input this layer knows nothing about.
func TestUnknownMetadataKeysAreNotDropped(t *testing.T) {
	cmd, vErrs := buildCreatePaymentCommand(dmodel.DynamicFields{
		paramPaymentMethodId: "01METHOD0000000000000000000",
		paramAmount:          "150000",
		paramMetadata:        map[string]any{"terminal_group": "G1", "till": 4},
	})

	assert.Zero(t, vErrs.Count())
	assert.Equal(t, "G1", cmd.Metadata["terminal_group"])
	assert.Equal(t, 4, cmd.Metadata["till"])
}

// An order with no method-specific input stores no metadata at all, rather than an empty map that
// would read as "this gateway required something and got nothing".
func TestNoMetadataMeansNilNotAnEmptyMap(t *testing.T) {
	cmd, _ := buildCreatePaymentCommand(dmodel.DynamicFields{
		paramPaymentMethodId: "01METHOD0000000000000000000",
		paramAmount:          "150000",
	})

	assert.Nil(t, cmd.Metadata)
}

func TestRefundNeedsAnOrderAndAnAmount(t *testing.T) {
	_, vErrs := buildRefundCommand(dmodel.DynamicFields{})

	assert.Equal(t, 2, vErrs.Count())
}

func TestARefundCommandIsBuiltFromTheBusinessOrderId(t *testing.T) {
	cmd, vErrs := buildRefundCommand(dmodel.DynamicFields{
		paramOrderId: "VDMCMOM0Q8HABCDEFGH",
		paramAmount:  "50000",
	})

	assert.Zero(t, vErrs.Count())
	assert.Equal(t, "VDMCMOM0Q8HABCDEFGH", cmd.OrderId)
	assert.Equal(t, "50000", cmd.Amount.String())
}
