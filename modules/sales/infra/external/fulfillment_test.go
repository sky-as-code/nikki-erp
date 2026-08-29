package external

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	invModels "github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	itStock "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/stock"

	itExt "github.com/sky-as-code/nikki-erp/modules/sales/interfaces/external"
)

// The invariant worth most here is the refusal/fault split: inventory answers "not enough stock" as
// ClientErrors with a nil Go error, and a genuine breakage as an error. A refusal reported as a
// fault turns an out-of-stock into an unactionable 500; a fault reported as a refusal marks a
// fulfilment rejected when inventory never got the message.

// stubTransfers records what it was asked and answers what the test told it to.
type stubTransfers struct {
	createResult *dyn.OpResult[dmodel.DynamicFields]
	createErr    error
	createdWith  dmodel.DynamicFields
	createdMoves []itStock.TransferMoveRequest

	confirmResult  *dyn.OpResult[dyn.MutateResultData]
	reserveResult  *dyn.OpResult[dyn.MutateResultData]
	validateResult *dyn.OpResult[dyn.MutateResultData]
	validateErr    error

	unreserveResult *dyn.OpResult[dyn.MutateResultData]

	// calls names the operations in order, so a test can assert a sequence stopped where it should.
	calls []string

	validatedWithKey string
}

func (this *stubTransfers) Create(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	this.calls = append(this.calls, "create")
	return this.createResult, this.createErr
}

func (this *stubTransfers) CreateWithMoves(
	ctx corectx.Context, params dmodel.DynamicFields, moves []itStock.TransferMoveRequest,
) (*dyn.OpResult[dmodel.DynamicFields], error) {
	this.calls = append(this.calls, "create_with_moves")
	this.createdWith = params
	this.createdMoves = moves
	return this.createResult, this.createErr
}

func (this *stubTransfers) Confirm(
	ctx corectx.Context, transferId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	this.calls = append(this.calls, "confirm")
	return this.confirmResult, nil
}

func (this *stubTransfers) Reserve(
	ctx corectx.Context, transferId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	this.calls = append(this.calls, "reserve")
	return this.reserveResult, nil
}

func (this *stubTransfers) Unreserve(
	ctx corectx.Context, transferId string,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	this.calls = append(this.calls, "unreserve")
	return this.unreserveResult, nil
}

func (this *stubTransfers) Validate(
	ctx corectx.Context, transferId string, idempotencyKey string, createBackorder *bool,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	this.calls = append(this.calls, "validate")
	this.validatedWithKey = idempotencyKey
	return this.validateResult, this.validateErr
}

func (this *stubTransfers) CreateReturn(
	ctx corectx.Context, transferId string, request itStock.TransferReturnRequest,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	this.calls = append(this.calls, "create_return")
	return nil, nil
}

// stubOperationTypes answers with whatever the deployment is pretending to have configured.
type stubOperationTypes struct {
	outgoing string
	incoming string
}

func (this *stubOperationTypes) OutgoingOperationTypeId(corectx.Context) (string, error) {
	return this.outgoing, nil
}

func (this *stubOperationTypes) IncomingOperationTypeId(corectx.Context) (string, error) {
	return this.incoming, nil
}

func okCreate(transferId string) *dyn.OpResult[dmodel.DynamicFields] {
	return &dyn.OpResult[dmodel.DynamicFields]{
		HasData: true,
		Data: dmodel.DynamicFields{
			invModels.StockTransferFieldId: transferId,
		},
	}
}

func okMutate() *dyn.OpResult[dyn.MutateResultData] {
	return &dyn.OpResult[dyn.MutateResultData]{HasData: true}
}

func refusedMutate(key, message string) *dyn.OpResult[dyn.MutateResultData] {
	vErrs := ft.NewClientErrors()
	vErrs.Append(*ft.NewBusinessViolation("stock", key, message))
	return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: *vErrs}
}

