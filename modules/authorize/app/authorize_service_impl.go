package app

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/convert"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/event"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	itAuthorize "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize"
	itAction "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/action"
	itAssign "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement_assignment"
	itResource "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/resource"
	itSuite "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role_suite"
)

func NewAuthorizeServiceImpl(cqrsBus cqrs.CqrsBus, eventBus event.EventBus, entAssignmentRepo itAssign.EntitlementAssignmentRepository, entSuiteRepo itSuite.RoleSuiteRepository, entActionRepo itAction.ActionRepository, entResourceRepo itResource.ResourceRepository) itAuthorize.AuthorizeService {
	return &AuthorizeServiceImpl{
		cqrsBus:           cqrsBus,
		eventBus:          eventBus,
		entAssignmentRepo: entAssignmentRepo,
		entSuiteRepo:      entSuiteRepo,
		entActionRepo:     entActionRepo,
		entResourceRepo:   entResourceRepo,
	}
}

type AuthorizeServiceImpl struct {
	cqrsBus           cqrs.CqrsBus
	eventBus          event.EventBus
	entAssignmentRepo itAssign.EntitlementAssignmentRepository
	entSuiteRepo      itSuite.RoleSuiteRepository
	entActionRepo     itAction.ActionRepository
	entResourceRepo   itResource.ResourceRepository
}

func (this *AuthorizeServiceImpl) IsAuthorized(ctx context.Context, query itAuthorize.IsAuthorizedQuery) (result *itAuthorize.IsAuthorizedResult, err error) {
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
			if err != nil {
				return err
			}
			return nil
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

func (this *AuthorizeServiceImpl) isUserAuthorized(ctx context.Context, query itAuthorize.IsAuthorizedQuery) (result *itAuthorize.IsAuthorizedResult, err error) {
	userAssignments, err := this.entAssignmentRepo.FindViewsById(ctx, itAssign.FindViewsByIdParam{
		SubjectType: itAuthorize.SubjectTypeUser.String(),
		SubjectRef:  query.SubjectRef,
	})
	fault.PanicOnErr(err)

	for _, assignment := range userAssignments {
		if this.matchAssignment(assignment, query) {
			return &itAuthorize.IsAuthorizedResult{
				Decision: convert.StringPtr(itAuthorize.DecisionAllow),
			}, nil
		}
	}

	return &itAuthorize.IsAuthorizedResult{
		Decision: convert.StringPtr(itAuthorize.DecisionDeny),
	}, nil
}

func (this *AuthorizeServiceImpl) isGroupAuthorized(ctx context.Context, query itAuthorize.IsAuthorizedQuery) (result *itAuthorize.IsAuthorizedResult, err error) {
	groupAssignments, err := this.entAssignmentRepo.FindViewsById(ctx, itAssign.FindViewsByIdParam{
		SubjectType: itAuthorize.SubjectTypeGroup.String(),
		SubjectRef:  query.SubjectRef,
	})
	fault.PanicOnErr(err)

	for _, assignment := range groupAssignments {
		if this.matchAssignment(assignment, query) {
			return &itAuthorize.IsAuthorizedResult{
				Decision: convert.StringPtr(itAuthorize.DecisionAllow),
			}, nil
		}
	}

	return &itAuthorize.IsAuthorizedResult{
		Decision: convert.StringPtr(itAuthorize.DecisionDeny),
	}, nil
}

func (this *AuthorizeServiceImpl) isAuthorizedLegacy(ctx context.Context, query itAuthorize.IsAuthorizedQuery) (result *itAuthorize.IsAuthorizedResult, err error) {
	subjects, err := this.expandSubjects(ctx, query.SubjectType.String(), query.SubjectRef)
	fault.PanicOnErr(err)

	subjects = uniqueSubjects(subjects)

	assignments, err := this.getAssignmentsForSubjects(ctx, subjects)
	fault.PanicOnErr(err)

	for _, assignment := range assignments {
		if this.matchAssignment(assignment, query) {
			return &itAuthorize.IsAuthorizedResult{
				Decision: convert.StringPtr(itAuthorize.DecisionAllow),
			}, nil
		}
	}

	return &itAuthorize.IsAuthorizedResult{
		Decision: convert.StringPtr(itAuthorize.DecisionDeny),
	}, nil
}

func (this *AuthorizeServiceImpl) expandSubjects(ctx context.Context, subjectType, subjectRef string) (subjects []itAuthorize.Subject, err error) {
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

func (this *AuthorizeServiceImpl) getSuiteRoles(ctx context.Context, suiteRef string) ([]itAuthorize.Subject, error) {
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

func (this *AuthorizeServiceImpl) getAssignmentsForSubjects(ctx context.Context, subjects []itAuthorize.Subject) ([]*domain.EntitlementAssignment, error) {
	assignments := []*domain.EntitlementAssignment{}

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
		if assignment.Entitlement.Resource.ScopeType != nil {
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
	}

	return false
}

func (this *AuthorizeServiceImpl) validateResource(ctx context.Context, resourceName string, vErrs *fault.ValidationErrors) (*domain.Resource, error) {
	resource, err := this.entResourceRepo.FindByName(ctx, itResource.FindByNameParam{Name: resourceName})
	fault.PanicOnErr(err)

	if resource == nil {
		vErrs.AppendNotFound("resource name", resourceName)
	}

	return resource, nil
}

func (this *AuthorizeServiceImpl) validateAction(ctx context.Context, actionName string, resource *domain.Resource, vErrs *fault.ValidationErrors) error {
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