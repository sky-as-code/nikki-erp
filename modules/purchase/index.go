package purchase

import (
	"errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/purchase/app"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain"
	repo "github.com/sky-as-code/nikki-erp/modules/purchase/infra/repository"
	"github.com/sky-as-code/nikki-erp/modules/purchase/transport"
)

var ModuleSingleton modules.InCodeModule = &PurchaseModule{}

type PurchaseModule struct {
}

func (*PurchaseModule) LabelKey() string {
	return "purchase.moduleLabel"
}

func (*PurchaseModule) Name() string {
	return "purchase"
}

func (*PurchaseModule) Deps() []string {
	return []string{"essential", "inventory"}
}

func (*PurchaseModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

func (*PurchaseModule) Init() error {
	return errors.Join(
		repo.InitRepositories(),
		app.InitServices(),
		transport.InitTransport(),
	)
}

func (*PurchaseModule) RegisterModels() error {
	return errors.Join(
		dmodel.RegisterSchemaB(domain.PurchaseOrderSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.PurchaseOrderItemSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.PurchaseRequestSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.RequestForProposalSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.RequestForQuoteSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.VendorSchemaBuilder()),
	)
}

func Init() error {
	return errors.Join()
}
