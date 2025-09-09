package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/database"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entEntitlement "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/entitlement"
	entAssign "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/entitlementassignment"
	entGrantRequest "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/grantrequest"
	entPermissionHistory "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/permissionhistory"
	entRevokeRequest "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/revokerequest"
	entRole "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/role"
	entRoleUser "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/roleuser"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
)

func NewRoleEntRepository(client *ent.Client) it.RoleRepository {
	return &RoleEntRepository{
		client: client,
	}
}

type RoleEntRepository struct {
	client *ent.Client
}

func (this *RoleEntRepository) Create(ctx context.Context, role domain.Role) (*domain.Role, error) {
	creation := this.client.Role.Create().
		SetID(*role.Id).
		SetEtag(*role.Etag).
		SetName(*role.Name).
		SetNillableDescription(role.Description).
		SetOwnerType(entRole.OwnerType(*role.OwnerType)).
		SetOwnerRef(*role.OwnerRef).
		SetIsRequestable(*role.IsRequestable).
		SetIsRequiredAttachment(*role.IsRequiredAttachment).
		SetIsRequiredComment(*role.IsRequiredComment).
		SetCreatedBy(*role.CreatedBy).
		SetCreatedAt(time.Now())

	return database.Mutate(ctx, creation, ent.IsNotFound, entToRole)
}

func (this *RoleEntRepository) CreateWithEntitlements(ctx context.Context, role domain.Role, entitlementIds []model.Id) (result *domain.Role, err error) {
	tx, err := this.client.Tx(ctx)
	fault.PanicOnErr(err)

	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "create role transaction"); e != nil {
			_ = tx.Rollback()
			err = e
		}
	}()

	createdRole, err := this.createRoleTx(ctx, tx, role)
	fault.PanicOnErr(err)

	// Create entitlement assignments for each entitlement ID
	if len(entitlementIds) > 0 {
		for _, entitlementId := range entitlementIds {
			err := this.createAssignmentTx(ctx, tx, createdRole.ID, entitlementId)
			fault.PanicOnErr(err)
		}
	}

	fault.PanicOnErr(tx.Commit())

	return entToRole(createdRole), nil
}

// There is currently no update reason in permission histories for users/groups with this role.
func (this *RoleEntRepository) UpdateTx(ctx context.Context, role domain.Role, prevEtag model.Etag, addEntitlementIds, removeEntitlementIds []model.Id) (*domain.Role, error) {
	tx, err := this.client.Tx(ctx)
	fault.PanicOnErr(err)

	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "update role transaction"); e != nil {
			_ = tx.Rollback()
			err = e
		}
	}()

	updatedRole, err := this.updateRole(ctx, tx, prevEtag, role)

	// Update assignment_id on permission history to nil before remove assignment
	for _, entId := range removeEntitlementIds {
		err = this.setAssignmentIdNull(ctx, tx, entId)
		fault.PanicOnErr(err)

		err := this.removeAssignmentTx(ctx, tx, *role.Id, entId)
		fault.PanicOnErr(err)
	}

	for _, entId := range addEntitlementIds {
		err := this.createAssignmentTx(ctx, tx, *role.Id, entId)
		fault.PanicOnErr(err)
	}

	fault.PanicOnErr(tx.Commit())
	return entToRole(updatedRole), nil
}

// There is currently no delete reason in permission histories for users/groups with this role.
func (this *RoleEntRepository) DeleteHardTx(ctx context.Context, param it.DeleteRoleHardParam) (int, error) {
	tx, err := this.client.Tx(ctx)
	fault.PanicOnErr(err)

	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "delete hard role transaction"); e != nil {
			_ = tx.Rollback()
			err = e
		}
	}()

	err = this.setGrantRequestBeforeDeleteTx(ctx, tx, param.Id, param.Name)
	fault.PanicOnErr(err)

	err = this.setRevokeRequestBeforeDeleteTx(ctx, tx, param.Id, param.Name)
	fault.PanicOnErr(err)

	err = this.setPermissionHistoryBeforeDeleteTx(ctx, tx, param.Id, param.Name)
	fault.PanicOnErr(err)

	deletedCount, err := this.deleteRoleTx(ctx, tx, param.Id)
	fault.PanicOnErr(err)

	fault.PanicOnErr(tx.Commit())
	return deletedCount, nil
}

func (this *RoleEntRepository) FindByName(ctx context.Context, param it.FindByNameParam) (*domain.Role, error) {
	query := this.client.Role.Query().
		Where(entRole.NameEQ(param.Name))

	return database.FindOne(ctx, query, ent.IsNotFound, entToRole)
}

