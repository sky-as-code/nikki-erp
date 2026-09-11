package dynamicengines

import (
	stdErr "errors"

	"go.bryk.io/pkg/errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/dynamicresource/composable"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/services"
)

// The association schemas get a repository and nothing else: no onion, no application service, no
// route, no IAM resource row. A _rel row is configured through its owner's capabilities, and
// exposing it as a CRUD resource would let a client rewrite a channel's payment mapping without
// the validation that mapping requires. transport_surface_test.go asserts a declared junction has
// no engine, so registering one here would fail the build's own surface check.
//
// Under the legacy engine they needed an engine anyway, because the query builder and database
// client were private to the registry and a repository could not be built without one. The
// composable BuildParam hands those over directly, so the workaround is gone.

// InitJunctionRepositories registers the repositories of the association schemas.
func InitJunctionRepositories() error {
	return stdErr.Join(
		registerChannelPaymentRelRepository(),
		registerChannelFulfillmentMethodRepository(),
	)
}

func registerChannelPaymentRelRepository() error {
	err := deps.Register(func(param composable.BuildParam) (*services.ChannelPaymentDomainServiceImpl, error) {
		repo, err := newJunctionRepository(param, models.SalesChannelPaymentRelSchemaName)
		if err != nil {
			return nil, err
		}
		return services.NewChannelPaymentDomainService(repo), nil
	})
	// Forced like the onions are: the mapping service is reached through this constructor alone, and
	// a channel's payment routes would otherwise resolve it on the first request rather than at boot.
	return stdErr.Join(err, deps.Invoke(func(_ *services.ChannelPaymentDomainServiceImpl) {}))
}

// The channel-to-method mapping is read through the resource hub by the fulfillment method
// service, which asks whether a channel permits a method. It needs no service of its own.
func registerChannelFulfillmentMethodRepository() error {
	return deps.Invoke(func(param composable.BuildParam) error {
		repo, err := newJunctionRepository(param, models.SalesChannelFulfillmentMethodSchemaName)
		if err != nil {
			return err
		}
		// Installed with a nil domain service: the hub's repoFor is the only way this schema is
		// reached, and domainServiceFor on it would be a wiring mistake worth a nil panic.
		services.InstallResource(models.SalesChannelFulfillmentMethodSchemaName, repo, nil)
		return nil
	})
}

func newJunctionRepository(
	param composable.BuildParam, schemaName string,
) (composable.CrudRepository, error) {
	schema := dmodel.GetSchema(schemaName)
	if schema == nil {
		return nil, errors.New("the '" + schemaName + "' schema is not registered")
	}
	return composable.NewDefaultCrudRepository(composable.NewRepositoryParam{
		Client:        param.Client,
		ConfigSvc:     param.ConfigSvc,
		QueryBuilder:  param.QueryBuilder,
		Logger:        param.Logger,
		NewBaseRepoFn: param.NewBaseRepoFn,
		Schema:        schema,
	}), nil
}
