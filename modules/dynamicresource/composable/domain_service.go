package composable

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	corecrud "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/crud"
)

// CrudDomainService holds the business processing of a resource. It performs validation and
// orchestration, and invokes the repository when it needs to touch the database. It performs
// no permission check: that belongs to the application service, which sits above it.
//
// A module extends it by embedding the default implementation into its own
// {Resource}DomainServiceImpl. An override of Create/Update/Delete adds its rules by passing
// options to the embedded default, so the rules run inside the crud helper and fire on a direct
// service call as well as on a REST call.
type CrudDomainService interface {
	Create(ctx corectx.Context, cmd CreateCommand, options ...CreateOptions) (*CreateResult, error)
	Update(ctx corectx.Context, cmd UpdateCommand, options ...UpdateOptions) (*MutateResult, error)
	Delete(ctx corectx.Context, cmd DeleteCommand, options ...DeleteOptions) (*MutateResult, error)
	SetArchived(ctx corectx.Context, cmd SetArchivedCommand) (*MutateResult, error)

	// GetById fetches one record by primary key. The query carries "id" and optional "fields".
	GetById(ctx corectx.Context, query GetByIdQuery) (*GetOneResult, error)

	// GetOne fetches one record by any unique key carried in the query. It has no REST route.
	GetOne(ctx corectx.Context, query GetOneQuery) (*GetOneResult, error)

	Search(ctx corectx.Context, query SearchQuery, options ...corecrud.ServiceSearchOptions) (*SearchResult, error)
	Exists(ctx corectx.Context, query ExistsQuery) (*ExistsResult, error)

	// ComputeField evaluates one "function"-kind computed field against the unsaved model the
	// command carries under "model", with optional "args".
	ComputeField(ctx corectx.Context, cmd ComputeFieldCommand) (*ComputeFieldResult, error)

	Schema() *dmodel.ModelSchema
	Repository() CrudRepository
}

// NewDomainServiceParam carries what a default domain service needs to work.
type NewDomainServiceParam struct {
	Schema     *dmodel.ModelSchema
	Repository CrudRepository

	// DefaultFields is returned by a search that specifies neither fields nor a resolvable view.
	// When empty, every column of the schema is returned.
	DefaultFields []string

	// CrudActions is the allow-list. Empty means every action is supported.
	CrudActions []CrudAction

	// ComputedFunction resolves the Go implementation of a "function"-kind computed field. Nil
	// when the resource registers none.
	ComputedFunction func(name string) (ComputedFieldFn, bool)
}

func NewDefaultDomainService(param NewDomainServiceParam) CrudDomainService {
	defaultFields := param.DefaultFields
	if len(defaultFields) == 0 {
		defaultFields = columnNames(param.Schema)
	}
	crudActions := make(map[CrudAction]bool, len(param.CrudActions))
	for _, action := range param.CrudActions {
		crudActions[action] = true
	}
	return &DefaultDomainServiceImpl{
		schema:           param.Schema,
		repository:       param.Repository,
		defaultFields:    defaultFields,
		crudActions:      crudActions,
		computedFunction: param.ComputedFunction,
	}
}

// DefaultDomainServiceImpl is the schema-agnostic domain service: what a feature module writes
// by hand in {entity}_domservice.go, minus anything resource-specific.
type DefaultDomainServiceImpl struct {
	schema        *dmodel.ModelSchema
	repository    CrudRepository
	defaultFields []string

	// crudActions is the allow-list, empty meaning "all supported".
	crudActions      map[CrudAction]bool
	computedFunction func(name string) (ComputedFieldFn, bool)
}

func (this *DefaultDomainServiceImpl) Schema() *dmodel.ModelSchema {
	return this.schema
}

func (this *DefaultDomainServiceImpl) Repository() CrudRepository {
	return this.repository
}

