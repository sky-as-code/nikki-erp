package app

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
	itUser "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	itAuthorize "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize"
	itAssignt "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement_assignment"
	itRole "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
	itSuite "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role_suite"
)

func NewAuthorizeServiceImpl(cqrsBus cqrs.CqrsBus, eventBus event.EventBus) itAuthorize.AuthorizeService {
	return &AuthorizeServiceImpl{
		cqrsBus:  cqrsBus,
		eventBus: eventBus,
	}
}

type AuthorizeServiceImpl struct {
	cqrsBus  cqrs.CqrsBus
	eventBus event.EventBus
}

func (this *AuthorizeServiceImpl) IsAuthorized(ctx context.Context, query itAuthorize.IsAuthorizedQuery) (result *itAuthorize.IsAuthorizedResult, err error) {
	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "failed to check authorization"); e != nil {
			err = e
		}
	}()

	subjects, err := this.expandSubjects(ctx, query.Subject.Type.String(), query.Subject.Ref)
	fault.PanicOnErr(err)

	subjects = uniqueSubjects(subjects)

	assignments, err := this.getAssignmentsForSubjects(ctx, subjects)
	fault.PanicOnErr(err)

	for _, assignment := range assignments {
		if this.matchAssignment(assignment, query) {
			return &itAuthorize.IsAuthorizedResult{Decision: itAuthorize.DecisionAllow}, nil
		}
	}

	return &itAuthorize.IsAuthorizedResult{
		Decision: itAuthorize.DecisionDeny,
	}, nil
}

func (this *AuthorizeServiceImpl) expandSubjects(ctx context.Context, subjectType, subjectRef string) (subjects []itAuthorize.Subject, err error) {
	subjects = append(subjects, itAuthorize.Subject{Type: itAuthorize.WrapSubjectType(subjectType), Ref: subjectRef})

	switch subjectType {
	case itAuthorize.SubjectTypeUser.String():
		userGroups, err := this.getUserGroups(ctx, subjectRef)
		if err != nil {
			return nil, err
		}
		subjects = append(subjects, userGroups...)

		userRoles, err := this.getUserRoles(ctx, subjectRef)
		if err != nil {
			return nil, err
		}
		subjects = append(subjects, userRoles...)

		userSuiteRoles, err := this.getUserSuiteRoles(ctx, subjectRef)
		if err != nil {
			return nil, err
		}
		subjects = append(subjects, userSuiteRoles...)

	case itAuthorize.SubjectTypeGroup.String():
		groupRoles, err := this.getGroupRoles(ctx, subjectRef)
		if err != nil {
			return nil, err
		}
		subjects = append(subjects, groupRoles...)

		groupSuiteRoles, err := this.getGroupSuiteRoles(ctx, subjectRef)
		if err != nil {
			return nil, err
		}
		subjects = append(subjects, groupSuiteRoles...)

	case itAuthorize.SubjectTypeSuite.String():
		suiteRoles, err := this.getSuiteRoles(ctx, subjectRef)
		if err != nil {
			return nil, err
		}
		subjects = append(subjects, suiteRoles...)
	}

	return subjects, nil
}

func (this *AuthorizeServiceImpl) getUserGroups(ctx context.Context, userRef string) ([]itAuthorize.Subject, error) {
	userRes := itUser.GetUserByIdResult{}
	err := this.cqrsBus.Request(ctx, itUser.GetUserByIdQuery{Id: model.Id(userRef)}, &userRes)
	fault.PanicOnErr(err)

	if userRes.ClientError != nil {
		return nil, userRes.ClientError
	}

	groups := make([]itAuthorize.Subject, 0, len(userRes.Data.Groups))
	for _, g := range userRes.Data.Groups {
		groups = append(groups, itAuthorize.Subject{
			Type: itAuthorize.WrapSubjectType(itAuthorize.SubjectTypeGroup.String()),
			Ref:  *g.Id,
		})
	}
	return groups, nil
}

func (this *AuthorizeServiceImpl) getUserRoles(ctx context.Context, userRef string) ([]itAuthorize.Subject, error) {
	rolesRes := itRole.GetRolesBySubjectResult{}
	err := this.cqrsBus.Request(ctx, itRole.GetRolesBySubjectQuery{
		SubjectType: *domain.WrapEntitlementAssignmentSubjectType(domain.EntitlementAssignmentSubjectTypeNikkiUser.String()),
		SubjectRef:  userRef,
	}, &rolesRes)
	fault.PanicOnErr(err)

	if rolesRes.ClientError != nil {
		return nil, rolesRes.ClientError
	}

	roles := make([]itAuthorize.Subject, 0, len(rolesRes.Data))
	for _, r := range rolesRes.Data {
		roles = append(roles, itAuthorize.Subject{
			Type: itAuthorize.WrapSubjectType(itAuthorize.SubjectTypeRole.String()),
			Ref:  *r.Id,
		})
	}
	return roles, nil
}

func (this *AuthorizeServiceImpl) getUserSuiteRoles(ctx context.Context, userRef string) ([]itAuthorize.Subject, error) {
	suitesRes := itSuite.GetRoleSuitesBySubjectResult{}
	err := this.cqrsBus.Request(ctx, itSuite.GetRoleSuitesBySubjectQuery{
		SubjectType: *domain.WrapEntitlementAssignmentSubjectType(domain.EntitlementAssignmentSubjectTypeNikkiUser.String()),
		SubjectRef:  userRef,
	}, &suitesRes)
	fault.PanicOnErr(err)

	if suitesRes.ClientError != nil {
		return nil, suitesRes.ClientError
	}

	roles := make([]itAuthorize.Subject, 0)
	for _, su := range suitesRes.Data {
		for _, r := range su.Roles {
			roles = append(roles, itAuthorize.Subject{
				Type: itAuthorize.WrapSubjectType(itAuthorize.SubjectTypeRole.String()),
				Ref:  *r.Id,
			})
		}
	}
	return roles, nil
}

