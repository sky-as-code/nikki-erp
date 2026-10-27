package composable

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
)

// The private scope in the default application service: `{action}:{resource}[.{field}]:private`
// answers for a record whose created_by is the caller. It is a fallback, consulted only when the
// resource's own scope refused, so a caller with an org or tenant grant never pays the extra read.

// recordAccess is what a built-in was authorized for: the org it is confined to, and whether it
// was let through as the owner of the record(s), which is what makes a private grant answer.
type recordAccess struct {
	orgId *model.Id
	owned bool
}

// ownerFallback answers whether a refused single-row action may proceed through a private grant:
// the caller holds one for permission, and the addressed record was created by them. The grant is
// checked first, in memory, so the record is read only for a caller who could pass.
func (this *DefaultApplicationServiceImpl) ownerFallback(
	ctx corectx.Context, permission string, params dmodel.DynamicFields, orgId *model.Id,
) (bool, error) {
	if !this.supportsPrivateScope(ctx) || this.gate(ctx, permission, recordAccess{orgId: orgId, owned: true}) != nil {
		return false, nil
	}
	return this.isCreatedByCaller(ctx, params, orgId)
}

// supportsPrivateScope reports whether ownership can be decided at all: the schema records who
// created a row, and the caller is a user with an id to compare it with.
func (this *DefaultApplicationServiceImpl) supportsPrivateScope(ctx corectx.Context) bool {
	schema := this.Schema()
	if schema == nil || ctx.GetPermissions().UserId == "" {
		return false
	}
	_, exists := schema.Field(basemodel.FieldCreatedBy)
	return exists
}

func (this *DefaultApplicationServiceImpl) isCreatedByCaller(
	ctx corectx.Context, params dmodel.DynamicFields, orgId *model.Id,
) (bool, error) {
	recordId := readString(params, basemodel.FieldId)
	if recordId == "" {
		return false, nil
	}
	keys := dmodel.DynamicFields{basemodel.FieldId: recordId}
	if orgId != nil {
		keys[basemodel.FieldOrgId] = string(*orgId)
	}
	vErrs := ft.ClientErrors{}
	if !this.assertKeysAreWellFormed(keys, &vErrs) {
		return false, nil
	}
	found, err := this.domSvc.Repository().FindByKeys(ctx, keys)
	if err != nil {
		return false, errors.Wrap(err, "isCreatedByCaller")
	}
	if !found.HasData || found.ClientErrors.Count() > 0 {
		return false, nil
	}
	return readString(found.Data, basemodel.FieldCreatedBy) == string(ctx.GetPermissions().UserId), nil
}

// searchAccess gates a search. Refused at the resource's scope, a caller holding a private read
// grant searches their own records only: constrainSearch then ANDs created_by onto the graph.
// A caller who passes at the resource's scope is not widened by their private grants - the rows
// a search returns are not all theirs, so a private grant cannot answer for them.
func (this *DefaultApplicationServiceImpl) searchAccess(
	ctx corectx.Context, params dmodel.DynamicFields,
) (recordAccess, *ft.ClientErrors) {
	orgId, cErrs := this.ResolveOrgScope(ctx, params)
	if cErrs != nil {
		return recordAccess{}, cErrs
	}
	access := recordAccess{orgId: orgId}
	refused := this.gate(ctx, PermissionRead, access)
	if refused == nil {
		return access, nil
	}
	owned := recordAccess{orgId: orgId, owned: true}
	if !this.supportsPrivateScope(ctx) || this.gate(ctx, PermissionRead, owned) != nil {
		return access, refused
	}
	return owned, nil
}

// constrainSearch confines a search to the caller's org and, under a private grant, to the rows
// they created.
func (this *DefaultApplicationServiceImpl) constrainSearch(
	ctx corectx.Context, params dmodel.DynamicFields, access recordAccess,
) error {
	if access.orgId != nil {
		if err := this.ConstrainSearchToOrg(params, *access.orgId); err != nil {
			return err
		}
	}
	if !access.owned {
		return nil
	}
	ownerNode := dmodel.NewSearchNode()
	ownerNode.NewCondition(basemodel.FieldCreatedBy, dmodel.Equals, string(ctx.GetPermissions().UserId))
	graph, err := dmodel.AndDecodedGraph(params[fieldNameGraph], *ownerNode)
	if err != nil {
		return errors.Wrap(err, "constrainSearch")
	}
	params[fieldNameGraph] = graph
	return nil
}
