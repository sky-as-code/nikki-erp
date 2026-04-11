package app

// import (
// 	"github.com/sky-as-code/nikki-erp/common/convert"
// 	"github.com/sky-as-code/nikki-erp/common/fault"
// 	"github.com/sky-as-code/nikki-erp/common/model"
// 	"github.com/sky-as-code/nikki-erp/common/validator"
// 	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
// 	"github.com/sky-as-code/nikki-erp/modules/core/crud"

// 	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
// 	itAuthorize "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces"
// 	itAction "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/action"
// 	itEntitlement "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/entitlement"
// 	itAssign "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/entitlement_assignment"
// 	itResource "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/resource"
// )

// func NewAuthorizeServiceImpl(
// 	cqrsBus cqrs.CqrsBus,
// 	entAssignmentRepo itAssign.EntitlementAssignmentRepository,
// 	entActionRepo itAction.ActionRepository,
// 	entResourceRepo itResource.ResourceRepository,
// 	entitlementRepo itEntitlement.EntitlementRepository,
// ) itAuthorize.AuthorizeService {
// 	return &AuthorizeServiceImpl{
// 		cqrsBus:           cqrsBus,
// 		entAssignmentRepo: entAssignmentRepo,
// 		entActionRepo:     entActionRepo,
// 		entResourceRepo:   entResourceRepo,
// 		entitlementRepo:   entitlementRepo,
// 	}
// }

// type AuthorizeServiceImpl struct {
// 	cqrsBus           cqrs.CqrsBus
// 	entAssignmentRepo itAssign.EntitlementAssignmentRepository
// 	entActionRepo     itAction.ActionRepository
// 	entResourceRepo   itResource.ResourceRepository
// 	entitlementRepo   itEntitlement.EntitlementRepository
// }

// func (this *AuthorizeServiceImpl) IsAuthorized(ctx crud.Context, query itAuthorize.IsAuthorizedQuery) (result *itAuthorize.IsAuthorizedResult, err error) {
// 	defer func() {
// 		if e := fault.RecoverPanicFailedTo(recover(), "check authorization"); e != nil {
// 			err = e
// 		}
// 	}()

// 	// resource variable is used to store the resource object after validation
// 	var resource *domain.Resource

// 	flow := validator.StartValidationFlow()
// 	vErrs, err := flow.
// 		Step(func(vErrs *fault.ValidationErrors) error {
// 			*vErrs = query.Validate()
// 			return nil
// 		}).
// 		Step(func(vErrs *fault.ValidationErrors) error {
// 			resource, err = this.validateResource(ctx, query.ResourceName, vErrs)
// 			return err
// 		}).
// 		Step(func(vErrs *fault.ValidationErrors) error {
// 			return this.validateAction(ctx, query.ActionName, resource, vErrs)
// 		}).
// 		End()
// 	fault.PanicOnErr(err)

// 	if vErrs.Count() > 0 {
// 		return &itAuthorize.IsAuthorizedResult{
// 			ClientError: vErrs.ToClientError(),
// 		}, nil
// 	}

// 	// Temporarily not available for private type
// 	switch query.SubjectType {
// 	case itAuthorize.SubjectTypeUser:
// 		return this.isUserAuthorized(ctx, query)
// 	case itAuthorize.SubjectTypeGroup:
// 		return this.isGroupAuthorized(ctx, query)
// 	default:
// 		return this.isAuthorizedLegacy(ctx, query)
// 	}
// }

// func (this *AuthorizeServiceImpl) PermissionSnapshot(ctx crud.Context, query itAuthorize.PermissionSnapshotQuery) (result *itAuthorize.PermissionSnapshotResult, err error) {
// 	defer func() {
// 		if e := fault.RecoverPanicFailedTo(recover(), "get permission snapshot"); e != nil {
// 			err = e
// 		}
// 	}()

