package app

import (
	"fmt"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/common/array"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	reguard "github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itOrgUnit "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/orgunit"
	itPerm "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/permission"
	itUser "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/user"
)

func NewPermissionApplicationServiceImpl(
	permissionDomSvc itPerm.PermissionDomainService,
	permissionRepo itPerm.PermissionRepository,
	orgUnitRepo itOrgUnit.OrgUnitRepository,
	userDomSvc itUser.UserDomainService,
) itPerm.PermissionAppService {
	return &PermissionApplicationServiceImpl{
		permissionDomSvc: permissionDomSvc,
		permissionRepo:   permissionRepo,
		orgUnitRepo:      orgUnitRepo,
		userDomSvc:       userDomSvc,
	}
}

type PermissionApplicationServiceImpl struct {
	permissionDomSvc itPerm.PermissionDomainService
	permissionRepo   itPerm.PermissionRepository
	// orgUnitRepo resolves a unit to its org, so a unit-scoped question can fall
	// back to an org-level grant the same way enforcement does.
	orgUnitRepo      itOrgUnit.OrgUnitRepository
	userDomSvc       itUser.UserDomainService
}

func (this *PermissionApplicationServiceImpl) IsAuthorized(ctx corectx.Context, query itPerm.IsAuthorizedQuery) (*itPerm.IsAuthorizedResult, error) {
	return this.permissionDomSvc.IsAuthorized(ctx, query)
}

func (this *PermissionApplicationServiceImpl) GetUserEntitlements(
	ctx corectx.Context, query itPerm.GetUserEntitlementsQuery,
) (*itPerm.GetUserEntitlementsResult, error) {
	sanitized, cErrs := query.GetSchema().ValidateStruct(query)
	if cErrs.Count() > 0 {
		return &itPerm.GetUserEntitlementsResult{
			ClientErrors: cErrs,
		}, nil
	}
	query = *sanitized.(*itPerm.GetUserEntitlementsQuery)

	resUser, err := this.getEnabledUser(ctx, query.UserEmail, query.UserId)
	if err != nil {
		return nil, err
	}

	if resUser == nil {
		return &itPerm.GetUserEntitlementsResult{
			HasData: false,
		}, nil
	}

	resEnt, err := this.permissionDomSvc.ListAllUserPermissions(ctx, itPerm.ListAllUserPermissionsQuery(query))
	if err != nil {
		return nil, err
	}
	if resEnt.ClientErrors.Count() > 0 {
		return &itPerm.GetUserEntitlementsResult{
			ClientErrors: resEnt.ClientErrors,
		}, nil
	}

	resUser.Entitlements = array.Map(resEnt.Data, func(item models.UserPermission) string {
		return item.MustGetEntExpression()
	})
	return &itPerm.GetUserEntitlementsResult{
		Data:    *resUser,
		HasData: true,
	}, nil
}

// TestMyPermissions answers "am I granted this?" for the caller, and names the
// grant paths that answer when they are.
//
// The subject is the caller, taken from the request context - there is no user id
// to pass, so there is nothing to authorize and nothing to leak. Everything else
// runs through the same machinery as enforcement: the same parser, the same
// candidate generator with the caller's own org/unit context, the same cache table
// and the same expiry filter. A probe that computed its own answer would
// eventually disagree with the guard, and a debugging tool that lies is worse than
// no tool at all.
func (this *PermissionApplicationServiceImpl) TestMyPermissions(
	ctx corectx.Context, query itPerm.TestMyPermissionsQuery,
) (*itPerm.TestMyPermissionsResult, error) {
	callerPerm := ctx.GetPermissions()

	parsed, cErrs := parseProbeExpression(query.Expression)
	if cErrs.Count() > 0 {
		return &itPerm.TestMyPermissionsResult{ClientErrors: *cErrs}, nil
	}

	// The owner short-circuits enforcement, so it must short-circuit the probe
	// too - otherwise the owner would be told they lack a permission they in fact
	// have, because they hold no cached rows at all.
	if callerPerm.IsOwner {
		return grantedResult([]itPerm.PermissionMatch{{
			SourceKind:    models.UserPermSourceKindOwner,
			SourceId:      callerPerm.UserId,
			SourceName:    displayNameOf(ctx),
			EntExpression: reguard.OmnipotentExpression(),
		}}), nil
	}

	required, err := this.probePermOf(ctx, *parsed)
	if err != nil {
		return nil, err
	}
	candidates := reguard.CandidateExpressions(*required, reguard.EvalContextFrom(callerPerm))

	rows, err := this.permissionRepo.FindMatchingPermissions(ctx, callerPerm.UserId, candidates)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		// A bare negative: no near misses, no hint about which roles exist.
		return &itPerm.TestMyPermissionsResult{
			Data:    itPerm.TestMyPermissionsResultData{IsGranted: false, Matches: []itPerm.PermissionMatch{}},
			HasData: true,
		}, nil
	}

	matches, err := this.describeMatches(ctx, rows)
	if err != nil {
		return nil, err
	}
	return grantedResult(matches), nil
}

