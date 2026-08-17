package v1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/infra/gateway/momo"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/infra/gateway/vietqr"
)

// TestAsPayloadMap preserves gateway field names for evidence.
func TestAsPayloadMap(t *testing.T) {
	payload := momo.IpnPayload{
		OrderId:    "TEST0001",
		ResultCode: momo.ResultCodeSuccess,
		TransId:    123456789,
		Message:    "Success",
	}

	mapped := asPayloadMap(payload)

	require.NotNil(t, mapped)
	// Verify the mapped output carries the gateway's field names and values
	assert.Equal(t, "TEST0001", mapped["orderId"])
	assert.Equal(t, float64(momo.ResultCodeSuccess), mapped["resultCode"])
	assert.Equal(t, float64(123456789), mapped["transId"])
}

// TestAsPayloadMapHandlesMarshalError returns nil when payload cannot be marshaled.
func TestAsPayloadMapHandlesMarshalError(t *testing.T) {
	// Create something that fails to marshal (e.g., a channel)
	payload := make(chan int)

	mapped := asPayloadMap(payload)

	assert.Nil(t, mapped, "unmarshalable payload should return nil")
}

// TestMomoIpnPayloadShape verifies the payload structure matches gateway contract.
func TestMomoIpnPayloadShape(t *testing.T) {
	jsonPayload := `{
		"orderId": "TEST0001",
		"resultCode": 0,
		"transId": 123456789,
		"message": "Success",
		"responseTime": 20260818152030
	}`

	var payload momo.IpnPayload
	err := json.Unmarshal([]byte(jsonPayload), &payload)

	require.NoError(t, err)
	assert.Equal(t, "TEST0001", payload.OrderId)
	assert.Equal(t, momo.ResultCodeSuccess, payload.ResultCode)
	assert.Equal(t, int64(123456789), payload.TransId)
}

// TestVietQrTokenResponseShape verifies the token response structure.
func TestVietQrTokenResponseShape(t *testing.T) {
	response := vietqr.TokenResponse{
		AccessToken: "test-token-abc123",
		TokenType:   vietqr.TokenTypeBearer,
		ExpiresIn:   3600,
	}

	data, err := json.Marshal(response)
	require.NoError(t, err)

	var unmarshaled vietqr.TokenResponse
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, "test-token-abc123", unmarshaled.AccessToken)
	assert.Equal(t, "Bearer", unmarshaled.TokenType)
	assert.Equal(t, int64(3600), unmarshaled.ExpiresIn)
}
