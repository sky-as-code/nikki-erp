package mpos

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The ciphertexts below were produced by the NestJS service this module supersedes
// (src/modules/mpos/mpos.service.ts) over the same fixed payloads under the same fixed key. They
// are the contract with the gateway captured as data.
//
// mPOS rejects anything it cannot decrypt, and the failure surfaces as a payment that simply
// never happens — so a divergence has to be caught here. Do not "fix" a failing expectation by
// pasting in whatever the Go now produces: regenerate it from the TypeScript, or the test stops
// testing anything.

// Exactly 16 bytes, as AES-128 requires.
const testSecretKey = "mpos-test-key-16"

func TestAddOrderEncryptionMatchesReference(t *testing.T) {
	encrypted, err := encrypt(addOrderRequest{
		ServiceName:   ServiceAddOrder,
		OrderId:       "ORD1234ABCD5",
		PosId:         "POS-0001",
		Amount:        150000,
		Description:   "Test payment",
		PaymentType:   paymentTypeOther,
		PaymentMethod: PaymentMethodCard,
	}, testSecretKey)

	require.NoError(t, err)
	assert.Equal(t,
		"wcvq+MhRLd0v+7y5J1Iki0C/CTPr7ohd5PMsehk37JPKDf7RiExcJBjkxgHYgfXN1a0Ye12OkIu0op8mT1UPr1qP"+
			"Fb6zHAv3DG/LAMhx0JR3D6F6aLX+6l6fSugj9tb1JPgTkUi/zfopMTal1RvOdAED4SJt/lhpL+OngdRIVNFp"+
			"GEfLeSGMM9+J9KmKT7Gl07wRqmDmb2kkGVUJAhNWyMs87NyTX4+XR2NdwP3TcjU=",
		encrypted)
}

func TestRefundEncryptionMatchesReference(t *testing.T) {
	encrypted, err := encrypt(refundRequest{
		ServiceName:  ServiceRefundTransaction,
		TransCode:    "TXN-99887766",
		RequestId:    "req-0002",
		RefundAmount: 50000,
	}, testSecretKey)

	require.NoError(t, err)
	assert.Equal(t,
		"wcvq+MhRLd0v+7y5J1Ikiz0M6WPThWwOZQQrG1Maahg6/+Vja8078TPCw6lGyUBJ5EeV2m3/EP6UH5A4uEiD6K8X"+
			"MrendbfMI651bbWWWDBdMhDoaQ/MBsZePJ8yPHY++EM9u4RoXMWAYK3BENlUPw==",
		encrypted)
}

// The webhook path is the one that matters most: this is real gateway ciphertext, and being able
// to decrypt it is what authenticates the callback.
func TestWebhookDecryptionMatchesReference(t *testing.T) {
	const reference = "wcvq+MhRLd0v+7y5J1Ikizkc5AxqgsWt/iWeIFTa7/vhwxVy5fb+9VX3Y+sP0bRIO1WrKh0we8aolVNp" +
		"RKGgxeb46gRIP1YMmutwz7pvFw3rgytqGNCi9GZ/pt4yQcjrcuIjcceRg4KW7yITFKccgp8OH6f/sGKk" +
		"vAlKqL3iUH/uMqObyVi4q17ztDKEax6Wbz6XHi1JhRnJzdMu6TAhjRPrdescDupRb/LrxJOOhboZeRYH" +
		"PpiqcdC1srkdIAGkiarydatWXctdF5tt+G/ASFwklHImPAwMIDVt6/GoTDY3HoXUMm+cEakKUImkXl4J"

	adapter := NewAdapter(Config{MerchantId: "M1", SecretKey: testSecretKey}, nil)
	payload, err := adapter.DecryptWebhook(reference)

	require.NoError(t, err)
	assert.Equal(t, ServiceUpdateTransaction, payload.ServiceName)
	assert.Equal(t, TransStatusApproved, payload.TransStatus)
	assert.Equal(t, "TXN-99887766", payload.TransCode)
	assert.Equal(t, "ORD1234ABCD5", payload.OrderId)
	assert.Equal(t, "POS-0001", payload.PosId)
	assert.Equal(t, int64(150000), payload.TransAmount)
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	original := refundRequest{
		ServiceName:  ServiceRefundTransaction,
		TransCode:    "TXN-1",
		RequestId:    "req-1",
		RefundAmount: 1,
	}

	encrypted, err := encrypt(original, testSecretKey)
	require.NoError(t, err)

	var decoded refundRequest
	require.NoError(t, decrypt(encrypted, testSecretKey, &decoded))
	assert.Equal(t, original, decoded)
}

// A key of the wrong length must fail loudly. Left unchecked it would surface as every single
// request being rejected by the gateway, with nothing pointing at the configuration.
func TestWrongKeyLengthIsRejected(t *testing.T) {
	_, err := encrypt(refundRequest{}, "too-short")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "16 bytes")
}

// The webhook path decrypts bytes that arrived from outside. These are the ways that can go wrong
// without a valid secret, and each must be refused rather than acted on: a callback we cannot
// decrypt did not come from mPOS.
func TestMalformedCiphertextIsRefused(t *testing.T) {
	adapter := NewAdapter(Config{MerchantId: "M1", SecretKey: testSecretKey}, nil)

	cases := map[string]string{
		"not base64":        "!!!not-base64!!!",
		"empty":             "",
		"not a whole block": base64.StdEncoding.EncodeToString([]byte("short")),
		"wrong key's ciphertext": base64.StdEncoding.EncodeToString(
			[]byte("0123456789abcdef0123456789abcdef")),
	}

	for name, reqData := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := adapter.DecryptWebhook(reqData)
			assert.Error(t, err)
		})
	}
}

// Padding is stripped by trusting a length byte in the decrypted plaintext, so it has to be
// validated: an attacker-chosen body decrypts to arbitrary bytes whose last byte could claim any
// length, and slicing on it unchecked would read outside the buffer.
func TestInvalidPaddingIsRefused(t *testing.T) {
	block := make([]byte, 16)

	block[15] = 0
	_, err := unpadPkcs7(block, 16)
	assert.Error(t, err, "a zero padding length is not valid PKCS#7")

	block[15] = 17
	_, err = unpadPkcs7(block, 16)
	assert.Error(t, err, "padding longer than the block cannot be valid")

	// Claims four padding bytes, but the preceding bytes do not agree.
	block[15] = 4
	block[14] = 9
	_, err = unpadPkcs7(block, 16)
	assert.Error(t, err, "padding bytes must all equal the padding length")
}
