package composable

import (
	"encoding/json"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

// fieldNameGraph is the search param carrying the caller's search graph.
const fieldNameGraph = "graph"

// The org-scoping and permission machinery of the default application service. Every helper is
// exported through CrudApplicationService so a derived application service applies the same
// rules to its custom actions instead of re-deriving them.

// AssertAction is the one call a built-in or custom action makes before touching data: it
// resolves the org the request is confined to, then asserts the caller holds actionCode on this
// resource inside that org. A nil org means the action is not org-scoped.
func (this *DefaultApplicationServiceImpl) AssertAction(
	ctx corectx.Context, actionCode string, params dmodel.DynamicFields,
) (*model.Id, *ft.ClientErrors) {
	orgId, cErrs := this.ResolveOrgScope(ctx, params)
	if cErrs != nil {
		return nil, cErrs
	}
	if cErrs := this.AssertPermission(ctx, actionCode, orgId); cErrs != nil {
		return nil, cErrs
	}
	return orgId, nil
}

// AssertPermission checks the caller's entitlement for actionCode on this resource.
//
// orgId names the org the record belongs to, and is nil for an action that is not org-scoped.
// Passing it through InOrg is what lets an org-scoped grant match: a Perm with no OrgId can only
// ever match an exact or domain grant. An empty actionCode skips the check.
func (this *DefaultApplicationServiceImpl) AssertPermission(
	ctx corectx.Context, actionCode string, orgId *model.Id,
) *ft.ClientErrors {
	if actionCode == "" {
		return nil
	}
	return requestguard.AssertPermission(ctx,
		requestguard.PermFor(actionCode, this.ResourceCode(), this.scope).InOrg(orgId),
	)
}

// ResolveOrgScope enforces the resource's org scoping and returns the org the request is
// confined to. A nil org means the resource is not org-scoped, and nothing downstream filters
// by org. The authoritative org_id is normalized back into params so that every downstream
// consumer - the search graph, the single-row key set, the create payload - reads one value.
func (this *DefaultApplicationServiceImpl) ResolveOrgScope(
	ctx corectx.Context, params dmodel.DynamicFields,
) (*model.Id, *ft.ClientErrors) {
	if !this.orgScoped || !this.schemaHasOrgId() {
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
	if !mayActInOrg(ctx.GetPermissions(), orgId) {
		cErrs := ft.ClientErrors{}
		cErrs.Append(*ft.NewValidationError(
			basemodel.FieldOrgId,
			string(ft.ErrorKey("err_org_id_not_a_member")),
			"you do not belong to this organization",
		))
		return nil, &cErrs
	}

	params[basemodel.FieldOrgId] = rawOrgId
	return &orgId, nil
}

// mayActInOrg reports whether the caller is permitted to act inside orgId.
//
// A user's reach is their org membership. A service principal holds none - it is minted for one
// org and may act only there - so it is checked against the org on the principal instead. Both
// are still subject to the entitlement check; this answers only "which org", never "may they".
func mayActInOrg(perms corectx.ContextPermissions, orgId model.Id) bool {
	if perms.Principal.Kind == corectx.PrincipalKindService {
		return perms.Principal.OrgId != nil && *perms.Principal.OrgId == orgId
	}
	return perms.UserOrgIds.Contains(orgId)
}

// schemaHasOrgId reports whether the resource declares an org column. A resource without one
// cannot be org-filtered and is left alone by the scoping machinery.
func (this *DefaultApplicationServiceImpl) schemaHasOrgId() bool {
	schema := this.Schema()
	if schema == nil {
		return false
	}
	_, exists := schema.Field(basemodel.FieldOrgId)
	return exists
}

// ConstrainSearchToOrg ANDs an org_id equality onto whatever graph the caller sent, so the org
// the caller named decides which rows a search can reach, not just which rows it asks for.
//
// The caller's graph arrives as a decoded map, so it is re-marshalled through SearchNode rather
// than manipulated in place: SearchNode owns the graph's JSON shape.
func (this *DefaultApplicationServiceImpl) ConstrainSearchToOrg(
	params dmodel.DynamicFields, orgId model.Id,
) error {
	orgNode := dmodel.NewSearchNode()
	orgNode.NewCondition(basemodel.FieldOrgId, dmodel.Equals, string(orgId))

	raw, hasGraph := params[fieldNameGraph]
	if !hasGraph || raw == nil {
		graph := &dmodel.SearchGraph{}
		graph.And(*orgNode)
		params[fieldNameGraph] = graph
		return nil
	}

	callerNode, err := toSearchNode(raw)
	if err != nil {
		// A malformed graph is the caller's mistake, and the REST binder already rejected the
		// unparseable case. Anything reaching here is a shape the node decoder refuses.
		return errors.Wrap(err, "ConstrainSearchToOrg")
	}

	graph := &dmodel.SearchGraph{}
	graph.And(*callerNode, *orgNode)
	params[fieldNameGraph] = graph
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

// AssertRecordInOrg refuses a single-row action whose target row belongs to another org.
//
// The crud commands behind get_by_id, update, delete and set_archived take an id and nothing
// else, so without this check a caller holding a grant in their own org could read or mutate
// any row whose id they knew, simply by naming their own org. A row in another org answers
// "not found" rather than "forbidden": the caller is not entitled to learn that the id exists.
// Params without an id are left to the action's own validation.
func (this *DefaultApplicationServiceImpl) AssertRecordInOrg(
	ctx corectx.Context, params dmodel.DynamicFields, orgId model.Id,
) (*ft.ClientErrors, error) {
	recordId := readString(params, basemodel.FieldId)
	if recordId == "" {
		return nil, nil
	}

	vErrs := ft.ClientErrors{}
	found, err := this.FetchByKeys(ctx, dmodel.DynamicFields{
		basemodel.FieldId:    recordId,
		basemodel.FieldOrgId: string(orgId),
	}, &vErrs)
	if err != nil {
		return nil, errors.Wrap(err, "AssertRecordInOrg")
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

// FetchByKeys loads the record a custom action wants to act on. A missing record is a client
// error reported through vErrs with a nil map, not a Go error.
//
// A key of the wrong shape can never match a row, and reporting it as "not found" hides the
// caller's actual mistake, so each key is first checked against its declared data type.
func (this *DefaultApplicationServiceImpl) FetchByKeys(
	ctx corectx.Context, keys dmodel.DynamicFields, vErrs *ft.ClientErrors,
) (dmodel.DynamicFields, error) {
	if len(keys) == 0 {
		return nil, errors.New("FetchByKeys received no key")
	}
	if !this.assertKeysAreWellFormed(keys, vErrs) {
		return nil, nil
	}

	found, err := this.domSvc.Repository().FindByKeys(ctx, keys)
	if err != nil {
		return nil, errors.Wrap(err, "FetchByKeys")
	}
	if found.ClientErrors.Count() > 0 {
		vErrs.Concat(found.ClientErrors)
		return nil, nil
	}
	if !found.HasData {
		vErrs.Append(*ft.NewAnonymousNotFoundError())
		return nil, nil
	}
	if found.Data == nil {
		// A found row is never reported as a nil map: callers tell "found" from "missing" by
		// nil-checking the result.
		return dmodel.DynamicFields{}, nil
	}
	return found.Data, nil
}

// assertKeysAreWellFormed validates each key against the data type its schema field declares. A
// key naming no schema field is left alone: a caller may legitimately pass a virtual key.
func (this *DefaultApplicationServiceImpl) assertKeysAreWellFormed(
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
