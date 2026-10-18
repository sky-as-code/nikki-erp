package composable

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

// ActionPermission names the permission a custom action asserts. It is the resource's own unless
// ResourceCode names another one: assigning roles to a group asserts manage_role_assignments on
// iam_group even when the route is served by iam_role.
type ActionPermission struct {
	ActionCode string

	// ResourceCode is empty for the serving resource itself.
	ResourceCode string

	// Scope applies only with ResourceCode. Nil means requestguard.ResourceScopeOrg.
	Scope *requestguard.ResourceScope
}

// AssertRecordAction authorizes a custom action addressed at one record by id: it resolves the
// serving resource's org scope, asserts the permission inside that org, then refuses a record of
// another org. It is the custom-action counterpart of what the built-in update does.
func AssertRecordAction(
	ctx corectx.Context, appSvc CrudApplicationService, params dmodel.DynamicFields, perm ActionPermission,
) (*ft.ClientErrors, error) {
	orgId, cErrs := AssertActionPermission(ctx, appSvc, params, perm)
	if cErrs != nil || orgId == nil {
		return cErrs, nil
	}
	return appSvc.AssertRecordInOrg(ctx, params, *orgId)
}

// AssertActionPermission is AssertRecordAction without the record check, for a collection action.
// It returns the org the request is confined to, nil when the resource is not org-scoped.
func AssertActionPermission(
	ctx corectx.Context, appSvc CrudApplicationService, params dmodel.DynamicFields, perm ActionPermission,
) (*model.Id, *ft.ClientErrors) {
	orgId, cErrs := appSvc.ResolveOrgScope(ctx, params)
	if cErrs != nil {
		return nil, cErrs
	}
	if perm.ResourceCode == "" {
		return orgId, appSvc.AssertPermission(ctx, perm.ActionCode, orgId)
	}
	scope := requestguard.ResourceScopeOrg
	if perm.Scope != nil {
		scope = *perm.Scope
	}
	return orgId, requestguard.AssertPermission(ctx,
		requestguard.PermFor(perm.ActionCode, perm.ResourceCode, scope).InOrg(orgId))
}

// SearchLinked lists the peers linked to one record through a many-to-many edge the peer declares:
// a search of the peer resource with "<peerEdge> linked <srcId>" ANDed onto the caller's graph.
// defaultFields applies when the query names none.
func SearchLinked(
	ctx corectx.Context, peer CrudDomainService, query SearchQuery, peerEdge string, srcId model.Id,
	defaultFields []string,
) (*SearchResult, error) {
	peerQuery := dmodel.DynamicFields{}
	for key, val := range query {
		peerQuery[key] = val
	}
	delete(peerQuery, "id")
	if fields, ok := peerQuery["fields"]; (!ok || fields == nil) && len(defaultFields) > 0 {
		peerQuery["fields"] = defaultFields
	}

	linkNode := dmodel.NewSearchNode()
	linkNode.NewCondition(peerEdge, dmodel.Linked, string(srcId))
	graph, err := dmodel.AndDecodedGraph(peerQuery[fieldNameGraph], *linkNode)
	if err != nil {
		vErrs := ft.ClientErrors{*ft.NewValidationError(fieldNameGraph, ft.ErrorKey("err_malformed_request"),
			"graph is not a valid search graph")}
		return &SearchResult{ClientErrors: vErrs}, nil
	}
	peerQuery[fieldNameGraph] = graph
	return peer.Search(ctx, peerQuery)
}