// 	// userRes := &itUser.GetUserByIdResult{}
// 	flow := validator.StartValidationFlow()
// 	vErrs, err := flow.
// 		Step(func(vErrs *fault.ValidationErrors) error {
// 			*vErrs = query.Validate()
// 			return nil
// 		}).
// 		// Step(func(vErrs *fault.ValidationErrors) error {
// 		// 	userRes, err = this.getUser(ctx, query.UserId, vErrs)
// 		// 	return err
// 		// }).
// 		End()
// 	fault.PanicOnErr(err)

// 	if vErrs.Count() > 0 {
// 		return &itAuthorize.PermissionSnapshotResult{
// 			ClientError: vErrs.ToClientError(),
// 		}, nil
// 	}

// 	assignments, err := this.entAssignmentRepo.FindViewsById(ctx, itAssign.FindViewsByIdParam{
// 		SubjectType: itAuthorize.SubjectTypeUser.String(),
// 		SubjectRef:  query.UserId,
// 	})
// 	fault.PanicOnErr(err)

// 	permissions, err := this.buildPermissionsSnapshot(ctx, assignments)
// 	fault.PanicOnErr(err)

// 	return &itAuthorize.PermissionSnapshotResult{
// 		Permissions: permissions,
// 	}, nil
// }

// func (this *AuthorizeServiceImpl) isUserAuthorized(ctx crud.Context, query itAuthorize.IsAuthorizedQuery) (result *itAuthorize.IsAuthorizedResult, err error) {
// 	userAssignments, err := this.entAssignmentRepo.FindViewsById(ctx, itAssign.FindViewsByIdParam{
// 		SubjectType: itAuthorize.SubjectTypeUser.String(),
// 		SubjectRef:  query.SubjectRef,
// 	})
// 	fault.PanicOnErr(err)

// 	for _, assignment := range userAssignments {
// 		if this.matchAssignment(ctx, &assignment, query) {
// 			return &itAuthorize.IsAuthorizedResult{
// 				Decision: convert.StringPtr(itAuthorize.DecisionAllow),
// 			}, nil
// 		}
// 	}

// 	return &itAuthorize.IsAuthorizedResult{
// 		Decision: convert.StringPtr(itAuthorize.DecisionDeny),
// 	}, nil
// }

// func (this *AuthorizeServiceImpl) isGroupAuthorized(ctx crud.Context, query itAuthorize.IsAuthorizedQuery) (result *itAuthorize.IsAuthorizedResult, err error) {
// 	groupAssignments, err := this.entAssignmentRepo.FindViewsById(ctx, itAssign.FindViewsByIdParam{
// 		SubjectType: itAuthorize.SubjectTypeGroup.String(),
// 		SubjectRef:  query.SubjectRef,
// 	})
// 	fault.PanicOnErr(err)

// 	for _, assignment := range groupAssignments {
// 		if this.matchAssignment(ctx, &assignment, query) {
// 			return &itAuthorize.IsAuthorizedResult{
// 				Decision: convert.StringPtr(itAuthorize.DecisionAllow),
// 			}, nil
// 		}
// 	}

// 	return &itAuthorize.IsAuthorizedResult{
// 		Decision: convert.StringPtr(itAuthorize.DecisionDeny),
// 	}, nil
// }

// func (this *AuthorizeServiceImpl) isAuthorizedLegacy(ctx crud.Context, query itAuthorize.IsAuthorizedQuery) (result *itAuthorize.IsAuthorizedResult, err error) {
// 	subjects, err := this.expandSubjects(ctx, query.SubjectType.String(), query.SubjectRef)
// 	fault.PanicOnErr(err)

// 	subjects = uniqueSubjects(subjects)

// 	assignments, err := this.getAssignmentsForSubjects(ctx, subjects)
// 	fault.PanicOnErr(err)

// 	for _, assignment := range assignments {
// 		if this.matchAssignment(ctx, &assignment, query) {
// 			return &itAuthorize.IsAuthorizedResult{
// 				Decision: convert.StringPtr(itAuthorize.DecisionAllow),
// 			}, nil
// 		}
// 	}

