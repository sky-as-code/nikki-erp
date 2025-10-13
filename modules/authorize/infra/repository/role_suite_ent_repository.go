package repository

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/database"

	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entRoleSuite "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/rolesuite"
	entRoleSuiteUser "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/rolesuiteuser"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role_suite"
)

func NewRoleSuiteEntRepository(client *ent.Client) it.RoleSuiteRepository {
	return &RoleSuiteEntRepository{
		client: client,
	}
}

func (this *RoleSuiteEntRepository) suiteClient(ctx crud.Context) *ent.RoleSuiteClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.RoleSuite
	}
	return this.client.RoleSuite
}

func (this *RoleSuiteEntRepository) BeginTransaction(ctx crud.Context) (*ent.Tx, error) {
	return this.client.Tx(ctx)
}

func (this *RoleSuiteEntRepository) Create(ctx crud.Context, roleSuite domain.RoleSuite, roleIds []model.Id) (*domain.RoleSuite, error) {
	creation := this.suiteClient(ctx).Create().
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
		SetNillableOrgID(roleSuite.OrgId).
		SetCreatedAt(time.Now())

	if len(roleIds) > 0 {
		creation.AddRoleIDs(roleIds...)
	}

	return database.Mutate(ctx, creation, ent.IsNotFound, entToRoleSuite)
}

func (this *RoleSuiteEntRepository) Update(ctx crud.Context, roleSuite domain.RoleSuite, prevEtag model.Etag, addRoleIds, removeRoleIds []model.Id) (*domain.RoleSuite, error) {
	update := this.suiteClient(ctx).UpdateOneID(*roleSuite.Id).
		SetNillableName(roleSuite.Name).
		SetNillableDescription(roleSuite.Description).
		Where(
			entRoleSuite.EtagEQ(prevEtag),
		)

	if len(addRoleIds) > 0 {
		update.AddRoleIDs(addRoleIds...)
	}

	if len(removeRoleIds) > 0 {
		update.RemoveRoleIDs(removeRoleIds...)
	}

	if len(update.Mutation().Fields()) > 0 {
		update.
			SetEtag(*roleSuite.Etag)
	}

	return database.Mutate(ctx, update, ent.IsNotFound, entToRoleSuite)
}

func (this *RoleSuiteEntRepository) DeleteHard(ctx crud.Context, param it.DeleteRoleSuiteParam) (int, error) {
	return this.suiteClient(ctx).Delete().
		Where(entRoleSuite.IDEQ(param.Id)).
		Exec(ctx)
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

	if param.OrgId != nil {
		query = query.Where(entRoleSuite.OrgIDEQ(*param.OrgId))
	} else {
		query = query.Where(entRoleSuite.OrgIDIsNil())
	}

	return database.FindOne(ctx, query, ent.IsNotFound, entToRoleSuite)
}

func (this *RoleSuiteEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors) {
	return database.ParseSearchGraphStr[ent.RoleSuite, domain.RoleSuite](criteria, entRoleSuite.Label)
}

func (this *RoleSuiteEntRepository) FindAllBySubject(ctx crud.Context, param it.FindAllBySubjectParam) ([]domain.RoleSuite, error) {
	query := this.suiteClient(ctx).Query().
		Where(entRoleSuite.HasRolesuiteUsersWith(entRoleSuiteUser.ReceiverRefEQ(param.SubjectRef))).
		WithRoles()

	return database.List(ctx, query, entToRoleSuites)
}

func (this *RoleSuiteEntRepository) Search(
	ctx crud.Context,
	param it.SearchParam,
) (*crud.PagedResult[domain.RoleSuite], error) {
	query := this.suiteClient(ctx).Query().
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

func (this *RoleSuiteEntRepository) ExistUserWithRoleSuite(ctx crud.Context, param it.ExistUserWithRoleSuiteParam) (bool, error) {
	return this.suiteClient(ctx).Query().
		Where(
			entRoleSuite.HasRolesuiteUsersWith(
				entRoleSuiteUser.ReceiverTypeEQ(entRoleSuiteUser.ReceiverType(param.ReceiverType)),
				entRoleSuiteUser.ReceiverRefEQ(param.ReceiverId),
				entRoleSuiteUser.RoleSuiteIDEQ(param.TargetId),
			),
		).
		Exist(ctx)
}

func (this *RoleSuiteEntRepository) AddRemoveUser(ctx crud.Context, param it.AddRemoveUserParam) error {
	creation := this.client.RoleSuiteUser.Create()
	deletion := this.client.RoleSuiteUser.Delete()
	tx := ctx.GetDbTranx()

	if tx != nil {
		creation = tx.(*ent.Tx).RoleSuiteUser.Create()
		deletion = tx.(*ent.Tx).RoleSuiteUser.Delete()
	} else {
		creation = this.client.RoleSuiteUser.Create()
		deletion = this.client.RoleSuiteUser.Delete()
	}

	if param.Add {
		_, err := creation.
			SetApproverID(param.ApproverID).
			SetReceiverRef(param.ReceiverID).
			SetReceiverType(entRoleSuiteUser.ReceiverType(param.ReceiverType)).
			SetRoleSuiteID(param.Id).
			Save(ctx)
		return err
	}

	_, err := deletion.
		Where(
			entRoleSuiteUser.ReceiverRefEQ(param.ReceiverID),
			entRoleSuiteUser.ReceiverTypeEQ(entRoleSuiteUser.ReceiverType(param.ReceiverType)),
			entRoleSuiteUser.RoleSuiteIDEQ(param.Id),
		).
		Exec(ctx)
	return err
}

type RoleSuiteEntRepository struct {
	client *ent.Client
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
		Field(entRoleSuite.FieldOrgID, entity.OrgID).
		Field(entRoleSuite.FieldCreatedAt, entity.CreatedAt)

	return builder.Descriptor()
}
