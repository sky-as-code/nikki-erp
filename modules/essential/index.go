package essential

import (
	"errors"
	"reflect"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/go-model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	"github.com/sky-as-code/nikki-erp/modules/essential/app"
	"github.com/sky-as-code/nikki-erp/modules/essential/domain"
	repo "github.com/sky-as-code/nikki-erp/modules/essential/infra/repository"
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

// RegisterModels implements DynamicModule.
func (*EssentialModule) RegisterModels() error {
	return errors.Join(
		dmodel.RegisterSchemaB(domain.ModuleMetadataSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.ContactSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.ContactChannelSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.ContactRelationshipSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.ModelMetadataSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.FieldMetadataSchemaBuilder()),
		dmodel.RegisterSchemaB(domain.LanguageSchemaBuilder()),
		// Unit schemas
		dmodel.RegisterSchemaB(domain.UnitCategorySchemaBuilder()),
		dmodel.RegisterSchemaB(domain.UnitSchemaBuilder()),
	)
}

// OnAppStarted implements NikkiModuleAppStarted.
func (*EssentialModule) OnAppStarted() error {
	// return deps.Invoke(func(modules []modules.InCodeModule, moduleSvc it.ModuleService) error {
	// 	ctx := corectx.NewRequestContext(context.Background())
	// 	_, err := moduleSvc.SyncModuleMetadata(ctx, modules)
	// 	return err
	// })
	return nil
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
