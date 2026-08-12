package engine

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	it "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
)

// ExecuteAction runs the full pipeline of the named action.
//
// A violation the caller can fix (missing permission, invalid params, missing record)
// comes back as ClientErrors inside the result, never as a Go error, so that the REST
// layer answers 400 rather than 500. A Go error means the request could not be processed.
func (this *DynamicResourceEngineImpl) ExecuteAction(
	ctx corectx.Context, actionName string, params dmodel.DynamicFields,
) (result *it.ActionResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), "execute action "+actionName); e != nil {
			err = e
		}
	}()

	definition, exists := this.Action(actionName)
	if !exists {
		return nil, errors.Errorf(
			"action '%s' is not defined on resource '%s'",
			actionName, this.ResourceName(),
		)
	}

	if params == nil {
		params = dmodel.DynamicFields{}
	}

	if cErrs := this.assertPermission(ctx, definition); cErrs != nil {
		return &it.ActionResult{ClientErrors: *cErrs}, nil
	}

	var foundModel *dmodel.DynamicFields
	flow := dyn.StartValidationFlow()
	clientErrs, err := flow.
		Step(func(vErrs *ft.ClientErrors) error {
			if definition.ParamSchema == nil || definition.BeforeValidation == nil {
				return nil
			}
			sanitized, err := definition.BeforeValidation(ctx, params, vErrs)
			if err == nil && vErrs.Count() == 0 && sanitized != nil {
				params = sanitized
			}
			return errors.Wrap(err, "ExecuteAction.BeforeValidation")
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if definition.ParamSchema == nil {
				return nil
			}
			sanitized, cErrs := definition.ParamSchema().Validate(params, definition.ValidateAsEdit)
			if cErrs != nil {
				*vErrs = cErrs
			} else {
				params = sanitized
			}
			return nil
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if definition.ParamSchema == nil || definition.AfterValidationSuccess == nil {
				return nil
			}
			err := definition.AfterValidationSuccess(ctx, params)
			return errors.Wrap(err, "ExecuteAction.AfterValidationSuccess")
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if definition.KeysToFetch == nil {
				return nil
			}
			fetched, err := this.fetchByKeys(ctx, definition.KeysToFetch(params), vErrs)
			if err != nil {
				return errors.Wrap(err, "ExecuteAction.KeysToFetch")
			}
			foundModel = fetched
			return nil
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if definition.ValidateExtra == nil {
				return nil
			}
			err := definition.ValidateExtra(ctx, params, foundModel, vErrs)
			return errors.Wrap(err, "ExecuteAction.ValidateExtra")
		}).
		End()

	if err != nil {
		return nil, err
	}
	if clientErrs.Count() > 0 {
		return &it.ActionResult{ClientErrors: clientErrs}, nil
	}

	result, err = definition.MainProcess(ctx, it.ProcessInput{
		Params:             params,
		FoundModel:         foundModel,
		ResourceService:    this.ResourceService(),
		ResourceRepository: this.ResourceRepository(),
	})
	return result, errors.Wrap(err, "ExecuteAction.MainProcess")
}

// assertPermission checks the action's permission against the caller's entitlements.
// An action with an empty Permission is open to anyone the middleware already authenticated.
func (this *DynamicResourceEngineImpl) assertPermission(
	ctx corectx.Context, definition it.DynamicActionDefinition,
) *ft.ClientErrors {
	if definition.Permission == "" {
		return nil
	}

	scope := this.DefaultPermissionScope()
	if definition.PermissionScope != nil {
		scope = *definition.PermissionScope
	}

	return requestguard.AssertPermission(ctx, requestguard.Perm{
		ActionCode:   definition.Permission,
		ResourceCode: this.ResourceName(),
		Scope:        scope,
	})
}

// fetchByKeys loads the record the action wants to validate against.
// A missing record is a client error, not a Go error.
func (this *DynamicResourceEngineImpl) fetchByKeys(
	ctx corectx.Context, keys dmodel.DynamicFields, vErrs *ft.ClientErrors,
) (*dmodel.DynamicFields, error) {
	if len(keys) == 0 {
		return nil, errors.New("KeysToFetch returned no key")
	}

	// A key of the wrong shape can never match a row, and reporting it as "not found" hides
	// the caller's actual mistake. Built-in actions declare no ParamSchema, so this is the
	// first and only place a path-supplied key is checked against its declared data type.
	if !this.assertKeysAreWellFormed(keys, vErrs) {
		return nil, nil
	}

	found, err := this.ResourceRepository().FindByKeys(ctx, keys)
	if err != nil {
		return nil, err
	}
	if found.ClientErrors.Count() > 0 {
		vErrs.Concat(found.ClientErrors)
		return nil, nil
	}
	if !found.HasData {
		vErrs.Append(*ft.NewAnonymousNotFoundError())
		return nil, nil
	}

	return &found.Data, nil
}

// assertKeysAreWellFormed validates each fetch key against the data type its schema field
// declares, and reports false as soon as one of them is malformed. A key naming no schema
// field is left alone: KeysToFetch may legitimately return a virtual or computed key.
func (this *DynamicResourceEngineImpl) assertKeysAreWellFormed(
	keys dmodel.DynamicFields, vErrs *ft.ClientErrors,
) bool {
	schema := this.Schema()
	if schema == nil {
		return true
	}

	wellFormed := true
	for name, val := range keys {
		field, exists := schema.Field(name)
		if !exists || val == nil {
			continue
		}
		if _, cErr := field.Validate(val); cErr != nil {
			vErrs.Append(*cErr)
			wellFormed = false
		}
	}
	return wellFormed
}
