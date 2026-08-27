package dynamicengines

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
)

// TestMain registers the base mixins once for the whole package.
//
// Building any schema here parses its JSON, and ParseModelJson panics when a named mixin is not in
// the registry — normally CoreModule.RegisterModels does this during start-up.
func TestMain(m *testing.M) {
	_ = basemodel.RegisterJsonBaseSchemas()
	os.Exit(m.Run())
}

// Every schema this module registers must have an engine, and vice versa. The two lists are
// maintained by hand in different files, so a resource added to one and forgotten in the other
// would be a resource with no HTTP surface — or an engine for a schema that does not exist, which
// fails at boot.
func TestEveryResourceHasAnEngine(t *testing.T) {
	assert.ElementsMatch(t, []string{
		models.ConfigurationSchemaName,
		models.SourcingGroupSchemaName,
		models.AgreementSchemaName,
		models.AgreementLineSchemaName,
		models.PurchaseOrderSchemaName,
		models.PurchaseOrderLineSchemaName,
		models.AuditEventSchemaName,
	}, EngineSchemaNames())
}

// A default field set that names a field the schema does not have produces an engine that fails on
// every listing request. Cheap to check here, expensive to discover through the API.
func TestDefaultSearchFieldsExistOnTheirSchemas(t *testing.T) {
	builders := map[string]func() *dmodel.ModelSchemaBuilder{
		models.ConfigurationSchemaName:     models.ConfigurationSchemaBuilder,
		models.SourcingGroupSchemaName:     models.SourcingGroupSchemaBuilder,
		models.AgreementSchemaName:         models.AgreementSchemaBuilder,
		models.AgreementLineSchemaName:     models.AgreementLineSchemaBuilder,
		models.PurchaseOrderSchemaName:     models.PurchaseOrderSchemaBuilder,
		models.PurchaseOrderLineSchemaName: models.PurchaseOrderLineSchemaBuilder,
		models.AuditEventSchemaName:        models.AuditEventSchemaBuilder,
	}

	for _, spec := range engineSpecs {
		builder, ok := builders[spec.SchemaName]
		require.True(t, ok, "no builder for %s", spec.SchemaName)
		schema := builder().Build()

		assert.NotEmptyf(t, schema.DefaultSearchFields(),
			"%s declares no default_search_fields, so its listing returns only primary keys",
			spec.SchemaName)
		for _, fieldName := range schema.DefaultSearchFields() {
			_, exists := schema.Field(fieldName)
			assert.True(t, exists, "%s lists default search field %q, which it does not have",
				spec.SchemaName, fieldName)
		}
	}
}

// PUR-R6: the audit trail is written by the system alone. A client-written event would sit in the
// same table as the real ones with no way to tell them apart, which destroys the value of the trail
// rather than adding to it.
func TestAuditEventWritesAreRefused(t *testing.T) {
	vErrs := &ft.ClientErrors{}

	err := rejectAuditEventWrite(nil, drif.NewDynamicEntityFrom(dmodel.DynamicFields{}), nil, vErrs)

	require.NoError(t, err, "a refusal is a client error, not a Go error")
	assert.Equal(t, 1, vErrs.Count())
}

// The sourcing group is created by adding an alternative and reaped when fewer than two remain, so
// a hand-made one would be an empty container nothing reaps.
func TestSourcingGroupWritesAreRefused(t *testing.T) {
	vErrs := &ft.ClientErrors{}

	err := rejectSourcingGroupWrite(nil, drif.NewDynamicEntityFrom(dmodel.DynamicFields{}), nil, vErrs)

	require.NoError(t, err)
	assert.Equal(t, 1, vErrs.Count())
}

// BR 24: only a cancelled order may be deleted. Deleting a confirmed one would remove the evidence
// that the business committed to a purchase; cancelling records that it did and then stopped.
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

			err := guardOrderDelete(nil, drif.NewDynamicEntityFrom(found), nil, vErrs)

			require.NoError(t, err)
			if testCase.allowed {
				assert.Equal(t, 0, vErrs.Count(), "a cancelled order must be deletable")
			} else {
				assert.Equal(t, 1, vErrs.Count(), "%s must not be deletable", testCase.status)
			}
		})
	}
}

// BR 46: a draft agreement is deletable where a draft order is not, because an agreement's code is
// not quoted to a vendor until it is confirmed.
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

			err := guardAgreementDelete(nil, drif.NewDynamicEntityFrom(found), nil, vErrs)

			require.NoError(t, err)
			if testCase.allowed {
				assert.Equal(t, 0, vErrs.Count(), "%s must be deletable", testCase.status)
			} else {
				assert.Equal(t, 1, vErrs.Count(), "%s must not be deletable", testCase.status)
			}
		})
	}
}

// The guard must fail CLOSED. An unreadable status means something is wrong with the record or the
// fetch, and defaulting to "delete it" there is the one failure mode a guard must not have.
func TestDeleteGuardsRefuseAnUnreadableStatus(t *testing.T) {
	for _, guard := range []struct {
		name string
		fn   func(dmodel.DynamicFields, *ft.ClientErrors) error
	}{
		{"order", func(f dmodel.DynamicFields, v *ft.ClientErrors) error {
			return guardOrderDelete(nil, drif.NewDynamicEntityFrom(f), nil, v)
		}},
		{"agreement", func(f dmodel.DynamicFields, v *ft.ClientErrors) error {
			return guardAgreementDelete(nil, drif.NewDynamicEntityFrom(f), nil, v)
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

// The status strings the guards compare against must be the ones the schema actually declares. A
// typo here is a condition that is silently never true, which no compiler catches.
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
