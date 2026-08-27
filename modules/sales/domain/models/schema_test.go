package models

import (
	"slices"
	"strings"
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// These tests pin the schema declarations against the rules the requirement states, so that an
// edit which quietly changes a constraint fails here rather than in production data.

// requireBaseSchemasRegistered registers the core.basemodel.* schemas the Sales models extend.
// Normally done by CoreModule.RegisterModels during app start-up; without it ParseModelJson
// panics on the first "extend_before" it cannot resolve.
func requireBaseSchemasRegistered(t *testing.T) {
	t.Helper()
	_ = basemodel.RegisterJsonBaseSchemas()
}

func buildSchema(t *testing.T, builder *dmodel.ModelSchemaBuilder) *dmodel.ModelSchema {
	t.Helper()
	schema := builder.Build()
	if schema == nil {
		t.Fatal("Build returned nil")
	}
	return schema
}

func fieldOf(t *testing.T, schema *dmodel.ModelSchema, name string) *dmodel.ModelField {
	t.Helper()
	field, ok := schema.Fields()[name]
	if !ok || field == nil {
		t.Fatalf("schema %q has no field %q", schema.Name(), name)
	}
	return field
}

// TestSalesChannelCodeIsMandatoryAndUnique pins DEC-001.
//
// The requirement originally made code nullable and "unique when not NULL". That could not be
// expressed: the partial_uniques builder scopes a required field by a nullable one, so a
// single-column version would also emit a unique index over the remaining fields where code IS
// NULL — capping the table at one NULL-code channel and contradicting the acceptance criterion
// that many were allowed. Code is therefore required, immutable and plainly unique.
func TestSalesChannelCodeIsMandatoryAndUnique(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesChannelSchemaBuilder())
	code := fieldOf(t, schema, SalesChannelFieldCode)

	if !code.IsRequiredForCreate() {
		t.Error("code must be required_for_create (DEC-001)")
	}
	if !code.IsUnique() {
		t.Error("code must be unique (DEC-001)")
	}
	if !code.IsNoUpdate() {
		t.Error("code must be no_update: it is a published integration contract (CR 7.2, 41)")
	}
	if len(schema.PartialUniques()) != 0 {
		t.Error("sales_channel must not use a partial unique of either kind; DEC-001 records why " +
			"a plain unique on a required column is what this constraint needs")
	}
}

// TestSalesPointUniquesAreStrict guards against a bug this module already made once.
//
// A LOOSE group emits a PAIR of indexes: one over not_null_fields + nullable_field where the
// nullable IS NOT NULL, and one over not_null_fields ALONE where it IS NULL. Declaring
// {not_null: [sales_channel_id], nullable: external_reference_id} as loose therefore also creates
// a unique index on sales_channel_id alone for rows with no external reference — allowing exactly
// one manually-created sales point per channel. The same trap applies to code.
//
// Both rules must be STRICT, which emits only the IS NOT NULL half.
func TestSalesPointUniquesAreStrict(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesPointSchemaBuilder())

	if got := len(schema.PartialUniquesLoose()); got != 0 {
		t.Errorf("sales_point declares %d LOOSE partial uniques; both must be strict — a loose "+
			"group also emits a unique index over sales_channel_id alone for NULL rows, capping "+
			"a channel at one point without an external_reference_id or without a code", got)
	}

	strict := schema.PartialUniquesStrict()
	if len(strict) != 2 {
		t.Fatalf("sales_point declares %d strict partial uniques, want 2 "+
			"(external_reference_id and code)", len(strict))
	}

	nullables := map[string]bool{}
	for _, group := range strict {
		nullables[group.NullableField] = true
		if len(group.NotNullFields) != 1 || group.NotNullFields[0] != SalesPointFieldSalesChannelId {
			t.Errorf("strict group on %q must be scoped by %q alone, got %v",
				group.NullableField, SalesPointFieldSalesChannelId, group.NotNullFields)
		}
	}
	for _, want := range []string{SalesPointFieldExternalReferenceId, SalesPointFieldCode} {
		if !nullables[want] {
			t.Errorf("no strict partial unique on %q", want)
		}
	}
}

