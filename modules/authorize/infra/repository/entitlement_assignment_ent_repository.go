package repository

import (
	"context"
	"time"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/database"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entEffectiveGroup "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/effectivegroupentitlement"
	entEffectiveUser "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/effectiveuserentitlement"
	entAssign "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/entitlementassignment"
	entPermissionHistory "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/permissionhistory"
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

func (this *EntitlementAssignmentEntRepository) FindAllByEntitlementId(ctx context.Context, param it.FindAllByEntitlementIdParam) ([]*domain.EntitlementAssignment, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	query := this.client.EntitlementAssignment.Query().
		Where(entAssign.EntitlementIDEQ(param.EntitlementId))

	return database.List(ctx, query, entToEntitlementAssignments)
}

func (this *EntitlementAssignmentEntRepository) DeleteHard(ctx context.Context, param it.DeleteEntitlementAssignmentByIdQuery) (int, error) {
	return this.client.EntitlementAssignment.Delete().
		Where(entAssign.IDEQ(param.Id)).
		Exec(ctx)
}

func (this *EntitlementAssignmentEntRepository) DeleteHardTx(ctx context.Context, param it.DeleteEntitlementAssignmentByIdQuery) (int, error) {
	tx, err := this.client.Tx(ctx)
	fault.PanicOnErr(err)

	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "delete entitlement assignment transaction"); e != nil {
			_ = tx.Rollback()
			err = e
		}
	}()

	err = this.setEntitlementAssignmentIdNullTx(ctx, tx, param.Id)
	fault.PanicOnErr(err)

	deletedCount, err := this.deleteEntitlementAssignmentTx(ctx, tx, param.Id)
	fault.PanicOnErr(err)

	fault.PanicOnErr(tx.Commit())
	return deletedCount, nil
}

func (this *EntitlementAssignmentEntRepository) deleteEntitlementAssignmentTx(ctx context.Context, tx *ent.Tx, entitlementAssignmentId model.Id) (int, error) {
	deletedCount, err := tx.EntitlementAssignment.
		Delete().
		Where(entAssign.IDEQ(entitlementAssignmentId)).
		Exec(ctx)
	return deletedCount, err
}

func (this *EntitlementAssignmentEntRepository) setEntitlementAssignmentIdNullTx(ctx context.Context, tx *ent.Tx, entitlementAssignmentId string) error {
	_, err := tx.PermissionHistory.
		Update().
		Where(entPermissionHistory.EntitlementAssignmentIDEQ(entitlementAssignmentId)).
		ClearEntitlementAssignmentID().
		Save(ctx)
	return err
}

func (this *EntitlementAssignmentEntRepository) getUserEffectiveEntitlements(ctx context.Context, userId model.Id) ([]*domain.EntitlementAssignment, error) {
	effectiveAssignments, err := this.client.EffectiveUserEntitlement.
		Query().
		Where(entEffectiveUser.UserIDEQ(userId)).
		All(ctx)

	if err != nil {
		return nil, err
	}

	return effectiveEntToEntitlementAssignments(effectiveAssignments, nil), nil
}

func (this *EntitlementAssignmentEntRepository) getGroupEffectiveEntitlements(ctx context.Context, groupId model.Id) ([]*domain.EntitlementAssignment, error) {
	effectiveAssignments, err := this.client.EffectiveGroupEntitlement.
		Query().
		Where(entEffectiveGroup.GroupIDEQ(groupId)).
		All(ctx)

	if err != nil {
		return nil, err
	}

	return effectiveEntToEntitlementAssignments(nil, effectiveAssignments), nil
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
