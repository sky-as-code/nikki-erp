package repository

import (
	"context"
	"time"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/database"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entEffectiveGroup "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/effectivegroupentitlement"
	entEffectiveUser "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/effectiveuserentitlement"
	entAssign "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/entitlementassignment"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement_assignment"
)

func NewEntitlementAssignmentEntRepository(client *ent.Client) it.EntitlementAssignmentRepository {
	return &EntitlementAssignmentEntRepository{
		client: client,
	}
}

type EntitlementAssignmentEntRepository struct {
	client *ent.Client
}

func (this *EntitlementAssignmentEntRepository) Create(ctx context.Context, assignment domain.EntitlementAssignment) (*domain.EntitlementAssignment, error) {
	creation := this.client.EntitlementAssignment.Create().
		SetID(*assignment.Id).
		SetEntitlementID(*assignment.EntitlementId).
		SetSubjectType(entAssign.SubjectType(*assignment.SubjectType)).
		SetSubjectRef(*assignment.SubjectRef).
		SetResolvedExpr(*assignment.ResolvedExpr).
		SetNillableActionName(assignment.ActionName).
		SetNillableResourceName(assignment.ResourceName)

	return database.Mutate(ctx, creation, ent.IsNotFound, entToEntitlementAssignment)
}

func (this *EntitlementAssignmentEntRepository) CreateBulk(ctx context.Context, assignments []domain.EntitlementAssignment) error {
	builders := make([]*ent.EntitlementAssignmentCreate, len(assignments))

	for i, assignment := range assignments {
		builders[i] = this.client.EntitlementAssignment.Create().
			SetID(*assignment.Id).
			SetEntitlementID(*assignment.EntitlementId).
			SetSubjectType(entAssign.SubjectType(*assignment.SubjectType)).
			SetSubjectRef(*assignment.SubjectRef).
			SetResolvedExpr(*assignment.ResolvedExpr).
			SetNillableActionName(assignment.ActionName).
			SetNillableResourceName(assignment.ResourceName)
	}

	_, err := this.client.EntitlementAssignment.CreateBulk(builders...).Save(ctx)
	return err
}

func (this *EntitlementAssignmentEntRepository) FindAllBySubject(ctx context.Context, param it.FindBySubjectParam) ([]*domain.EntitlementAssignment, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	countQuery := this.client.EntitlementAssignment.Query().
		Where(entAssign.SubjectTypeEQ(entAssign.SubjectType(param.SubjectType))).
		Where(entAssign.SubjectRefEQ(param.SubjectRef))

	count, err := countQuery.Count(ctx)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return []*domain.EntitlementAssignment{}, nil
	}

	query := this.client.EntitlementAssignment.Query().
		Where(entAssign.SubjectTypeEQ(entAssign.SubjectType(param.SubjectType))).
		Where(entAssign.SubjectRefEQ(param.SubjectRef)).
		WithEntitlement(func(eq *ent.EntitlementQuery) {
			eq.WithResource()
		})

	return database.List(ctx, query, func(assignments []*ent.EntitlementAssignment) []*domain.EntitlementAssignment {
		result := make([]*domain.EntitlementAssignment, len(assignments))
		for i, assignment := range assignments {
			result[i] = entToEntitlementAssignment(assignment)
		}
		return result
	})
}

func (this *EntitlementAssignmentEntRepository) FindViewsById(ctx context.Context, param it.FindViewsByIdParam) ([]*domain.EntitlementAssignment, error) {
	assignments := make([]*domain.EntitlementAssignment, 0)

	switch param.SubjectType {
	case domain.EntitlementAssignmentSubjectTypeNikkiUser.String():
		userAssignments, err := this.getUserEffectiveEntitlements(ctx, model.Id(param.SubjectRef))
		if err != nil {
			return nil, err
		}

		assignments = append(assignments, userAssignments...)
	case domain.EntitlementAssignmentSubjectTypeNikkiGroup.String():
		groupAssignments, err := this.getGroupEffectiveEntitlements(ctx, model.Id(param.SubjectRef))
		if err != nil {
			return nil, err
		}

		assignments = append(assignments, groupAssignments...)
	}

	return assignments, nil
}

func (this *EntitlementAssignmentEntRepository) getUserEffectiveEntitlements(ctx context.Context, userId model.Id) ([]*domain.EntitlementAssignment, error) {
	effectiveAssignments, err := this.client.EffectiveUserEntitlement.
		Query().
		Where(entEffectiveUser.UserIDEQ(string(userId))).
		All(ctx)

	if err != nil {
		return nil, err
	}

	return entToEntitlementAssignments(effectiveAssignments, nil), nil
}

func (this *EntitlementAssignmentEntRepository) getGroupEffectiveEntitlements(ctx context.Context, groupId model.Id) ([]*domain.EntitlementAssignment, error) {
	effectiveAssignments, err := this.client.EffectiveGroupEntitlement.
		Query().
		Where(entEffectiveGroup.GroupIDEQ(string(groupId))).
		All(ctx)

	if err != nil {
		return nil, err
	}

	return entToEntitlementAssignments(nil, effectiveAssignments), nil
}

func BuildEntitlementAssignmentDescriptor() *orm.EntityDescriptor {
	entity := ent.EntitlementAssignment{}
	builder := orm.DescribeEntity(entAssign.Label).
		Aliases("entitlement_assignments").
		Field(entAssign.FieldID, entity.ID).
		Field(entAssign.FieldSubjectType, entity.SubjectType).
		Field(entAssign.FieldSubjectRef, entity.SubjectRef).
		Field(entAssign.FieldActionName, entity.ActionName).
		Field(entAssign.FieldResourceName, entity.ResourceName).
		Field(entAssign.FieldResolvedExpr, entity.ResolvedExpr).
		Field(entAssign.FieldEntitlementID, entity.EntitlementID)

	return builder.Descriptor()
}
