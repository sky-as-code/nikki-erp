package repository

import (
	"context"
	"time"

	"github.com/sky-as-code/nikki-erp/common/crud"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
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

func (this *OrganizationEntRepository) Create(ctx context.Context, org domain.Organization) (*domain.Organization, error) {
	creation := this.client.Organization.Create().
		SetID(*org.Id).
		SetNillableAddress(org.Address).
		SetDisplayName(*org.DisplayName).
		SetNillableLegalName(org.LegalName).
		SetNillablePhoneNumber(org.PhoneNumber).
		SetSlug(*org.Slug).
		SetStatus(entOrg.Status(*org.Status)).
		SetEtag(*org.Etag)

	return db.Mutate(ctx, creation, ent.IsNotFound, entToOrganization)
}

func (this *OrganizationEntRepository) Update(ctx context.Context, org domain.Organization, prevEtag model.Etag) (*domain.Organization, error) {
	update := this.client.Organization.UpdateOneID(*org.Id).
		SetNillableAddress(org.Address).
		SetNillableDisplayName(org.DisplayName).
		SetNillableLegalName(org.LegalName).
		SetNillablePhoneNumber(org.PhoneNumber).
		SetNillableStatus((*entOrg.Status)(org.Status)).
		// IMPORTANT: Must have!
		Where(entOrg.EtagEQ(prevEtag))

	if len(update.Mutation().Fields()) > 0 {
		update.
			SetEtag(*org.Etag).
			SetUpdatedAt(time.Now())
	}

	return db.Mutate(ctx, update, ent.IsNotFound, entToOrganization)
}

func (this *OrganizationEntRepository) DeleteSoft(ctx context.Context, id model.Id) (*domain.Organization, error) {
	update := this.client.Organization.UpdateOneID(id).
		SetDeletedAt(time.Now())

	return db.Mutate(ctx, update, ent.IsNotFound, entToOrganization)
}

func (this *OrganizationEntRepository) DeleteHard(ctx context.Context, id model.Id) error {
	return db.Delete[ent.Organization](ctx, this.client.Organization.DeleteOneID(id))
}

func (this *OrganizationEntRepository) FindById(ctx context.Context, id model.Id) (*domain.Organization, error) {
	query := this.client.Organization.Query().
		Where(entOrg.ID(id)).
		WithUsers()

	return db.FindOne(ctx, query, ent.IsNotFound, entToOrganization)
}

func (this *OrganizationEntRepository) FindBySlug(ctx context.Context, query it.GetOrganizationBySlugQuery) (*domain.Organization, error) {
	builder := this.client.Organization.Query().
		Where(entOrg.Slug(query.Slug))

	if !query.IncludeDeleted {
		builder = builder.Where(entOrg.DeletedAtIsNil())
	}

	builder = builder.WithUsers()
	return db.FindOne(ctx, builder, ent.IsNotFound, entToOrganization)
}

func (this *OrganizationEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.Organization, domain.Organization](criteria, entOrg.Label)
}

func (this *OrganizationEntRepository) Search(
	ctx context.Context,
	predicate *orm.Predicate,
	order []orm.OrderOption,
	opts it.SearchOrganizationsQuery,
) (*crud.PagedResult[domain.Organization], error) {
	query := this.client.Organization.Query()
	if !opts.IncludeDeleted {
		query = query.Where(entOrg.DeletedAtIsNil())
	}

	return db.Search(
		ctx,
		predicate,
		order,
		crud.PagingOptions{
			Page: *opts.Page,
			Size: *opts.Size,
		},
		query,
		entToOrganizations,
	)
}

func BuildOrganizationDescriptor() *orm.EntityDescriptor {
	entity := ent.Organization{}
	builder := orm.DescribeEntity(entOrg.Label).
		Aliases("orgs").
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
