package composable

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// managedDeleteMode says which records a delete accepts, by their managed_by_type.
type managedDeleteMode int

const (
	// managedDeleteRefuse is Delete: a managed record belongs to its manager, which removes it
	// through DeleteManaged when the managing record goes.
	managedDeleteRefuse managedDeleteMode = iota
	// managedDeleteRequire is DeleteManaged: it must not become a back door to client records.
	managedDeleteRequire
)

// withManagedDeleteRule chains the managed check after the caller's own delete rule, which runs
// first so a resource-specific refusal (e.g. iam's system-managed role) keeps its own error. A
// schema without managed_by_type has no managed records: Delete runs unchanged on it, and
// DeleteManaged refuses every record.
func (this *DefaultDomainServiceImpl) withManagedDeleteRule(
	mode managedDeleteMode, hook ValidateExtraFn,
) ValidateExtraFn {
	_, hasField := this.schema.Field(basemodel.FieldManagedByType)
	if mode == managedDeleteRefuse && !hasField {
		return hook
	}
	return func(
		ctx corectx.Context, input *DynamicEntity, found *DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		if hook != nil {
			if err := hook(ctx, input, found, vErrs); err != nil || vErrs.Count() > 0 {
				return err
			}
		}
		assertManagedDelete(mode, found, vErrs)
		return nil
	}
}

func assertManagedDelete(mode managedDeleteMode, found *DynamicEntity, vErrs *ft.ClientErrors) {
	isManaged := found != nil && basemodel.IsManaged(found.GetFieldData())
	switch {
	case mode == managedDeleteRefuse && isManaged:
		vErrs.Append(*ft.NewBusinessViolation(basemodel.FieldManagedByType,
			ft.ErrorKey("err_record_managed_delete"), "a managed record is deleted by its manager only"))
	case mode == managedDeleteRequire && !isManaged:
		vErrs.Append(*ft.NewBusinessViolation(basemodel.FieldManagedByType,
			ft.ErrorKey("err_record_not_managed"), "only a managed record is deleted this way"))
	}
}
