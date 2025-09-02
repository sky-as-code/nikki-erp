package repository

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entGrantRequest "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/grantrequest"
	entPermissionHistory "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/permissionhistory"
	entRevokeRequest "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/revokerequest"
	entRoleSuite "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/rolesuite"
	entRoleSuiteUser "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/rolesuiteuser"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role_suite"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/database"
)

func NewRoleSuiteEntRepository(client *ent.Client) it.RoleSuiteRepository {
	return &RoleSuiteEntRepository{
		client: client,
	}
}

type RoleSuiteEntRepository struct {
	client *ent.Client
}

func (this *RoleSuiteEntRepository) Create(ctx crud.Context, roleSuite domain.RoleSuite, roleIds []model.Id) (*domain.RoleSuite, error) {
	creation := this.client.RoleSuite.Create().
		SetID(*roleSuite.Id).
		SetName(*roleSuite.Name).
		SetNillableDescription(roleSuite.Description).
		SetEtag(*roleSuite.Etag).
		SetOwnerType(entRoleSuite.OwnerType(*roleSuite.OwnerType)).
		SetOwnerRef(*roleSuite.OwnerRef).
		SetIsRequestable(*roleSuite.IsRequestable).
		SetIsRequiredAttachment(*roleSuite.IsRequiredAttachment).
		SetIsRequiredComment(*roleSuite.IsRequiredComment).
		SetCreatedBy(*roleSuite.CreatedBy).
		SetCreatedAt(time.Now())

	if len(roleIds) > 0 {
		creation.AddRoleIDs(roleIds...)
	}

	return database.Mutate(ctx, creation, ent.IsNotFound, entToRoleSuite)
}

// There is currently no update reason in permission histories for users/groups with this role_suite.
func (this *RoleSuiteEntRepository) UpdateTx(ctx crud.Context, roleSuite domain.RoleSuite, prevEtag model.Etag, addRoleIds, removeRoleIds []model.Id) (*domain.RoleSuite, error) {
	tx, err := this.client.Tx(ctx)
	fault.PanicOnErr(err)

	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "update role suite transaction"); e != nil {
			_ = tx.Rollback()
			err = e
		}
	}()

	updatedRoleSuite, err := this.updateRoleSuiteTx(ctx, tx, roleSuite, prevEtag, addRoleIds, removeRoleIds)
	fault.PanicOnErr(err)

	fault.PanicOnErr(tx.Commit())
	return updatedRoleSuite, nil
}

// There is currently no delete reason in permission histories for users/groups with this role_suite.
func (this *RoleSuiteEntRepository) DeleteHardTx(ctx crud.Context, param it.DeleteRoleSuiteParam) (int, error) {
	tx, err := this.client.Tx(ctx)
	fault.PanicOnErr(err)

	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "delete role suite transaction"); e != nil {
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

	deletedCount, err := this.deleteRoleSuiteTx(ctx, tx, param.Id)
	fault.PanicOnErr(err)

	fault.PanicOnErr(tx.Commit())
	return deletedCount, nil
}

func (this *RoleSuiteEntRepository) FindById(ctx crud.Context, param it.FindByIdParam) (*domain.RoleSuite, error) {
	query := this.client.RoleSuite.Query().
		Where(entRoleSuite.IDEQ(param.Id)).
		WithRoles()

	return database.FindOne(ctx, query, ent.IsNotFound, entToRoleSuite)
}

func (this *RoleSuiteEntRepository) FindByName(ctx crud.Context, param it.FindByNameParam) (*domain.RoleSuite, error) {
	query := this.client.RoleSuite.Query().
		Where(entRoleSuite.NameEQ(param.Name))

	return database.FindOne(ctx, query, ent.IsNotFound, entToRoleSuite)
}

func (this *RoleSuiteEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors) {
	return database.ParseSearchGraphStr[ent.RoleSuite, domain.RoleSuite](criteria, entRoleSuite.Label)
}

func (this *RoleSuiteEntRepository) Search(
	ctx crud.Context,
	param it.SearchParam,
) (*crud.PagedResult[domain.RoleSuite], error) {
	query := this.client.RoleSuite.Query().
		WithRoles()

	return database.Search(
		ctx,
		param.Predicate,
		param.Order,
		crud.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		entToRoleSuites,
	)
}

