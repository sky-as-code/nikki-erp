package essential

import (
	"context"
	"errors"
	"reflect"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	"github.com/sky-as-code/nikki-erp/common/go-model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/core/crud"
	"github.com/sky-as-code/nikki-erp/modules/essential/app"
	repo "github.com/sky-as-code/nikki-erp/modules/essential/infra/repository"
	it "github.com/sky-as-code/nikki-erp/modules/essential/interfaces/module"
	"github.com/sky-as-code/nikki-erp/modules/essential/transport"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.InCodeModule = &EssentialModule{}

type EssentialModule struct {
}

// LabelKey implements NikkiModule.
func (*EssentialModule) LabelKey() string {
	return "essentialzz.moduleLabel"
}

// Name implements NikkiModule.
func (*EssentialModule) Name() string {
	return "essential"
}

// Deps implements NikkiModule.
func (*EssentialModule) Deps() []string {
	return nil
}

// Version implements NikkiModule.
func (*EssentialModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.1")
}

// Init implements NikkiModule.
func (*EssentialModule) Init() error {
	err := errors.Join(
		repo.InitRepositories(),
		app.InitServices(),
		transport.InitTransport(),
	)
	return err
}

// OnAppStarted implements NikkiModuleAppStarted.
func (*EssentialModule) OnAppStarted() error {
	return deps.Invoke(func(modules []modules.InCodeModule, moduleSvc it.ModuleService) error {
		ctx := crud.NewRequestContext(context.Background())
		_, err := moduleSvc.SyncModuleMetadata(ctx, modules)
		return err
	})
}

func init() {
	model.AddConversion[*string, *semver.SemVer](func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf((*semver.SemVer)(nil)), nil
		}

		result := semver.MustParseSemVer(in.Interface().(string))
		return reflect.ValueOf(&result), nil
	})
	model.AddConversion[*semver.SemVer, *string](func(in reflect.Value) (reflect.Value, error) {
		if in.IsNil() {
			return reflect.ValueOf((*string)(nil)), nil
		}

		result := in.Interface().(*semver.SemVer).String()
		return reflect.ValueOf(&result), nil
	})
}
