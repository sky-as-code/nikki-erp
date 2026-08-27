package engine

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// ActionLookupFn resolves an action definition by name. An engine's Action method is one, and
// is what the registry hands the service: the hooks a module declares with DefineAction or
// ModifyAction are then the hooks the service runs, with no copy in between to fall out of step.
type ActionLookupFn func(actionName string) (it.DynamicActionDefinition, bool)

// An action declares its hooks against DynamicFields, while the crud helpers are generic over a
// domain type and receive DynamicEntity. The adapters below carry one shape into the other so
// that a module writes its hook once, in the terms the engine speaks, and the helper still gets
// the signature it expects.

// action returns the definition of the named action, or the zero definition when the service was
// built without a lookup. A zero definition has nil hooks, which every adapter treats as
// "no hook", so a lookup-less service simply runs none.
func (this *DynamicResourceServiceImpl) action(actionName string) it.DynamicActionDefinition {
	if this.actionLookup == nil {
		return it.DynamicActionDefinition{}
	}
	definition, _ := this.actionLookup(actionName)
	return definition
}

// beforeValidationFn adapts BeforeValidation for the crud helpers of create and update.
//
// It returns a NEW entity around the sanitized map rather than mutating the one it was handed:
// crud.Create only adopts the returned field data when the pointer differs from the one it
// passed in, so mutating in place would silently discard what the hook produced.
func beforeValidationFn(hook it.ActionBeforeValidationFn) corecrud.BeforeValidationFn[*it.DynamicEntity] {
	if hook == nil {
		return nil
	}
	return func(
		ctx corectx.Context, model *it.DynamicEntity, vErrs *ft.ClientErrors,
	) (*it.DynamicEntity, error) {
		sanitized, err := hook(ctx, model.GetFieldData(), vErrs)
		if err != nil || vErrs.Count() > 0 || sanitized == nil {
			// Never nil: crud.Update dereferences the result whenever err is nil.
			return model, err
		}
		return it.NewDynamicEntityFrom(sanitized), nil
	}
}

// afterValidationFn adapts AfterValidationSuccess for the crud helpers of create and update.
// The action's hook returns no map, so whatever it changed it changed in the field map it was
// given, and returning the same entity carries that through.
func afterValidationFn(hook it.ActionAfterValidationFn) corecrud.AfterValidationSuccessFn[*it.DynamicEntity] {
	if hook == nil {
		return nil
	}
	return func(ctx corectx.Context, model *it.DynamicEntity) (*it.DynamicEntity, error) {
		return model, hook(ctx, model.GetFieldData())
	}
}

// createValidateExtraFn adapts ValidateExtra for crud.Create, which has no record to hand over:
// a create validates against what is being written and nothing that already exists.
func createValidateExtraFn(hook it.ActionValidateExtraFn) corecrud.CreateValidateExtraFn[*it.DynamicEntity] {
	if hook == nil {
		return nil
	}
	return func(ctx corectx.Context, inputModel *it.DynamicEntity, vErrs *ft.ClientErrors) error {
		return hook(ctx, inputModel.GetFieldData(), nil, vErrs)
	}
}

// updateValidateExtraFn adapts ValidateExtra for crud.Update.
//
// The foundModel it passes on is the record crud.Update has already read and etag-checked, so an
// update hook sees the stored row without the engine reading it a second time.
func updateValidateExtraFn(hook it.ActionValidateExtraFn) corecrud.UpdateValidateExtraFn[*it.DynamicEntity] {
	if hook == nil {
		return nil
	}
	return func(
		ctx corectx.Context, inputModel, foundModel *it.DynamicEntity, vErrs *ft.ClientErrors,
	) error {
		var found *dmodel.DynamicFields
		if foundModel != nil {
			data := foundModel.GetFieldData()
			found = &data
		}
		return hook(ctx, inputModel.GetFieldData(), found, vErrs)
	}
}

// deleteValidateExtraFn adapts ValidateExtra for crud.DeleteOne, which passes only the key
// fields. The params and the record the service already resolved are closed over instead, so a
// delete guard sees the same two arguments every other guard does.
func deleteValidateExtraFn(
	hook it.ActionValidateExtraFn, params dmodel.DynamicFields, foundModel *dmodel.DynamicFields,
) corecrud.DeleteValidateExtraFn {
	if hook == nil {
		return nil
	}
	return func(ctx corectx.Context, _ dmodel.DynamicFields, vErrs *ft.ClientErrors) error {
		return hook(ctx, params, foundModel, vErrs)
	}
}

// deleteAfterValidationFn adapts AfterValidationSuccess for crud.DeleteOne, whose hook is typed
// on the delete command. The action's hook speaks params, so the command passes through untouched.
func deleteAfterValidationFn(
	hook it.ActionAfterValidationFn, params dmodel.DynamicFields,
) corecrud.AfterValidationSuccessFn[dyn.DeleteOneCommand] {
	if hook == nil {
		return nil
	}
	return func(ctx corectx.Context, cmd dyn.DeleteOneCommand) (dyn.DeleteOneCommand, error) {
		return cmd, hook(ctx, params)
	}
}

// prepare runs the part of an action's hook set that must happen before its crud helper is
// called at all: the sanitizing hook, and the read that resolves the record a guard validates
// against. It returns the params the operation should carry on with.
func (this *DynamicResourceServiceImpl) prepare(
	ctx corectx.Context, definition it.DynamicActionDefinition, params dmodel.DynamicFields,
) (dmodel.DynamicFields, *dmodel.DynamicFields, ft.ClientErrors, error) {
	var foundModel *dmodel.DynamicFields

	cErrs, err := dyn.StartValidationFlow().
		Step(func(vErrs *ft.ClientErrors) error {
			if definition.BeforeValidation == nil {
				return nil
			}
			sanitized, err := definition.BeforeValidation(ctx, params, vErrs)
			if err == nil && vErrs.Count() == 0 && sanitized != nil {
				params = sanitized
			}
			return errors.Wrap(err, "BeforeValidation")
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if definition.KeysToFetch == nil {
				return nil
			}
			fetched, err := fetchByKeys(ctx, this.repository, this.schema, definition.KeysToFetch(params), vErrs)
			if err != nil {
				return errors.Wrap(err, "KeysToFetch")
			}
			foundModel = fetched
			return nil
		}).
		End()

	return params, foundModel, cErrs, err
}

// runHooks runs the whole hook set of an operation whose crud helper takes none of them:
// set_archived, the reads and exists. Those helpers validate a typed command rather than the
// resource schema, so there is no slot to hand the hooks to.
//
// The order is the crud helpers' own — ValidateExtra first, AfterValidationSuccess last — so
// that a hook set behaves the same whichever operation carries it.
func (this *DynamicResourceServiceImpl) runHooks(
	ctx corectx.Context, actionName string, params dmodel.DynamicFields,
) (dmodel.DynamicFields, ft.ClientErrors, error) {
	definition := this.action(actionName)

	params, foundModel, cErrs, err := this.prepare(ctx, definition, params)
	if err != nil || cErrs.Count() > 0 {
		return params, cErrs, err
	}

	cErrs, err = dyn.StartValidationFlow().
		Step(func(vErrs *ft.ClientErrors) error {
			if definition.ValidateExtra == nil {
				return nil
			}
			return errors.Wrap(definition.ValidateExtra(ctx, params, foundModel, vErrs), "ValidateExtra")
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if definition.AfterValidationSuccess == nil {
				return nil
			}
			return errors.Wrap(definition.AfterValidationSuccess(ctx, params), "AfterValidationSuccess")
		}).
		End()

	return params, cErrs, err
}