func (this *RoleSuiteEntRepository) FindAllBySubject(ctx crud.Context, param it.FindAllBySubjectParam) ([]domain.RoleSuite, error) {
	query := this.client.RoleSuite.Query().
		Where(entRoleSuite.HasRolesuiteUsersWith(entRoleSuiteUser.ReceiverRefEQ(param.SubjectRef))).
		WithRoles()

	return database.List(ctx, query, entToRoleSuites)
}

func (this *RoleSuiteEntRepository) deleteRoleSuiteTx(ctx crud.Context, tx *ent.Tx, roleSuiteId model.Id) (int, error) {
	deletedCount, err := tx.RoleSuite.
		Delete().
		Where(entRoleSuite.IDEQ(roleSuiteId)).
		Exec(ctx)
	return deletedCount, err
}

func (this *RoleSuiteEntRepository) setGrantRequestBeforeDeleteTx(ctx crud.Context, tx *ent.Tx, roleSuiteId, nameSuite string) error {
	_, err := tx.GrantRequest.
		Update().
		SetTargetSuiteName(nameSuite).
		Where(entGrantRequest.TargetSuiteID(roleSuiteId)).
		ClearTargetSuiteID().
		Save(ctx)

	return err
}

func (this *RoleSuiteEntRepository) setRevokeRequestBeforeDeleteTx(ctx crud.Context, tx *ent.Tx, roleSuiteId, nameSuite string) error {
	_, err := tx.RevokeRequest.
		Update().
		SetTargetSuiteName(nameSuite).
		Where(entRevokeRequest.TargetSuiteID(roleSuiteId)).
		ClearTargetSuiteID().
		Save(ctx)

	return err
}

func (this *RoleSuiteEntRepository) setPermissionHistoryBeforeDeleteTx(ctx crud.Context, tx *ent.Tx, roleSuiteId, nameSuite string) error {
	_, err := tx.PermissionHistory.
		Update().
		SetRoleSuiteName(nameSuite).
		Where(entPermissionHistory.RoleSuiteID(roleSuiteId)).
		ClearRoleSuiteID().
		Save(ctx)

	return err
}

func (this *RoleSuiteEntRepository) updateRoleSuiteTx(ctx crud.Context, tx *ent.Tx, roleSuite domain.RoleSuite, prevEtag model.Etag, addRoleIds, removeRoleIds []model.Id) (*domain.RoleSuite, error) {
	updation := tx.RoleSuite.UpdateOneID(*roleSuite.Id).
		SetNillableName(roleSuite.Name).
		SetNillableDescription(roleSuite.Description).
		SetEtag(*roleSuite.Etag).
		Where(
			entRoleSuite.IDEQ(*roleSuite.Id),
			entRoleSuite.EtagEQ(prevEtag),
		)

	if len(addRoleIds) > 0 {
		updation.AddRoleIDs(addRoleIds...)
	}

	if len(removeRoleIds) > 0 {
		updation.RemoveRoleIDs(removeRoleIds...)
	}

	updatedRoleSuite, err := updation.Save(ctx)
	fault.PanicOnErr(err)

	return entToRoleSuite(updatedRoleSuite), nil
}

func BuildRoleSuiteDescriptor() *orm.EntityDescriptor {
	entity := ent.RoleSuite{}
	builder := orm.DescribeEntity(entRoleSuite.Label).
		Aliases("role_suites").
		Field(entRoleSuite.FieldID, entity.ID).
		Field(entRoleSuite.FieldName, entity.Name).
		Field(entRoleSuite.FieldDescription, entity.Description).
		Field(entRoleSuite.FieldEtag, entity.Etag).
		Field(entRoleSuite.FieldOwnerType, entity.OwnerType).
		Field(entRoleSuite.FieldOwnerRef, entity.OwnerRef).
		Field(entRoleSuite.FieldIsRequestable, entity.IsRequestable).
		Field(entRoleSuite.FieldIsRequiredAttachment, entity.IsRequiredAttachment).
		Field(entRoleSuite.FieldIsRequiredComment, entity.IsRequiredComment).
		Field(entRoleSuite.FieldCreatedBy, entity.CreatedBy).
		Field(entRoleSuite.FieldCreatedAt, entity.CreatedAt)

	return builder.Descriptor()
}
