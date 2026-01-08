package repository

import (
	"time"

	"github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/core/database"

	"github.com/sky-as-code/nikki-erp/modules/authorize/domain"
	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
	entGrantRequest "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/grantrequest"
	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/authorize/grant_request"
)

func NewGrantRequestEntRepository(client *ent.Client) it.GrantRequestRepository {
	return &GrantRequestEntRepository{
		client: client,
	}
}

func (this *GrantRequestEntRepository) BeginTransaction(ctx crud.Context) (*ent.Tx, error) {
	return this.client.Tx(ctx)
}

func (this *GrantRequestEntRepository) grantRequestClient(ctx crud.Context) *ent.GrantRequestClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.GrantRequest
	}
	return this.client.GrantRequest
}

func (this *GrantRequestEntRepository) Create(ctx crud.Context, grantRequest *domain.GrantRequest) (*domain.GrantRequest, error) {
	creation := this.grantRequestClient(ctx).Create().
		SetID(*grantRequest.Id).
		SetEtag(*grantRequest.Etag).
		SetNillableAttachmentURL(grantRequest.AttachmentURL).
		SetNillableComment(grantRequest.Comment).
		SetCreatedBy(*grantRequest.RequestorId).
		SetReceiverType(entGrantRequest.ReceiverType(*grantRequest.ReceiverType)).
		SetReceiverID(*grantRequest.ReceiverId).
		SetTargetType(entGrantRequest.TargetType(*grantRequest.TargetType)).
		SetStatus(entGrantRequest.Status(*grantRequest.Status)).
		SetNillableTargetRoleName(grantRequest.TargetRoleName).
		SetNillableTargetSuiteName(grantRequest.TargetSuiteName).
		SetCreatedAt(time.Now()).
		SetNillableOrgID(grantRequest.OrgId)

	switch *grantRequest.TargetType {
	case domain.GrantRequestTargetTypeRole:
		creation = creation.SetNillableTargetRoleID(grantRequest.TargetRef)
	case domain.GrantRequestTargetTypeSuite:
		creation = creation.SetNillableTargetSuiteID(grantRequest.TargetRef)
	}

	return database.Mutate(ctx, creation, ent.IsNotFound, entToGrantRequest)
}

func (this *GrantRequestEntRepository) FindAllByTarget(ctx crud.Context, param it.FindAllByTargetParam) ([]domain.GrantRequest, error) {
	query := this.grantRequestClient(ctx).Query()

	switch param.TargetType {
	case domain.GrantRequestTargetTypeRole:
		query = query.Where(
			entGrantRequest.TargetTypeEQ(entGrantRequest.TargetTypeRole),
			entGrantRequest.TargetRoleIDEQ(param.TargetRef),
		)
	case domain.GrantRequestTargetTypeSuite:
		query = query.Where(
			entGrantRequest.TargetTypeEQ(entGrantRequest.TargetTypeSuite),
			entGrantRequest.TargetSuiteIDEQ(param.TargetRef),
		)
	}

	return database.List(ctx, query, entToGrantRequests)
}

func (this *GrantRequestEntRepository) FindById(ctx crud.Context, param it.FindByIdParam) (*domain.GrantRequest, error) {
	query := this.grantRequestClient(ctx).Query().
		Where(entGrantRequest.IDEQ(param.Id)).
		WithGrantResponses().
		WithRole().
		WithRoleSuite()

	return database.FindOne(ctx, query, ent.IsNotFound, entToGrantRequest)
}

func (this *GrantRequestEntRepository) Update(ctx crud.Context, grantRequest *domain.GrantRequest, prevEtag model.Etag) (*domain.GrantRequest, error) {
	update := this.grantRequestClient(ctx).UpdateOneID(*grantRequest.Id).
		SetStatus(entGrantRequest.Status(*grantRequest.Status)).
		Where(entGrantRequest.EtagEQ(prevEtag))

	if len(update.Mutation().Fields()) > 0 {
		update = update.SetEtag(*grantRequest.Etag)
	}

	return database.Mutate(ctx, update, ent.IsNotFound, entToGrantRequest)
}

