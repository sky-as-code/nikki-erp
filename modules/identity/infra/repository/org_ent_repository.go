package repository

import (
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	"github.com/sky-as-code/nikki-erp/modules/identity/infra/ent"
	entOrg "github.com/sky-as-code/nikki-erp/modules/identity/infra/ent/organization"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/organization"
)

func NewOrganizationEntRepository(client *ent.Client) it.OrganizationRepository {
	return &OrganizationEntRepository{
		client: client,
	}
}

type OrganizationEntRepository struct {
	client *ent.Client
}

func (this *OrganizationEntRepository) orgClient(ctx crud.Context) *ent.OrganizationClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.Organization
	}
	return this.client.Organization
}

func (this *OrganizationEntRepository) Create(ctx crud.Context, org *domain.Organization) (*domain.Organization, error) {
	creation := this.orgClient(ctx).Create().
		SetID(*org.Id).
		SetNillableAddress(org.Address).
		SetDisplayName(*org.DisplayName).
		SetEtag(*org.Etag).
		SetNillableLegalName(org.LegalName).
		SetNillablePhoneNumber(org.PhoneNumber).
		SetSlug(*org.Slug).
		SetStatus(string(*org.Status))

	return db.Mutate(ctx, creation, ent.IsNotFound, entToOrganization)
}

func (this *OrganizationEntRepository) Update(ctx crud.Context, org *domain.Organization, prevEtag model.Etag) (*domain.Organization, error) {
	update := this.orgClient(ctx).UpdateOneID(*org.Id).
		SetNillableAddress(org.Address).
		SetNillableDisplayName(org.DisplayName).
		SetNillableLegalName(org.LegalName).
		SetNillablePhoneNumber(org.PhoneNumber).
		SetNillableStatus((*string)(org.Status)).
		// IMPORTANT: Must have!
		Where(entOrg.EtagEQ(prevEtag))

	if len(update.Mutation().Fields()) > 0 {
		update.
			SetEtag(*org.Etag).
			SetUpdatedAt(time.Now())
	}

	return db.Mutate(ctx, update, ent.IsNotFound, entToOrganization)
}

func (this *OrganizationEntRepository) DeleteSoft(ctx crud.Context, param it.DeleteParam) (*domain.Organization, error) {
	update := this.orgClient(ctx).UpdateOneID(param.Slug).
		SetDeletedAt(time.Now())

	return db.Mutate(ctx, update, ent.IsNotFound, entToOrganization)
}

func (this *OrganizationEntRepository) DeleteHard(ctx crud.Context, param it.DeleteParam) (int, error) {
	return this.orgClient(ctx).Delete().
		Where(entOrg.Slug(param.Slug)).
		Exec(ctx)
}

func (this *OrganizationEntRepository) FindById(ctx crud.Context, id model.Id) (*domain.Organization, error) {
	query := this.orgClient(ctx).Query().
		Where(entOrg.ID(id)).
		WithUsers()

	return db.FindOne(ctx, query, ent.IsNotFound, entToOrganization)
}

func (this *OrganizationEntRepository) FindBySlug(ctx crud.Context, param it.FindBySlugParam) (*domain.Organization, error) {
	query := this.orgClient(ctx).Query().
		Where(entOrg.Slug(param.Slug)).
		WithUsers()

	query = this.queryIncludeDeleted(query, param.IncludeDeleted)

	return db.FindOne(ctx, query, ent.IsNotFound, entToOrganization)
}

func (this *OrganizationEntRepository) queryIncludeDeleted(query *ent.OrganizationQuery, includeDeleted bool) *ent.OrganizationQuery {
	if includeDeleted {
		return query.Where(entOrg.Or(
			entOrg.DeletedAtNotNil(),
			entOrg.DeletedAtIsNil(),
		))
	}
	return query.Where(entOrg.DeletedAtIsNil())
}

func (this *OrganizationEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.Organization, domain.Organization](criteria, entOrg.Label)
}

func (this *OrganizationEntRepository) Search(
	ctx crud.Context, param it.SearchParam,
) (*crud.PagedResult[domain.Organization], error) {
	query := this.orgClient(ctx).Query()

	query = this.queryIncludeDeleted(query, param.IncludeDeleted)

	return db.Search(
		ctx,
		param.Predicate,
		param.Order,
		crud.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		entToOrganizations,
	)
}

func (this *OrganizationEntRepository) AddRemoveUsers(ctx crud.Context, param it.AddRemoveUsersParam) (*ft.ClientError, error) {
	if len(param.Add) == 0 && len(param.Remove) == 0 {
		return nil, nil
	}
	err := this.orgClient(ctx).UpdateOneID(param.OrgId).
		AddUserIDs(param.Add...).
		RemoveUserIDs(param.Remove...).
		SetEtag(param.Etag).
		Exec(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return &ft.ClientError{
				Code:    "not_found",
				Details: "some resource doesn't exist",
			}, nil
		}
		return nil, err
	}

	return nil, nil
}

func (this *OrganizationEntRepository) Exists(ctx crud.Context, id model.Id) (bool, error) {
	return this.orgClient(ctx).Query().
		Where(entOrg.ID(id)).
		Exist(ctx)
}

func BuildOrganizationDescriptor() *orm.EntityDescriptor {
	entity := ent.Organization{}
	builder := orm.DescribeEntity(entOrg.Label).
		Aliases("orgs", "org").
		Field(entOrg.FieldCreatedAt, entity.CreatedAt).
		Field(entOrg.FieldDisplayName, entity.DisplayName).
		Field(entOrg.FieldEtag, entity.Etag).
		Field(entOrg.FieldID, entity.ID).
		Field(entOrg.FieldSlug, entity.Slug).
		Field(entOrg.FieldStatus, entity.Status).
		Field(entOrg.FieldUpdatedAt, entity.UpdatedAt).
		Edge(entOrg.EdgeUsers, orm.ToEdgePredicate(entOrg.HasUsersWith))

	return builder.Descriptor()
}
