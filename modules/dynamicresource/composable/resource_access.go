package composable

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/resourceaccess"
)

const paramResourceCode = "resource_code"

// NewResourceAccess implements the core/resourceaccess port from the engine registry: the record
// is read through the target's own application GetById, as the caller, so the answer is exactly
// what opening the record would have been.
func NewResourceAccess() resourceaccess.ResourceAccess {
	return &engineResourceAccess{resolve: resolveApplicationService, schemaOf: dmodel.GetSchema}
}

type engineResourceAccess struct {
	resolve  func(schemaName string) (CrudApplicationService, error)
	schemaOf func(schemaName string) *dmodel.ModelSchema
}

func resolveApplicationService(schemaName string) (CrudApplicationService, error) {
	engines, err := ResolveEngines(schemaName)
	if err != nil {
		return nil, err
	}
	return engines[0].ApplicationService(), nil
}

func (this *engineResourceAccess) ReadAsCaller(
	ctx corectx.Context, query resourceaccess.ReadAsCallerQuery,
) (*resourceaccess.ReadAsCallerResult, ft.ClientErrors, error) {
	schema := this.schemaOf(query.ResourceCode)
	if schema == nil {
		return nil, unknownResource(), nil
	}
	appSvc, err := this.resolve(query.ResourceCode)
	if err != nil {
		// A schema with no engine (a legacy resource, or a junction table) has no read to run as
		// the caller, so it is answered like an unknown resource rather than a server fault.
		return nil, unknownResource(), nil
	}
	found, err := appSvc.GetById(ctx, readAsCallerParams(schema, query))
	if err != nil {
		return nil, nil, errors.Wrap(err, "ReadAsCaller.GetById")
	}
	if isNotFound(found) {
		return this.readGone(ctx, appSvc, query)
	}
	if found.ClientErrors.Count() > 0 {
		return nil, found.ClientErrors, nil
	}
	return &resourceaccess.ReadAsCallerResult{Exists: true, MaskedFields: found.Data.Meta.MaskedFields}, nil, nil
}

// readGone answers for a record that no longer exists: the caller must hold read on the resource
// in the org they named, the only scope a deleted record can still be checked against.
func (this *engineResourceAccess) readGone(
	ctx corectx.Context, appSvc CrudApplicationService, query resourceaccess.ReadAsCallerQuery,
) (*resourceaccess.ReadAsCallerResult, ft.ClientErrors, error) {
	params := dmodel.DynamicFields{}
	if query.OrgId != nil {
		params[basemodel.FieldOrgId] = string(*query.OrgId)
	}
	if _, cErrs := appSvc.AssertAction(ctx, PermissionRead, params); cErrs != nil {
		return nil, *cErrs, nil
	}
	return &resourceaccess.ReadAsCallerResult{Exists: false}, nil, nil
}

// readAsCallerParams asks for every column, so the masking covers each field an entry may carry.
func readAsCallerParams(schema *dmodel.ModelSchema, query resourceaccess.ReadAsCallerQuery) GetByIdQuery {
	params := GetByIdQuery{
		basemodel.FieldId:     string(query.ResourceId),
		basemodel.FieldFields: columnNames(schema),
	}
	if query.OrgId != nil {
		params[basemodel.FieldOrgId] = string(*query.OrgId)
	}
	return params
}

func isNotFound(result *GetOneResult) bool {
	if result == nil {
		return true
	}
	if result.ClientErrors.Count() == 0 {
		return !result.HasData
	}
	notFoundKey := ft.ErrorKey("err_not_found")
	for _, item := range result.ClientErrors {
		if item.Key != notFoundKey {
			return false
		}
	}
	return true
}

func unknownResource() ft.ClientErrors {
	return ft.ClientErrors{*ft.NewNotFoundError(paramResourceCode)}
}