func adapterWith(transfers *stubTransfers) *fulfillmentAdapter {
	return &fulfillmentAdapter{
		transfers: transfers,
		operationTypes: &stubOperationTypes{
			outgoing: "01OUTGOING000000000000000",
			incoming: "01INCOMING000000000000000",
		},
	}
}

func oneLine() itExt.FulfillmentRequest {
	return itExt.FulfillmentRequest{
		SalesOrderId:              "01ORDER00000000000000000",
		SalesFulfillmentRequestId: "01REQ0000000000000000000",
		IdempotencyKey:            "01REQ0000000000000000000",
		Lines: []itExt.FulfillmentLine{{
			SalesOrderLineId: "01LINE000000000000000000",
			ProductVariantId: "01VARIANT0000000000000000",
			UomId:            "01UOM0000000000000000000",
			Quantity:         decimal.NewFromInt(2),
		}},
	}
}

// A reservation stops after Reserve: validating would move goods a confirmed-but-unshipped sale
// must leave where they are.
func TestReservationHoldsStockWithoutMovingIt(t *testing.T) {
	transfers := &stubTransfers{
		createResult:  okCreate("01TRANSFER0000000000000"),
		confirmResult: okMutate(),
		reserveResult: okMutate(),
	}

	response, err := adapterWith(transfers).RequestReservation(nil, oneLine())

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.True(t, response.Accepted)
	assert.False(t, response.Completed, "a reservation must not report goods as moved")
	assert.Equal(t, "01TRANSFER0000000000000", response.InventoryReference)
	assert.Equal(t, []string{"create_with_moves", "confirm", "reserve"}, transfers.calls,
		"a reservation must not validate: validating moves the goods")
}

// A goods issue runs the whole sequence and reports the goods moved.
func TestGoodsIssueValidatesAndReportsCompletion(t *testing.T) {
	transfers := &stubTransfers{
		createResult:   okCreate("01TRANSFER0000000000000"),
		confirmResult:  okMutate(),
		reserveResult:  okMutate(),
		validateResult: okMutate(),
	}

	response, err := adapterWith(transfers).RequestGoodsIssue(nil, oneLine())

	require.NoError(t, err)
	require.True(t, response.Accepted)
	assert.True(t, response.Completed)
	assert.Equal(t,
		[]string{"create_with_moves", "confirm", "reserve", "validate"}, transfers.calls)

	// The idempotency key must reach Validate, or a retry after a timeout moves the goods twice.
	assert.Equal(t, "01REQ0000000000000000000", transfers.validatedWithKey,
		"validate must carry the idempotency key, or a retry issues the goods a second time")
}

// Not enough stock is a REFUSAL, not a fault: no Go error, Accepted false, reason carried across.
func TestInsufficientStockIsARefusalRatherThanAnError(t *testing.T) {
	transfers := &stubTransfers{
		createResult:  okCreate("01TRANSFER0000000000000"),
		confirmResult: okMutate(),
		reserveResult: refusedMutate(
			"stock_transfer.insufficient", "only 1 of 2 units are available"),
	}

	response, err := adapterWith(transfers).RequestReservation(nil, oneLine())

	require.NoError(t, err, "stock running out is an ordinary outcome, not a server fault")
	require.NotNil(t, response)
	assert.False(t, response.Accepted)
	assert.Contains(t, response.FailureReason, "only 1 of 2 units are available",
		"an operator told only 'rejected' cannot tell whether to wait, re-route or refund")
	assert.Contains(t, response.FailureReason, "stock_transfer.insufficient",
		"the translation key must survive, so a UI can localise the refusal")

	// The transfer really was created, so its id must come back or the operator cannot find it.
	assert.Equal(t, "01TRANSFER0000000000000", response.InventoryReference)
}

// A genuine breakage stays a Go error; reporting it as a refusal would mark the fulfilment rejected
// when inventory never answered.
func TestATransportFailureStaysAnError(t *testing.T) {
	transfers := &stubTransfers{
		createResult:   okCreate("01TRANSFER0000000000000"),
		confirmResult:  okMutate(),
		reserveResult:  okMutate(),
		validateErr:    assert.AnError,
		validateResult: nil,
	}

	response, err := adapterWith(transfers).RequestGoodsIssue(nil, oneLine())

	require.Error(t, err, "a broken call must not be recorded as a business refusal")
	assert.Nil(t, response)
}