// parseProbeExpression validates the question. Both failures are client errors:
// this endpoint takes a request body, so malformed input must be a 400, never a
// 500 and never a panic.
func parseProbeExpression(expression string) (*reguard.ParsedExpression, *ft.ClientErrors) {
	cErrs := ft.NewClientErrors()

	parsed, err := reguard.ParseExpression(expression)
	if err != nil {
		cErrs.Append(*ft.NewValidationError(
			"expression", "permission.test_expression_invalid",
			"the expression is not a valid entitlement expression",
		))
		return nil, cErrs
	}
	// Wildcards are legitimate in a stored GRANT - that is what makes them useful.
	// In a QUESTION they are not: a real requirement is always concrete, and
	// allowing them would turn this endpoint into a grant-enumeration tool.
	if parsed.HasWildcard() {
		cErrs.Append(*ft.NewValidationError(
			"expression", "permission.test_expression_wildcard",
			"the expression under test must not contain wildcards",
		))
		return nil, cErrs
	}
	return parsed, cErrs
}

func (this *PermissionApplicationServiceImpl) probePermOf(
	ctx corectx.Context, parsed reguard.ParsedExpression,
) (*reguard.Perm, error) {
	required := reguard.PermFor(parsed.ActionCode, parsed.ResourceCode, parsed.Scope)
	switch parsed.Scope {
	case reguard.ResourceScopeOrg:
		required = required.InOrg(parsed.ScopeId)

	case reguard.ResourceScopeOrgUnit:
		required = required.InOrgUnit(parsed.ScopeId)
		// The unit's own org must be resolved here, exactly as the SQL matcher does
		// it, or the org-level fallback cannot fire and the probe would deny a
		// permission that enforcement grants - the precise disagreement this design
		// exists to prevent.
		orgId, err := this.orgIdOfUnit(ctx, parsed.ScopeId)
		if err != nil {
			return nil, err
		}
		required = required.InOrg(orgId)

	case reguard.ResourceScopePrivate:
		// A private-scope question is about the caller's own records by
		// definition - the caller is asking about themself.
		required = required.OwnedByCaller(true)
	}
	return &required, nil
}

func (this *PermissionApplicationServiceImpl) orgIdOfUnit(
	ctx corectx.Context, orgUnitId *model.Id,
) (*model.Id, error) {
	if orgUnitId == nil {
		return nil, nil
	}
	res, err := this.orgUnitRepo.GetOne(ctx, dyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{models.OrgUnitFieldId: *orgUnitId},
		Fields: []string{models.OrgUnitFieldOrgId},
	})
	if err != nil || res == nil || !res.HasData {
		return nil, err
	}
	return res.Data.GetOrgId(), nil
}

