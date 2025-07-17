package app

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	itAuthorize "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize"
	entitlementAssignIt "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement_assignment"
	roleIt "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
	roleSuiteIt "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role_suite"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
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
		if e := ft.RecoverPanic(recover(), "failed to check authorization"); e != nil {
			err = e
		}
	}()

	subjects, err := this.expandSubjects(ctx, query.Subjects.Type.String(), query.Subjects.Ref)
	ft.PanicOnErr(err)

	subjects = uniqueSubjects(subjects)

	assignments, err := this.getAssignmentsForSubjects(ctx, subjects)
	ft.PanicOnErr(err)

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
	// subjects = append(subjects, itAuthorize.Subject{Type: itAuthorize.WrapSubjectType(subjectType), Ref: subjectRef})

	switch subjectType {
	case itAuthorize.SubjectTypeUser.String():
		// Lấy group của user (Tạm thời chưa làm)
		// groupsRes := groupIt.GetGroupsBySubjectResult{}
		// err := this.cqrsBus.Request(ctx, groupIt.GetGroupsBySubjectQuery{
		// 	SubjectType: domain.WrapEntitlementAssignmentSubjectType(domain.EntitlementAssignmentSubjectTypeNikkiUser.String()),
		// 	SubjectRef:  subjectRef,
		// }, &groupsRes)
		// ft.PanicOnErr(err)

		// if groupsRes.ClientError != nil {
		// 	return nil, groupsRes.ClientError
		// }
		// for _, g := range groupsRes.Data {
		// 	subjects = append(subjects, itAuthorize.Subject{SubjectType: domain.WrapEntitlementAssignmentSubjectType(domain.EntitlementAssignmentSubjectTypeNikkiGroup.String()), SubjectRef: g})
		// }

		// Get role of user
		rolesRes := roleIt.GetRolesBySubjectResult{}
		err := this.cqrsBus.Request(ctx, roleIt.GetRolesBySubjectQuery{
			SubjectType: domain.WrapEntitlementAssignmentSubjectType(domain.EntitlementAssignmentSubjectTypeNikkiUser.String()),
			SubjectRef:  subjectRef,
		}, &rolesRes)
		ft.PanicOnErr(err)

		if rolesRes.ClientError != nil {
			return nil, err
		}
		for _, r := range rolesRes.Data {
			subjects = append(subjects, itAuthorize.Subject{Type: itAuthorize.WrapSubjectType(itAuthorize.SubjectTypeRole.String()), Ref: *r.Id})
		}

		// Get suite of user (it is means get many roles of user)
		suitesRes := roleSuiteIt.GetRoleSuitesBySubjectResult{}
		err = this.cqrsBus.Request(ctx, roleSuiteIt.GetRoleSuitesBySubjectQuery{
			SubjectType: domain.WrapEntitlementAssignmentSubjectType(domain.EntitlementAssignmentSubjectTypeNikkiUser.String()),
			SubjectRef:  subjectRef,
		}, &suitesRes)
		ft.PanicOnErr(err)

		if suitesRes.ClientError != nil {
			return nil, suitesRes.ClientError
		}
		for _, su := range suitesRes.Data {
			for _, r := range su.Roles {
				subjects = append(subjects, itAuthorize.Subject{Type: itAuthorize.WrapSubjectType(itAuthorize.SubjectTypeRole.String()), Ref: *r.Id})
			}
		}
	case itAuthorize.SubjectTypeGroup.String():
		// Get role, suite of group
		rolesRes := roleIt.GetRolesBySubjectResult{}
		err := this.cqrsBus.Request(ctx, roleIt.GetRolesBySubjectQuery{
			SubjectType: domain.WrapEntitlementAssignmentSubjectType(domain.EntitlementAssignmentSubjectTypeNikkiGroup.String()),
			SubjectRef:  subjectRef,
		}, &rolesRes)
		ft.PanicOnErr(err)

		if rolesRes.ClientError != nil {
			return nil, err
		}
		for _, r := range rolesRes.Data {
			subjects = append(subjects, itAuthorize.Subject{Type: itAuthorize.WrapSubjectType(itAuthorize.SubjectTypeRole.String()), Ref: *r.Id})
		}

		suitesRes := roleSuiteIt.GetRoleSuitesBySubjectResult{}
		err = this.cqrsBus.Request(ctx, roleSuiteIt.GetRoleSuitesBySubjectQuery{
			SubjectType: domain.WrapEntitlementAssignmentSubjectType(domain.EntitlementAssignmentSubjectTypeNikkiGroup.String()),
			SubjectRef:  subjectRef,
		}, &suitesRes)
		ft.PanicOnErr(err)

		if suitesRes.ClientError != nil {
			return nil, suitesRes.ClientError
		}
		for _, su := range suitesRes.Data {
			for _, r := range su.Roles {
				subjects = append(subjects, itAuthorize.Subject{Type: itAuthorize.WrapSubjectType(itAuthorize.SubjectTypeRole.String()), Ref: *r.Id})
			}
		}
	case itAuthorize.SubjectTypeRole.String():
		subjects = append(subjects, itAuthorize.Subject{Type: itAuthorize.WrapSubjectType(itAuthorize.SubjectTypeRole.String()), Ref: subjectRef})
	case itAuthorize.SubjectTypeSuite.String():
		suiteRes := roleSuiteIt.GetRoleSuiteByIdResult{}
		err = this.cqrsBus.Request(ctx, roleSuiteIt.GetRoleSuiteByIdQuery{
			Id: model.Id(subjectRef),
		}, &suiteRes)
		ft.PanicOnErr(err)

		if suiteRes.ClientError != nil {
			return nil, suiteRes.ClientError
		}

		for _, r := range suiteRes.Data.Roles {
			subjects = append(subjects, itAuthorize.Subject{Type: itAuthorize.WrapSubjectType(itAuthorize.SubjectTypeRole.String()), Ref: *r.Id})
		}
	}

	return subjects, nil
}