// 	return &itAuthorize.IsAuthorizedResult{
// 		Decision: convert.StringPtr(itAuthorize.DecisionDeny),
// 	}, nil
// }

// func (this *AuthorizeServiceImpl) expandSubjects(ctx crud.Context, subjectType, subjectRef string) (subjects []itAuthorize.Subject, err error) {
// 	subjects = append(subjects, itAuthorize.Subject{Type: itAuthorize.SubjectTypeAuthorize(subjectType), Ref: subjectRef})

// 	return subjects, nil
// }

// func (this *AuthorizeServiceImpl) getAssignmentsForSubjects(ctx crud.Context, subjects []itAuthorize.Subject) ([]domain.EntitlementGrant, error) {
// 	assignments := []domain.EntitlementGrant{}

// 	for _, subject := range subjects {
// 		res, err := this.entAssignmentRepo.FindAllBySubject(ctx, itAssign.FindBySubjectParam{
// 			SubjectType: domain.EntitlementAssignmentSubjectType(subject.Type.String()),
// 			SubjectRef:  model.Id(subject.Ref),
// 		})
// 		fault.PanicOnErr(err)

// 		assignments = append(assignments, res...)
// 	}

// 	return assignments, nil
// }

// func (this *AuthorizeServiceImpl) matchAssignment(ctx crud.Context, assignment *domain.EntitlementGrant, query itAuthorize.IsAuthorizedQuery) bool {
// 	if na := assignment.GetActionName(); na != nil && *na != query.ActionName {
// 		return false
// 	}
// 	if nr := assignment.GetResourceName(); nr != nil && *nr != query.ResourceName {
// 		return false
// 	}
// 	if assignment.EffectiveResourceScopeType != nil {
// 		return this.matchEffectiveAssignmentScope(assignment, query)
// 	}
// 	eid := assignment.GetEntitlementId()
// 	if eid == nil {
// 		return false
// 	}
// 	ent, err := this.entitlementRepo.FindById(ctx, itEntitlement.FindByIdParam{Id: *eid})
// 	fault.PanicOnErr(err)
// 	if ent == nil {
// 		return false
// 	}
// 	rid := ent.GetResourceId()
// 	if rid == nil {
// 		return assignment.GetScopeRef() == nil
// 	}
// 	res, err := this.entResourceRepo.FindById(ctx, itResource.FindByIdParam{Id: *rid})
// 	fault.PanicOnErr(err)
// 	if res == nil {
// 		return false
// 	}
// 	st := res.GetScopeType()
// 	if st == nil {
// 		return true
// 	}
// 	switch *st {
// 	case domain.ResourceMinScopeDomain:
// 		return true
// 	case domain.ResourceMinScopeOrg, domain.ResourceMinScopeHierarchy:
// 		if assignment.GetScopeRef() == nil {
// 			return true
// 		}
// 		return query.ScopeRef != "" && *assignment.GetScopeRef() == query.ScopeRef
// 	case domain.ResourceMinTypePrivate:
// 		if assignment.GetScopeRef() == nil {
// 			return false
// 		}
// 		return query.ScopeRef != "" && *assignment.GetScopeRef() == query.ScopeRef
// 	default:
// 		return false
// 	}
// }

// func (this *AuthorizeServiceImpl) matchEffectiveAssignmentScope(assignment *domain.EntitlementGrant, query itAuthorize.IsAuthorizedQuery) bool {
// 	st := assignment.EffectiveResourceScopeType
// 	if st == nil {
// 		return true
// 	}
// 	switch *st {
// 	case domain.ResourceMinScopeDomain:
// 		return true
// 	case domain.ResourceMinScopeOrg, domain.ResourceMinScopeHierarchy:
// 		if assignment.EffectiveEntitlementScopeRef == nil {
// 			return true
// 		}
// 		return query.ScopeRef != "" && *assignment.EffectiveEntitlementScopeRef == query.ScopeRef
// 	default:
// 		return false
// 	}
// }

