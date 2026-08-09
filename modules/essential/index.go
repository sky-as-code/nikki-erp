package essential

import (
	"context"
	"errors"
	"reflect"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/go-model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/essential/app"
	modconstants "github.com/sky-as-code/nikki-erp/modules/essential/constants"
	models "github.com/sky-as-code/nikki-erp/modules/essential/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/essential/dynamicengines"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain/services"
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
	return "essential.moduleLabel"
}

// Name implements NikkiModule.
func (*EssentialModule) Name() string {
	return modconstants.EssentialModuleName
}

// Deps implements NikkiModule.
func (*EssentialModule) Deps() []string {
	return []string{
		"dynamicresource",
	}
}

// IsInternal implements InCodeModule.
func (*EssentialModule) IsInternal() bool {
	return false
}

// Version implements NikkiModule.
func (*EssentialModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.1.0")
}

// Init implements NikkiModule.
func (*EssentialModule) Init() error {
	// The engines must exist before transport registers their routes.
	if err := dynamicengines.InitDynamicEngines(); err != nil {
		return err
	}

	err := errors.Join(
		repo.InitRepositories(),
		services.InitDomainServices(),
		app.InitApplicationServices(),
		transport.InitTransport(),
	)
	return err
}

// RegisterModels implements DynamicModule.
func (*EssentialModule) RegisterModels() error {
	return errors.Join(
		dmodel.RegisterSchemaB(models.EnumSchemaBuilder()),
		dmodel.RegisterSchemaB(models.FieldMetadataSchemaBuilder()),
		dmodel.RegisterSchemaB(models.ModuleMetadataSchemaBuilder()),
		dmodel.RegisterSchemaB(models.ModelMetadataSchemaBuilder()),
		dmodel.RegisterSchemaB(models.LanguageSchemaBuilder()),
		dmodel.RegisterSchemaB(models.TagSchemaBuilder()),
		// The category must be registered before the UoM: the UoM's edge points at it.
		dmodel.RegisterSchemaB(models.UomCatSchemaBuilder()),
		dmodel.RegisterSchemaB(models.UomSchemaBuilder()),
	)
}

// OnAppStarted implements NikkiModuleAppStarted.
func (*EssentialModule) OnAppStarted() error {
	return deps.Invoke(func(modules []modules.InCodeModule, moduleSvc it.ModuleAppService) error {
		ctx := corectx.NewRequestContext(context.Background())
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
