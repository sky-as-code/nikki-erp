package notification

import (
	"context"
	"errors"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/common/semver"
	"github.com/sky-as-code/nikki-erp/modules"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	"github.com/sky-as-code/nikki-erp/modules/notification/app"
	modconstants "github.com/sky-as-code/nikki-erp/modules/notification/constants"
	models "github.com/sky-as-code/nikki-erp/modules/notification/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/notification/dynamicengines"
	external "github.com/sky-as-code/nikki-erp/modules/notification/infra/external"
	itExt "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/external"
	"github.com/sky-as-code/nikki-erp/modules/notification/transport"
)

// ModuleSingleton is the exported symbol that will be looked up by the plugin loader
var ModuleSingleton modules.InCodeModule = &NotificationModule{}

type NotificationModule struct {
	running *runningServices
}

// LabelKey implements NikkiModule.
func (*NotificationModule) LabelKey() string {
	return "notification.moduleLabel"
}

// Name implements NikkiModule.
func (*NotificationModule) Name() string {
	return modconstants.NotificationModuleName
}

// ModelPrefix implements DynamicModule.
func (*NotificationModule) ModelPrefix() string {
	return modconstants.NotificationModuleName
}

// Deps implements NikkiModule.
func (*NotificationModule) Deps() []string {
	return []string{
		"dynamicresource",
		// Notification registers the channel configuration an organization sets for itself. The
		// edge is safe: settings depends only on dynamicresource, so nothing routes back here.
		"settings",
	}
}

// IsInternal implements InCodeModule.
func (*NotificationModule) IsInternal() bool {
	return false
}

// Version implements NikkiModule.
func (*NotificationModule) Version() semver.SemVer {
	return *semver.MustParseSemVer("v1.0.0")
}

// Init implements NikkiModule.
func (*NotificationModule) Init() error {
	// The engines must exist before transport registers their routes.
	if err := dynamicengines.InitDynamicEngines(); err != nil {
		return err
	}

	return errors.Join(
		external.InitExternalServices(),
		app.InitApplicationServices(),
		transport.InitTransport(),
	)
}

// RegisterModels implements DynamicModule.
func (*NotificationModule) RegisterModels() error {
	return errors.Join(
		// The notification is registered before the recipient and the delivery, which carry edges
		// pointing at it.
		dmodel.RegisterSchemaB(models.NotificationSchemaBuilder()),
		dmodel.RegisterSchemaB(models.RecipientSchemaBuilder()),
		dmodel.RegisterSchemaB(models.DeliverySchemaBuilder()),
	)
}

// OnAppStarted implements NikkiModuleAppStarted.
//
// The settings schema is registered here rather than in Init() because peer module init order is
// nondeterministic: Init() cannot assume the settings module has built its engines yet, while
// OnAppStarted runs after every module has initialized. The stream dispatcher starts here for the
// same reason — it needs the recipient repository, which is resolved lazily.
func (this *NotificationModule) OnAppStarted() error {
	if err := this.registerSettings(); err != nil {
		return err
	}
	return this.startServices()
}

func (this *NotificationModule) registerSettings() error {
	return deps.Invoke(func(settingsSvc itExt.SettingsRegistrationExtService) error {
		return registerSettings(corectx.NewRequestContext(context.Background()), settingsSvc)
	})
}

// OnAppStopping implements NikkiModuleAppStopping.
func (this *NotificationModule) OnAppStopping() error {
	if this.running == nil {
		return nil
	}
	this.running.stop()
	this.running = nil
	return nil
}
