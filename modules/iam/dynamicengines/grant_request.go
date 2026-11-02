package dynamicengines

import (
	stdErr "errors"

	"go.uber.org/dig"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/util"
	"github.com/sky-as-code/nikki-erp/modules/core/requestguard"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource"
	"github.com/sky-as-code/nikki-erp/modules/iam/domain/models"
	itRoleReq "github.com/sky-as-code/nikki-erp/modules/iam/interfaces/rolerequest"
)

type grantRequestEngineParam struct {
	dig.In

	Engine dynamicresource.DynamicResourceEngineOnion `name:"dynengine_iam_grant_request"`
}

// registerGrantRequestEngine declares the iam_grant_request onion, tenant-scoped. Grant requests
// are plain CRUD, so every layer is the composable default; a rule gets its own layer when one
// is needed.
func registerGrantRequestEngine() error {
	err := deps.RegisterNamed(
		dynamicresource.EngineDependencyName(models.RoleRequestSchemaName),
		func(param dynamicresource.BuildParam) dynamicresource.DynamicResourceEngineOnion {
			return dynamicresource.MustBuild(&dynamicresource.DynamicResourceEngineOnionImpl{
				SchemaName:      models.RoleRequestSchemaName,
				PermissionScope: util.ToPtr(requestguard.ResourceScopeTenant),
			}, param)
		},
	)
	return stdErr.Join(err, deps.Register(
		func(p grantRequestEngineParam) itRoleReq.RoleRequestCrudRepository {
			return p.Engine.Repository()
		},
		func(p grantRequestEngineParam) itRoleReq.RoleRequestCrudDomainService {
			return p.Engine.DomainService()
		},
		func(p grantRequestEngineParam) itRoleReq.RoleRequestCrudApplicationService {
			return p.Engine.ApplicationService()
		},
	))
}
