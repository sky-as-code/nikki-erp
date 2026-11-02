package dynamicresource

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// managedMode says which records an update or delete accepts, by their managed_by_type.
type managedMode int

const (
	// managedRefuse is Update/Delete: a managed record belongs to its manager, which changes it
	// through UpdateManaged and removes it through DeleteManaged.
	managedRefuse managedMode = iota
	// managedRequire is UpdateManaged/DeleteManaged: they must not become a back door to client records.
	managedRequire
)

// withManagedRule chains the managed check after the caller's own rule, which runs first so a
// resource-specific refusal (e.g. iam's system-managed role) keeps its own error. A schema without
// managed_by_type has no managed records: Update/Delete run unchanged on it, and
// UpdateManaged/DeleteManaged refuse every record.
func (this *DefaultDomainServiceImpl) withManagedRule(
	mode managedMode, action CrudAction, hook ValidateExtraFn,
) ValidateExtraFn {
	_, hasField := this.schema.Field(basemodel.FieldManagedByType)
	if mode == managedRefuse && !hasField {
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
		assertManaged(mode, action, found, vErrs)
		return nil
	}
}

func assertManaged(mode managedMode, action CrudAction, found *DynamicEntity, vErrs *ft.ClientErrors) {
	isManaged := found != nil && basemodel.IsManaged(found.GetFieldData())
	switch {
	case mode == managedRefuse && isManaged:
		vErrs.Append(*ft.NewBusinessViolation(basemodel.FieldManagedByType,
			ft.ErrorKey("err_record_managed_"+string(action)),
			"a managed record is changed and deleted by its manager only"))
	case mode == managedRequire && !isManaged:
		vErrs.Append(*ft.NewBusinessViolation(basemodel.FieldManagedByType,
			ft.ErrorKey("err_record_not_managed"), "only a managed record is written this way"))
	}
}
