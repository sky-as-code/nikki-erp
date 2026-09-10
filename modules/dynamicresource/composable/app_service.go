package composable

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
)

// CrudApplicationService is the layer that authorizes. Every built-in method resolves the org
// scope, asserts the permission and only then delegates to the domain service, which is what a
// hand-written {entity}_appservice.go does.
//
// A module extends it by embedding the default implementation into its own
// {Resource}ApplicationServiceImpl. A custom action is a method on that struct which calls
// AssertAction (and AssertRecordInOrg for a single-row action) before delegating. Whether the
// action also gets a REST route is decided in the transport layer with RestEngine.AddRoute; a
// method without a route is the resource's service-only API.
type CrudApplicationService interface {
	Create(ctx corectx.Context, cmd CreateCommand) (*CreateResult, error)
	Update(ctx corectx.Context, cmd UpdateCommand) (*MutateResult, error)
	Delete(ctx corectx.Context, cmd DeleteCommand) (*MutateResult, error)
	SetArchived(ctx corectx.Context, cmd SetArchivedCommand) (*MutateResult, error)
	GetById(ctx corectx.Context, query GetByIdQuery) (*GetOneResult, error)
	GetOne(ctx corectx.Context, query GetOneQuery) (*GetOneResult, error)
	Search(ctx corectx.Context, query SearchQuery) (*SearchResult, error)
	Exists(ctx corectx.Context, query ExistsQuery) (*ExistsResult, error)
	GetSchema(ctx corectx.Context, query GetSchemaQuery) (*GetSchemaResult, error)
	ComputeField(ctx corectx.Context, cmd ComputeFieldCommand) (*ComputeFieldResult, error)

	DomainService() CrudDomainService
	Schema() *dmodel.ModelSchema

	// ResourceCode is the permission resource code, which is the schema name.
	ResourceCode() string

	// The authorization machinery, reusable by a derived service's custom actions.
	AssertAction(ctx corectx.Context, actionCode string, params dmodel.DynamicFields) (*model.Id, *ft.ClientErrors)
	AssertPermission(ctx corectx.Context, actionCode string, orgId *model.Id) *ft.ClientErrors
	ResolveOrgScope(ctx corectx.Context, params dmodel.DynamicFields) (*model.Id, *ft.ClientErrors)
	ConstrainSearchToOrg(params dmodel.DynamicFields, orgId model.Id) error
	AssertRecordInOrg(ctx corectx.Context, params dmodel.DynamicFields, orgId model.Id) (*ft.ClientErrors, error)
	FetchByKeys(ctx corectx.Context, keys dmodel.DynamicFields, vErrs *ft.ClientErrors) (dmodel.DynamicFields, error)
}

// NewAppServiceParam carries what a default application service needs to work.
type NewAppServiceParam struct {
	DomainService CrudDomainService

	// PermissionScope applies to every permission check. Nil means requestguard.ResourceScopeOrg.
	PermissionScope *requestguard.ResourceScope

	// IsOrgScoped confines every action to one organization: the caller must supply org_id, the
	// value must name an org the caller belongs to, and the action only ever sees records of
	// that org. Nil means true, because the unsafe direction is the silent one. It has no effect
	// on a resource whose schema declares no org_id field.
	IsOrgScoped *bool

	// CrudActions is the allow-list. Empty means every action is supported.
	CrudActions []CrudAction
}

func NewDefaultApplicationService(param NewAppServiceParam) CrudApplicationService {
	scope := requestguard.ResourceScopeOrg
	if param.PermissionScope != nil {
		scope = *param.PermissionScope
	}
	crudActions := make(map[CrudAction]bool, len(param.CrudActions))
	for _, action := range param.CrudActions {
		crudActions[action] = true
	}
	return &DefaultApplicationServiceImpl{
		domSvc:      param.DomainService,
		scope:       scope,
		orgScoped:   param.IsOrgScoped == nil || *param.IsOrgScoped,
		crudActions: crudActions,
	}
}

type DefaultApplicationServiceImpl struct {
	domSvc      CrudDomainService
	scope       requestguard.ResourceScope
	orgScoped   bool
	crudActions map[CrudAction]bool
}