func (this *RoleEntRepository) FindById(ctx context.Context, param it.FindByIdParam) (*domain.Role, error) {
	query := this.client.Role.Query().
		Where(entRole.IDEQ(param.Id))

	return database.FindOne(ctx, query, ent.IsNotFound, entToRole)
}

func (this *RoleEntRepository) FindAllBySubject(ctx context.Context, param it.FindAllBySubjectParam) ([]domain.Role, error) {
	query := this.client.Role.Query().
		Where(entRole.HasRoleUsersWith(entRoleUser.ReceiverRefEQ(param.SubjectRef)))

	return database.List(ctx, query, entToRoles)
}

func (this *RoleEntRepository) Exist(ctx context.Context, param it.ExistRoleParam) (bool, error) {
	return this.client.Role.Query().
		Where(entRole.IDEQ(param.Id)).
		Exist(ctx)
}

func (this *RoleEntRepository) ExistUserWithRole(ctx context.Context, param it.ExistUserWithRoleParam) (bool, error) {
	return this.client.Role.Query().
		Where(entRole.HasRoleUsersWith(entRoleUser.ReceiverTypeEQ(entRoleUser.ReceiverType(param.ReceiverType)))).
		Where(entRole.HasRoleUsersWith(entRoleUser.ReceiverRefEQ(param.ReceiverId))).
		Exist(ctx)
}

func (this *RoleEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors) {
	return database.ParseSearchGraphStr[ent.Role, domain.Role](criteria, entRole.Label)
}

func (this *RoleEntRepository) Search(
	ctx context.Context,
	param it.SearchParam,
) (*crud.PagedResult[domain.Role], error) {
	query := this.client.Role.Query()

	return database.Search(
		ctx,
		param.Predicate,
		param.Order,
		crud.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		entToRoles,
	)
}

func (this *RoleEntRepository) AddRemoveUser(ctx context.Context, param it.AddRemoveUserParam) error {
	if param.Add {
		_, err := this.client.RoleUser.Create().
			SetApproverID(param.ApproverID).
			SetReceiverRef(param.ReceiverID).
			SetReceiverType(entRoleUser.ReceiverType(param.ReceiverType)).
			SetRoleID(param.Id).
			Save(ctx)
		return err
	}

	_, err := this.client.RoleUser.Delete().
		Where(
			entRoleUser.ReceiverRefEQ(param.ReceiverID),
			entRoleUser.ReceiverTypeEQ(entRoleUser.ReceiverType(param.ReceiverType)),
			entRoleUser.RoleIDEQ(param.Id),
		).
		Exec(ctx)
	return err
}

func (this *RoleEntRepository) createRoleTx(ctx context.Context, tx *ent.Tx, role domain.Role) (*ent.Role, error) {
	return tx.Role.
		Create().
		SetID(*role.Id).
		SetEtag(*role.Etag).
		SetName(*role.Name).
		SetNillableDescription(role.Description).
		SetOwnerType(entRole.OwnerType(*role.OwnerType)).
		SetOwnerRef(*role.OwnerRef).
		SetIsRequestable(*role.IsRequestable).
		SetIsRequiredAttachment(*role.IsRequiredAttachment).
		SetIsRequiredComment(*role.IsRequiredComment).
		SetCreatedBy(*role.CreatedBy).
		Save(ctx)
}

func (this *RoleEntRepository) createAssignmentTx(ctx context.Context, tx *ent.Tx, roleID model.Id, entitlementID model.Id) error {
	entitlement, err := tx.Entitlement.
		Query().
		Where(entEntitlement.IDEQ(entitlementID)).
		WithAction().
		WithResource().
		Only(ctx)
	fault.PanicOnErr(err)

	var actionName *string
	if entitlement.Edges.Action != nil {
		actionName = &entitlement.Edges.Action.Name
	}

	var resourceName *string
	if entitlement.Edges.Resource != nil {
		resourceName = &entitlement.Edges.Resource.Name
	}

	scopeRef := "*"
	if entitlement.ScopeRef != nil {
		scopeRef = *entitlement.ScopeRef
	}

	actionExpr := "*"
	if actionName != nil {
		actionExpr = *actionName
	}

	resourceExpr := "*"
	if resourceName != nil {
		resourceExpr = *resourceName
	}

	resolvedExpr := fmt.Sprintf("%s:%s:%s.%s", roleID, actionExpr, scopeRef, resourceExpr)

	// Generate new ID for entitlement assignment
	assignmentID, err := model.NewId()
	fault.PanicOnErr(err)

	_, err = tx.EntitlementAssignment.
		Create().
		SetID(*assignmentID).
		SetSubjectRef(roleID).
		SetSubjectType(entAssign.SubjectTypeNikkiRole).
		SetEntitlementID(entitlementID).
		SetResolvedExpr(resolvedExpr).
		SetNillableActionName(actionName).
		SetNillableResourceName(resourceName).
		Save(ctx)

	return err
}

