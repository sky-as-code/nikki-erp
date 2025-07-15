package repository

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	"github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entEntitlementAssignment "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/entitlementassignment"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/entitlement_assignment"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
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
		SetSubjectType(entEntitlementAssignment.SubjectType(*assignment.SubjectType)).
		SetSubjectRef(*assignment.SubjectRef).
		SetResolvedExpr(*assignment.ResolvedExpr).
		SetNillableActionName(assignment.ActionName).
		SetNillableResourceName(assignment.ResourceName)

	return db.Mutate(ctx, creation, entToEntitlementAssignment)
}

func (this *EntitlementAssignmentEntRepository) CreateBulk(ctx context.Context, assignments []domain.EntitlementAssignment) error {
	builders := make([]*ent.EntitlementAssignmentCreate, len(assignments))

	for i, assignment := range assignments {
		builders[i] = this.client.EntitlementAssignment.Create().
			SetID(*assignment.Id).
			SetEntitlementID(*assignment.EntitlementId).
			SetSubjectType(entEntitlementAssignment.SubjectType(*assignment.SubjectType)).
			SetSubjectRef(*assignment.SubjectRef).
			SetResolvedExpr(*assignment.ResolvedExpr).
			SetNillableActionName(assignment.ActionName).
			SetNillableResourceName(assignment.ResourceName)
	}

	_, err := this.client.EntitlementAssignment.CreateBulk(builders...).Save(ctx)
	return err
}

func (this *EntitlementAssignmentEntRepository) FindAllBySubject(ctx context.Context, param it.FindBySubjectParam) ([]*domain.EntitlementAssignment, error) {
	query := this.client.EntitlementAssignment.Query().
		Where(entEntitlementAssignment.SubjectTypeEQ(entEntitlementAssignment.SubjectType(param.SubjectType))).
		Where(entEntitlementAssignment.SubjectRefEQ(param.SubjectRef))

	return db.List(ctx, query, func(assignments []*ent.EntitlementAssignment) []*domain.EntitlementAssignment {
		result := make([]*domain.EntitlementAssignment, len(assignments))
		for i, assignment := range assignments {
			result[i] = entToEntitlementAssignment(assignment)
		}
		return result
	})
}

func BuildEntitlementAssignmentDescriptor() *orm.EntityDescriptor {
	entity := ent.EntitlementAssignment{}
	builder := orm.DescribeEntity(entEntitlementAssignment.Label).
		Aliases("entitlement_assignments").
		Field(entEntitlementAssignment.FieldID, entity.ID).
		Field(entEntitlementAssignment.FieldSubjectType, entity.SubjectType).
		Field(entEntitlementAssignment.FieldSubjectRef, entity.SubjectRef).
		Field(entEntitlementAssignment.FieldActionName, entity.ActionName).
		Field(entEntitlementAssignment.FieldResourceName, entity.ResourceName).
		Field(entEntitlementAssignment.FieldResolvedExpr, entity.ResolvedExpr).
		Field(entEntitlementAssignment.FieldEntitlementID, entity.EntitlementID)

	return builder.Descriptor()
}