// func (this *AuthorizeServiceImpl) validateResource(ctx crud.Context, resourceName string, vErrs *fault.ValidationErrors) (*domain.Resource, error) {
// 	resource, err := this.entResourceRepo.FindByName(ctx, itResource.FindByNameParam{Name: resourceName})
// 	fault.PanicOnErr(err)

// 	if resource == nil {
// 		vErrs.AppendNotFound("resource name", resourceName)
// 	}

// 	return resource, nil
// }

// func (this *AuthorizeServiceImpl) validateAction(ctx crud.Context, actionName string, resource *domain.Resource, vErrs *fault.ValidationErrors) error {
// 	if resource == nil {
// 		return nil
// 	}

// 	action, err := this.entActionRepo.FindByName(ctx, itAction.FindByNameParam{Name: actionName, ResourceId: *resource.GetId()})
// 	fault.PanicOnErr(err)

// 	if action == nil {
// 		vErrs.AppendNotFound("action name", actionName)
// 	}

// 	return nil
// }

// func uniqueSubjects(subjects []itAuthorize.Subject) []itAuthorize.Subject {
// 	seen := make(map[string]struct{})
// 	result := make([]itAuthorize.Subject, 0, len(subjects))
// 	for _, s := range subjects {
// 		key := s.Type.String() + ":" + s.Ref
// 		if _, ok := seen[key]; !ok {
// 			seen[key] = struct{}{}
// 			result = append(result, s)
// 		}
// 	}
// 	return result
// }

// func (this *AuthorizeServiceImpl) buildPermissionsSnapshot(ctx crud.Context, assignments []domain.EntitlementGrant) (map[string][]itAuthorize.ResourceScopePermissions, error) {
// 	permissions := make(map[string][]itAuthorize.ResourceScopePermissions)

// 	for _, assignment := range assignments {
// 		resourceName, actionName, err := this.resolveResourceAndAction(ctx, assignment)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if resourceName == "" {
// 			continue
// 		}

// 		scopeType, scopeRef, err := this.resourceScopeFromAssignment(ctx, assignment)
// 		if err != nil {
// 			return nil, err
// 		}

// 		resourcePerms := permissions[resourceName]
// 		var scopePerms *itAuthorize.ResourceScopePermissions
// 		for i := range resourcePerms {
// 			if resourcePerms[i].ScopeType == scopeType && resourcePerms[i].ScopeRef == scopeRef {
// 				scopePerms = &resourcePerms[i]
// 				break
// 			}
// 		}

// 		if scopePerms == nil {
// 			resourcePerms = append(resourcePerms, itAuthorize.ResourceScopePermissions{
// 				ScopeType: scopeType,
// 				ScopeRef:  scopeRef,
// 				Actions:   []string{},
// 			})
// 			scopePerms = &resourcePerms[len(resourcePerms)-1]
// 		}

// 		if actionName != "" {
// 			hasWildcard := false
// 			for _, a := range scopePerms.Actions {
// 				if a == "*" {
// 					hasWildcard = true
// 					break
// 				}
// 			}
// 			if actionName == "*" {
// 				scopePerms.Actions = []string{"*"}
// 			} else if !hasWildcard {
// 				found := false
// 				for _, existing := range scopePerms.Actions {
// 					if existing == actionName {
// 						found = true
// 						break
// 					}
// 				}
// 				if !found {
// 					scopePerms.Actions = append(scopePerms.Actions, actionName)
// 				}
// 			}
// 		}

// 		permissions[resourceName] = resourcePerms
// 	}

// 	return permissions, nil
// }

