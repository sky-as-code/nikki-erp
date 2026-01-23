package app

import (
	"github.com/sky-as-code/nikki-erp/common/convert"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	itAuthorize "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces"
	itAction "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/action"
	itAssign "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/entitlement_assignment"
	itResource "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/resource"
	itSuite "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/role_suite"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewAuthorizeServiceImpl(
	cqrsBus cqrs.CqrsBus,
	entAssignmentRepo itAssign.EntitlementAssignmentRepository,
	entSuiteRepo itSuite.RoleSuiteRepository,
	entActionRepo itAction.ActionRepository,
	entResourceRepo itResource.ResourceRepository,
) itAuthorize.AuthorizeService {
	return &AuthorizeServiceImpl{
		cqrsBus:           cqrsBus,
		entAssignmentRepo: entAssignmentRepo,
		entSuiteRepo:      entSuiteRepo,
		entActionRepo:     entActionRepo,
		entResourceRepo:   entResourceRepo,
	}
}

type AuthorizeServiceImpl struct {
	cqrsBus           cqrs.CqrsBus
	entAssignmentRepo itAssign.EntitlementAssignmentRepository
	entSuiteRepo      itSuite.RoleSuiteRepository
	entActionRepo     itAction.ActionRepository
	entResourceRepo   itResource.ResourceRepository
}

func (this *AuthorizeServiceImpl) IsAuthorized(ctx crud.Context, query itAuthorize.IsAuthorizedQuery) (result *itAuthorize.IsAuthorizedResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "check authorization"); e != nil {
			err = e
		}
	}()

	// resource variable is used to store the resource object after validation
	var resource *domain.Resource

	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			resource, err = this.validateResource(ctx, query.ResourceName, vErrs)
			return err
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			return this.validateAction(ctx, query.ActionName, resource, vErrs)
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itAuthorize.IsAuthorizedResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	// Temporarily not available for private type
	switch query.SubjectType {
	case itAuthorize.SubjectTypeUser:
		return this.isUserAuthorized(ctx, query)
	case itAuthorize.SubjectTypeGroup:
		return this.isGroupAuthorized(ctx, query)
	default:
		return this.isAuthorizedLegacy(ctx, query)
	}
}

func (this *AuthorizeServiceImpl) PermissionSnapshot(ctx crud.Context, query itAuthorize.PermissionSnapshotQuery) (result *itAuthorize.PermissionSnapshotResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "get permission snapshot"); e != nil {
			err = e
		}
	}()

	userRes := &itUser.GetUserByIdResult{}
	flow := validator.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *fault.ValidationErrors) error {
			*vErrs = query.Validate()
			return nil
		}).
		Step(func(vErrs *fault.ValidationErrors) error {
			userRes, err = this.getUser(ctx, query.UserId, vErrs)
			return err
		}).
		End()
	fault.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &itAuthorize.PermissionSnapshotResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	assignments, err := this.entAssignmentRepo.FindViewsById(ctx, itAssign.FindViewsByIdParam{
		SubjectType: itAuthorize.SubjectTypeUser.String(),
		SubjectRef:  query.UserId,
	})
	fault.PanicOnErr(err)

	permissions := this.buildPermissionsSnapshot(assignments)

	return &itAuthorize.PermissionSnapshotResult{
		AvatarUrl:   userRes.Data.AvatarUrl,
		DisplayName: userRes.Data.DisplayName,
		Permissions: permissions,
	}, nil
}

func (this *AuthorizeServiceImpl) isUserAuthorized(ctx crud.Context, query itAuthorize.IsAuthorizedQuery) (result *itAuthorize.IsAuthorizedResult, err error) {
	userAssignments, err := this.entAssignmentRepo.FindViewsById(ctx, itAssign.FindViewsByIdParam{
		SubjectType: itAuthorize.SubjectTypeUser.String(),
		SubjectRef:  query.SubjectRef,
	})
	fault.PanicOnErr(err)

	for _, assignment := range userAssignments {
		if this.matchAssignment(&assignment, query) {
			return &itAuthorize.IsAuthorizedResult{
				Decision: convert.StringPtr(itAuthorize.DecisionAllow),
			}, nil
		}
	}

	return &itAuthorize.IsAuthorizedResult{
		Decision: convert.StringPtr(itAuthorize.DecisionDeny),
	}, nil
}