// describeMatches turns provenance rows into the reportable shape, resolving the
// role and group names in one query per kind rather than one per row.
func (this *PermissionApplicationServiceImpl) describeMatches(
	ctx corectx.Context, rows []models.UserPermission,
) ([]itPerm.PermissionMatch, error) {
	names, err := this.resolveSourceNames(ctx, rows)
	if err != nil {
		return nil, err
	}

	matches := make([]itPerm.PermissionMatch, 0, len(rows))
	for _, row := range rows {
		sourceKind := safe.GetVal(row.GetSourceKind(), "")
		sourceId := safe.GetVal(row.GetSourceId(), model.Id(""))
		matches = append(matches, itPerm.PermissionMatch{
			SourceKind:    sourceKind,
			SourceId:      sourceId,
			SourceName:    names[sourceId],
			EntExpression: row.MustGetEntExpression(),
		})
	}
	return matches, nil
}

// resolveSourceNames maps each assignment id to the name a person would recognise:
// the role name for a direct grant, the group name for a grant through a group.
//
// Two queries at most, regardless of how many rows matched - one per source kind.
// Resolving per row would make a wide-open account the slowest to answer.
func (this *PermissionApplicationServiceImpl) resolveSourceNames(
	ctx corectx.Context, rows []models.UserPermission,
) (map[model.Id]string, error) {
	directIds, groupIds := partitionSourceIds(rows)
	return this.permissionRepo.ResolveGrantSourceNames(ctx, directIds, groupIds)
}

func partitionSourceIds(rows []models.UserPermission) (direct []model.Id, group []model.Id) {
	seen := make(map[model.Id]struct{}, len(rows))
	for _, row := range rows {
		sourceId := row.GetSourceId()
		if sourceId == nil {
			continue
		}
		if _, dup := seen[*sourceId]; dup {
			continue
		}
		seen[*sourceId] = struct{}{}

		switch safe.GetVal(row.GetSourceKind(), "") {
		case models.UserPermSourceKindDirect:
			direct = append(direct, *sourceId)
		case models.UserPermSourceKindGroup:
			group = append(group, *sourceId)
		}
	}
	return direct, group
}

func grantedResult(matches []itPerm.PermissionMatch) *itPerm.TestMyPermissionsResult {
	return &itPerm.TestMyPermissionsResult{
		Data:    itPerm.TestMyPermissionsResultData{IsGranted: true, Matches: matches},
		HasData: true,
	}
}

func displayNameOf(ctx corectx.Context) string {
	user := ctx.GetUser()
	if user == nil {
		return ""
	}
	return safe.GetVal(user.GetString(models.UserFieldDisplayName), "")
}

func (this *PermissionApplicationServiceImpl) getEnabledUser(ctx corectx.Context, userEmail *string, userId *model.Id) (*itPerm.GetUserEntitlementsResultData, error) {
	result, err := this.userDomSvc.GetEnabledUser(ctx, itUser.GetUserQuery{
		Email: userEmail,
		Id:    userId,
		Fields: []string{
			models.UserFieldId,
			models.UserFieldAvatarUrl,
			models.UserFieldDisplayName,
			models.UserFieldEmail,
			models.UserFieldIsOwner,
			models.UserFieldOrgUnitId,
			fmt.Sprintf("%s.%s", models.UserEdgeOrgUnit, models.OrgUnitFieldOrgId),
			fmt.Sprintf("%s.%s", models.UserEdgeOrgs, models.OrgFieldId),
			fmt.Sprintf("%s.%s", models.UserEdgeOrgs, models.OrgFieldDisplayName),
			fmt.Sprintf("%s.%s", models.UserEdgeOrgs, models.OrgFieldSlug),
		},
	})
	if err != nil {
		return nil, err
	}
	if result.ClientErrors.Count() > 0 || !result.HasData {
		return nil, errors.Wrap(result.ClientErrors.ToError(), "getEnabledUser")
	}

	return &itPerm.GetUserEntitlementsResultData{
		IsOwner:    result.Data.IsOwner(),
		UserId:     result.Data.MustGetId(),
		UserOrgIds: result.Data.GetOrgIds(),
		OrgUnitId:  result.Data.GetOrgUnitId(),
		User:       result.Data.GetFieldData(),
		// Carried into the request context so a unit-scoped check can fall back to
		// an org-level grant for the unit's own org.
		OrgUnitOrgId: result.Data.GetOrgUnitOrgId(),
	}, nil
}
