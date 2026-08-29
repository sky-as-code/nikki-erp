package dynamicengines

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"

	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
)

const statusField = models.TaxDefinitionVersionFieldLifecycleStatus

func storedWithStatus(status models.LifecycleStatus) *dmodel.DynamicFields {
	found := dmodel.DynamicFields{statusField: string(status)}
	return &found
}

// hasViolation lets a test assert the specific refusal rather than merely that something was
// refused.
func hasViolation(cErrs *ft.ClientErrors, key string) bool {
	for _, item := range *cErrs {
		if item.Key == key {
			return true
		}
	}
	return false
}

// A draft is freely editable.
func TestDraftMaterialFieldsAreEditable(t *testing.T) {
	cErrs := ft.NewClientErrors()
	params := dmodel.DynamicFields{
		models.TaxDefinitionVersionFieldCalculationType: string(models.CalculationPercentage),
		models.TaxDefinitionVersionFieldTaxTreatment:    string(models.TaxTreatmentTaxable),
	}

	assertMaterialFieldsImmutable(models.TaxDefinitionVersionSchemaName, statusField,
		params, storedWithStatus(models.LifecycleDraft), cErrs)

	if cErrs.Count() > 0 {
		t.Fatalf("expected a draft to be editable, got %v", cErrs.ToError())
	}
}

// Once published, every field that decides an amount is frozen; editing one in place would
// silently reinterpret transactions already priced under it.
func TestPublishedMaterialFieldsAreImmutable(t *testing.T) {
	for _, field := range materialFieldsBySchema[models.TaxDefinitionVersionSchemaName] {
		cErrs := ft.NewClientErrors()
		params := dmodel.DynamicFields{field: "anything"}

		assertMaterialFieldsImmutable(models.TaxDefinitionVersionSchemaName, statusField,
			params, storedWithStatus(models.LifecyclePublished), cErrs)

		if !hasViolation(cErrs, "tax.published_field_immutable") {
			t.Errorf("expected published field %q to be immutable", field)
		}
	}
}

// Descriptive fields stay editable after publication.
func TestPublishedDescriptiveFieldsStayEditable(t *testing.T) {
	cErrs := ft.NewClientErrors()
	params := dmodel.DynamicFields{
		models.TaxDefinitionVersionFieldLegalReference: "corrected citation",
	}

	assertMaterialFieldsImmutable(models.TaxDefinitionVersionSchemaName, statusField,
		params, storedWithStatus(models.LifecyclePublished), cErrs)

	if cErrs.Count() > 0 {
		t.Fatalf("expected a descriptive field to stay editable, got %v", cErrs.ToError())
	}
}

// The rate gets its own assertion on its own schema rather than relying on the loop above.
func TestPublishedRateIsImmutable(t *testing.T) {
	cErrs := ft.NewClientErrors()
	params := dmodel.DynamicFields{models.TaxRateVersionFieldRate: "8"}
	found := dmodel.DynamicFields{models.TaxRateVersionFieldLifecycleStatus: string(models.LifecyclePublished)}

	assertMaterialFieldsImmutable(models.TaxRateVersionSchemaName,
		models.TaxRateVersionFieldLifecycleStatus, params, &found, cErrs)

	if !hasViolation(cErrs, "tax.published_field_immutable") {
		t.Fatalf("expected a published rate to be immutable, got %v", cErrs.ToError())
	}
}

func TestLifecycleTransitionsThatAreAllowed(t *testing.T) {
	allowed := []struct {
		from models.LifecycleStatus
		to   models.LifecycleStatus
	}{
		{models.LifecycleDraft, models.LifecyclePublished},
		{models.LifecycleDraft, models.LifecycleWithdrawn},
		{models.LifecyclePublished, models.LifecycleWithdrawn},
	}

	for _, transition := range allowed {
		cErrs := ft.NewClientErrors()
		params := dmodel.DynamicFields{statusField: string(transition.to)}

		assertLifecycleTransition(statusField, params, storedWithStatus(transition.from), cErrs)

		if cErrs.Count() > 0 {
			t.Errorf("expected %s -> %s to be allowed, got %v",
				transition.from, transition.to, cErrs.ToError())
		}
	}
}

// Withdrawn is terminal: returning to draft would reopen closed history, and republishing would
// silently re-price transactions.
func TestWithdrawnIsTerminal(t *testing.T) {
	for _, target := range []models.LifecycleStatus{models.LifecycleDraft, models.LifecyclePublished} {
		cErrs := ft.NewClientErrors()
		params := dmodel.DynamicFields{statusField: string(target)}

		assertLifecycleTransition(statusField, params, storedWithStatus(models.LifecycleWithdrawn), cErrs)

		if !hasViolation(cErrs, "tax.invalid_lifecycle_transition") {
			t.Errorf("expected withdrawn -> %s to be refused", target)
		}
	}
}

// Publication is not reversible.
func TestPublishedCannotReturnToDraft(t *testing.T) {
	cErrs := ft.NewClientErrors()
	params := dmodel.DynamicFields{statusField: string(models.LifecycleDraft)}

	assertLifecycleTransition(statusField, params, storedWithStatus(models.LifecyclePublished), cErrs)

	if !hasViolation(cErrs, "tax.invalid_lifecycle_transition") {
		t.Fatalf("expected published -> draft to be refused, got %v", cErrs.ToError())
	}
}

// A no-op status write is not a transition, so a client resubmitting the record unchanged must not
// be refused.
func TestUnchangedStatusIsNotATransition(t *testing.T) {
	cErrs := ft.NewClientErrors()
	params := dmodel.DynamicFields{statusField: string(models.LifecyclePublished)}

	assertLifecycleTransition(statusField, params, storedWithStatus(models.LifecyclePublished), cErrs)

	if cErrs.Count() > 0 {
		t.Fatalf("expected an unchanged status to pass, got %v", cErrs.ToError())
	}
}

// A draft has priced nothing, so deleting it destroys no history.
func TestDraftIsDeletable(t *testing.T) {
	cErrs := ft.NewClientErrors()

	assertDeletableLifecycle(statusField, storedWithStatus(models.LifecycleDraft), cErrs)

	if cErrs.Count() > 0 {
		t.Fatalf("expected a draft to be deletable, got %v", cErrs.ToError())
	}
}

// Publication is the boundary Tax enforces: anything ever published may have priced a transaction
// whose snapshot still references it.
func TestPublishedAndWithdrawnAreNotDeletable(t *testing.T) {
	for _, status := range []models.LifecycleStatus{models.LifecyclePublished, models.LifecycleWithdrawn} {
		cErrs := ft.NewClientErrors()

		assertDeletableLifecycle(statusField, storedWithStatus(status), cErrs)

		if !hasViolation(cErrs, "tax.published_not_deletable") {
			t.Errorf("expected %s configuration to be undeletable", status)
		}
	}
}

// A schema reaching the lifecycle rules with an empty material list would accept edits to a
// published configuration while appearing to be governed.
func TestEveryLifecycleResourceDeclaresMaterialFields(t *testing.T) {
	lifecycleSchemas := []string{
		models.TaxDefinitionVersionSchemaName,
		models.TaxRateVersionSchemaName,
		models.TaxRoundingPolicySchemaName,
		models.TaxMappingSchemaName,
		models.TaxRuleSchemaName,
	}

	for _, schemaName := range lifecycleSchemas {
		if len(materialFieldsBySchema[schemaName]) == 0 {
			t.Errorf("schema %q has a lifecycle but freezes no fields on publication", schemaName)
		}
	}
}
