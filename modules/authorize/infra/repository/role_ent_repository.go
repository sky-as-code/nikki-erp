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
	entRole "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/role"
	entRoleUser "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/roleuser"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/role"
)

func NewRoleEntRepository(client *ent.Client) it.RoleRepository {
	return &RoleEntRepository{
		client: client,
	}
}

func (this *RoleEntRepository) BeginTransaction(ctx crud.Context) (*ent.Tx, error) {
	return this.client.Tx(ctx)
}

func (this *RoleEntRepository) roleClient(ctx crud.Context) *ent.RoleClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.Role
	}
	return this.client.Role
}

func (this *RoleEntRepository) Create(ctx crud.Context, role *domain.Role) (*domain.Role, error) {
	creation := this.roleClient(ctx).Create().
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
		SetNillableOrgID(role.OrgId).
		SetCreatedAt(time.Now())

	return database.Mutate(ctx, creation, ent.IsNotFound, entToRole)
}

func (this *RoleEntRepository) Update(ctx crud.Context, role *domain.Role, prevEtag model.Etag) (*domain.Role, error) {
	updation := this.roleClient(ctx).UpdateOneID(*role.Id).
		SetNillableDescription(role.Description).
		SetName(*role.Name).
		Where(entRole.EtagEQ(prevEtag))

	if len(updation.Mutation().Fields()) > 0 {
		updation.
			SetEtag(*role.Etag)
	}

	return database.Mutate(ctx, updation, ent.IsNotFound, entToRole)
}

func (this *RoleEntRepository) DeleteHard(ctx crud.Context, param it.DeleteRoleHardParam) (int, error) {
	return this.roleClient(ctx).Delete().
		Where(entRole.IDEQ(param.Id)).
		Exec(ctx)
}

func (this *RoleEntRepository) FindByName(ctx crud.Context, param it.FindByNameParam) (*domain.Role, error) {
	query := this.roleClient(ctx).Query().
		Where(entRole.NameEQ(param.Name))

	if param.OrgId != nil {
		query = query.Where(entRole.OrgIDEQ(*param.OrgId))
	} else {
		query = query.Where(entRole.OrgIDIsNil())
	}

	return database.FindOne(ctx, query, ent.IsNotFound, entToRole)
}

func (this *RoleEntRepository) FindById(ctx crud.Context, param it.FindByIdParam) (*domain.Role, error) {
	query := this.roleClient(ctx).Query().
		Where(entRole.IDEQ(param.Id))

	return database.FindOne(ctx, query, ent.IsNotFound, entToRole)
}

func (this *RoleEntRepository) Exist(ctx crud.Context, param it.ExistRoleParam) (bool, error) {
	return this.roleClient(ctx).Query().
		Where(entRole.IDEQ(param.Id)).
		Exist(ctx)
}

func (this *RoleEntRepository) ExistUserWithRole(ctx crud.Context, param it.ExistUserWithRoleParam) (bool, error) {
	return this.roleClient(ctx).Query().
		Where(
			entRole.HasRoleUsersWith(
				entRoleUser.ReceiverTypeEQ(entRoleUser.ReceiverType(param.ReceiverType)),
				entRoleUser.ReceiverRefEQ(param.ReceiverId),
				entRoleUser.RoleIDEQ(param.TargetId),
			),
		).
		Exist(ctx)
}

func (this *RoleEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors) {
	return database.ParseSearchGraphStr[ent.Role, domain.Role](criteria, entRole.Label)
}

func (this *RoleEntRepository) Search(
	ctx crud.Context,
	param it.SearchParam,
) (*crud.PagedResult[domain.Role], error) {
	query := this.roleClient(ctx).Query()

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

func (this *RoleEntRepository) AddRemoveUser(ctx crud.Context, param it.AddRemoveUserParam) error {
	var creation *ent.RoleUserCreate
	var deletion *ent.RoleUserDelete
	tx := ctx.GetDbTranx()

	if tx != nil {
		creation = tx.(*ent.Tx).RoleUser.Create()
		deletion = tx.(*ent.Tx).RoleUser.Delete()
	} else {
		creation = this.client.RoleUser.Create()
		deletion = this.client.RoleUser.Delete()
	}

	if param.Add {
		_, err := creation.
			SetApproverID(param.ApproverID).
			SetReceiverRef(param.ReceiverID).
			SetReceiverType(entRoleUser.ReceiverType(param.ReceiverType)).
			SetRoleID(param.Id).
			Save(ctx)
		return err
	}

	_, err := deletion.
		Where(
			entRoleUser.ReceiverRefEQ(param.ReceiverID),
			entRoleUser.ReceiverTypeEQ(entRoleUser.ReceiverType(param.ReceiverType)),
			entRoleUser.RoleIDEQ(param.Id),
		).
		Exec(ctx)
	return err
}

func (this *RoleEntRepository) FindAllBySubject(ctx crud.Context, param it.FindAllBySubjectParam) ([]domain.Role, error) {
	query := this.roleClient(ctx).Query().
		Where(entRole.HasRoleUsersWith(entRoleUser.ReceiverRefEQ(param.SubjectRef)))

	return database.List(ctx, query, entToRoles)
}

type RoleEntRepository struct {
	client *ent.Client
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
		Field(entRole.FieldCreatedAt, entity.CreatedAt).
		Edge(entRole.EdgeRoleUsers, orm.ToEdgePredicate(entRole.HasRoleUsersWith))

	return builder.Descriptor()
}
