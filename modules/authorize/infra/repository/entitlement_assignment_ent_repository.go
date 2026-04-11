package repository

// import (
// 	"errors"

// 	"github.com/sky-as-code/nikki-erp/common/model"
// 	"github.com/sky-as-code/nikki-erp/common/orm"
// 	"github.com/sky-as-code/nikki-erp/modules/core/crud"
// 	"github.com/sky-as-code/nikki-erp/modules/core/database"

// 	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
// 	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
// 	entEffectiveGroup "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/effectivegroupentitlement"
// 	entEffectiveUser "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/effectiveuserentitlement"
// 	entAssign "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/entitlementassignment"
// 	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/entitlement_assignment"
// )

// func NewEntitlementAssignmentEntRepository(client *ent.Client) it.EntitlementAssignmentRepository {
// 	return &EntitlementAssignmentEntRepository{
// 		client: client,
// 	}
// }

// func (this *EntitlementAssignmentEntRepository) assignmentClient(ctx crud.Context) *ent.EntitlementAssignmentClient {
// 	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
// 	if isOk {
// 		return tx.EntitlementAssignment
// 	}
// 	return this.client.EntitlementAssignment
// }

// func (this *EntitlementAssignmentEntRepository) Create(ctx crud.Context, assignment *domain.EntitlementGrant) (*domain.EntitlementGrant, error) {
// 	st := assignment.GetSubjectType()
// 	if st == nil {
// 		return nil, errors.New("authorize: entitlement assignment subject type is required")
// 	}
// 	if assignment.GetId() == nil {
// 		return nil, errors.New("authorize: entitlement assignment id is required")
// 	}
// 	creation := this.assignmentClient(ctx).Create().
// 		SetID(string(*assignment.GetId())).
// 		SetSubjectType(entAssign.SubjectType(string(*st))).
// 		SetSubjectRef(*assignment.GetSubjectRef()).
// 		SetResolvedExpr(*assignment.GetResolvedExpr()).
// 		SetNillableActionName(assignment.GetActionName()).
// 		SetNillableResourceName(assignment.GetResourceName()).
// 		SetEntitlementID(string(*assignment.GetEntitlementId())).
// 		SetNillableScopeRef(assignment.GetScopeRef())

// 	return database.Mutate(ctx, creation, ent.IsNotFound, entToEntitlementAssignment)
// }

// func (this *EntitlementAssignmentEntRepository) FindAllBySubject(ctx crud.Context, param it.FindBySubjectParam) ([]domain.EntitlementGrant, error) {
// 	query := this.assignmentClient(ctx).Query().
// 		Where(entAssign.SubjectTypeEQ(entAssign.SubjectType(param.SubjectType))).
// 		Where(entAssign.SubjectRefEQ(param.SubjectRef)).
// 		WithEntitlement(func(eq *ent.EntitlementQuery) {
// 			eq.WithResource()
// 		})

// 	return database.List(ctx, query, func(assignments []*ent.EntitlementAssignment) []domain.EntitlementGrant {
// 		result := make([]domain.EntitlementGrant, len(assignments))
// 		for i, assignment := range assignments {
// 			result[i] = *entToEntitlementAssignment(assignment)
// 		}
// 		return result
// 	})
// }

// func (this *EntitlementAssignmentEntRepository) FindViewsById(ctx crud.Context, param it.FindViewsByIdParam) ([]domain.EntitlementGrant, error) {
// 	assignments := make([]domain.EntitlementGrant, 0)

// 	switch param.SubjectType {
// 	case domain.EntitlementAssignmentSubjectTypeNikkiUser.String():
// 		userAssignments, err := this.getUserEffectiveEntitlements(ctx, model.Id(param.SubjectRef))
// 		if err != nil {
// 			return nil, err
// 		}

// 		assignments = append(assignments, userAssignments...)
// 	case domain.EntitlementAssignmentSubjectTypeNikkiGroup.String():
// 		groupAssignments, err := this.getGroupEffectiveEntitlements(ctx, model.Id(param.SubjectRef))
// 		if err != nil {
// 			return nil, err
// 		}

// 		assignments = append(assignments, groupAssignments...)
// 	}

// 	return assignments, nil
// }

