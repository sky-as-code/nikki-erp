package services

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
)

// A system-managed role (managed_by_type = system, e.g. the "User" role every person holds) is
// maintained by the platform through seeds, so the domain service refuses to create, change or
// remove one.

func withRoleManagedCreateGuard(options []composable.CreateOptions) composable.CreateOptions {
	opts := composable.CreateOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	opts.ValidateExtra = composable.ChainValidateExtra(assertRoleCreateNotSystemManaged, opts.ValidateExtra)
	return opts
}

func assertRoleCreateNotSystemManaged(
	_ corectx.Context, input *composable.DynamicEntity, _ *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	if input != nil && basemodel.IsSystemManaged(input.GetFieldData()) {
		vErrs.Append(*ft.NewBusinessViolation(basemodel.FieldManagedByType,
			ft.ErrorKey("err_role_system_managed", "iam"), "a system-managed role is created by the platform only"))
	}
	return nil
}

func withRoleManagedUpdateGuard(options []composable.UpdateOptions) composable.UpdateOptions {
	opts := composable.UpdateOptions{}
	if len(options) > 0 {
		opts = options[0]
	}
	opts.ValidateExtra = composable.ChainValidateExtra(assertRoleNotSystemManaged, opts.ValidateExtra)
	return opts
}

func assertRoleNotSystemManaged(
	_ corectx.Context, _ *composable.DynamicEntity, found *composable.DynamicEntity, vErrs *ft.ClientErrors,
) error {
	if found != nil && basemodel.IsSystemManaged(found.GetFieldData()) {
		vErrs.Append(*ft.NewBusinessViolation(basemodel.FieldManagedByType,
			ft.ErrorKey("err_role_system_managed", "iam"), "a system-managed role cannot be changed or deleted"))
	}
	return nil
}