func (this *GrantRequestEntRepository) ConfigTargetFields(ctx crud.Context, grantRequest *domain.GrantRequest, name string, prevEtag model.Etag) error {
	update := this.grantRequestClient(ctx).Update().
		Where(entGrantRequest.IDEQ(*grantRequest.Id)).
		Where(entGrantRequest.EtagEQ(prevEtag)).
		SetStatus(entGrantRequest.Status(*grantRequest.Status))

	switch *grantRequest.TargetType {
	case domain.GrantRequestTargetTypeRole:
		update = update.
			ClearTargetRoleID().
			SetTargetRoleName(name)
	case domain.GrantRequestTargetTypeSuite:
		update = update.
			ClearTargetSuiteID().
			SetTargetSuiteName(name)
	}

	if len(update.Mutation().Fields()) > 0 {
		update = update.SetEtag(*grantRequest.Etag)
	}

	return update.Exec(ctx)
}

func (this *GrantRequestEntRepository) Delete(ctx crud.Context, param it.DeleteParam) (int, error) {
	return this.grantRequestClient(ctx).Delete().
		Where(entGrantRequest.IDEQ(param.Id)).
		Exec(ctx)
}

func (this *GrantRequestEntRepository) FindPendingByReceiverAndTarget(ctx crud.Context, receiverId model.Id, targetId model.Id, targetType domain.GrantRequestTargetType) ([]domain.GrantRequest, error) {
	query := this.grantRequestClient(ctx).Query().
		Where(
			entGrantRequest.ReceiverIDEQ(receiverId),
			entGrantRequest.StatusEQ(entGrantRequest.StatusPending),
		).
		WithGrantResponses().
		WithRole().
		WithRoleSuite()

	switch targetType {
	case domain.GrantRequestTargetTypeRole:
		query = query.Where(
			entGrantRequest.TargetTypeEQ(entGrantRequest.TargetTypeRole),
			entGrantRequest.TargetRoleIDEQ(targetId),
		)
	case domain.GrantRequestTargetTypeSuite:
		query = query.Where(
			entGrantRequest.TargetTypeEQ(entGrantRequest.TargetTypeSuite),
			entGrantRequest.TargetSuiteIDEQ(targetId),
		)
	}

	return database.List(ctx, query, entToGrantRequests)
}

func (this *GrantRequestEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors) {
	return database.ParseSearchGraphStr[ent.GrantRequest, domain.GrantRequest](criteria, entGrantRequest.Label)
}

func (this *GrantRequestEntRepository) Search(
	ctx crud.Context,
	param it.SearchParam,
) (*crud.PagedResult[domain.GrantRequest], error) {
	query := this.grantRequestClient(ctx).Query().
		WithGrantResponses().
		WithRole().
		WithRoleSuite()

	return database.Search(
		ctx,
		param.Predicate,
		param.Order,
		crud.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		entToGrantRequests,
	)
}

type GrantRequestEntRepository struct {
	client *ent.Client
}

func BuildGrantRequestDescriptor() *orm.EntityDescriptor {
	entity := ent.GrantRequest{}
	builder := orm.DescribeEntity(entGrantRequest.Label).
		Aliases("grant_requests").
		Field(entGrantRequest.FieldID, entity.ID).
		Field(entGrantRequest.FieldComment, entity.Comment).
		Field(entGrantRequest.FieldCreatedAt, entity.CreatedAt).
		Field(entGrantRequest.FieldCreatedBy, entity.CreatedBy).
		Field(entGrantRequest.FieldEtag, entity.Etag).
		Field(entGrantRequest.FieldReceiverID, entity.ReceiverID).
		Field(entGrantRequest.FieldTargetType, entity.TargetType).
		Field(entGrantRequest.FieldTargetRoleID, entity.TargetRoleID).
		Field(entGrantRequest.FieldTargetSuiteID, entity.TargetSuiteID).
		Field(entGrantRequest.FieldStatus, entity.Status).
		Field(entGrantRequest.FieldTargetRoleName, entity.TargetRoleName).
		Field(entGrantRequest.FieldTargetSuiteName, entity.TargetSuiteName)

	return builder.Descriptor()
}
