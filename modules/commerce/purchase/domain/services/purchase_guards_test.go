package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// The guards receive the stored record as foundModel, the third argument, which is where the
// domain service's Delete hands it over.

// The audit trail is written by the system alone: a client-written event would sit in the same
// table as the real ones with no way to tell them apart.
func TestAuditEventWritesAreRefused(t *testing.T) {
	vErrs := &ft.ClientErrors{}

	err := rejectAuditEventWrite(nil, nil, composable.NewDynamicEntityFrom(dmodel.DynamicFields{}), vErrs)

	require.NoError(t, err, "a refusal is a client error, not a Go error")
	assert.Equal(t, 1, vErrs.Count())
}

// The sourcing group is created by adding an alternative and reaped when fewer than two remain, so
// a hand-made one would be an empty container nothing reaps.
func TestSourcingGroupWritesAreRefused(t *testing.T) {
	vErrs := &ft.ClientErrors{}

	err := rejectSourcingGroupWrite(nil, nil, composable.NewDynamicEntityFrom(dmodel.DynamicFields{}), vErrs)

	require.NoError(t, err)
	assert.Equal(t, 1, vErrs.Count())
}

// Only a cancelled order may be deleted: deleting a confirmed one would remove the evidence that
// the business committed to a purchase, where cancelling records that it did and then stopped.
func TestOrderDeleteIsRefusedUnlessCancelled(t *testing.T) {
	testCases := []struct {
		status  models.PurchaseOrderStatus
		allowed bool
	}{
		{models.PurchaseOrderStatusCancelled, true},
		{models.PurchaseOrderStatusRfq, false},
		{models.PurchaseOrderStatusRfqSent, false},
		{models.PurchaseOrderStatusToApprove, false},
		{models.PurchaseOrderStatusPurchaseOrder, false},
	}

	for _, testCase := range testCases {
		t.Run(string(testCase.status), func(t *testing.T) {
			vErrs := &ft.ClientErrors{}
			found := dmodel.DynamicFields{
				models.PurchaseOrderFieldStatus: string(testCase.status),
			}

			err := GuardOrderDelete(nil, nil, composable.NewDynamicEntityFrom(found), vErrs)

			require.NoError(t, err)
			if testCase.allowed {
				assert.Equal(t, 0, vErrs.Count(), "a cancelled order must be deletable")
			} else {
				assert.Equal(t, 1, vErrs.Count(), "%s must not be deletable", testCase.status)
			}
		})
	}
}

// A draft agreement is deletable where a draft order is not, because an agreement's code is not
// quoted to a vendor until it is confirmed.
func TestAgreementDeleteIsRefusedUnlessDraftOrCancelled(t *testing.T) {
	testCases := []struct {
		status  models.AgreementStatus
		allowed bool
	}{
		{models.AgreementStatusDraft, true},
		{models.AgreementStatusCancelled, true},
		{models.AgreementStatusConfirmed, false},
		{models.AgreementStatusClosed, false},
	}

	for _, testCase := range testCases {
		t.Run(string(testCase.status), func(t *testing.T) {
			vErrs := &ft.ClientErrors{}
			found := dmodel.DynamicFields{
				models.AgreementFieldStatus: string(testCase.status),
			}

			err := GuardAgreementDelete(nil, nil, composable.NewDynamicEntityFrom(found), vErrs)

			require.NoError(t, err)
			if testCase.allowed {
				assert.Equal(t, 0, vErrs.Count(), "%s must be deletable", testCase.status)
			} else {
				assert.Equal(t, 1, vErrs.Count(), "%s must not be deletable", testCase.status)
			}
		})
	}
}

// The guard must fail closed: defaulting to "delete it" on an unreadable status is the one failure
// mode a guard must not have.
func TestDeleteGuardsRefuseAnUnreadableStatus(t *testing.T) {
	for _, guard := range []struct {
		name string
		fn   func(dmodel.DynamicFields, *ft.ClientErrors) error
	}{
		{"order", func(f dmodel.DynamicFields, v *ft.ClientErrors) error {
			return GuardOrderDelete(nil, nil, composable.NewDynamicEntityFrom(f), v)
		}},
		{"agreement", func(f dmodel.DynamicFields, v *ft.ClientErrors) error {
			return GuardAgreementDelete(nil, nil, composable.NewDynamicEntityFrom(f), v)
		}},
	} {
		t.Run(guard.name, func(t *testing.T) {
			vErrs := &ft.ClientErrors{}

			err := guard.fn(dmodel.DynamicFields{}, vErrs)

			require.NoError(t, err)
			assert.Equal(t, 1, vErrs.Count(), "a missing status must refuse the delete, not allow it")
		})
	}
}

// The status strings the guards compare against must be the ones the schema declares; a typo is a
// condition silently never true, which no compiler catches.
func TestGuardStatusesAreDeclaredByTheSchemas(t *testing.T) {
	orderStatuses := enumValues(t, models.PurchaseOrderSchemaBuilder(), models.PurchaseOrderFieldStatus)
	for status := range deletableOrderStatuses {
		assert.Contains(t, orderStatuses, status)
	}

	agreementStatuses := enumValues(t, models.AgreementSchemaBuilder(), models.AgreementFieldStatus)
	for status := range deletableAgreementStatuses {
		assert.Contains(t, agreementStatuses, status)
	}
}

func enumValues(t *testing.T, builder *dmodel.ModelSchemaBuilder, fieldName string) []string {
	t.Helper()
	field, ok := builder.Build().Field(fieldName)
	require.True(t, ok)
	raw, ok := field.DataType().Options()["enumValues"]
	require.True(t, ok)
	values, ok := raw.([]string)
	require.True(t, ok)
	return values
}

// An update is validated as the record will be: the change overlays the stored row, never the
// other way round, or a price change would be validated against the price it replaces.
func TestTheChangeOverlaysTheStoredVendorPrice(t *testing.T) {
	merged := MergedRecord(
		dmodel.DynamicFields{"unit_price": "10", "vendor_id": "01VENDOR"},
		dmodel.DynamicFields{"unit_price": "12"},
	)

	assert.Equal(t, "12", merged["unit_price"])
	assert.Equal(t, "01VENDOR", merged["vendor_id"])
}

// A create has no stored row, and must still be validated: the input alone is the record.
func TestACreatedVendorPriceIsValidated(t *testing.T) {
	svc := &VendorProductPriceDomainServiceImpl{validator: NewVendorPriceValidator(nil, nil, nil)}
	vErrs := &ft.ClientErrors{}

	err := svc.validateVendorPrice(nil, composable.NewDynamicEntityFrom(dmodel.DynamicFields{
		models.VendorProductPriceFieldValidFrom: "2026-02-01T00:00:00Z",
		models.VendorProductPriceFieldValidTo:   "2026-01-01T00:00:00Z",
	}), nil, vErrs)

	require.NoError(t, err)
	assert.Positive(t, vErrs.Count(), "a create must reach the validator")
}
