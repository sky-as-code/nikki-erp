package dynamicengines

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/accounting/domain/models"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// mergeFields overlays a submitted patch onto the stored row.
//
// An update carries only the fields the caller is changing, so a rule that spans several fields —
// "this period must not overlap", "this treatment must match that calculation type" — cannot be
// evaluated against the patch alone. Merging first means those rules see the record as it will be
// after the write rather than the fragment of it that arrived.
func mergeFields(found *dmodel.DynamicFields, params dmodel.DynamicFields) dmodel.DynamicFields {
	merged := dmodel.DynamicFields{}
	if found != nil {
		for key, value := range *found {
			merged[key] = value
		}
	}
	for key, value := range params {
		merged[key] = value
	}
	return merged
}

// assertWellFormedPeriod rejects a period that ends before it starts.
//
// An inverted period is not merely odd: it matches no date at all, so a rate configured that way
// would silently never apply and the failure would surface much later as an unresolved tax on a
// transaction nobody connected to this row.
func assertWellFormedPeriod(
	from *model.ModelDate, to *model.ModelDate, field string, vErrs *ft.ClientErrors,
) {
	if from == nil || to == nil {
		return
	}
	if !models.PeriodIsWellFormed(from, to) {
		vErrs.Append(*ft.NewBusinessViolation(field, "tax.period_ends_before_it_starts",
			"the effective end date must be on or after the effective start date"))
	}
}

// entityFields exposes a fetched entity's fields as a pointer, preserving nil.
//
// The validation hooks receive nil when there is no stored record; the assertions below distinguish
// that from a record whose fields happen to be empty, so the nil must survive the unwrapping.
func entityFields(entity *drif.DynamicEntity) *dmodel.DynamicFields {
	if entity == nil {
		return nil
	}
	fields := entity.GetFieldData()
	return &fields
}