func (this *DefaultApplicationServiceImpl) DomainService() CrudDomainService {
	return this.domSvc
}

func (this *DefaultApplicationServiceImpl) Schema() *dmodel.ModelSchema {
	return this.domSvc.Schema()
}

func (this *DefaultApplicationServiceImpl) ResourceCode() string {
	return this.domSvc.Schema().Name()
}

// Create defers a missing org_id to schema validation rather than reporting it here.
//
// org_id is a declared required-for-create field on an org-scoped resource, so a create that
// omits it is a missing required field like any other, and must answer the same
// err_missing_required_field the rest of the payload would. Reporting err_org_id_required first
// would make one required field explain itself differently from its neighbours. An org_id that
// IS present still resolves here, so a caller naming an org they do not belong to is refused
// before the record is written.
func (this *DefaultApplicationServiceImpl) Create(
	ctx corectx.Context, cmd CreateCommand,
) (*CreateResult, error) {
	if cErrs := assertActionSupported(this.crudActions, this.ResourceCode(), CrudActionCreate); cErrs != nil {
		return &CreateResult{ClientErrors: *cErrs}, nil
	}
	if this.orgIdIsSchemaRequired() && readString(cmd, basemodel.FieldOrgId) == "" {
		if cErrs := this.AssertPermission(ctx, PermissionCreate, nil); cErrs != nil {
			return &CreateResult{ClientErrors: *cErrs}, nil
		}
		return this.domSvc.Create(ctx, cmd)
	}
	if _, cErrs := this.AssertAction(ctx, PermissionCreate, cmd); cErrs != nil {
		return &CreateResult{ClientErrors: *cErrs}, nil
	}
	return this.domSvc.Create(ctx, cmd)
}

// orgIdIsSchemaRequired reports whether the schema itself demands org_id on create, which is
// what makes deferring to schema validation the correct report rather than a silent skip.
func (this *DefaultApplicationServiceImpl) orgIdIsSchemaRequired() bool {
	schema := this.Schema()
	if schema == nil {
		return false
	}
	field, exists := schema.Field(basemodel.FieldOrgId)
	return exists && field.IsRequiredForCreate()
}

