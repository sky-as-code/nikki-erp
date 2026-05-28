package purchase

import (
	"errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/purchase/app"
	modconstants "github.com/sky-as-code/nikki-erp/modules/purchase/constants"
	models "github.com/sky-as-code/nikki-erp/modules/purchase/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/purchase/domain/services"
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
	return modconstants.PurchaseModuleName
}

func (*PurchaseModule) Deps() []string {
	return []string{
		"essential",
		"inventory",
	}
}

func (*PurchaseModule) IsInternal() bool {
	return false
}

func (*PurchaseModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

func (*PurchaseModule) Init() error {
	return errors.Join(
		repo.InitRepositories(),
		services.InitDomainServices(),
		app.InitApplicationServices(),
		transport.InitTransport(),
	)
}

func (*PurchaseModule) RegisterModels() error {
	return errors.Join(
		dmodel.RegisterSchemaB(models.PurchaseOrderSchemaBuilder()),
		dmodel.RegisterSchemaB(models.PurchaseOrderItemSchemaBuilder()),
		dmodel.RegisterSchemaB(models.PurchaseRequestSchemaBuilder()),
		dmodel.RegisterSchemaB(models.RequestForProposalSchemaBuilder()),
		dmodel.RegisterSchemaB(models.RequestForQuoteSchemaBuilder()),
		dmodel.RegisterSchemaB(models.VendorSchemaBuilder()),
	)
}

func Init() error {
	return errors.Join()
}