func (this *AuthorizeServiceImpl) isGroupAuthorized(ctx crud.Context, query itAuthorize.IsAuthorizedQuery) (result *itAuthorize.IsAuthorizedResult, err error) {
	groupAssignments, err := this.entAssignmentRepo.FindViewsById(ctx, itAssign.FindViewsByIdParam{
		SubjectType: itAuthorize.SubjectTypeGroup.String(),
		SubjectRef:  query.SubjectRef,
	})
	fault.PanicOnErr(err)

	for _, assignment := range groupAssignments {
		if this.matchAssignment(&assignment, query) {
			return &itAuthorize.IsAuthorizedResult{
				Decision: convert.StringPtr(itAuthorize.DecisionAllow),
			}, nil
		}
	}

	return &itAuthorize.IsAuthorizedResult{
		Decision: convert.StringPtr(itAuthorize.DecisionDeny),
	}, nil
}

func (this *AuthorizeServiceImpl) isAuthorizedLegacy(ctx crud.Context, query itAuthorize.IsAuthorizedQuery) (result *itAuthorize.IsAuthorizedResult, err error) {
	subjects, err := this.expandSubjects(ctx, query.SubjectType.String(), query.SubjectRef)
	fault.PanicOnErr(err)

	subjects = uniqueSubjects(subjects)

	assignments, err := this.getAssignmentsForSubjects(ctx, subjects)
	fault.PanicOnErr(err)

	for _, assignment := range assignments {
		if this.matchAssignment(&assignment, query) {
			return &itAuthorize.IsAuthorizedResult{
				Decision: convert.StringPtr(itAuthorize.DecisionAllow),
			}, nil
		}
	}

	return &itAuthorize.IsAuthorizedResult{
		Decision: convert.StringPtr(itAuthorize.DecisionDeny),
	}, nil
}

func (this *AuthorizeServiceImpl) expandSubjects(ctx crud.Context, subjectType, subjectRef string) (subjects []itAuthorize.Subject, err error) {
	subjects = append(subjects, itAuthorize.Subject{Type: itAuthorize.SubjectTypeAuthorize(subjectType), Ref: subjectRef})

	if subjectType == itAuthorize.SubjectTypeSuite.String() {
		suiteRoles, err := this.getSuiteRoles(ctx, subjectRef)
		fault.PanicOnErr(err)

		if suiteRoles != nil {
			subjects = append(subjects, suiteRoles...)
		}
	}

	return subjects, nil
}

func (this *AuthorizeServiceImpl) getSuiteRoles(ctx crud.Context, suiteRef string) ([]itAuthorize.Subject, error) {
	suiteRes, err := this.entSuiteRepo.FindById(ctx, itSuite.FindByIdParam{
		Id: suiteRef,
	})
	fault.PanicOnErr(err)

	if suiteRes == nil {
		return nil, nil
	}

	roles := make([]itAuthorize.Subject, 0, len(suiteRes.Roles))
	for _, role := range suiteRes.Roles {
		roles = append(roles, itAuthorize.Subject{
			Type: itAuthorize.SubjectTypeAuthorize(itAuthorize.SubjectTypeRole.String()),
			Ref:  *role.Id,
		})
	}
	return roles, nil
}

func (this *AuthorizeServiceImpl) getAssignmentsForSubjects(ctx crud.Context, subjects []itAuthorize.Subject) ([]domain.EntitlementAssignment, error) {
	assignments := []domain.EntitlementAssignment{}

	for _, subject := range subjects {
		res, err := this.entAssignmentRepo.FindAllBySubject(ctx, itAssign.FindBySubjectParam{
			SubjectType: domain.EntitlementAssignmentSubjectType(subject.Type.String()),
			SubjectRef:  model.Id(subject.Ref),
		})
		fault.PanicOnErr(err)

		assignments = append(assignments, res...)
	}

	return assignments, nil
}

func (this *AuthorizeServiceImpl) matchAssignment(assignment *domain.EntitlementAssignment, query itAuthorize.IsAuthorizedQuery) bool {
	if assignment.Entitlement == nil {
		return false
	}

	if assignment.ActionName != nil {
		if *assignment.ActionName != query.ActionName {
			return false
		}
	}

	if assignment.ResourceName != nil {
		if *assignment.ResourceName != query.ResourceName {
			return false
		}
	}

	if assignment.Entitlement.Resource != nil {
		if assignment.Entitlement.Resource.ScopeType == nil {
			return true
		}

		switch *assignment.Entitlement.Resource.ScopeType {
		case "domain":
			return true
		case "org", "hierarchy":
			if assignment.Entitlement.ScopeRef == nil {
				return true
			}

			if *assignment.Entitlement.ScopeRef == query.ScopeRef {
				return true
			}
			return false
		default:
			return false
		}
	}

	return true
}

