package repository

import (
	"context"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/identity/domain"
	"github.com/sky-as-code/nikki-erp/modules/identity/infra/ent"
	entOrg "github.com/sky-as-code/nikki-erp/modules/identity/infra/ent/organization"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/user"
)

func NewOrganizationEntRepository(client *ent.Client) it.OrganizationRepository {
	return &OrganizationEntRepository{
		client: client,
	}
}

type OrganizationEntRepository struct {
	client *ent.Client
}

func (this *OrganizationEntRepository) Create(ctx context.Context, org domain.Organization) (*domain.Organization, error) {
	creation := this.client.Organization.Create().
		SetID(org.Id.String()).
		SetDisplayName(*org.DisplayName).
		SetSlug(org.Slug.String()).
		SetEtag(org.Etag.String()).
		SetStatus(entOrg.Status(*org.Status))

	return db.Mutate(ctx, creation, entToOrganization)
}

func (this *OrganizationEntRepository) Update(ctx context.Context, org domain.Organization) (*domain.Organization, error) {
	update := this.client.Organization.UpdateOneID(org.Id.String()).
		SetDisplayName(*org.DisplayName).
		SetEtag(org.Etag.String()).
		SetStatus(entOrg.Status(*org.Status))

	return db.Mutate(ctx, update, entToOrganization)
}

func (this *OrganizationEntRepository) Delete(ctx context.Context, id model.Id) error {
	return db.Delete[ent.Organization](ctx, this.client.Organization.DeleteOneID(id.String()))
}

func (this *OrganizationEntRepository) FindById(ctx context.Context, id model.Id) (*domain.Organization, error) {
	query := this.client.Organization.Query().
		Where(entOrg.ID(id.String())).
		WithUsers()

	return db.FindOne(ctx, query, entToOrganization)
}

func (this *OrganizationEntRepository) FindBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	org, err := this.client.Organization.Query().
		Where(entOrg.Slug(slug)).
		WithUsers().
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entToOrganization(org), nil
}

func (this *OrganizationEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.Organization, domain.Organization](criteria)
}

func (this *OrganizationEntRepository) Search(
	ctx context.Context,
	predicate *orm.Predicate,
	order []orm.OrderOption,
	opts crud.PagingOptions,
) (*crud.PagedResult[domain.Organization], error) {
	return db.Search(
		ctx,
		predicate,
		order,
		opts,
		this.client.Organization.Query(),
		entToOrganizations,
	)
}

func BuildOrganizationDescriptor() *orm.EntityDescriptor {
	entity := ent.Organization{}
	builder := orm.DescribeEntity(entOrg.Label).
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