func (this *DefaultDomainServiceImpl) Create(
	ctx corectx.Context, cmd CreateCommand, options ...CreateOptions,
) (*CreateResult, error) {
	if cErrs := this.assertActionSupported(CrudActionCreate); cErrs != nil {
		return &CreateResult{ClientErrors: *cErrs}, nil
	}
	opts := firstOrZero(options)

	// corecrud's create hook takes no foundModel: a create has no stored record to compare
	// against. The variable stays nil when there is no hook, because corecrud skips the step on
	// nil and an always-non-nil closure would turn a no-op into a SetFieldData round trip.
	var validateExtra corecrud.CreateValidateExtraFn[*DynamicEntity]
	if opts.ValidateExtra != nil {
		hook := opts.ValidateExtra
		validateExtra = func(ctx corectx.Context, inputModel *DynamicEntity, vErrs *ft.ClientErrors) error {
			return hook(ctx, inputModel, nil, vErrs)
		}
	}

	result, err := corecrud.Create[DynamicEntity](ctx, corecrud.CreateParam[DynamicEntity, *DynamicEntity]{
		Action:                 this.actionName("create"),
		BaseRepoGetter:         this.repository,
		Data:                   NewDynamicEntityFrom(cmd),
		BeforeValidation:       opts.BeforeValidation,
		AfterValidationSuccess: opts.AfterValidationSuccess,
		ValidateExtra:          validateExtra,
	})
	if err != nil {
		return nil, errors.Wrap(err, "CrudDomainService.Create")
	}
	return unwrapEntityResult(result), nil
}

func (this *DefaultDomainServiceImpl) Update(
	ctx corectx.Context, cmd UpdateCommand, options ...UpdateOptions,
) (*MutateResult, error) {
	if cErrs := this.assertActionSupported(CrudActionUpdate); cErrs != nil {
		return &MutateResult{ClientErrors: *cErrs}, nil
	}
	opts := firstOrZero(options)

	// No shim on this side: UpdateParam.ValidateExtra *is* ValidateExtraFn, and corecrud fetches
	// the stored record itself to fill foundModel.
	result, err := corecrud.Update[DynamicEntity](ctx, corecrud.UpdateParam[DynamicEntity, *DynamicEntity]{
		Action:                 this.actionName("update"),
		DbRepoGetter:           this.repository,
		Data:                   NewDynamicEntityFrom(cmd),
		BeforeValidation:       opts.BeforeValidation,
		AfterValidationSuccess: opts.AfterValidationSuccess,
		ValidateExtra:          opts.ValidateExtra,
	})
	return result, errors.Wrap(err, "CrudDomainService.Update")
}

func (this *DefaultDomainServiceImpl) Delete(
	ctx corectx.Context, cmd DeleteCommand, options ...DeleteOptions,
) (*MutateResult, error) {
	if cErrs := this.assertActionSupported(CrudActionDelete); cErrs != nil {
		return &MutateResult{ClientErrors: *cErrs}, nil
	}
	opts := firstOrZero(options)

	// corecrud's delete hook sees only the key fields - it never reads the row it is about to
	// remove - but a delete guard's whole job is to inspect the stored record ("only a cancelled
	// order may be deleted"). So the service fetches it here and hands it over as foundModel.
	//
	// A guard that cannot see the record would fail open or, worse, fail closed on every delete,
	// so a fetch that errors aborts rather than validating against a blank.
	var validateExtra corecrud.DeleteValidateExtraFn
	if opts.ValidateExtra != nil {
		hook := opts.ValidateExtra
		validateExtra = func(ctx corectx.Context, keyFields dmodel.DynamicFields, vErrs *ft.ClientErrors) error {
			foundModel, err := this.findByKeys(ctx, keyFields, vErrs)
			if err != nil {
				return errors.Wrap(err, "Delete.ValidateExtra")
			}
			if vErrs.Count() > 0 {
				return nil
			}
			return hook(ctx, NewDynamicEntityFrom(keyFields), foundModel, vErrs)
		}
	}

	result, err := corecrud.DeleteOne(ctx, corecrud.DeleteOneParam{
		Action:        this.actionName("delete"),
		DbRepoGetter:  this.repository,
		Cmd:           paramsToDeleteCommand(cmd),
		ValidateExtra: validateExtra,
	})
	return result, errors.Wrap(err, "CrudDomainService.Delete")
}

func (this *DefaultDomainServiceImpl) SetArchived(
	ctx corectx.Context, cmd SetArchivedCommand,
) (*MutateResult, error) {
	if cErrs := this.assertActionSupported(CrudActionSetArchived); cErrs != nil {
		return &MutateResult{ClientErrors: *cErrs}, nil
	}
	// corecrud.SetIsArchived takes its arguments positionally and accepts no hooks. A derived
	// service that needs a rule here overrides SetArchived itself and checks before delegating.
	result, err := corecrud.SetIsArchived(ctx, this.repository, paramsToSetArchivedCommand(cmd))
	return result, errors.Wrap(err, "CrudDomainService.SetArchived")
}

func (this *DefaultDomainServiceImpl) actionName(verb string) string {
	return verb + " " + this.schema.Name()
}
