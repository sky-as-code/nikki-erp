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
	entAssignt "github.com/sky-as-code/nikki-erp/modules/authorize/infra/ent/entitlementassignment"
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

func (this *RoleEntRepository) CreateWithEntitlements(ctx context.Context, role domain.Role, entitlementIds []model.Id) (result *domain.Role, err error) {
	tx, err := this.client.Tx(ctx)
	fault.PanicOnErr(err)

	defer func() {
		if e := fault.RecoverPanicFailedTo(recover(), "failed to create role transaction"); e != nil {
			_ = tx.Rollback()
			err = e
		}
	}()

	// Create role
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
	if err != nil {
		return err
	}

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
	if err != nil {
		return err
	}

	_, err = tx.EntitlementAssignment.
		Create().
		SetID(*assignmentID).
		SetSubjectRef(roleID).
		SetSubjectType(entAssignt.SubjectTypeNikkiRole).
		SetEntitlementID(entitlementID).
		SetResolvedExpr(resolvedExpr).
		SetNillableActionName(actionName).
		SetNillableResourceName(resourceName).
		Save(ctx)

	return err
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

func (this *RoleEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, fault.ValidationErrors) {
	return database.ParseSearchGraphStr[ent.Role, domain.Role](criteria, entRole.Label)
}

func (this *RoleEntRepository) Search(
	ctx context.Context,
	param it.SearchParam,
) (*crud.PagedResult[*domain.Role], error) {
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

func (this *RoleEntRepository) FindAllBySubject(ctx context.Context, param it.FindAllBySubjectParam) ([]*domain.Role, error) {
	query := this.client.Role.Query().
		Where(entRole.HasRoleUsersWith(entRoleUser.ReceiverRefEQ(param.SubjectRef)))

	return database.List(ctx, query, entToRoles)
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