// TestSalesPointExternalReferenceIsQualified pins the id/type pair.
//
// An external_reference_id alone is ambiguous: it is another module's ulid, and more than one
// module may register points on the same channel, so nothing in the row says which resource to
// resolve it against. The type carries that, as a readable "{module}.{resource}" pair.
func TestSalesPointExternalReferenceIsQualified(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesPointSchemaBuilder())

	refId := fieldOf(t, schema, SalesPointFieldExternalReferenceId)
	if refId.IsRequiredForCreate() {
		t.Error("external_reference_id must stay optional: a point created by a human has none")
	}
	if !refId.IsNoUpdate() {
		t.Error("external_reference_id must be no_update: it is the idempotency key a retrying " +
			"caller resolves against (CR 48)")
	}

	refType := fieldOf(t, schema, SalesPointFieldExternalReferenceType)
	if refType.IsRequiredForCreate() {
		t.Error("external_reference_type must stay optional: it qualifies an optional id")
	}
	if !refType.IsNoUpdate() {
		t.Error("external_reference_type must be no_update: changing it would repoint the reference")
	}

	if KioskReferenceType != "vending_machine.kiosk" {
		t.Errorf("KioskReferenceType = %q, want %q", KioskReferenceType, "vending_machine.kiosk")
	}
}

// TestSalesPointImmutableFields pins the two fields a sales point may never change.
func TestSalesPointImmutableFields(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesPointSchemaBuilder())

	channelId := fieldOf(t, schema, SalesPointFieldSalesChannelId)
	if !channelId.IsNoUpdate() {
		t.Error("sales_channel_id must be no_update: a point is archived and recreated rather " +
			"than moved between channels (CR 72)")
	}
	if !channelId.IsRequiredForCreate() {
		t.Error("sales_channel_id must be required_for_create: every point belongs to a channel")
	}

}

// TestSalesPointCodeIsOptionalAndMutable states the deliberate contrast with a channel code.
//
// A channel code is an integration identity, frozen once published. A sales point code is a
// display reference nothing resolves by, so it can be corrected (CR 13).
func TestSalesPointCodeIsOptionalAndMutable(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesPointSchemaBuilder())
	code := fieldOf(t, schema, SalesPointFieldCode)

	if code.IsRequiredForCreate() {
		t.Error("sales point code must be optional (CR 13)")
	}
	if code.IsNoUpdate() {
		t.Error("sales point code must be mutable: it is a display code, not an identity (CR 13)")
	}
	if code.IsUnique() {
		t.Error("sales point code must not carry a table-wide unique: it is unique per channel, " +
			"enforced by a strict partial unique")
	}
}

// TestStatusValuesMatchConstants keeps the JSON enum and the Go constants from drifting.
//
// A drift here is invisible: a comparison against a constant that no longer appears in the enum is
// simply never true, and no compiler or schema check catches it.
func TestStatusValuesMatchConstants(t *testing.T) {
	requireBaseSchemasRegistered(t)

	cases := []struct {
		schemaName string
		builder    *dmodel.ModelSchemaBuilder
		field      string
		want       []string
	}{
		{
			SalesChannelSchemaName, SalesChannelSchemaBuilder(), SalesChannelFieldStatus,
			[]string{
				string(SalesChannelStatusActive),
				string(SalesChannelStatusSuspended),
			},
		},
		{
			SalesPointSchemaName, SalesPointSchemaBuilder(), SalesPointFieldStatus,
			[]string{
				string(SalesPointStatusActive),
				string(SalesPointStatusSuspended),
			},
		},
		{
			SalesVoucherCodeSchemaName, SalesVoucherCodeSchemaBuilder(),
			SalesVoucherCodeFieldStatus,
			[]string{
				string(VoucherCodeStatusActive),
				string(VoucherCodeStatusDisabled),
				string(VoucherCodeStatusExhausted),
			},
		},
		{
			SalesBillSchemaName, SalesBillSchemaBuilder(), SalesBillFieldStatus,
			[]string{
				string(SalesBillStatusOpen),
				string(SalesBillStatusSettled),
				string(SalesBillStatusCancelled),
			},
		},
		{
			SalesBillRelationSchemaName, SalesBillRelationSchemaBuilder(),
			SalesBillRelationFieldType,
			[]string{
				string(SalesBillRelationSplitInto),
				string(SalesBillRelationMergedInto),
			},
		},
		{
			SalesVoucherRedemptionSchemaName, SalesVoucherRedemptionSchemaBuilder(),
			SalesVoucherRedemptionFieldStatus,
			[]string{
				string(VoucherRedemptionStatusReserved),
				string(VoucherRedemptionStatusRedeemed),
				string(VoucherRedemptionStatusReleased),
				string(VoucherRedemptionStatusReversed),
			},
		},
	}

	for _, tc := range cases {
		schema := buildSchema(t, tc.builder)
		field := fieldOf(t, schema, tc.field)
		declared := enumValuesOf(t, field)
		if len(declared) != len(tc.want) {
			t.Errorf("%s.%s: schema declares %v but the Go constants are %v",
				tc.schemaName, tc.field, declared, tc.want)
			continue
		}
		for _, want := range tc.want {
			if !slices.Contains(declared, want) {
				t.Errorf("%s.%s: Go constant %q is absent from the schema enum %v",
					tc.schemaName, tc.field, want, declared)
			}
		}
	}
}

