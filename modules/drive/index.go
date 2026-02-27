package drive

import (
	"errors"

	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/drive/adapter"
	app "github.com/sky-as-code/nikki-erp/modules/drive/app"
	repo "github.com/sky-as-code/nikki-erp/modules/drive/infra/repository"
	transport "github.com/sky-as-code/nikki-erp/modules/drive/transports"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.InCodeModule = &DriveModule{}

type DriveModule struct {
}

// LabelKey implements NikkiModule.
func (*DriveModule) LabelKey() string {
	return "drive.moduleLabel"
}

// Name implements NikkiModule.
func (*DriveModule) Name() string {
	return "drive"
}

// Deps implements NikkiModule.
func (*DriveModule) Deps() []string {
	return []string{}
}

// Version implements NikkiModule.
func (*DriveModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements NikkiModule.
func (*DriveModule) Init() error {
	err := errors.Join(
		adapter.InitAdapters(),
		repo.InitRepositories(),
		app.InitServices(),
		transport.InitTransport(),
	)

	return err
}
