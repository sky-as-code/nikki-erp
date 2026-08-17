package momo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// The signatures below are not hand-computed: each was produced by running the same algorithm as
// the NestJS service this module supersedes (src/modules/momo/momo.service.ts) over the same
// fixed inputs under the same fixed secret. They are the contract with MoMo captured as data.
//
// A signature MoMo rejects fails with a message that says only that the signature is wrong, and
// the payment simply never happens — so this must be caught here rather than in production. Do
// not "fix" a failing expectation by pasting in whatever the code now produces: regenerate it
// from the TypeScript, or the test stops testing anything.

const (
	testSecretKey = "test-secret-key-do-not-use"
	testAccessKey = "test-access-key"
	testPartner   = "MOMOTEST01"
)

func TestCreatePaymentSignatureMatchesReference(t *testing.T) {
	fields := signingFields{
		"accessKey":   testAccessKey,
		"amount":      "150000",
		"extraData":   "",
		"ipnUrl":      "https://example.test/v1/paymentinvoice/webhooks/momo",
		"orderId":     "ORD1234ABCD5",
		"orderInfo":   "Test payment",
		"partnerCode": testPartner,
		"redirectUrl": "https://example.test/done",
		"requestId":   "req-0001",
		"requestType": "captureWallet",
	}

	assert.Equal(t,
		"accessKey=test-access-key&amount=150000&extraData=&ipnUrl=https://example.test/v1/paymentinvoice/webhooks/momo"+
			"&orderId=ORD1234ABCD5&orderInfo=Test payment&partnerCode=MOMOTEST01"+
			"&redirectUrl=https://example.test/done&requestId=req-0001&requestType=captureWallet",
		fields.rawSignature())
	assert.Equal(t,
		"a6b49a43bd9ab22bdd498edb437f367e1fa6e518b1f9212f9c970d7dddb87495",
		fields.sign(testSecretKey))
}

func TestRefundSignatureMatchesReference(t *testing.T) {
	fields := signingFields{
		"accessKey":   testAccessKey,
		"amount":      "50000",
		"description": "Refund for ORD1234ABCD5",
		"orderId":     "RFD0001",
		"partnerCode": testPartner,
		"requestId":   "req-0002",
		"transId":     "2147483647",
	}

	assert.Equal(t,
		"8906e5a022badf076b0fb19570d04302d294be5619a88b8075916f6308919f34",
		fields.sign(testSecretKey))
}

func TestQuerySignatureMatchesReference(t *testing.T) {
	fields := signingFields{
		"accessKey":   testAccessKey,
		"orderId":     "ORD1234ABCD5",
		"partnerCode": testPartner,
		"requestId":   "req-0003",
	}

	assert.Equal(t,
		"c33162ed354971864cab771f529ccf8e61a01189cb7d9d7233347e83b292c32b",
		fields.sign(testSecretKey))
}

// The IPN set is not the same as any request's, so it gets its own vector: it signs message,
// orderType, payType, responseTime and resultCode, none of which a request signs.
func TestVerifyIpnAcceptsAGenuineCallback(t *testing.T) {
	adapter := newTestAdapter()

	assert.True(t, adapter.VerifyIpn(genuineIpnPayload()))
}

func TestVerifyIpnAcceptsAnUpperCaseSignature(t *testing.T) {
	adapter := newTestAdapter()
	payload := genuineIpnPayload()
	payload.Signature = "C6C8E1C0DFBC246057D7A0CB6917A78399F919638AF7C6C12C24AF96FD47823A"

	// A callback that differed only in case would otherwise be discarded as forged, leaving a
	// payment that really happened permanently unsettled.
	assert.True(t, adapter.VerifyIpn(payload))
}

// Every field is signed, so tampering with any of them must invalidate the callback — otherwise
// anyone who learned an order code could post a "successful payment" for it.
func TestVerifyIpnRejectsTampering(t *testing.T) {
	adapter := newTestAdapter()

	tampered := map[string]func(*IpnPayload){
		"amount":     func(p *IpnPayload) { p.Amount = 999999 },
		"resultCode": func(p *IpnPayload) { p.ResultCode = 1 },
		"orderId":    func(p *IpnPayload) { p.OrderId = "SOMEONE-ELSES-ORDER" },
		"transId":    func(p *IpnPayload) { p.TransId = 1 },
		"signature":  func(p *IpnPayload) { p.Signature = "00" },
	}

	for name, tamper := range tampered {
		t.Run(name, func(t *testing.T) {
			payload := genuineIpnPayload()
			tamper(&payload)
			assert.False(t, adapter.VerifyIpn(payload))
		})
	}
}

func newTestAdapter() *Adapter {
	return NewAdapter(Config{
		PartnerCode: testPartner,
		AccessKey:   testAccessKey,
		SecretKey:   testSecretKey,
	}, nil)
}

func genuineIpnPayload() IpnPayload {
	return IpnPayload{
		PartnerCode:  testPartner,
		OrderId:      "ORD1234ABCD5",
		RequestId:    "req-0001",
		Amount:       150000,
		OrderInfo:    "Test payment",
		OrderType:    "momo_wallet",
		TransId:      2147483647,
		ResultCode:   0,
		Message:      "Successful.",
		PayType:      "qr",
		ResponseTime: 1755400000000,
		ExtraData:    "",
		Signature:    "c6c8e1c0dfbc246057d7a0cb6917a78399f919638af7c6c12c24af96fd47823a",
	}
}