// enumValuesOf reads the declared values of an enum_string field.
func enumValuesOf(t *testing.T, field *dmodel.ModelField) []string {
	t.Helper()
	raw, ok := field.DataType().Options()[dmodel.FieldDataTypeOptEnumValues]
	if !ok {
		t.Fatalf("field %q declares no enum values", field.Name())
	}
	switch values := raw.(type) {
	case []string:
		return values
	case []any:
		out := make([]string, 0, len(values))
		for _, item := range values {
			text, ok := item.(string)
			if !ok {
				t.Fatalf("field %q has a non-string enum value %v", field.Name(), item)
			}
			out = append(out, text)
		}
		return out
	default:
		t.Fatalf("field %q has unexpected enum value type %T", field.Name(), raw)
		return nil
	}
}

// TestSchemaNamesMatchTablePrefix pins the assumption cmd/application.go's schemaPrefixesOf relies
// on: every Sales schema starts with "sales_", so the module needs no explicit case there. A
// schema that broke this would silently emit no migration.
func TestSchemaNamesMatchTablePrefix(t *testing.T) {
	for _, name := range []string{SalesChannelSchemaName, SalesPointSchemaName} {
		if !strings.HasPrefix(name, "sales_") {
			t.Errorf("schema %q must start with \"sales_\"; otherwise add a case to "+
				"schemaPrefixesOf in cmd/application.go", name)
		}
	}
}

// TestChannelPaymentRelIsUniquePerPair pins the constraint that makes enabling idempotent.
//
// Without it a lost response followed by a retry writes a second mapping row, and disabling then
// removes only one of them — leaving a channel that still accepts a method an administrator
// believes they turned off. The application service checks first, but the check and the insert are
// not atomic; this constraint is what actually decides the race.
func TestChannelPaymentRelIsUniquePerPair(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesChannelPaymentRelSchemaBuilder())

	found := false
	for _, unique := range schema.CompositeUniques() {
		if len(unique.Fields) != 2 {
			continue
		}
		if slices.Contains(unique.Fields, SalesChannelPaymentRelFieldSalesChannelId) &&
			slices.Contains(unique.Fields, SalesChannelPaymentRelFieldPaymentMethodId) {
			found = true
		}
	}
	if !found {
		t.Error("sales_channel_payment_rel must be unique on " +
			"(sales_channel_id, payment_method_id): without it a retried enable writes a second row")
	}
}

// TestChannelPaymentRelHasNoEnabledFlag pins CR 27: the row IS the state.
//
// A boolean would create a second way to say "not enabled" — a row set to false versus no row —
// and every reader would then have to handle both with nothing making them agree. It would also
// break the default-deny rule, since a reader finding no row would have to decide what the absence
// meant.
func TestChannelPaymentRelHasNoEnabledFlag(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesChannelPaymentRelSchemaBuilder())
	for name := range schema.Fields() {
		if strings.Contains(name, "enabled") || strings.Contains(name, "is_active") {
			t.Errorf("sales_channel_payment_rel must carry no %q flag (CR 27): "+
				"the presence of the row is the state", name)
		}
	}
}

// TestChannelPaymentRelKeysAreImmutable guards against a mapping being repointed by an update.
//
// Moving a mapping between channels or between methods is the same thing as disabling one and
// enabling the other, and doing it as an update would skip the validation the enable path runs
// against paymentinvoice.
func TestChannelPaymentRelKeysAreImmutable(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesChannelPaymentRelSchemaBuilder())
	for _, name := range []string{
		SalesChannelPaymentRelFieldSalesChannelId,
		SalesChannelPaymentRelFieldPaymentMethodId,
	} {
		if !fieldOf(t, schema, name).IsNoUpdate() {
			t.Errorf("%s must be no_update: repointing a mapping would skip the enable validation",
				name)
		}
	}
}

// TestChannelPaymentRelHasNoEdgeToPaymentMethod pins CR 25.
//
// payment_method_id names a row in another module. A foreign key across that boundary would make
// Sales unable to write a mapping while paymentinvoice was mid-migration, and would make retiring a
// payment method fail on Sales data. The cost is that a mapping can outlive its method, which is
// handled explicitly as the stale case rather than prevented.
func TestChannelPaymentRelHasNoEdgeToPaymentMethod(t *testing.T) {
	requireBaseSchemasRegistered(t)

	schema := buildSchema(t, SalesChannelPaymentRelSchemaBuilder())
	for _, edge := range schema.ToRelations() {
		if strings.Contains(edge.DestSchemaName, "payment") {
			t.Errorf("sales_channel_payment_rel must not declare edge %q onto %q (CR 25): "+
				"a constraint across a module boundary couples the two schemas' migrations",
				edge.Edge, edge.DestSchemaName)
		}
	}
}