func (this *AuthorizeServiceImpl) getGroupRoles(ctx context.Context, groupRef string) ([]itAuthorize.Subject, error) {
	rolesRes := itRole.GetRolesBySubjectResult{}
	err := this.cqrsBus.Request(ctx, itRole.GetRolesBySubjectQuery{
		SubjectType: *domain.WrapEntitlementAssignmentSubjectType(domain.EntitlementAssignmentSubjectTypeNikkiGroup.String()),
		SubjectRef:  groupRef,
	}, &rolesRes)
	fault.PanicOnErr(err)

	if rolesRes.ClientError != nil {
		return nil, rolesRes.ClientError
	}

	roles := make([]itAuthorize.Subject, 0, len(rolesRes.Data))
	for _, r := range rolesRes.Data {
		roles = append(roles, itAuthorize.Subject{
			Type: itAuthorize.WrapSubjectType(itAuthorize.SubjectTypeRole.String()),
			Ref:  *r.Id,
		})
	}
	return roles, nil
}

func (this *AuthorizeServiceImpl) getGroupSuiteRoles(ctx context.Context, groupRef string) ([]itAuthorize.Subject, error) {
	suitesRes := itSuite.GetRoleSuitesBySubjectResult{}
	err := this.cqrsBus.Request(ctx, itSuite.GetRoleSuitesBySubjectQuery{
		SubjectType: *domain.WrapEntitlementAssignmentSubjectType(domain.EntitlementAssignmentSubjectTypeNikkiGroup.String()),
		SubjectRef:  groupRef,
	}, &suitesRes)
	fault.PanicOnErr(err)

	if suitesRes.ClientError != nil {
		return nil, suitesRes.ClientError
	}

	roles := make([]itAuthorize.Subject, 0)
	for _, su := range suitesRes.Data {
		for _, r := range su.Roles {
			roles = append(roles, itAuthorize.Subject{
				Type: itAuthorize.WrapSubjectType(itAuthorize.SubjectTypeRole.String()),
				Ref:  *r.Id,
			})
		}
	}
	return roles, nil
}

func (this *AuthorizeServiceImpl) getSuiteRoles(ctx context.Context, suiteRef string) ([]itAuthorize.Subject, error) {
	suiteRes := itSuite.GetRoleSuiteByIdResult{}
	err := this.cqrsBus.Request(ctx, itSuite.GetRoleSuiteByIdQuery{
		Id: suiteRef,
	}, &suiteRes)
	fault.PanicOnErr(err)

	if suiteRes.ClientError != nil {
		return nil, suiteRes.ClientError
	}

	roles := make([]itAuthorize.Subject, 0, len(suiteRes.Data.Roles))
	for _, r := range suiteRes.Data.Roles {
		roles = append(roles, itAuthorize.Subject{
			Type: itAuthorize.WrapSubjectType(itAuthorize.SubjectTypeRole.String()),
			Ref:  *r.Id,
		})
	}
	return roles, nil
}

func (this *AuthorizeServiceImpl) getAssignmentsForSubjects(ctx context.Context, subjects []itAuthorize.Subject) ([]*domain.EntitlementAssignment, error) {
	assignments := []*domain.EntitlementAssignment{}

	for _, subject := range subjects {
		query := itAssignt.GetAllEntitlementAssignmentBySubjectQuery{
			SubjectType: domain.EntitlementAssignmentSubjectType(subject.Type.String()),
			SubjectRef:  subject.Ref,
		}
		assignmentsRes := itAssignt.GetAllEntitlementAssignmentBySubjectResult{}
		err := this.cqrsBus.Request(ctx, query, &assignmentsRes)
		fault.PanicOnErr(err)

		if assignmentsRes.ClientError != nil {
			return nil, assignmentsRes.ClientError
		}

		assignments = append(assignments, assignmentsRes.Data...)
	}

	return assignments, nil
}

func (this *AuthorizeServiceImpl) matchAssignment(assignment *domain.EntitlementAssignment, query itAuthorize.IsAuthorizedQuery) bool {
	// 1. Action: If ActionName is nil, allow all action. If not nil, must match actionName
	if assignment.ActionName != nil && *assignment.ActionName != *query.ActionName {
		return false
	}

	// 2. Resource: If ResourceName is nil, allow all resource. If not nil, must match resourceName
	if assignment.ResourceName != nil && *assignment.ResourceName != *query.ResourceName {
		return false
	}

	// 3. Scope: If assignment doesn't have Entitlement or Entitlement doesn't have ScopeRef, allow all scope
	if assignment.Entitlement == nil || assignment.Entitlement.ScopeRef == nil {
		return true
	}

	if assignment.Entitlement.Resource != nil && assignment.Entitlement.Resource.ScopeType != nil {
		switch *assignment.Entitlement.Resource.ScopeType {
		case "domain":
			return true
		case "org":
			if query.ScopeRef != nil && *assignment.Entitlement.ScopeRef == *query.ScopeRef {
				return true
			}
			return false
		case "hierarchy":
			if query.ScopeRef == nil && *assignment.Entitlement.ScopeRef == *query.ScopeRef {
				return true
			}
			return false
		default:
			return false
		}
	}

	if query.ScopeRef != nil && *assignment.Entitlement.ScopeRef == *query.ScopeRef {
		return true
	}

	return false
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
