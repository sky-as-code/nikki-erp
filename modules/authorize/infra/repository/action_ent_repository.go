package repository

// import (
// 	"time"

// 	"github.com/sky-as-code/nikki-erp/common/fault"
// 	"github.com/sky-as-code/nikki-erp/common/model"
// 	"github.com/sky-as-code/nikki-erp/common/orm"
// 	"github.com/sky-as-code/nikki-erp/modules/core/crud"
// 	"github.com/sky-as-code/nikki-erp/modules/core/database"

// 	domain "github.com/sky-as-code/nikki-erp/modules/authorize/domain"
// 	ent "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent"
// 	entAction "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/action"
// 	entEntitlement "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/entitlement"
// 	it "github.com/sky-as-code/nikki-erp/modules/authorize/interfaces/action"
// )

// func NewActionEntRepository(client *ent.Client) it.ActionRepository {
// 	return &ActionEntRepository{
// 		client: client,
// 	}
// }

// // Deprecated: Must create dynamic model repository instead
// type ActionEntRepository struct {
// 	client *ent.Client
// }

// func (this *ActionEntRepository) Create(ctx crud.Context, action *domain.Action) (*domain.Action, error) {
// 	creation := this.client.Action.Create().
// 		SetID(*action.GetId()).
// 		SetEtag(*action.GetEtag()).
// 		SetName(*action.GetName()).
// 		SetResourceID(*action.GetResourceId()).
// 		SetNillableDescription(action.GetDescription()).
// 		SetCreatedBy(string(*action.GetCreatedBy())).
// 		SetCreatedAt(time.Now())

// 	return database.Mutate(ctx, creation, ent.IsNotFound, entToAction)
// }

// func (this *ActionEntRepository) FindById(ctx crud.Context, param it.FindByIdParam) (*domain.Action, error) {
// 	query := this.client.Action.Query().
// 		Where(entAction.IDEQ(param.Id)).
// 		WithResource().
// 		WithEntitlements()

// 	return database.FindOne(ctx, query, ent.IsNotFound, entToAction)
// }

// func (this *ActionEntRepository) FindByName(ctx crud.Context, param it.FindByNameParam) (*domain.Action, error) {
// 	query := this.client.Action.Query().
// 		Where(entAction.NameEQ(param.Name)).
// 		Where(entAction.ResourceIDEQ(param.ResourceId))

// 	return database.FindOne(ctx, query, ent.IsNotFound, entToAction)
// }

// func (this *ActionEntRepository) Update(ctx crud.Context, action *domain.Action, prevEtag model.Etag) (*domain.Action, error) {
// 	update := this.client.Action.UpdateOneID(*action.GetId()).
// 		SetNillableDescription(action.GetDescription()).
// 		Where(entAction.EtagEQ(prevEtag))

// 	if len(update.Mutation().Fields()) > 0 {
// 		update.
// 			SetEtag(*action.GetEtag())
// 	}

// 	return database.Mutate(ctx, update, ent.IsNotFound, entToAction)
// }

// func (this *ActionEntRepository) DeleteHard(ctx crud.Context, param it.DeleteParam) (int, error) {
// 	return this.client.Action.Delete().
// 		Where(entAction.IDEQ(param.Id)).
// 		Exec(ctx)
// }

// func (this *ActionEntRepository) ListEntitlementNamesForActionId(ctx crud.Context, actionId model.Id) ([]string, error) {
// 	rows, err := this.client.Entitlement.Query().
// 		Where(entEntitlement.ActionIDEQ(string(actionId))).
// 		All(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	names := make([]string, 0, len(rows))
// 	for _, row := range rows {
// 		names = append(names, row.Name)
// 	}
// 	return names, nil
// }

// func (this *ActionEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors) {
// 	return database.ParseSearchGraphStr[ent.Action, domain.Action](criteria, entAction.Label)
// }

// func (this *ActionEntRepository) Search(
// 	ctx crud.Context,
// 	param it.SearchParam,
// ) (*crud.PagedResult[domain.Action], error) {
// 	query := this.client.Action.Query().
// 		WithResource()

// 	return database.Search(
// 		ctx,
// 		param.Predicate,
// 		param.Order,
// 		crud.PagingOptions{
// 			Page: param.Page,
// 			Size: param.Size,
// 		},
// 		query,
// 		entToActions,
// 	)
// }

// func BuildActionDescriptor() *orm.EntityDescriptor {
// 	entity := ent.Action{}
// 	builder := orm.DescribeEntity(entAction.Label).
// 		Aliases("actions").
// 		Field(entAction.FieldID, entity.ID).
// 		Field(entAction.FieldName, entity.Name).
// 		Field(entAction.FieldDescription, entity.Description).
// 		Field(entAction.FieldEtag, entity.Etag).
// 		Field(entAction.FieldResourceID, entity.ResourceID).
// 		Field(entAction.FieldCreatedAt, entity.CreatedAt)

// 	return builder.Descriptor()
// }
