package notification

import (
	"context"

	deps "github.com/sky-as-code/nikki-erp/common/deps_inject"
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/services"
	"github.com/sky-as-code/nikki-erp/modules/notification/domain/settings"
	it "github.com/sky-as-code/nikki-erp/modules/notification/interfaces/delivery"
)

// runningServices is what this module started and must stop again.
//
// The dispatcher is not registered in Init() for the same reason the settings schema is not: it
// needs peer modules fully built, and Init() runs while they may not be.
type runningServices struct {
	cancel context.CancelFunc
}

func (this *runningServices) stop() {
	if this.cancel != nil {
		this.cancel()
	}
}

// startServices starts the bridge from the broker to this instance's streams.
//
// Without it a notification created on another instance never reaches a client connected here, and
// nothing would report the gap: the inbox would still be correct, and only realtime delivery would
// quietly stop working for everyone but the instance that happened to handle the write.
func (this *NotificationModule) startServices() error {
	return deps.Invoke(func(
		broker it.RealtimeNotificationBroker,
		registry it.StreamRegistry,
		recipients it.RecipientRepository,
		logger logging.LoggerService,
	) error {
		ctx, cancel := context.WithCancel(context.Background())

		dispatcher := services.NewStreamDispatcher(broker, registry, recipients, logger)
		if err := dispatcher.Run(ctx); err != nil {
			cancel()
			return err
		}

		this.running = &runningServices{cancel: cancel}
		return nil
	})
}

func orgSettingsSchema() *dmodel.ModelSchema {
	return settings.OrgSettingsSchemaBuilder().Build()
}
