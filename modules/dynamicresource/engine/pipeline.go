package engine

import (
	"encoding/json"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
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

	orgId, cErrs := this.resolveOrgScope(ctx, definition, params)
	if cErrs != nil {
		return &it.ActionResult{ClientErrors: *cErrs}, nil
	}

	if cErrs := this.assertPermission(ctx, definition, orgId); cErrs != nil {
		return &it.ActionResult{ClientErrors: *cErrs}, nil
	}

	if orgId != nil {
		if err := applyOrgConstraint(actionName, params, *orgId); err != nil {
			return nil, err
		}
		cErrs, err := this.assertRecordInOrg(ctx, actionName, params, *orgId)
		if err != nil {
			return nil, err
		}
		if cErrs != nil {
			return &it.ActionResult{ClientErrors: *cErrs}, nil
		}
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
//
// orgId names the org the record belongs to, and is nil for an action that is not org-scoped.
// Passing it through InOrg is what lets an org-scoped grant match: a Perm with no OrgId can
// only ever match an exact or domain grant, which is how org-scoped checks silently degraded.
func (this *DynamicResourceEngineImpl) assertPermission(
	ctx corectx.Context, definition it.DynamicActionDefinition, orgId *model.Id,
) *ft.ClientErrors {
	if definition.Permission == "" {
		return nil
	}

	scope := this.DefaultPermissionScope()
	if definition.PermissionScope != nil {
		scope = *definition.PermissionScope
	}

	return requestguard.AssertPermission(ctx,
		requestguard.PermFor(definition.Permission, this.ResourceName(), scope).InOrg(orgId),
	)
}

// resolveOrgScope enforces the action's org scoping and returns the org the request is confined
// to. A nil org means the action is not org-scoped, and nothing downstream filters by org.
//
// It is a pipeline step rather than a REST concern so that the rule holds for every caller of
// ExecuteAction, and so that an action cannot be org-scoped over HTTP but unscoped over CQRS.
func (this *DynamicResourceEngineImpl) resolveOrgScope(
	ctx corectx.Context, definition it.DynamicActionDefinition, params dmodel.DynamicFields,
) (*model.Id, *ft.ClientErrors) {
	if !definition.OrgScoped() || !this.schemaHasOrgId() {
		return nil, nil
	}

	rawOrgId := readString(params, basemodel.FieldOrgId)
	if rawOrgId == "" {
		cErrs := ft.ClientErrors{}
		cErrs.Append(*ft.NewValidationError(
			basemodel.FieldOrgId,
			string(ft.ErrorKey("err_org_id_required")),
			"this resource is scoped to an organization, so 'org_id' is required",
		))
		return nil, &cErrs
	}

	// A caller may only act inside an org they belong to. Answering "not found" instead would
	// hide the caller's own mistake behind the same response an empty org produces.
	orgId := model.Id(rawOrgId)
	if !ctx.GetPermissions().UserOrgIds.Contains(orgId) {
		cErrs := ft.ClientErrors{}
		cErrs.Append(*ft.NewValidationError(
			basemodel.FieldOrgId,
			string(ft.ErrorKey("err_org_id_not_a_member")),
			"you do not belong to this organization",
		))
		return nil, &cErrs
	}

	// Normalized back into params so that every downstream consumer - the search graph, the
	// single-row key set, the create payload - reads one authoritative value.
	params[basemodel.FieldOrgId] = rawOrgId
	return &orgId, nil
}

// schemaHasOrgId reports whether this engine's resource declares an org column.
// A resource without one cannot be org-filtered and is left alone by the scoping machinery,
// which is why a settings or metadata engine needs no per-action opt-out.
func (this *DynamicResourceEngineImpl) schemaHasOrgId() bool {
	schema := this.Schema()
	if schema == nil {
		return false
	}
	_, exists := schema.Field(basemodel.FieldOrgId)
	return exists
}

// applyOrgConstraint confines an action's params to one organization, so that the org the
// caller named decides which rows the action can reach - not just which rows it asks for.
//
// It works on params rather than on the service, because a module may replace the service
// wholesale with Engine.SetResourceService. Constraining here means an override inherits the
// filter instead of having to remember it.
//
// The three shapes it has to cover:
//   - search: AND the caller's graph with org_id = <org>, so a caller-supplied graph narrows
//     the result rather than widening it.
//   - single-row reads and mutations: carry org_id alongside the id, so a row in another org
//     simply is not found.
//   - create: org_id is already in params, and resolveOrgScope has normalized it.
func applyOrgConstraint(actionName string, params dmodel.DynamicFields, orgId model.Id) error {
	if actionName != it.ActionSearch {
		// Every other action identifies rows by key, and resolveOrgScope has already written
		// the authoritative org_id into params for the key set to pick up.
		return nil
	}
	return constrainSearchGraph(params, orgId)
}

// constrainSearchGraph ANDs an org_id equality onto whatever graph the caller sent.
//
// The caller's graph arrives as a decoded map, so it is re-marshalled through SearchNode
// rather than manipulated in place: SearchNode owns the graph's JSON shape, and rebuilding it
// by hand here would fork that shape.
func constrainSearchGraph(params dmodel.DynamicFields, orgId model.Id) error {
	orgNode := dmodel.NewSearchNode()
	orgNode.NewCondition(basemodel.FieldOrgId, dmodel.Equals, string(orgId))

	raw, hasGraph := params[queryParamGraph]
	if !hasGraph || raw == nil {
		graph := &dmodel.SearchGraph{}
		graph.And(*orgNode)
		params[queryParamGraph] = graph
		return nil
	}

	callerNode, err := toSearchNode(raw)
	if err != nil {
		// A malformed graph is the caller's mistake, and searchParams already rejected the
		// unparseable case. Anything reaching here is a shape the node decoder refuses.
		return errors.Wrap(err, "constrainSearchGraph")
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*callerNode, *orgNode)
	params[queryParamGraph] = graph
	return nil
}

// toSearchNode reinterprets an already-decoded graph value as a SearchNode.
func toSearchNode(raw any) (*dmodel.SearchNode, error) {
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil, errors.Wrap(err, "toSearchNode.Marshal")
	}
	node := dmodel.NewSearchNode()
	if err := json.Unmarshal(encoded, node); err != nil {
		return nil, errors.Wrap(err, "toSearchNode.Unmarshal")
	}
	return node, nil
}

// singleRecordActions are the built-in actions that identify exactly one existing row by id.
// Each one reaches its row through a crud command that carries only the id, so the org has to
// be checked before the command runs rather than inside it.
var singleRecordActions = map[string]bool{
	it.ActionGetById:     true,
	it.ActionUpdate:      true,
	it.ActionDelete:      true,
	it.ActionSetArchived: true,
}

// assertRecordInOrg refuses a single-row action whose target row belongs to another org.
//
// The crud commands behind get_by_id, update, delete and set_archived take an id and nothing
// else (dyn.DeleteOneCommand and dyn.GetOneQuery have no org field), so without this check a
// caller holding a grant in their own org could read or mutate any row whose id they knew,
// simply by naming their own org in the query string.
//
// A row in another org answers "not found" rather than "forbidden": the caller is not entitled
// to learn that the id exists at all.
func (this *DynamicResourceEngineImpl) assertRecordInOrg(
	ctx corectx.Context, actionName string, params dmodel.DynamicFields, orgId model.Id,
) (*ft.ClientErrors, error) {
	if !singleRecordActions[actionName] {
		return nil, nil
	}
	recordId := readString(params, basemodel.FieldId)
	if recordId == "" {
		// No id to check. The action's own validation reports the missing key.
		return nil, nil
	}

	vErrs := ft.ClientErrors{}
	found, err := this.fetchByKeys(ctx, dmodel.DynamicFields{
		basemodel.FieldId:    recordId,
		basemodel.FieldOrgId: string(orgId),
	}, &vErrs)
	if err != nil {
		return nil, errors.Wrap(err, "assertRecordInOrg")
	}
	if vErrs.Count() > 0 {
		return &vErrs, nil
	}
	if found == nil {
		notFound := ft.ClientErrors{}
		notFound.Append(*ft.NewAnonymousNotFoundError())
		return &notFound, nil
	}
	return nil, nil
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
