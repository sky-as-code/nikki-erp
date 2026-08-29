package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"

	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/domain/models"
)

// A payment method and an order may each be named two ways, because the standalone service this
// module supersedes handed out both identifiers and its callers kept whichever one suited them.
// What must not happen is a lookup that silently picks the wrong column.

func TestAMethodIsLookedUpByIdWhenGiven(t *testing.T) {
	key, named := methodLookupKey(CreatePaymentCommand{
		PaymentMethodId:   "01METHOD0000000000000000000",
		PaymentMethodCode: "momo",
	})

	assert.True(t, named)
	assert.Equal(t, dmodel.DynamicFields{
		models.PaymentMethodFieldId: "01METHOD0000000000000000000",
	}, key, "the id wins over the code")
}

func TestAMethodIsLookedUpByCodeWhenThatIsAllThereIs(t *testing.T) {
	key, named := methodLookupKey(CreatePaymentCommand{PaymentMethodCode: "momo"})

	assert.True(t, named)
	assert.Equal(t, dmodel.DynamicFields{models.PaymentMethodFieldCode: "momo"}, key)
}

func TestAMethodNamedNoWayIsRefused(t *testing.T) {
	key, named := methodLookupKey(CreatePaymentCommand{})

	assert.False(t, named)
	assert.Nil(t, key)
}

// The "not found" message has to say which identifier missed, or an operator cannot tell a wrong
// id from a code this deployment does not have a row for.
func TestTheNotFoundMessageNamesWhatWasAskedFor(t *testing.T) {
	assert.Contains(t, describeMethodKey(CreatePaymentCommand{PaymentMethodId: "01M"}), "id '01M'")
	assert.Contains(t, describeMethodKey(CreatePaymentCommand{PaymentMethodCode: "momo"}), "code 'momo'")

	assert.Contains(t, describeRefundTarget(RefundCommand{OrderId: "VDMC-1"}), "id 'VDMC-1'")
	assert.Contains(t, describeRefundTarget(RefundCommand{OrderCode: "ORD1"}), "code 'ORD1'")
	assert.Contains(t, describeRefundTarget(RefundCommand{}), "no identifier")
}