func (this *AuthorizeServiceImpl) getAssignmentsForSubjects(ctx context.Context, subjects []itAuthorize.Subject) ([]*domain.EntitlementAssignment, error) {
	assignments := []*domain.EntitlementAssignment{}

	for _, subject := range subjects {
		query := entitlementAssignIt.GetAllEntitlementAssignmentBySubjectQuery{
			SubjectType: subject.Type.String(),
			SubjectRef:  subject.Ref,
		}
		assignmentsRes := entitlementAssignIt.GetAllEntitlementAssignmentBySubjectResult{}
		err := this.cqrsBus.Request(ctx, query, &assignmentsRes)
		ft.PanicOnErr(err)

		if assignmentsRes.ClientError != nil {
			return nil, assignmentsRes.ClientError
		}

		assignments = append(assignments, assignmentsRes.Data...)
	}

	return assignments, nil
}

func (this *AuthorizeServiceImpl) matchAssignment(assignment *domain.EntitlementAssignment, query itAuthorize.IsAuthorizedQuery) bool {
	// 1. Action: If ActionName is nil, allow all action. If not nil, must match actionName
	if assignment.ActionName != nil && *assignment.ActionName != query.ActionName {
		return false
	}

	// 2. Resource: If ResourceName is nil, allow all resource. If not nil, must match resourceName
	if assignment.ResourceName != nil && *assignment.ResourceName != query.ResourceName {
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
			if query.ScopeRef != "" && *assignment.Entitlement.ScopeRef == query.ScopeRef {
				return true
			}
			return false
		case "hierarchy":
			if isHigherOrEqual(query.ScopeRef, *assignment.Entitlement.ScopeRef) {
				return true
			}
			return false
		default:
			return false
		}
	}

	if query.ScopeRef != "" && *assignment.Entitlement.ScopeRef == query.ScopeRef {
		return true
	}

	return false
}

func isHigherOrEqual(scopeA, scopeB string) bool {

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