// func (this *EntitlementAssignmentEntRepository) FindAllByEntitlementId(ctx crud.Context, param it.FindAllByEntitlementIdParam) ([]domain.EntitlementGrant, error) {
// 	query := this.assignmentClient(ctx).Query().
// 		Where(entAssign.EntitlementIDEQ(param.EntitlementId))

// 	return database.List(ctx, query, entToEntitlementAssignments)
// }

// func (this *EntitlementAssignmentEntRepository) DeleteHard(ctx crud.Context, param it.DeleteHardParam) (int, error) {
// 	return this.assignmentClient(ctx).Delete().
// 		Where(entAssign.IDEQ(param.Id)).
// 		Exec(ctx)
// }

// func (this *EntitlementAssignmentEntRepository) DeleteHardByEntitlementId(ctx crud.Context, param it.DeleteHardByEntitlementIdParam) (int, error) {
// 	return this.assignmentClient(ctx).Delete().
// 		Where(entAssign.EntitlementIDEQ(param.EntitlementId)).
// 		Exec(ctx)
// }

// func (this *EntitlementAssignmentEntRepository) FindById(ctx crud.Context, param it.FindByIdParam) (*domain.EntitlementGrant, error) {
// 	query := this.assignmentClient(ctx).Query().
// 		Where(entAssign.IDEQ(param.Id)).
// 		WithEntitlement()

// 	return database.FindOne(ctx, query, ent.IsNotFound, entToEntitlementAssignment)
// }

// func (this *EntitlementAssignmentEntRepository) FindByFilter(ctx crud.Context, param it.FindByFilterParam) (*domain.EntitlementGrant, error) {
// 	query := this.assignmentClient(ctx).Query().
// 		Where(
// 			entAssign.SubjectTypeEQ(entAssign.SubjectType(param.SubjectType)),
// 			entAssign.SubjectRefEQ(param.SubjectRef),
// 			entAssign.EntitlementIDEQ(param.EntitlementId),
// 		).
// 		WithEntitlement()

// 	if param.ScopeRef != nil {
// 		query = query.Where(entAssign.ScopeRefEQ(*param.ScopeRef))
// 	} else {
// 		query = query.Where(entAssign.ScopeRefIsNil())
// 	}

// 	return database.FindOne(ctx, query, ent.IsNotFound, entToEntitlementAssignment)
// }

// func (this *EntitlementAssignmentEntRepository) getUserEffectiveEntitlements(ctx crud.Context, userId model.Id) ([]domain.EntitlementGrant, error) {
// 	effectiveAssignments, err := this.client.EffectiveUserEntitlement.
// 		Query().
// 		Where(entEffectiveUser.UserIDEQ(userId)).
// 		All(ctx)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return effectiveEntToEntitlementAssignments(effectiveAssignments, nil), nil
// }

// func (this *EntitlementAssignmentEntRepository) getGroupEffectiveEntitlements(ctx crud.Context, groupId model.Id) ([]domain.EntitlementGrant, error) {
// 	effectiveAssignments, err := this.client.EffectiveGroupEntitlement.
// 		Query().
// 		Where(entEffectiveGroup.GroupIDEQ(groupId)).
// 		All(ctx)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return effectiveEntToEntitlementAssignments(nil, effectiveAssignments), nil
// }

// // Deprecated: Must create dynamic model repository instead
// type EntitlementAssignmentEntRepository struct {
// 	client *ent.Client
// }

// func BuildEntitlementAssignmentDescriptor() *orm.EntityDescriptor {
// 	entity := ent.EntitlementAssignment{}
// 	builder := orm.DescribeEntity(entAssign.Label).
// 		Aliases("entitlement_assignments").
// 		Field(entAssign.FieldID, entity.ID).
// 		Field(entAssign.FieldSubjectType, entity.SubjectType).
// 		Field(entAssign.FieldSubjectRef, entity.SubjectRef).
// 		Field(entAssign.FieldActionName, entity.ActionName).
// 		Field(entAssign.FieldResourceName, entity.ResourceName).
// 		Field(entAssign.FieldResolvedExpr, entity.ResolvedExpr).
// 		Field(entAssign.FieldEntitlementID, entity.EntitlementID).
// 		Field(entAssign.FieldScopeRef, entity.ScopeRef)

// 	return builder.Descriptor()
// }
