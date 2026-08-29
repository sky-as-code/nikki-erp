package dynamicengines

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// testOrgId stands in for the organization every order belongs to. It is required on every create,
// so the fixtures below carry it and each test isolates the one rule it is about.
const testOrgId = "01ORG00000000000000000000"

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
		params := dmodel.DynamicFields{paramPaymentMethodId: "01METHOD0000000000000000000", paramOrgId: testOrgId}
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
		paramOrgId:           testOrgId,
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
			paramOrgId:           testOrgId,
			paramAmount:          value,
		})

		assert.Zero(t, vErrs.Count(), name)
		assert.Equal(t, "150000", cmd.Amount.String(), name)
	}
}

func TestAPaymentMethodIsRequired(t *testing.T) {
	_, vErrs := buildCreatePaymentCommand(dmodel.DynamicFields{paramAmount: "150000", paramOrgId: testOrgId})

	assert.Equal(t, 1, vErrs.Count())
}

// pos_id is accepted at the top level as well as inside metadata, because that is where the old
// service's callers put it and the vending machines still do. Either spelling has to reach the
// adapter, or every card-terminal payment from an un-updated caller fails validation.
func TestATopLevelPosIdIsFoldedIntoTheMetadata(t *testing.T) {
	cmd, vErrs := buildCreatePaymentCommand(dmodel.DynamicFields{
		paramPaymentMethodId: "01METHOD0000000000000000000",
		paramOrgId:           testOrgId,
		paramAmount:          "150000",
		paramPosId:           "POS-01",
	})

	assert.Zero(t, vErrs.Count())
	assert.Equal(t, "POS-01", cmd.Metadata[models.OrderMetaPosId])
}

func TestAMetadataPosIdIsPassedThrough(t *testing.T) {
	cmd, vErrs := buildCreatePaymentCommand(dmodel.DynamicFields{
		paramPaymentMethodId: "01METHOD0000000000000000000",
		paramOrgId:           testOrgId,
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
		paramOrgId:           testOrgId,
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
		paramOrgId:           testOrgId,
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

// The merchant account is optional: a caller that names none is collected with the credentials in
// this deployment's configuration, which is what every caller did before profiles existed. Making
// it required would have broken every existing vending machine on the day profiles shipped.
func TestAPaymentProfileIsOptional(t *testing.T) {
	cmd, vErrs := buildCreatePaymentCommand(dmodel.DynamicFields{
		paramPaymentMethodId: "01METHOD0000000000000000000",
		paramOrgId:           testOrgId,
		paramAmount:          "150000",
	})

	assert.Zero(t, vErrs.Count())
	assert.Empty(t, cmd.PaymentProfileId)
}

func TestAPaymentProfileIsCarriedIntoTheCommand(t *testing.T) {
	cmd, vErrs := buildCreatePaymentCommand(dmodel.DynamicFields{
		paramPaymentMethodId:  "01METHOD0000000000000000000",
		paramOrgId:            testOrgId,
		paramPaymentProfileId: "01PROFILE000000000000000000",
		paramAmount:           "150000",
	})

	assert.Zero(t, vErrs.Count())
	assert.Equal(t, "01PROFILE000000000000000000", cmd.PaymentProfileId)
}

// Every record this module writes carries an organization, and it cannot be derived from the
// request context: a caller may belong to several. Refusing here names the missing field; letting
// it through would have the order schema reject the composed record as a server fault instead.
func TestAnOrganizationIsRequired(t *testing.T) {
	_, vErrs := buildCreatePaymentCommand(dmodel.DynamicFields{
		paramPaymentMethodId: "01METHOD0000000000000000000",
		paramAmount:          "150000",
	})

	assert.Equal(t, 1, vErrs.Count())
}

// A caller migrating off the standalone service has "momo", not one of our ids. Naming the method
// by code has to be accepted, or every such caller would need a lookup it has no way to perform.
func TestAPaymentMethodMayBeNamedByCode(t *testing.T) {
	cmd, vErrs := buildCreatePaymentCommand(dmodel.DynamicFields{
		paramPaymentMethodCode: "momo",
		paramOrgId:             testOrgId,
		paramAmount:            "150000",
	})

	assert.Zero(t, vErrs.Count())
	assert.Equal(t, "momo", cmd.PaymentMethodCode)
	assert.Empty(t, cmd.PaymentMethodId)
}

// Both spellings reaching the command is not an error — they name the same row when the caller is
// consistent. The id is carried through as well so the service can prefer it.
func TestBothMethodSpellingsAreCarriedThrough(t *testing.T) {
	cmd, vErrs := buildCreatePaymentCommand(dmodel.DynamicFields{
		paramPaymentMethodId:   "01METHOD0000000000000000000",
		paramPaymentMethodCode: "momo",
		paramOrgId:             testOrgId,
		paramAmount:            "150000",
	})

	assert.Zero(t, vErrs.Count())
	assert.Equal(t, "01METHOD0000000000000000000", cmd.PaymentMethodId)
	assert.Equal(t, "momo", cmd.PaymentMethodCode)
}

// The standalone service handed out order_id and order_code together, and callers stored one or
// the other. A refund must not fail because the caller kept the one we did not ask for.
func TestARefundMayQuoteTheOrderCode(t *testing.T) {
	cmd, vErrs := buildRefundCommand(dmodel.DynamicFields{
		paramOrderCode: "ORD1234ABCD5",
		paramAmount:    "50000",
	})

	assert.Zero(t, vErrs.Count())
	assert.Equal(t, "ORD1234ABCD5", cmd.OrderCode)
	assert.Empty(t, cmd.OrderId)
}

func TestARefundStillNeedsOneOfTheTwoIdentifiers(t *testing.T) {
	_, vErrs := buildRefundCommand(dmodel.DynamicFields{paramAmount: "50000"})

	assert.Equal(t, 1, vErrs.Count())
}
