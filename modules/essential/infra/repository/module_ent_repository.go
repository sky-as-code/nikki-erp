package repository

import (
	"math"

	"github.com/sky-as-code/nikki-erp/common/array"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	db "github.com/sky-as-code/nikki-erp/modules/core/database"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
	"github.com/sky-as-code/nikki-erp/modules/essential/infra/ent"
	entMod "github.com/sky-as-code/nikki-erp/modules/essential/infra/ent/module"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/module"
)

const lockKey = math.MaxInt64

func NewModuleEntRepository(client *ent.Client) it.ModuleRepository {
	return &ModuleEntRepository{
		client: client,
	}
}

type ModuleEntRepository struct {
	client *ent.Client
}

func (this *ModuleEntRepository) AcquireLock(ctx crud.Context) (bool, error) {
	var acquired bool
	err := this.client.DB().QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", lockKey).Scan(&acquired)
	if err != nil {
		return false, err
	}
	return acquired, nil
}

func (this *ModuleEntRepository) ReleaseLock(ctx crud.Context) error {
	_, err := this.client.DB().ExecContext(ctx, "SELECT pg_advisory_unlock($1)", lockKey)
	return err
}

func (this *ModuleEntRepository) IncludeTransaction(ctx crud.Context) (crud.Context, error) {
	if ctx.GetDbTranx() != nil {
		return ctx, nil
	}
	newCtx := crud.CloneRequestContext(ctx)
	tx, err := this.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	newCtx.SetDbTranx(tx)
	return newCtx, nil
}

func (this *ModuleEntRepository) moduleClient(ctx crud.Context) *ent.ModuleClient {
	tx, isOk := ctx.GetDbTranx().(*ent.Tx)
	if isOk {
		return tx.Module
	}
	return this.client.Module
}

func (this *ModuleEntRepository) Create(ctx crud.Context, module *domain.ModuleMetadata) (*domain.ModuleMetadata, error) {
	creation := this.entCreation(ctx, module).
		SetID(*module.Id).
		SetLabel(*module.Label).
		SetName(*module.Name).
		SetIsOrphaned(*module.IsOrphaned).
		SetVersion(module.Version.String())

	return db.Mutate(ctx, creation, ent.IsNotFound, entToModule)
}

func (this *ModuleEntRepository) CreateBulk(ctx crud.Context, modules []*domain.ModuleMetadata) ([]*domain.ModuleMetadata, error) {
	creations := array.Map(modules, func(module *domain.ModuleMetadata) *ent.ModuleCreate {
		return this.entCreation(ctx, module)
	})
	creation := this.moduleClient(ctx).CreateBulk(creations...)

	return db.MutateBulk(ctx, creation, ent.IsNotFound, entToModules)
}

func (this *ModuleEntRepository) entCreation(ctx crud.Context, module *domain.ModuleMetadata) *ent.ModuleCreate {
	return this.moduleClient(ctx).Create().
		SetID(*module.Id).
		SetLabel(*module.Label).
		SetName(*module.Name).
		SetIsOrphaned(*module.IsOrphaned).
		SetVersion(module.Version.String())
}

func (this *ModuleEntRepository) Update(ctx crud.Context, module *domain.ModuleMetadata, prevEtag model.Etag) (*domain.ModuleMetadata, error) {
	update := this.moduleClient(ctx).UpdateOneID(*module.Id).
		SetLabel(*module.Label).
		SetNillableName(module.Name).
		SetNillableIsOrphaned(module.IsOrphaned)

	if module.Version != nil {
		update.SetVersion(module.Version.String())
	}

	return db.Mutate(ctx, update, ent.IsNotFound, entToModule)
}

func (this *ModuleEntRepository) DeleteById(ctx crud.Context, param it.DeleteByIdParam) (int, error) {
	return this.moduleClient(ctx).Delete().
		Where(entMod.ID(param.Id)).
		Exec(ctx)
}

func (this *ModuleEntRepository) Exists(ctx crud.Context, param it.ExistsParam) (bool, error) {
	return this.moduleClient(ctx).Query().
		Where(entMod.ID(param.Id)).
		Exist(ctx)
}

func (this *ModuleEntRepository) ExistsByName(ctx crud.Context, param it.ExistsByNameParam) (bool, error) {
	return this.moduleClient(ctx).Query().
		Where(entMod.Name(param.Name)).
		Exist(ctx)
}

func (this *ModuleEntRepository) FindById(ctx crud.Context, param it.FindByIdParam) (*domain.ModuleMetadata, error) {
	query := this.moduleClient(ctx).Query().
		Where(entMod.ID(param.Id))

	return db.FindOne(ctx, query, ent.IsNotFound, entToModule)
}

func (this *ModuleEntRepository) FindByName(ctx crud.Context, param it.FindByNameParam) (*domain.ModuleMetadata, error) {
	query := this.moduleClient(ctx).Query().
		Where(entMod.Name(param.Name))

	return db.FindOne(ctx, query, ent.IsNotFound, entToModule)
}

func (this *ModuleEntRepository) List(ctx crud.Context, param it.ListParam) ([]domain.ModuleMetadata, error) {
	query := this.moduleClient(ctx).Query()

	result, err := db.Search(
		ctx,
		nil,
		nil,
		crud.PagingOptions{},
		query,
		entToModules,
	)
	if err != nil {
		return nil, err
	}
	items := array.Map(result.Items, func(item *domain.ModuleMetadata) domain.ModuleMetadata {
		return *item
	})
	return items, nil
}

func (this *ModuleEntRepository) ParseSearchGraph(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors) {
	return db.ParseSearchGraphStr[ent.Module, domain.ModuleMetadata](criteria, entMod.Label)
}

func (this *ModuleEntRepository) Search(
	ctx crud.Context,
	param it.SearchParam,
) (*crud.PagedResult[domain.ModuleMetadata], error) {
	query := this.moduleClient(ctx).Query()

	result, err := db.Search(
		ctx,
		param.Predicate,
		param.Order,
		crud.PagingOptions{
			Page: param.Page,
			Size: param.Size,
		},
		query,
		entToModules,
	)
	if err != nil {
		return nil, err
	}
	items := array.Map(result.Items, func(item *domain.ModuleMetadata) domain.ModuleMetadata {
		return *item
	})
	return &crud.PagedResult[domain.ModuleMetadata]{
		Items: items,
		Total: result.Total,
		Page:  result.Page,
		Size:  result.Size,
	}, nil
}

func BuildModuleDescriptor() *orm.EntityDescriptor {
	return GetModuleDescriptorBuilder(entMod.Label).
		Aliases("modules").
		Descriptor()
}

func GetModuleDescriptorBuilder(entityName string) *orm.EntityDescriptorBuilder {
	entity := ent.Module{}
	builder := orm.DescribeEntity(entityName).
		Field(entMod.FieldID, entity.ID).
		Field(entMod.FieldLabel, entity.Label).
		Field(entMod.FieldName, entity.Name).
		Field(entMod.FieldIsOrphaned, entity.IsOrphaned).
		Field(entMod.FieldVersion, entity.Version)
	return builder
}
