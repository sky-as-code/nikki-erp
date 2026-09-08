package models

import (
	"slices"
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
)

// These tests pin the fulfillment method schema against the rules that make a snapshot meaningful.
// The recurring theme: a method is copied onto the fulfillment that uses it, so what may change
// afterwards, and what may not, decides whether history stays readable.

// TestFulfillmentMethodCodeIsMandatoryUniqueAndImmutable. The code is what seeds, channel
// configuration and integration callers name a method by, rather than a database id. A plain unique
// on a required column for the reason recorded on sales_channel.code: the partial builder scopes a
// required field by a nullable one, so a single-column partial would also cap the table at one
// NULL-code row.
func TestFulfillmentMethodCodeIsMandatoryUniqueAndImmutable(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesFulfillmentMethodSchemaBuilder())
	code := fieldOf(t, schema, SalesFulfillmentMethodFieldCode)

	if !code.IsRequiredForCreate() {
		t.Error("code must be required_for_create: it is how a method is named by everything " +
			"outside the database")
	}
	if !code.IsUnique() {
		t.Error("code must be unique: two methods answering to one code makes resolution ambiguous")
	}
	if !code.IsNoUpdate() {
		t.Error("code must be no_update: renaming it would repoint every seed and caller that " +
			"resolved the old spelling at nothing")
	}
}

// TestFulfillmentTypeIsImmutable. The type chooses which execution workflow runs, so changing it
// under a live fulfillment would reinterpret snapshots taken under one set of rules against
// another's.
func TestFulfillmentTypeIsImmutable(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesFulfillmentMethodSchemaBuilder())
	fulfillmentType := fieldOf(t, schema, SalesFulfillmentMethodFieldFulfillmentType)

	if !fulfillmentType.IsNoUpdate() {
		t.Error("fulfillment_type must be no_update: it selects the workflow, and a method that " +
			"changed type mid-life would run one policy's fulfillments under another's rules")
	}
	if !fulfillmentType.IsRequiredForCreate() {
		t.Error("fulfillment_type must be required_for_create: there is no sensible default " +
			"between dispensing at a machine and shipping by carrier")
	}
}

// TestFulfillmentMethodEnumsMatchConstants keeps the four new enums and their Go constants in step.
// Drift is invisible otherwise: a comparison against a constant absent from the enum is simply
// never true, and the policy it guards silently never applies.
func TestFulfillmentMethodEnumsMatchConstants(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesFulfillmentMethodSchemaBuilder())

	cases := []struct {
		field string
		want  []string
	}{
		{
			SalesFulfillmentMethodFieldFulfillmentType,
			[]string{
				string(FulfillmentTypeKioskDispense),
				string(FulfillmentTypeCarrierShipping),
				string(FulfillmentTypeInternalDelivery),
				string(FulfillmentTypeStorePickup),
				string(FulfillmentTypeExternalFulfillment),
			},
		},
		{
			SalesFulfillmentMethodFieldInitialTargetSelection,
			[]string{
				string(InitialTargetSelectionCurrentSalesOutlet),
				string(InitialTargetSelectionCustomerSelectedOutlet),
			},
		},
		{
			SalesFulfillmentMethodFieldFailureAction,
			[]string{
				string(FulfillmentFailureActionAutoRefund),
				string(FulfillmentFailureActionCustomerActionRequired),
				string(FulfillmentFailureActionManualResolution),
			},
		},
	}

	for _, tc := range cases {
		field := fieldOf(t, schema, tc.field)
		declared := enumValuesOf(t, field)
		if len(declared) != len(tc.want) {
			t.Errorf("%s: schema declares %v but the Go constants are %v",
				tc.field, declared, tc.want)
			continue
		}
		for _, want := range tc.want {
			if !slices.Contains(declared, want) {
				t.Errorf("%s: Go constant %q is absent from the schema enum %v",
					tc.field, want, declared)
			}
		}
	}
}