func (this *DefaultApplicationServiceImpl) Update(
	ctx corectx.Context, cmd UpdateCommand,
) (*MutateResult, error) {
	if cErrs, err := this.guardRecord(ctx, CrudActionUpdate, PermissionUpdate, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return this.domSvc.Update(ctx, cmd)
}

func (this *DefaultApplicationServiceImpl) Delete(
	ctx corectx.Context, cmd DeleteCommand,
) (*MutateResult, error) {
	if cErrs, err := this.guardRecord(ctx, CrudActionDelete, PermissionDelete, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return this.domSvc.Delete(ctx, cmd)
}

func (this *DefaultApplicationServiceImpl) SetArchived(
	ctx corectx.Context, cmd SetArchivedCommand,
) (*MutateResult, error) {
	if cErrs, err := this.guardRecord(ctx, CrudActionSetArchived, PermissionSetArchived, cmd); err != nil || cErrs != nil {
		return mutateFailure(cErrs, err)
	}
	return this.domSvc.SetArchived(ctx, cmd)
}

func (this *DefaultApplicationServiceImpl) GetById(
	ctx corectx.Context, query GetByIdQuery,
) (*GetOneResult, error) {
	cErrs, err := this.guardRecord(ctx, CrudActionGetById, PermissionRead, query)
	if err != nil {
		return nil, err
	}
	if cErrs != nil {
		return &GetOneResult{ClientErrors: *cErrs}, nil
	}
	return this.domSvc.GetById(ctx, query)
}

// GetOne is org-confined through the normalized org_id itself: the domain service turns every
// column-named param, org_id included, into an equality condition.
func (this *DefaultApplicationServiceImpl) GetOne(
	ctx corectx.Context, query GetOneQuery,
) (*GetOneResult, error) {
	if cErrs := this.guard(ctx, CrudActionGetByUnique, PermissionRead, query); cErrs != nil {
		return &GetOneResult{ClientErrors: *cErrs}, nil
	}
	return this.domSvc.GetOne(ctx, query)
}

func (this *DefaultApplicationServiceImpl) Search(
	ctx corectx.Context, query SearchQuery,
) (*SearchResult, error) {
	if cErrs := assertActionSupported(this.crudActions, this.ResourceCode(), CrudActionSearch); cErrs != nil {
		return &SearchResult{ClientErrors: *cErrs}, nil
	}
	orgId, cErrs := this.AssertAction(ctx, PermissionRead, query)
	if cErrs != nil {
		return &SearchResult{ClientErrors: *cErrs}, nil
	}
	if orgId != nil {
		if err := this.ConstrainSearchToOrg(query, *orgId); err != nil {
			return nil, err
		}
	}
	return this.domSvc.Search(ctx, query)
}

func (this *DefaultApplicationServiceImpl) Exists(
	ctx corectx.Context, query ExistsQuery,
) (*ExistsResult, error) {
	if cErrs := this.guard(ctx, CrudActionExists, PermissionRead, query); cErrs != nil {
		return &ExistsResult{ClientErrors: *cErrs}, nil
	}
	return this.domSvc.Exists(ctx, query)
}

// GetSchema serves the resource schema in the simplified shape the clients cache. Schema
// metadata belongs to no org, so only the permission is checked.
func (this *DefaultApplicationServiceImpl) GetSchema(
	ctx corectx.Context, _ GetSchemaQuery,
) (*GetSchemaResult, error) {
	if cErrs := assertActionSupported(this.crudActions, this.ResourceCode(), CrudActionGetSchema); cErrs != nil {
		return &GetSchemaResult{ClientErrors: *cErrs}, nil
	}
	if cErrs := this.AssertPermission(ctx, PermissionRead, nil); cErrs != nil {
		return &GetSchemaResult{ClientErrors: *cErrs}, nil
	}
	return &GetSchemaResult{Data: this.Schema().ToSimplized(), HasData: true}, nil
}

// ComputeField evaluates a function-kind computed field over an unsaved model. The model travels
// in the command and there is no stored row to confine to an org, so only the permission is
// checked: computing a derived value discloses no more than reading the record would.
func (this *DefaultApplicationServiceImpl) ComputeField(
	ctx corectx.Context, cmd ComputeFieldCommand,
) (*ComputeFieldResult, error) {
	if cErrs := assertActionSupported(this.crudActions, this.ResourceCode(), CrudActionComputeField); cErrs != nil {
		return &ComputeFieldResult{ClientErrors: *cErrs}, nil
	}
	if cErrs := this.AssertPermission(ctx, PermissionRead, nil); cErrs != nil {
		return &ComputeFieldResult{ClientErrors: *cErrs}, nil
	}
	return this.domSvc.ComputeField(ctx, cmd)
}

// guard is the allow-list check followed by AssertAction, for actions that touch no single row.
func (this *DefaultApplicationServiceImpl) guard(
	ctx corectx.Context, action CrudAction, permission string, params dmodel.DynamicFields,
) *ft.ClientErrors {
	if cErrs := assertActionSupported(this.crudActions, this.ResourceCode(), action); cErrs != nil {
		return cErrs
	}
	_, cErrs := this.AssertAction(ctx, permission, params)
	return cErrs
}

// guardRecord is guard plus AssertRecordInOrg, for the actions that identify exactly one existing
// row by id: the org has to be checked before the crud command runs, because the command itself
// carries only the id.
func (this *DefaultApplicationServiceImpl) guardRecord(
	ctx corectx.Context, action CrudAction, permission string, params dmodel.DynamicFields,
) (*ft.ClientErrors, error) {
	if cErrs := assertActionSupported(this.crudActions, this.ResourceCode(), action); cErrs != nil {
		return cErrs, nil
	}
	orgId, cErrs := this.AssertAction(ctx, permission, params)
	if cErrs != nil {
		return cErrs, nil
	}
	if orgId == nil {
		return nil, nil
	}
	return this.AssertRecordInOrg(ctx, params, *orgId)
}

func mutateFailure(cErrs *ft.ClientErrors, err error) (*MutateResult, error) {
	if err != nil {
		return nil, err
	}
	return &MutateResult{ClientErrors: *cErrs}, nil
}