func (this *AuthorizeServiceImpl) validateResource(ctx crud.Context, resourceName string, vErrs *fault.ValidationErrors) (*domain.Resource, error) {
	resource, err := this.entResourceRepo.FindByName(ctx, itResource.FindByNameParam{Name: resourceName})
	fault.PanicOnErr(err)

	if resource == nil {
		vErrs.AppendNotFound("resource name", resourceName)
	}

	return resource, nil
}

func (this *AuthorizeServiceImpl) validateAction(ctx crud.Context, actionName string, resource *domain.Resource, vErrs *fault.ValidationErrors) error {
	if resource == nil {
		return nil
	}

	action, err := this.entActionRepo.FindByName(ctx, itAction.FindByNameParam{Name: actionName, ResourceId: *resource.Id})
	fault.PanicOnErr(err)

	if action == nil {
		vErrs.AppendNotFound("action name", actionName)
	}

	return nil
}

func uniqueSubjects(subjects []itAuthorize.Subject) []itAuthorize.Subject {
	seen := make(map[string]struct{})
	result := make([]itAuthorize.Subject, 0, len(subjects))
	for _, s := range subjects {
		key := s.Type.String() + ":" + s.Ref
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			result = append(result, s)
		}
	}
	return result
}

func (this *AuthorizeServiceImpl) buildPermissionsSnapshot(assignments []domain.EntitlementAssignment) map[string][]itAuthorize.ResourceScopePermissions {
	permissions := make(map[string][]itAuthorize.ResourceScopePermissions)

	for _, assignment := range assignments {
		var resourceName string
		if assignment.ResourceName != nil && *assignment.ResourceName != "" {
			resourceName = *assignment.ResourceName
		} else if assignment.Entitlement != nil && assignment.Entitlement.Resource != nil && assignment.Entitlement.Resource.Name != nil {
			resourceName = *assignment.Entitlement.Resource.Name
		} else {
			continue
		}

		scopeType := "domain"
		if assignment.Entitlement != nil && assignment.Entitlement.Resource != nil && assignment.Entitlement.Resource.ScopeType != nil {
			scopeType = string(*assignment.Entitlement.Resource.ScopeType)
		}

		scopeRef := ""
		if assignment.ScopeRef != nil {
			scopeRef = *assignment.ScopeRef
		} else if assignment.Entitlement != nil && assignment.Entitlement.ScopeRef != nil {
			scopeRef = *assignment.Entitlement.ScopeRef
		}

		actionName := ""
		if assignment.ActionName != nil {
			actionName = *assignment.ActionName
		}

		resourcePerms := permissions[resourceName]
		var scopePerms *itAuthorize.ResourceScopePermissions
		for i := range resourcePerms {
			if resourcePerms[i].ScopeType == scopeType && resourcePerms[i].ScopeRef == scopeRef {
				scopePerms = &resourcePerms[i]
				break
			}
		}

		if scopePerms == nil {
			resourcePerms = append(resourcePerms, itAuthorize.ResourceScopePermissions{
				ScopeType: scopeType,
				ScopeRef:  scopeRef,
				Actions:   []string{},
			})
			scopePerms = &resourcePerms[len(resourcePerms)-1]
		}

		if actionName != "" {
			found := false
			for _, existing := range scopePerms.Actions {
				if existing == actionName {
					found = true
					break
				}
			}
			if !found {
				scopePerms.Actions = append(scopePerms.Actions, actionName)
			}
		}

		permissions[resourceName] = resourcePerms
	}

	return permissions
}

func (this *AuthorizeServiceImpl) getUser(ctx crud.Context, userId model.Id, vErrs *fault.ValidationErrors) (*itUser.GetUserByIdResult, error) {
	userRes := &itUser.GetUserByIdResult{}
	err := this.cqrsBus.Request(ctx, itUser.GetUserByIdQuery{Id: userId}, &userRes)
	fault.PanicOnErr(err)

	if userRes.ClientError != nil {
		if !vErrs.MergeClientError(userRes.ClientError) {
			vErrs.AppendNotFound("userId", "user")
		}
		return nil, nil
	}

	return userRes, nil
}