// TestFulfillmentCountersRejectZero. Both nullable counters mean "no limit" when absent, so zero
// must not be reachable: a method allowing zero attempts could never deliver anything, and a
// reservation valid for zero minutes expires before the customer has walked to the machine. Not
// offering the method is how "never" is expressed.
func TestFulfillmentCountersRejectZero(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesFulfillmentMethodSchemaBuilder())

	for _, name := range []string{
		SalesFulfillmentMethodFieldMaxAttempts,
		SalesFulfillmentMethodFieldReservationTtlMinutes,
	} {
		field := fieldOf(t, schema, name)
		if field.IsRequiredForCreate() {
			t.Errorf("%s must stay optional: NULL is what 'no limit' means", name)
		}
		if got := minimumOf(t, field); got != 1 {
			t.Errorf("%s declares minimum %v, want 1: zero would be a policy that can never "+
				"succeed, which is what not offering the method already says", name, got)
		}
	}
}

// TestChannelFulfillmentMappingIsUniquePerPair pins the constraint that makes allowing a method
// idempotent. Without it a retried call writes a second row and disallowing removes only one.
func TestChannelFulfillmentMappingIsUniquePerPair(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesChannelFulfillmentMethodSchemaBuilder())

	found := false
	for _, unique := range schema.CompositeUniques() {
		if len(unique.Fields) != 2 {
			continue
		}
		if slices.Contains(unique.Fields, SalesChannelFulfillmentMethodFieldSalesChannelId) &&
			slices.Contains(unique.Fields, SalesChannelFulfillmentMethodFieldFulfillmentMethodId) {
			found = true
		}
	}
	if !found {
		t.Error("sales_channel_fulfillment_method must be unique on " +
			"(sales_channel_id, fulfillment_method_id): without it a retried allow writes a " +
			"second row and a later disallow removes only one of them")
	}
}

// TestChannelFulfillmentMappingHasNoEnabledFlag: the row is the state, exactly as it is for the
// payment mapping beside it. A boolean would create two ways to say "not allowed" — a row set false
// versus no row — and break default-deny.
func TestChannelFulfillmentMappingHasNoEnabledFlag(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesChannelFulfillmentMethodSchemaBuilder())
	for name := range schema.Fields() {
		if name == "is_enabled" || name == "enabled" || name == "is_active" {
			t.Errorf("sales_channel_fulfillment_method must carry no %q flag: the presence of "+
				"the row is the state", name)
		}
	}
	if _, ok := schema.Fields()["is_archived"]; ok {
		t.Error("sales_channel_fulfillment_method must not be archivable: a mapping is deleted " +
			"rather than retired, and an archived row would be a third way to say 'not allowed'")
	}
}

// TestSalesPointFulfillmentTargetFields. A point becomes a fulfillment target only when somebody
// says so, and the Inventory location is a plain identifier: a foreign key onto another module's
// table would couple the two schemas' migrations, the rule sales_channel_payment_rel already
// follows.
func TestSalesPointFulfillmentTargetFields(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesPointSchemaBuilder())

	enabled := fieldOf(t, schema, SalesPointFieldFulfillmentEnabled)
	if !enabled.IsRequiredForCreate() {
		t.Error("fulfillment_enabled must be required_for_create so it defaults explicitly to " +
			"false; a nullable flag would make 'not configured' and 'not allowed' indistinguishable")
	}

	location := fieldOf(t, schema, SalesPointFieldInventoryLocationId)
	if location.IsRequiredForCreate() {
		t.Error("inventory_location_id must stay optional: most sales points never fulfil anything")
	}

	for _, edge := range schema.ToRelations() {
		if edge.DestSchemaName == "inventory_location" {
			t.Errorf("sales_point must not declare edge %q onto inventory_location: a constraint "+
				"across a module boundary couples the two schemas' migrations", edge.Edge)
		}
	}
}

// minimumOf reads the declared lower bound of a numeric field. The range arrives as the [min, max]
// pair the int32 builder stores, so the test asks for element zero rather than for a "min" key that
// does not exist.
func minimumOf(t *testing.T, field *dmodel.ModelField) int32 {
	t.Helper()
	raw, ok := field.DataType().Options()[dmodel.FieldDataTypeOptRange]
	if !ok {
		t.Fatalf("field %q declares no numeric range", field.Name())
	}
	limits, ok := raw.([]int32)
	if !ok || len(limits) != 2 {
		t.Fatalf("field %q has an unexpected range %v (%T)", field.Name(), raw, raw)
	}
	return limits[0]
}
