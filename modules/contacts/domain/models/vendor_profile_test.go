package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The 1-1 is enforced by a composite unique on (party_id, org_id), following
// inventory_stock_product_config. A second profile row for one party would give a supplier two sets
// of payment terms with no way to say which applies.
//
// Note this CAN be a composite unique where the party's tax_id could not: both columns are
// requiredForCreate, which is exactly the condition the builder enforces.
func TestVendorProfileIsOnePerPartyPerOrg(t *testing.T) {
	requireBaseSchemasRegistered(t)
	schema := VendorProfileSchemaBuilder().Build()

	composites := schema.CompositeUniques()
	require.Len(t, composites, 1)
	assert.Equal(t,
		[]string{VendorProfileFieldPartyId, VendorProfileFieldOrgId},
		composites[0].Fields)
}

// party_id is immutable. Moving a profile to a different party would silently reassign every
// purchase order that named the first one.
func TestVendorProfilePartyIdIsImmutable(t *testing.T) {
	requireBaseSchemasRegistered(t)

	field, ok := VendorProfileSchemaBuilder().Build().Field(VendorProfileFieldPartyId)

	require.True(t, ok)
	assert.True(t, field.IsNoUpdate(), "party_id must be no_update")
	assert.True(t, field.IsRequiredForCreate(), "party_id must be required")
}

// default_currency_id crosses a module boundary, so it is a plain ulid with no edge. A foreign key
// here would make Contacts' schema depend on Essential's table.
func TestVendorProfileCurrencyIsNotAnEdge(t *testing.T) {
	requireBaseSchemasRegistered(t)
	schema := VendorProfileSchemaBuilder().Build()

	for _, relation := range schema.ToRelations() {
		assert.NotEqual(t, VendorProfileFieldDefaultCurrencyId, relation.SrcField,
			"default_currency_id must not be a foreign key: the currency belongs to Essential")
	}

	field, ok := schema.Field(VendorProfileFieldDefaultCurrencyId)
	require.True(t, ok)
	assert.False(t, field.IsRequiredForCreate(),
		"a vendor may be recorded before its currency is agreed")
}

// The profile hangs off the party and cannot outlive it.
func TestVendorProfileCascadesFromParty(t *testing.T) {
	requireBaseSchemasRegistered(t)
	relations := VendorProfileSchemaBuilder().Build().ToRelations()

	require.Len(t, relations, 1)
	assert.Equal(t, VendorProfileEdgeParty, relations[0].Edge)
	assert.Equal(t, PartySchemaName, relations[0].DestSchemaName)
	assert.Equal(t, dmodel.RelationCascadeCascade, relations[0].OnDelete)
}

// IsOrderable is the definition of "may be ordered from", kept in one place so that a caller
// comparing Status itself does not have to be found and changed when a fifth status is added.
func TestIsOrderable(t *testing.T) {
	testCases := []struct {
		name     string
		status   string
		archived bool
		want     bool
	}{
		{"active is orderable", "active", false, true},
		{"proposed is not: qualification is unfinished", "proposed", false, false},
		{"suspended is not", "suspended", false, false},
		{"blacklisted is not", "blacklisted", false, false},
		{"archived is not, even when active", "active", true, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			profile := NewVendorProfileFrom(dmodel.DynamicFields{
				VendorProfileFieldStatus:  testCase.status,
				basemodel.FieldIsArchived: testCase.archived,
				VendorProfileFieldPartyId: "01JPARTY00000000000000000",
			})

			assert.Equal(t, testCase.want, profile.IsOrderable())
		})
	}
}

// A profile with no status at all must not be orderable. The schema defaults it to "proposed", so
// this is only reachable through a row written directly — and defaulting to orderable would let an
// unqualified supplier through.
func TestIsOrderableIsFalseWithoutStatus(t *testing.T) {
	profile := NewVendorProfileFrom(dmodel.DynamicFields{
		VendorProfileFieldPartyId: "01JPARTY00000000000000000",
	})

	assert.Nil(t, profile.GetStatus())
	assert.False(t, profile.IsOrderable())
}

// The defaults a purchase order reads are all optional: a vendor may be recorded before its terms
// are agreed, so a caller must distinguish "not stated" from a value.
func TestVendorProfileDefaultsAreOptional(t *testing.T) {
	profile := NewVendorProfileFrom(dmodel.DynamicFields{
		VendorProfileFieldPartyId: "01JPARTY00000000000000000",
		VendorProfileFieldStatus:  string(VendorStatusActive),
	})

	assert.Nil(t, profile.GetDefaultCurrencyId())
	assert.Nil(t, profile.GetLeadTimeDays())
	assert.Equal(t, "", util.ValueOrZeroOf(profile.GetPaymentTerms()))
	assert.True(t, profile.IsOrderable(), "missing defaults must not make a vendor unorderable")
}
