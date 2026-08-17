package vietqr

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The checksums below were produced by the NestJS service this module supersedes
// (src/modules/vietqr/vietqr.service.ts) over the same fixed inputs under the same fixed secret.
//
// A wrong checksum is refused by the gateway with no useful explanation, so a divergence has to
// be caught here. Do not "fix" a failing expectation by pasting in whatever the Go now produces:
// regenerate it from the TypeScript, or the test stops testing anything.

const (
	testSecretKey  = "vietqr-test-secret"
	testBankNumber = "0011002233445"
	testUsername   = "vietqr-test-user"
)

func TestRefundChecksumMatchesReference(t *testing.T) {
	checksum := refundChecksum(testSecretKey, "FT2426000123456", "150000", testBankNumber)

	assert.Equal(t, "72b1590225b1bc9154cc75e33990efd0", checksum)
}

func TestCheckOrderChecksumMatchesReference(t *testing.T) {
	checksum := checkOrderChecksum(testBankNumber, testUsername)

	assert.Equal(t, "fc4a5c13b1f4cfb155b04b2d8bfbb75f", checksum)
}

// The amount is checksummed as the exact text the request carries. Rendering it differently on
// either side — "150000.00" for "150000" — produces a checksum the gateway rejects, and the two
// have to be built from one string for that reason.
func TestRefundChecksumIsSensitiveToAmountFormatting(t *testing.T) {
	plain := refundChecksum(testSecretKey, "FT2426000123456", "150000", testBankNumber)
	scaled := refundChecksum(testSecretKey, "FT2426000123456", "150000.00", testBankNumber)

	assert.NotEqual(t, plain, scaled)
}

// The order of the concatenated parts is the gateway's, and rearranging it into something that
// reads more naturally would silently break every refund.
func TestRefundChecksumOrderIsLoadBearing(t *testing.T) {
	correct := refundChecksum(testSecretKey, "REF1", "100", testBankNumber)
	swapped := refundChecksum(testSecretKey, testBankNumber, "100", "REF1")

	assert.NotEqual(t, correct, swapped)
}

// The bank reads this reply's body, not the HTTP status, and keys off errorReason. Serialising it
// differently — renaming a field, omitting `object`, returning a bare 200 — breaks the
// integration silently: the bank treats the reply as a failure and keeps retrying.
func TestAcceptedWebhookReplyIsByteExact(t *testing.T) {
	encoded, err := json.Marshal(NewWebhookAccepted("01ABCDEF"))

	require.NoError(t, err)
	assert.JSONEq(t,
		`{"error":false,"errorReason":"000","toastMessage":"","object":{"reftransactionid":"01ABCDEF"}}`,
		string(encoded))
}

func TestNotFoundWebhookReplyIsByteExact(t *testing.T) {
	encoded, err := json.Marshal(NewWebhookNotFound())

	require.NoError(t, err)
	assert.JSONEq(t,
		`{"error":true,"errorReason":"001","toastMessage":"Giao dịch không tồn tại",`+
			`"object":{"reftransactionid":""}}`,
		string(encoded))
}

// The bank's field names are run-together lower case that no style guide would produce, so it is
// worth pinning that they survive a decode.
func TestWebhookPayloadDecodesTheBanksFieldNames(t *testing.T) {
	const body = `{"bankaccount":"0011002233445","amount":150000,"transType":"C",` +
		`"content":"ORD1234ABCD5","transactionid":"TXN1","transactiontime":1755400000,` +
		`"referencenumber":"FT2426000123456","orderId":"ORD1234ABCD5"}`

	var payload WebhookPayload
	require.NoError(t, json.Unmarshal([]byte(body), &payload))

	assert.Equal(t, "0011002233445", payload.BankAccount)
	assert.Equal(t, int64(150000), payload.Amount)
	assert.Equal(t, "FT2426000123456", payload.ReferenceNumber)
	assert.Equal(t, "ORD1234ABCD5", payload.OrderId)
}