func (this *RoleEntRepository) updateRole(ctx context.Context, tx *ent.Tx, prevEtag model.Etag, role domain.Role) (*ent.Role, error) {
	_, err := tx.Role.
		UpdateOneID(*role.Id).
		SetName(*role.Name).
		SetNillableDescription(role.Description).
		Where(entRole.EtagEQ(prevEtag)).
		Save(ctx)
	fault.PanicOnErr(err)

	updatedRole, err := tx.Role.
		UpdateOneID(*role.Id).
		SetEtag(*role.Etag).
		Save(ctx)
	fault.PanicOnErr(err)

	return updatedRole, nil
}

func (r *RoleEntRepository) setAssignmentIdNull(ctx context.Context, tx *ent.Tx, assignmentId string) error {
	_, err := tx.PermissionHistory.
		Update().
		Where(entPermissionHistory.EntitlementAssignmentIDEQ(assignmentId)).
		ClearEntitlementAssignmentID().
		Save(ctx)
	return err
}

func (this *RoleEntRepository) removeAssignmentTx(ctx context.Context, tx *ent.Tx, roleID model.Id, entitlementID model.Id) error {
	_, err := tx.EntitlementAssignment.
		Delete().
		Where(
			entAssign.SubjectTypeEQ(entAssign.SubjectTypeNikkiRole),
			entAssign.SubjectRefEQ(roleID),
			entAssign.EntitlementIDEQ(entitlementID),
		).
		Exec(ctx)
	return err
}

func (this *RoleEntRepository) deleteRoleTx(ctx context.Context, tx *ent.Tx, roleId model.Id) (int, error) {
	deletedCount, err := tx.Role.
		Delete().
		Where(entRole.IDEQ(roleId)).
		Exec(ctx)
	return deletedCount, err
}

func (this *RoleEntRepository) setGrantRequestBeforeDeleteTx(ctx context.Context, tx *ent.Tx, roleId, roleName string) error {
	_, err := tx.GrantRequest.
		Update().
		SetTargetRoleName(roleName).
		Where(entGrantRequest.TargetRoleID(roleId)).
		ClearTargetRoleID().
		Save(ctx)
	return err
}

func (this *RoleEntRepository) setRevokeRequestBeforeDeleteTx(ctx context.Context, tx *ent.Tx, roleId, roleName string) error {
	_, err := tx.RevokeRequest.
		Update().
		SetTargetRoleName(roleName).
		Where(entRevokeRequest.TargetRoleID(roleId)).
		ClearTargetRoleID().
		Save(ctx)
	return err
}

func (this *RoleEntRepository) setPermissionHistoryBeforeDeleteTx(ctx context.Context, tx *ent.Tx, roleId, roleName string) error {
	_, err := tx.PermissionHistory.
		Update().
		SetRoleName(roleName).
		Where(entPermissionHistory.RoleIDEQ(roleId)).
		ClearRoleID().
		Save(ctx)
	return err
}

func BuildRoleDescriptor() *orm.EntityDescriptor {
	entity := ent.Role{}
	builder := orm.DescribeEntity(entRole.Label).
		Aliases("roles").
		Field(entRole.FieldID, entity.ID).
		Field(entRole.FieldEtag, entity.Etag).
		Field(entRole.FieldName, entity.Name).
		Field(entRole.FieldDescription, entity.Description).
		Field(entRole.FieldOwnerType, entity.OwnerType).
		Field(entRole.FieldOwnerRef, entity.OwnerRef).
		Field(entRole.FieldIsRequestable, entity.IsRequestable).
		Field(entRole.FieldIsRequiredAttachment, entity.IsRequiredAttachment).
		Field(entRole.FieldIsRequiredComment, entity.IsRequiredComment).
		Field(entRole.FieldCreatedBy, entity.CreatedBy).
		Field(entRole.FieldCreatedAt, entity.CreatedAt)

	return builder.Descriptor()
}
