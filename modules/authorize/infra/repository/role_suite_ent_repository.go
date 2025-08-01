package repository

import (
	"context"
	"time"

	"github.com/sky-as-code/nikki-erp/common/array"
	"github.com/sky-as-code/nikki-erp/common/crud"
	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/orm"
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

type RoleSuiteEntRepository struct {
	client *ent.Client
}

func (this *RoleSuiteEntRepository) Create(ctx context.Context, roleSuite domain.RoleSuite) (*domain.RoleSuite, error) {
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

	if len(roleSuite.Roles) > 0 {
		roleIds := array.Map(roleSuite.Roles, func(role domain.Role) string {
			return *role.Id
		})
		creation.AddRoleIDs(roleIds...)
	}

	return database.Mutate(ctx, creation, ent.IsNotFound, entToRoleSuite)
}

func (this *RoleSuiteEntRepository) FindById(ctx context.Context, param it.FindByIdParam) (*domain.RoleSuite, error) {
	query := this.client.RoleSuite.Query().
		Where(entRoleSuite.IDEQ(param.Id)).
		WithRoles()

	return database.FindOne(ctx, query, ent.IsNotFound, entToRoleSuite)
}

func (this *RoleSuiteEntRepository) FindByName(ctx context.Context, param it.FindByNameParam) (*domain.RoleSuite, error) {
	query := this.client.RoleSuite.Query().
		Where(entRoleSuite.NameEQ(param.Name))

	return database.FindOne(ctx, query, ent.IsNotFound, entToRoleSuite)
}

func (this *RoleSuiteEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors) {
	return database.ParseSearchGraphStr[ent.RoleSuite, domain.RoleSuite](criteria, entRoleSuite.Label)
}

func (this *RoleSuiteEntRepository) Search(
	ctx context.Context,
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

func (this *RoleSuiteEntRepository) FindAllBySubject(ctx context.Context, param it.FindAllBySubjectParam) ([]domain.RoleSuite, error) {
	query := this.client.RoleSuite.Query().
		Where(entRoleSuite.HasRolesuiteUsersWith(entRoleSuiteUser.ReceiverRefEQ(param.SubjectRef))).
		WithRoles()

	return database.List(ctx, query, entToRoleSuites)
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