// func (this *AuthorizeServiceImpl) resourceScopeFromAssignment(ctx crud.Context, a domain.EntitlementGrant) (scopeType, scopeRef string, err error) {
// 	scopeType = domain.ResourceMinScopeDomain.String()
// 	if a.EffectiveResourceScopeType != nil {
// 		scopeType = string(*a.EffectiveResourceScopeType)
// 		if a.EffectiveEntitlementScopeRef != nil {
// 			scopeRef = *a.EffectiveEntitlementScopeRef
// 		}
// 		return scopeType, scopeRef, nil
// 	}
// 	if sr := a.GetScopeRef(); sr != nil {
// 		scopeRef = *sr
// 	}
// 	eid := a.GetEntitlementId()
// 	if eid == nil {
// 		return scopeType, scopeRef, nil
// 	}
// 	ent, err := this.entitlementRepo.FindById(ctx, itEntitlement.FindByIdParam{Id: *eid})
// 	if err != nil {
// 		return "", "", err
// 	}
// 	if ent == nil {
// 		return scopeType, scopeRef, nil
// 	}
// 	rid := ent.GetResourceId()
// 	if rid == nil {
// 		return scopeType, scopeRef, nil
// 	}
// 	res, err := this.entResourceRepo.FindById(ctx, itResource.FindByIdParam{Id: *rid})
// 	if err != nil {
// 		return "", "", err
// 	}
// 	if res == nil {
// 		return scopeType, scopeRef, nil
// 	}
// 	if st := res.GetScopeType(); st != nil {
// 		scopeType = string(*st)
// 	}
// 	return scopeType, scopeRef, nil
// }

// // resolveResourceAndAction derives resource and action from assignment.
// // 1. *:*: resource nil, action nil → "*", "*"
// // 2. Resource:*: resource set, action nil → resource, "*"
// // 3. Resource:Action: both set → resource, action
// func (this *AuthorizeServiceImpl) resolveResourceAndAction(ctx crud.Context, assignment domain.EntitlementGrant) (resourceName, actionName string, err error) {
// 	an := assignment.GetActionName()
// 	hasAction := an != nil && *an != ""
// 	rn := assignment.GetResourceName()
// 	hasResource := rn != nil && *rn != ""
// 	if !hasResource && assignment.GetEntitlementId() != nil {
// 		ent, e := this.entitlementRepo.FindById(ctx, itEntitlement.FindByIdParam{Id: *assignment.GetEntitlementId()})
// 		if e != nil {
// 			return "", "", e
// 		}
// 		if ent != nil {
// 			resId := ent.GetResourceId()
// 			if resId != nil {
// 				res, e := this.entResourceRepo.FindById(ctx, itResource.FindByIdParam{Id: *resId})
// 				if e != nil {
// 					return "", "", e
// 				}
// 				if res != nil && res.GetName() != nil {
// 					resourceName = *res.GetName()
// 					hasResource = resourceName != ""
// 				}
// 			}
// 		}
// 	}
// 	if hasResource && resourceName == "" && rn != nil {
// 		resourceName = *rn
// 	}
// 	if !hasResource {
// 		if !hasAction {
// 			return "*", "*", nil
// 		}
// 		return "", "", nil
// 	}
// 	if hasAction {
// 		actionName = *an
// 	} else {
// 		actionName = "*"
// 	}
// 	return resourceName, actionName, nil
// }

// // func (this *AuthorizeServiceImpl) getUser(ctx crud.Context, userId model.Id, vErrs *fault.ValidationErrors) (*itUser.GetUserByIdResult, error) {
// // 	userRes := &itUser.GetUserByIdResult{}
// // 	err := this.cqrsBus.Request(ctx, itUser.GetUserByIdQuery{Id: userId}, &userRes)
// // 	fault.PanicOnErr(err)

// // 	if userRes.ClientError != nil {
// // 		if !vErrs.MergeClientError(userRes.ClientError) {
// // 			vErrs.AppendNotFound("userId", "user")
// // 		}
// // 		return nil, nil
// // 	}

// // 	return userRes, nil
// // }
