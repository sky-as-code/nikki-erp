package repository

import (
	"context"
	"time"

	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/database"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entAssignt "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/entitlementassignment"
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
		SetSubjectType(entAssignt.SubjectType(*assignment.SubjectType)).
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
			SetSubjectType(entAssignt.SubjectType(*assignment.SubjectType)).
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
		Where(entAssignt.SubjectTypeEQ(entAssignt.SubjectType(param.SubjectType))).
		Where(entAssignt.SubjectRefEQ(param.SubjectRef))

	count, err := countQuery.Count(ctx)
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return []*domain.EntitlementAssignment{}, nil
	}

	query := this.client.EntitlementAssignment.Query().
		Where(entAssignt.SubjectTypeEQ(entAssignt.SubjectType(param.SubjectType))).
		Where(entAssignt.SubjectRefEQ(param.SubjectRef)).
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

func BuildEntitlementAssignmentDescriptor() *orm.EntityDescriptor {
	entity := ent.EntitlementAssignment{}
	builder := orm.DescribeEntity(entAssignt.Label).
		Aliases("entitlement_assignments").
		Field(entAssignt.FieldID, entity.ID).
		Field(entAssignt.FieldSubjectType, entity.SubjectType).
		Field(entAssignt.FieldSubjectRef, entity.SubjectRef).
		Field(entAssignt.FieldActionName, entity.ActionName).
		Field(entAssignt.FieldResourceName, entity.ResourceName).
		Field(entAssignt.FieldResolvedExpr, entity.ResolvedExpr).
		Field(entAssignt.FieldEntitlementID, entity.EntitlementID)

	return builder.Descriptor()
}
