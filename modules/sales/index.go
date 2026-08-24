// Package sales owns the commercial transaction: the sales order and its lines, pricing and
// promotion, billing, payment allocation, return and refund — per
// docs/requirements/sales/00-business-requirement-md and the channel/sales-point change request in
// docs/requirements/sales/01-sales-channel.md.
//
// It is deliberately channel-agnostic. A vending kiosk, a future POS and a future storefront all
// create the same sales_order through the same pricing engine; nothing in this module branches on
// which one is calling. What differs between them is data — a sales channel row and its sales
// points — not code.
//
// It owns none of Product, UoM, Warehouse, Stock, Payment Method, VAT Invoice or Accounting. It
// holds ids into those modules and reads them through ports, and it sends business intent rather
// than instructions: a fulfilment request rather than a stock movement, a fiscal document request
// rather than a call to a tax provider.
package sales

import (
	stdErr "errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/sales/app"
	modconstants "github.com/sky-as-code/nikki-erp/modules/sales/constants"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/sales/dynamicengines"
	"github.com/sky-as-code/nikki-erp/modules/sales/infra/external"
	"github.com/sky-as-code/nikki-erp/modules/sales/transport/restful"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader.
//
// It is typed DynamicModule rather than InCodeModule so that dropping RegisterModels fails the
// build. Under the wider interface the method is found by a type assertion instead, and a module
// that has lost it still compiles, still loads, and silently registers no schemas at all.
var ModuleSingleton modules.DynamicModule = &SalesModule{}

type SalesModule struct{}

func (*SalesModule) LabelKey() string {
	return "sales.moduleLabel"
}

func (*SalesModule) Name() string {
	return modconstants.SalesModuleName
}

// Deps names every module Sales reads through a port.
//
// dynamicresource hosts the resource engines. inventory supplies the product variant a line sells
// and receives the fulfilment requests an order raises. essential supplies UoM conversion and
// currency. contacts supplies the party a customer_reference points at. paymentinvoice owns
// payment method master data and the invoice capability a fiscal request is delegated to.
//
// essential and core are injected automatically by buildDependencyGraph; essential is named anyway
// because Sales consumes it directly, and a reader should not have to know the implicit rule to
// see that.
func (*SalesModule) Deps() []string {
	return []string{
		"dynamicresource",
		"essential",
		"inventory",
		"contacts",
		"paymentinvoice",
	}
}

func (*SalesModule) IsInternal() bool {
	return false
}

func (*SalesModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements DynamicModule.
//
// The order is load-bearing three times over, and is fixed now rather than when the first resource
// arrives: the external ports bind first, because a derived service resolves its ports when it is
// constructed; the engines are created before the derived services, because a derived service
// wraps the engine's own; and the REST layer is registered last, because it registers the engines'
// routes and so cannot run before they exist.
func (*SalesModule) Init() error {
	if err := external.InitExternal(); err != nil {
		return err
	}
	if err := dynamicengines.InitDynamicEngines(); err != nil {
		return err
	}
	if err := dynamicengines.InitDomainServices(); err != nil {
		return err
	}
	if err := app.InitApplicationServices(); err != nil {
		return err
	}
	return restful.InitRestfulHandlers()
}

// RegisterModels implements DynamicModule.
//
// Registration order is load-bearing: an edge is resolved against the schema registry at
// registration time, so a referenced schema must be registered before the one pointing at it —
// the sales channel before its sales points, the order before its lines.
//
// The schemas are listed here rather than scattered across the packages that own them, so that
// the order is visible in a single place and a missing registration is a gap in a list rather than
// an absence nobody can see.
func (*SalesModule) RegisterModels() error {
	return stdErr.Join(
		dmodel.RegisterSchemaB(models.SalesChannelSchemaBuilder()),
		dmodel.RegisterSchemaB(models.SalesPointSchemaBuilder()),
	)
}