// With no operation type configured the request is refused with a reason and nothing is created:
// guessing a type would move real goods out of the wrong place.
func TestAnUnconfiguredOperationTypeRefusesWithoutCreatingAnything(t *testing.T) {
	transfers := &stubTransfers{createResult: okCreate("01TRANSFER0000000000000")}
	adapter := &fulfillmentAdapter{
		transfers:      transfers,
		operationTypes: &stubOperationTypes{outgoing: "", incoming: ""},
	}

	response, err := adapter.RequestReservation(nil, oneLine())

	require.NoError(t, err, "a configuration gap is somebody's to close, not a 500")
	require.NotNil(t, response)
	assert.False(t, response.Accepted)
	assert.Contains(t, response.FailureReason, "outgoing")
	assert.Empty(t, transfers.calls,
		"nothing may be created before the operation type is known: picking one at random "+
			"ships goods from a warehouse nobody chose")
}

// Sales' order id and idempotency key must reach the transfer, and the lines must arrive as moves.
func TestTheOrderAndItsLinesReachInventory(t *testing.T) {
	transfers := &stubTransfers{
		createResult:  okCreate("01TRANSFER0000000000000"),
		confirmResult: okMutate(),
		reserveResult: okMutate(),
	}

	_, err := adapterWith(transfers).RequestReservation(nil, oneLine())
	require.NoError(t, err)

	assert.Equal(t, "01ORDER00000000000000000",
		transfers.createdWith[invModels.StockTransferFieldOriginReference],
		"the sales order must be traceable from the transfer")
	assert.Equal(t, "01REQ0000000000000000000",
		transfers.createdWith[invModels.StockTransferFieldIdempotencyKey],
		"without the key, a retried request raises a second transfer")

	require.Len(t, transfers.createdMoves, 1)
	assert.Equal(t, "01VARIANT0000000000000000", transfers.createdMoves[0].ProductVariantId)
	assert.True(t, decimal.NewFromInt(2).Equal(transfers.createdMoves[0].Quantity))
}

// A return is received as an incoming transfer and is not reserved: an incoming draws from outside
// the business, so there is no balance to claim.
func TestAReturnIsReceivedWithoutReserving(t *testing.T) {
	transfers := &stubTransfers{
		createResult:   okCreate("01RETURN000000000000000"),
		confirmResult:  okMutate(),
		validateResult: okMutate(),
	}

	response, err := adapterWith(transfers).RequestReturnReceipt(nil, oneLine())

	require.NoError(t, err)
	require.True(t, response.Accepted)
	assert.True(t, response.Completed)
	assert.Equal(t, []string{"create_with_moves", "confirm", "validate"}, transfers.calls,
		"an incoming transfer has no balance to reserve against")
}

// Releasing a hold that was never taken succeeds: the caller's intent is already true.
func TestReleasingNothingSucceeds(t *testing.T) {
	transfers := &stubTransfers{}

	response, err := adapterWith(transfers).ReleaseReservation(nil, "")

	require.NoError(t, err)
	assert.True(t, response.Accepted)
	assert.Empty(t, transfers.calls, "there is nothing to ask Inventory to release")
}

// Releasing unreserves without cancelling, so the document records that it existed and was
// released.
func TestReleasingUnreservesWithoutCancelling(t *testing.T) {
	transfers := &stubTransfers{unreserveResult: okMutate()}

	response, err := adapterWith(transfers).ReleaseReservation(nil, "01TRANSFER0000000000000")

	require.NoError(t, err)
	assert.True(t, response.Accepted)
	assert.Equal(t, []string{"unreserve"}, transfers.calls)
}

// The adapter really does satisfy the port Sales depends on.
var _ itExt.FulfillmentExtService = (*fulfillmentAdapter)(nil)
