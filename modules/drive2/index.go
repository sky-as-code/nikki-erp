package drive2

import (
	"errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/drive2/adapter"
	app "github.com/sky-as-code/nikki-erp/modules/drive2/app"
	"github.com/sky-as-code/nikki-erp/modules/drive2/domain"
	repo "github.com/sky-as-code/nikki-erp/modules/drive2/infra/repository"
	transport "github.com/sky-as-code/nikki-erp/modules/drive2/transport"
)

var ModuleSingleton modules.InCodeModule = &DriveModule{}

type DriveModule struct{}

func (*DriveModule) LabelKey() string {
	return "drive2.moduleLabel"
}

func (*DriveModule) Name() string {
	return "drive2"
}

func (*DriveModule) Deps() []string {
	return []string{}
}

func (*DriveModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

func (*DriveModule) Init() error {
	return errors.Join(
		adapter.InitAdapters(),
		repo.InitRepositories(),
		app.InitServices(),
		transport.InitTransport(),
	)
}

func (*DriveModule) RegisterModels() error {
	return errors.Join(
		dmodel.RegisterSchemaB(domain.DriveFileSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.DriveFileShareSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.DriveFileAncestorSchemaBuilder()),
	)
}
