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
	entHierarchy "github.com/sky-as-code/nikki-erp/modules/identity/infra/ent/hierarchylevel"
	it "github.com/sky-as-code/nikki-erp/modules/identity/interfaces/hierarchy"
)

func NewHierarchyLevelEntRepository(client *ent.Client) it.HierarchyRepository {
	return &HierarchyLevelEntRepository{
		client: client,
	}
}

type HierarchyLevelEntRepository struct {
	client *ent.Client
}

func (this *HierarchyLevelEntRepository) hierarchyClient(ctx crud.Context) *ent.HierarchyLevelClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.HierarchyLevel
	}
	return this.client.HierarchyLevel
}

func (this *HierarchyLevelEntRepository) Create(ctx crud.Context, hierarchyLevel *domain.HierarchyLevel) (*domain.HierarchyLevel, error) {
	creation := this.hierarchyClient(ctx).Create().
		SetID(*hierarchyLevel.Id).
		SetName(*hierarchyLevel.Name).
		SetOrgID(string(*hierarchyLevel.OrgId)).
		SetNillableParentID(hierarchyLevel.ParentId).
		SetEtag(*hierarchyLevel.Etag)

	return db.Mutate(ctx, creation, ent.IsNotFound, entToHierarchyLevel)
}

func (this *HierarchyLevelEntRepository) Update(ctx crud.Context, hierarchyLevel *domain.HierarchyLevel, prevEtag model.Etag) (*domain.HierarchyLevel, error) {
	update := this.hierarchyClient(ctx).UpdateOneID(*hierarchyLevel.Id).
		SetNillableName(hierarchyLevel.Name).
		SetNillableParentID(hierarchyLevel.ParentId).
		SetEtag(*hierarchyLevel.Etag).
		Where(entHierarchy.EtagEQ(prevEtag))

	if len(update.Mutation().Fields()) > 0 {
		update.
			SetEtag(*hierarchyLevel.Etag).
			SetUpdatedAt(time.Now())
	}

	return db.Mutate(ctx, update, ent.IsNotFound, entToHierarchyLevel)
}

func (this *HierarchyLevelEntRepository) DeleteHard(ctx crud.Context, param it.DeleteParam) (int, error) {
	return this.hierarchyClient(ctx).Delete().
		Where(entHierarchy.ID(param.Id)).
		Exec(ctx)
}

func (this *HierarchyLevelEntRepository) FindById(ctx crud.Context, param it.FindByIdParam) (*domain.HierarchyLevel, error) {
	dbQuery := this.hierarchyClient(ctx).Query().
		Where(entHierarchy.ID(param.Id))
	if param.ScopeRef != nil {
		dbQuery = dbQuery.Where(entHierarchy.OrgIDEQ(string(*param.ScopeRef)))
	}

	if param.WithChildren {
		dbQuery = dbQuery.WithChildren()
	}

	// Add soft delete check if needed
	if !param.IncludeDeleted {
		dbQuery = dbQuery.Where(entHierarchy.DeletedAtIsNil())
	}

	return db.FindOne(ctx, dbQuery, ent.IsNotFound, entToHierarchyLevel)
}

func (this *HierarchyLevelEntRepository) FindByName(ctx crud.Context, param it.FindByNameParam) (*domain.HierarchyLevel, error) {
	return db.FindOne(
		ctx,
		this.hierarchyClient(ctx).Query().Where(entHierarchy.Name(param.Name)),
		ent.IsNotFound,
		entToHierarchyLevel,
	)
}

func (this *HierarchyLevelEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.HierarchyLevel, domain.HierarchyLevel](criteria, entHierarchy.Label)
}

func (this *HierarchyLevelEntRepository) Search(
	ctx crud.Context,
	param it.SearchParam,
) (*crud.PagedResult[domain.HierarchyLevel], error) {
	query := this.hierarchyClient(ctx).Query()

	if param.WithOrg {
		query = query.WithOrg()
	}
	if param.OrgId != nil {
		query = query.Where(entHierarchy.OrgIDEQ(string(*param.OrgId)))
	}

	if param.WithParent {
		query = query.WithParent()
	}

	if param.WithChildren {
		query = query.WithChildren()
	}

	// Add soft delete check
	if !param.IncludeDeleted {
		query = query.Where(entHierarchy.DeletedAtIsNil())
	}

	return db.Search(
		ctx,
		param.Predicate,
		param.Order,
		crud.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		entToHierarchyLevels,
	)
}

func (this *HierarchyLevelEntRepository) AddRemoveUsers(ctx crud.Context, param it.AddRemoveUsersParam) (*ft.ClientError, error) {
	if len(param.Add) == 0 && len(param.Remove) == 0 {
		return nil, nil
	}

	err := this.hierarchyClient(ctx).UpdateOneID(param.HierarchyId).
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

func (this *HierarchyLevelEntRepository) Exists(ctx crud.Context, param it.ExistsParam) (bool, error) {
	return this.hierarchyClient(ctx).Query().
		Where(entHierarchy.ID(param.Id)).
		Exist(ctx)
}

func BuildHierarchyLevelDescriptor() *orm.EntityDescriptor {
	entity := ent.HierarchyLevel{}
	builder := orm.DescribeEntity(entHierarchy.Label).
		Aliases("hierarchy_levels", "hierarchies").
		Field(entHierarchy.FieldCreatedAt, entity.CreatedAt).
		Field(entHierarchy.FieldDeletedAt, entity.DeletedAt).
		Field(entHierarchy.FieldID, entity.ID).
		Field(entHierarchy.FieldName, entity.Name).
		Field(entHierarchy.FieldOrgID, entity.OrgID).
		Field(entHierarchy.FieldParentID, entity.ParentID).
		Field(entHierarchy.FieldUpdatedAt, entity.UpdatedAt).
		Edge(entHierarchy.EdgeUsers, orm.ToEdgePredicate(entHierarchy.HasUsersWith)).
		Edge(entHierarchy.EdgeOrg, orm.ToEdgePredicate(entHierarchy.HasOrgWith)).
		Edge(entHierarchy.EdgeParent, orm.ToEdgePredicate(entHierarchy.HasParentWith)).
		Edge(entHierarchy.EdgeChildren, orm.ToEdgePredicate(entHierarchy.HasChildrenWith))

	return builder.Descriptor()
}
