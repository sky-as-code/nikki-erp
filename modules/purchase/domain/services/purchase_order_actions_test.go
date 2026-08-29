package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// Unlocking without a reason is refused before anything is read. The reason is mandatory because
// unlocking undoes a deliberately applied control, and a trail of unexplained unlocks is no trail.
func TestUnlockRequiresAReason(t *testing.T) {
	service := &PurchaseOrderDomainServiceImpl{}

	result, err := service.Unlock(nil, "01ABC", "")

	require.NoError(t, err, "a refusal is a client error, not a Go error")
	require.Equal(t, 1, result.ClientErrors.Count())
	assert.Equal(t, "purchase_order.unlock_reason_required", result.ClientErrors[0].Key)
}

// The duplicate carries the order's terms and none of its history, or it would claim to have been
// approved by someone who never saw it.
func TestCopyableOrderFieldsDropsTheHistory(t *testing.T) {
	order := dmodel.DynamicFields{
		models.PurchaseOrderFieldId:                 "01ORIGINAL",
		models.PurchaseOrderFieldCode:               "PO-01ORIGINAL",
		models.PurchaseOrderFieldStatus:             string(models.PurchaseOrderStatusPurchaseOrder),
		models.PurchaseOrderFieldVendorId:           "01VENDOR",
		models.PurchaseOrderFieldBuyerId:            "01BUYER",
		models.PurchaseOrderFieldCurrencyId:         "01CURRENCY",
		models.PurchaseOrderFieldPriority:           string(models.PurchasePriorityUrgent),
		models.PurchaseOrderFieldTermsConditions:    "net 30",
		models.PurchaseOrderFieldConfirmedAt:        "2026-01-01T00:00:00Z",
		models.PurchaseOrderFieldApprovedBy:         "01APPROVER",
		models.PurchaseOrderFieldApprovedAt:         "2026-01-01T00:00:00Z",
		models.PurchaseOrderFieldApprovalRequired:   true,
		models.PurchaseOrderFieldIsLocked:           true,
		models.PurchaseOrderFieldVendorAcknowledged: true,
		models.PurchaseOrderFieldTotalAmount:        dec("500"),
		models.PurchaseOrderFieldOrderDeadline:      "2026-01-01T00:00:00Z",
		models.PurchaseOrderFieldSourcingGroupId:    "01GROUP",
		basemodel.FieldOrgId:                        "01ORG",
		basemodel.FieldEtag:                         "12345",
	}

	copied := copyableOrderFields(order)

	// The terms come across.
	assert.Equal(t, "01VENDOR", copied[models.PurchaseOrderFieldVendorId])
	assert.Equal(t, "01BUYER", copied[models.PurchaseOrderFieldBuyerId])
	assert.Equal(t, "01CURRENCY", copied[models.PurchaseOrderFieldCurrencyId])
	assert.Equal(t, string(models.PurchasePriorityUrgent), copied[models.PurchaseOrderFieldPriority])
	assert.Equal(t, "net 30", copied[models.PurchaseOrderFieldTermsConditions])
	assert.Equal(t, "01ORG", copied[basemodel.FieldOrgId])

	// The history does not.
	for _, field := range []string{
		models.PurchaseOrderFieldId,
		models.PurchaseOrderFieldCode,
		models.PurchaseOrderFieldStatus,
		models.PurchaseOrderFieldConfirmedAt,
		models.PurchaseOrderFieldApprovedBy,
		models.PurchaseOrderFieldApprovedAt,
		models.PurchaseOrderFieldApprovalRequired,
		models.PurchaseOrderFieldIsLocked,
		models.PurchaseOrderFieldVendorAcknowledged,
		models.PurchaseOrderFieldTotalAmount,
		models.PurchaseOrderFieldUntaxedAmount,
		models.PurchaseOrderFieldTaxAmount,
		basemodel.FieldEtag,
	} {
		assert.NotContains(t, copied, field, "%s must not survive a duplicate", field)
	}

	// A deadline already in the past is worse than none, and a duplicate is a new requirement rather
	// than another alternative for the one being compared.
	assert.NotContains(t, copied, models.PurchaseOrderFieldOrderDeadline)
	assert.NotContains(t, copied, models.PurchaseOrderFieldSourcingGroupId)
}

// The allowlist must name only fields the schema has; a typo would silently drop a term from every
// duplicate.
func TestCopyableFieldsExistOnTheSchema(t *testing.T) {
	schema := models.PurchaseOrderSchemaBuilder().Build()

	source := dmodel.DynamicFields{}
	for _, field := range []string{
		models.PurchaseOrderFieldVendorId,
		models.PurchaseOrderFieldBuyerId,
		models.PurchaseOrderFieldCurrencyId,
		models.PurchaseOrderFieldVendorReference,
		models.PurchaseOrderFieldSourceReference,
		models.PurchaseOrderFieldExpectedArrival,
		models.PurchaseOrderFieldPriority,
		models.PurchaseOrderFieldTermsConditions,
		models.PurchaseOrderFieldAgreementId,
		basemodel.FieldOrgId,
	} {
		source[field] = "value"
	}

	for field := range copyableOrderFields(source) {
		_, exists := schema.Field(field)
		assert.True(t, exists, "the duplicate allowlist names %q, which purchase_order does not have", field)
	}
}

// formatStatus stands in for a one-verb fmt.Sprintf, so it must behave like one on the cases the
// refusal messages use.
func TestFormatStatus(t *testing.T) {
	assert.Equal(t, "this one is 'rfq'", formatStatus("this one is '%s'", "rfq"))
	assert.Equal(t, "no verb here", formatStatus("no verb here", "rfq"))
	assert.Equal(t, "rfq at the front", formatStatus("%s at the front", "rfq"))
	// Only the first is substituted, which is all these messages need.
	assert.Equal(t, "rfq and %s", formatStatus("%s and %s", "rfq"))
	// A trailing bare % must not run off the end of the string.
	assert.Equal(t, "trailing %", formatStatus("trailing %", "rfq"))
}
